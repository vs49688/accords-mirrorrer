package download

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"os"

	"github.com/cavaliergopher/grab/v3"
)

type DownloadInfo struct {
	URL           string `json:"url"`
	ContentLength int64  `json:"content_length,omitempty"`
	Size          int64  `json:"size,omitempty"`
	OutPath       string `json:"out_path"`
	SHA256        string `json:"sha256,omitempty"`
	Completed     bool   `json:"completed"`
}

type Request struct {
	grabRequest *grab.Request
}

func (r *Request) GetURL() *url.URL {
	return r.grabRequest.URL()
}

func (r *Request) GetDownloadInfo() *DownloadInfo {
	return r.grabRequest.Tag.(*DownloadInfo)
}

type Options struct {
	Requests          []*Request
	Parallelism       int
	Client            *http.Client
	UserAgent         string
	Logger            *slog.Logger
	DontHandleSignals bool
	SignalHandler     func(ctx context.Context, sig os.Signal, cancel func()) error
}
