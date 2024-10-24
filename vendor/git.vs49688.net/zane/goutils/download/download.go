package download

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"

	"git.vs49688.net/zane/goutils"

	"github.com/cavaliergopher/grab/v3"

	"git.vs49688.net/zane/goutils/parallel"
)

// MakeRequest creates a Request. Pass to either RunOne(), or RunAll()
func MakeRequest(ctx context.Context, di *DownloadInfo, le *slog.Logger) (*Request, error) {
	req, err := grab.NewRequest(di.OutPath, di.URL)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	req.Tag = di
	req.Size = di.Size

	if di.SHA256 != "" {
		digest, err := hex.DecodeString(di.SHA256)
		if err != nil {
			le.WarnContext(ctx, "invalid sha256 digest, ignoring")
		}
		req.SetChecksum(sha256.New(), digest, true)
	}

	return &Request{grabRequest: req}, nil
}

func do(ctx context.Context, client *grab.Client, req *Request, logger *slog.Logger) error {
	di := req.grabRequest.Tag.(*DownloadInfo)

	logger = logger.With(slog.String("url", di.URL), slog.String("out_path", di.OutPath))

	logger.InfoContext(ctx, "downloading")

	resp := client.Do(req.grabRequest)
	if err := resp.Err(); err != nil {
		logger.With(slog.Any("error", err)).ErrorContext(ctx, "download failed")
		return err
	}

	logger.InfoContext(ctx, "download completed")

	if sz := resp.Size(); sz >= 0 {
		di.Size = sz
	}

	if di.SHA256 == "" {
		h := sha256.New()

		fp, err := resp.Open()
		if err != nil {
			return err
		}

		defer func() { _ = fp.Close() }()
		if _, err := io.CopyBuffer(h, fp, nil); err != nil {
			return err
		}

		di.SHA256 = hex.EncodeToString(h.Sum(nil))
	}

	di.Completed = true

	return nil
}

// RunOne downloads a single Request.
func RunOne(ctx context.Context, client *http.Client, req *Request, logger *slog.Logger) error {
	if client == nil {
		client = http.DefaultClient
	}

	return do(ctx, &grab.Client{HTTPClient: client}, req, logger)
}

// RunAll is a convenience function to download many Request's in parallel.
// It installs signal handling
func RunAll(ctx context.Context, opts Options) (error, []error) {
	if len(opts.Requests) == 0 {
		return nil, nil
	}

	if opts.Client == nil {
		opts.Client = http.DefaultClient
	}

	client := &grab.Client{
		HTTPClient: opts.Client,
		UserAgent:  opts.UserAgent,
	}

	if opts.Parallelism <= 0 {
		opts.Parallelism = runtime.NumCPU()
	}

	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}

	if opts.SignalHandler == nil {
		opts.SignalHandler = func(ctx context.Context, sig os.Signal, cancel func()) error { return nil }
	}

	sigChan := make(chan os.Signal, 10)
	if len(opts.Signals) > 0 {
		signal.Notify(sigChan, opts.Signals...)
	}

	doneChan := make(chan struct{}, 1)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var errs []error

	go func() {
		errs = parallel.Parallel(ctx, opts.Parallelism, opts.Requests, func(ctx context.Context, req *Request) error {
			err := do(ctx, client, req, opts.Logger)

			di := req.GetDownloadInfo()

			l := opts.Logger.With(
				slog.Any("url", di.URL),
				slog.String("out_path", di.OutPath),
			)

			// Abort if out-of-space.
			if errors.Is(err, syscall.ENOSPC) {
				l.With("error", err).ErrorContext(ctx, "out of disk space, cancelling")
				cancel()
			}

			return err
		})
		doneChan <- struct{}{}
	}()

	for {
		select {
		case sig := <-sigChan:
			opts.Logger.With(slog.String("signal", sig.String())).InfoContext(ctx, "caught signal")

			if err := opts.SignalHandler(ctx, sig, cancel); err != nil {
				cancel()
				return err, nil
			}
		case _ = <-doneChan:
			opts.Logger.InfoContext(ctx, "terminated")
			return ctx.Err(), errs
		}
	}
}

type MaybeOptions struct {
	Client          *http.Client
	Logger          *slog.Logger
	ValidMediaTypes []string
	BuildRequest    func(ctx context.Context, method string, loc string, reader io.Reader) (*http.Request, error)
}

func isFileAlreadyDownloaded(di *DownloadInfo) bool {
	sum, err := hex.DecodeString(di.SHA256)
	if err != nil {
		return false
	}

	f, err := os.OpenFile(di.OutPath, os.O_RDONLY, 0)
	if err != nil {
		return false
	}
	defer func() { _ = f.Close() }()

	hr := goutils.NewHashReader(f, sha256.New)
	if _, err := io.Copy(io.Discard, hr); err != nil {
		return false
	}

	return bytes.Equal(sum, hr.Hash())
}

// MaybeDownloadFile is a legacy version of RunOne(), kept for posterity.
func MaybeDownloadFile(ctx context.Context, di *DownloadInfo, opts MaybeOptions) error {
	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}

	if opts.BuildRequest == nil {
		opts.BuildRequest = http.NewRequestWithContext
	}

	if opts.Client == nil {
		opts.Client = http.DefaultClient
	}

	logger = opts.Logger.With(slog.String("url", di.URL))

	if isFileAlreadyDownloaded(di) {
		di.Completed = true
		logger.InfoContext(ctx, "already downloaded, skipping")
		return nil
	}

	logger.InfoContext(ctx, "downloading")

	req, err := opts.BuildRequest(ctx, http.MethodGet, di.URL, nil)
	if err != nil {
		return err
	}

	resp, err := opts.Client.Do(req)
	if err != nil {
		logger.With(slog.Any("error", err)).ErrorContext(ctx, "download error")
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		logger.With(slog.Any("error", err), slog.Any("http_status", resp.StatusCode)).ErrorContext(ctx, resp.Status)
		return errors.New(resp.Status)
	}

	logger.With(slog.Any("headers", resp.Header)).DebugContext(ctx, "dumping request headers")

	ct := resp.Header.Get("Content-Type")
	mt, _, err := mime.ParseMediaType(ct)
	if err != nil {
		logger.With(slog.Any("error", err), slog.String("content-type", ct)).ErrorContext(ctx, "invalid content type")
		return err
	}

	if len(opts.ValidMediaTypes) > 0 {
		found := false
		for _, mmt := range opts.ValidMediaTypes {
			if mt == mmt {
				found = true
				break
			}
		}

		if !found {
			logger.With(slog.Any("content-type", ct)).ErrorContext(ctx, "unsupported content type")
			return fmt.Errorf("unsupported content type: %v", ct)
		}
	}

	if size, err := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64); err == nil {
		di.ContentLength = size
	}

	// #nosec G302 - Because these actually need to be readable.
	f, err := os.OpenFile(di.OutPath+".tmp", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		logger.With(slog.Any("error", err)).ErrorContext(ctx, "error opening output file")
		return err
	}
	defer func() { _ = f.Close() }()

	hr := goutils.NewHashReader(resp.Body, sha256.New)

	if _, err := io.Copy(f, hr); err != nil {
		logger.With(slog.Any("error", err)).ErrorContext(ctx, "read/write error")
		return err
	}

	if err := f.Sync(); err != nil {
		logger.With(slog.Any("error", err)).ErrorContext(ctx, "sync error")
		return err
	}

	if err := os.Rename(di.OutPath+".tmp", di.OutPath); err != nil {
		logger.With(slog.Any("error", err)).ErrorContext(ctx, "error renaming temp file")
		return err
	}

	di.Size = hr.GetSize()
	di.SHA256 = hex.EncodeToString(hr.Hash())

	di.Completed = true

	logger.With(
		slog.Int64("content_length", di.ContentLength),
		slog.Int64("size", di.Size),
		slog.String("sha256", di.SHA256),
	).InfoContext(ctx, "downloaded")

	return nil
}
