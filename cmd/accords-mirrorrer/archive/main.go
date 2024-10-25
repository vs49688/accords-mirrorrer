package archive

import (
	"context"
	"errors"

	"log/slog"
	"net/http"
	"net/url"
	"os"
	"sort"
	"syscall"

	"github.com/urfave/cli/v2"

	"git.vs49688.net/zane/goutils/download"

	accords_mirrorrer "github.com/vs49688/accords-mirrorrer"
	"github.com/vs49688/accords-mirrorrer/cmd/accords-mirrorrer/config"
	"github.com/vs49688/accords-mirrorrer/library"
)

type configuration struct {
	*config.Configuration
	StateFile    string
	Parallelism  int
	DontRefresh  bool
	DontDownload bool
}

func RegisterCommand(app *cli.App, globalCfg *config.Configuration) *cli.App {
	cfg := configuration{
		Configuration: globalCfg,
		StateFile:     "state.json",
		Parallelism:   0,
		DontRefresh:   false,
		DontDownload:  false,
	}

	app.Commands = append(app.Commands, &cli.Command{
		Name:  "archive",
		Usage: "archive the library",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "state-file",
				Usage:       "state file",
				Value:       cfg.StateFile,
				Destination: &cfg.StateFile,
			},
			&cli.IntFlag{
				Name:        "parallelism",
				Usage:       "parallelism, <1 for GOMAXPROCS",
				Value:       cfg.Parallelism,
				Destination: &cfg.Parallelism,
			},
			&cli.BoolFlag{
				Name:        "dont-refresh",
				Usage:       "don't refresh the index, only download what we've got",
				Value:       cfg.DontRefresh,
				Destination: &cfg.DontRefresh,
			},
			&cli.BoolFlag{
				Name:        "dont-download",
				Usage:       "don't download files, only refresh the index",
				Value:       cfg.DontDownload,
				Destination: &cfg.DontDownload,
			},
		},
		Action: func(c *cli.Context) error { return archive(c.Context, &cfg) },
	})

	return app
}

func refreshIndex(ctx context.Context, state *accords_mirrorrer.State, client library.Client, hc *http.Client, log *slog.Logger) error {
	if err := collectFolders(ctx, state, client, log); err != nil {
		return err
	}

	if err := gatherContents(ctx, state, client, log); err != nil {
		return err
	}

	if err := searchLibrary(ctx, state, client, log); err != nil {
		return err
	}

	if err := searchWiki(ctx, state, client, log); err != nil {
		return err
	}

	if err := searchWeaponStories(ctx, state, client, log); err != nil {
		return err
	}

	{
		chron, err := client.GetChronology(ctx)
		if err != nil {
			return err
		}
		state.RawChronology = chron.Raw
		state.Chronology = chron
	}

	if err := searchPosts(ctx, state, client, log); err != nil {
		return err
	}

	if err := gatherChronicles(ctx, state, client, log); err != nil {
		return err
	}

	if err := searchVideos(ctx, state, client, log); err != nil {
		return err
	}

	if err := gatherVideoChannels(ctx, state, client, log); err != nil {
		return err
	}

	// Now do a brute-force scrape of the actual storage, just in case. Thankfully, they've got directory listing enabled.
	{
		resURLs, err := download.ScrapeIndex(ctx, &url.URL{Scheme: "https", Host: "resha.re", Path: "/accords/"}, hc, log)
		if err != nil {
			log.With(slog.Any("error", err)).ErrorContext(ctx, "error scraping urls")
			return err
		}

		for _, u := range resURLs {
			if _, err := state.AddDownload(u.String()); err != nil {
				log.With(slog.Any("error", err), slog.String("url", u.String())).ErrorContext(ctx, "error scraping urls")
				continue
			}
		}
	}

	return nil
}

type transport struct {
	cfg *configuration
}

func (t transport) RoundTrip(request *http.Request) (*http.Response, error) {
	request.Header.Set("User-Agent", t.cfg.UserAgent)
	return http.DefaultTransport.RoundTrip(request)
}

func archive(ctx context.Context, cfg *configuration) error {
	l := cfg.Logger

	state, err := accords_mirrorrer.LoadState(cfg.StateFile)
	if err != nil {
		return err
	}
	l.InfoContext(ctx, "loaded state")

	hc := &http.Client{Transport: transport{cfg: cfg}}
	lc := library.NewClient(hc)

	if !cfg.DontRefresh {
		if err := refreshIndex(ctx, state, lc, hc, l); err != nil {
			if err2 := accords_mirrorrer.SaveState(state, cfg.StateFile); err2 != nil {
				err = errors.Join(err, err2)
			}
			return err
		}
	} else {
		l.InfoContext(ctx, "skipping index refresh by request")
	}

	processStateOpenGraph(state, l)

	for _, folder := range state.Folders {
		for _, ct := range folder.Contents {
			addUploadImageFragment(ct.Thumbnail, lc, state, l)
		}
	}

	for _, ct := range state.Content {
		extractContents(ct.Content, lc, state, l)
	}

	for _, item := range state.Library {
		extractLibraryItem(item.Item, lc, state, l)
	}

	for _, item := range state.Reader {
		extractLibraryItem(item.Item, lc, state, l)
	}

	for _, item := range state.Wiki {
		addUploadFileEntityResponse(item.Page.Thumbnail, lc, state, l)
	}

	for _, item := range state.WeaponStories {
		extractWeaponStory(item.Weapon, lc, state, l)
	}

	for _, item := range state.Posts {
		addUploadFileEntityResponse(item.Post.Thumbnail, lc, state, l)

		for _, tl := range item.Post.Translations {
			addUploadFileEntityResponse(tl.Thumbnail, lc, state, l)
		}

	}

	for _, item := range state.Videos {
		extractVideo(item, lc, state, l)
	}

	for _, item := range state.Chronicles.Entries {
		extractChronicle(&item.Chronicle, lc, state, l)
	}

	l.InfoContext(ctx, "index update finished...")

	if !cfg.DontDownload {
		reqs := make([]*download.Request, 0, len(state.Downloads))
		for _, di := range state.Downloads {
			if !di.Completed {
				req, _ := download.MakeRequest(ctx, di, l)
				reqs = append(reqs, req)
			}
		}

		sort.Slice(reqs, func(i, j int) bool {
			return reqs[i].GetURL().String() < reqs[j].GetURL().String()
		})

		numInterrupts := 0

		err, _ = download.RunAll(ctx, download.Options{
			Requests:    reqs,
			Parallelism: cfg.Parallelism,
			Client:      http.DefaultClient,
			UserAgent:   accords_mirrorrer.UserAgent,
			Logger:      l,
			Signals:     []os.Signal{syscall.SIGINT, syscall.SIGTERM, SIGUSR1},
			SignalHandler: func(ctx context.Context, sig os.Signal, cancel func()) error {
				if sig == SIGUSR1 {
					if err := accords_mirrorrer.SaveState(state, cfg.StateFile); err != nil {
						return err
					}
				} else {
					cancel()

					numInterrupts += 1

					if numInterrupts == 1 {
						l.InfoContext(ctx, "interrupt received, stopping new downloads")
						l.InfoContext(ctx, "interrupt 2 more times to stop current")
					}

					if numInterrupts >= 3 {
						return errors.New("aborting")
					}
				}
				return nil
			},
		})
	} else {
		l.InfoContext(ctx, "skipping download by request")
	}

	if err2 := accords_mirrorrer.SaveState(state, cfg.StateFile); err2 != nil {
		err = errors.Join(err, err2)
	}

	return err
}
