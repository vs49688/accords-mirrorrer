package archive

import (
	"context"
	"fmt"
	"log/slog"

	accords_mirrorrer "github.com/vs49688/accords-mirrorrer"
	"github.com/vs49688/accords-mirrorrer/library"
)

func fetchVideo(ctx context.Context, state *accords_mirrorrer.State, client library.Client, uid string) (*library.VideoProps, error) {
	if v, exists := state.Videos[uid]; exists {
		return v, nil
	}

	v, err := client.GetVideo(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("error fetching video %v: %w", uid, err)
	}

	state.RawVideos[uid] = v.Raw
	state.Videos[uid] = v
	return v, nil
}

func searchVideos(ctx context.Context, state *accords_mirrorrer.State, client library.Client, log *slog.Logger) error {
	uids, err := searchEntity[library.Video](ctx, log, "uid", client.SearchVideos)
	if err != nil {
		return err
	}

	for _, slug := range uids {
		l := log.With(slog.String("uid", slug))
		l.InfoContext(ctx, "fetching video")

		if _, err := fetchVideo(ctx, state, client, slug); err != nil {
			l.With(slog.Any("error", err)).ErrorContext(ctx, "error fetching video")
			return err
		}
	}

	return nil
}

func extractVideo(item *library.VideoProps, client library.Client, state *accords_mirrorrer.State, l *slog.Logger) {
	if item == nil || item.Video.Uid == "" {
		return
	}

	l = l.With(slog.String("uid", item.Video.Uid))

	thumbURL := client.BuildVideoThumbnailURL(item.Video.Uid)

	if _, err := state.AddDownload(thumbURL.String()); err != nil {
		l.With(slog.Any("error", err)).Warn("error adding video thumbnail")
	}

	videoURL := client.BuildVideoFileURL(item.Video.Uid)
	if _, err := state.AddDownload(videoURL.String()); err != nil {
		l.With(slog.Any("error", err)).Warn("error adding video")
	}
}
