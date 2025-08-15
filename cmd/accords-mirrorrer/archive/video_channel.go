package archive

import (
	"context"
	"fmt"
	"log/slog"
	"path"
	"strings"

	accords_mirrorrer "git.vs49688.net/zane/accords-mirrorrer"
	"git.vs49688.net/zane/accords-mirrorrer/library"
)

func fetchVideoChannel(ctx context.Context, state *accords_mirrorrer.State, client library.Client, uid string) (*library.VideoChannelProps, error) {
	if v, exists := state.VideoChannels[uid]; exists {
		return v, nil
	}

	v, err := client.GetVideoChannel(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("error fetching video channel %v: %w", uid, err)
	}

	state.RawVideoChannels[uid] = v.Raw
	state.VideoChannels[uid] = v
	return v, nil
}

func gatherVideoChannels(ctx context.Context, state *accords_mirrorrer.State, client library.Client, log *slog.Logger) error {
	// There's no endpoint to list/search these that I can find, so infer them from the video list.

	uids := map[string]struct{}{}

	for _, v := range state.Videos {

		if !strings.HasPrefix(v.Channel.Href, "/archives/videos/c/") {
			continue
		}

		_, uid := path.Split(v.Channel.Href)
		uids[uid] = struct{}{}
	}

	for uid := range uids {
		l := log.With(slog.String("uid", uid))
		l.InfoContext(ctx, "fetching video channel")

		if _, err := fetchVideoChannel(ctx, state, client, uid); err != nil {
			l.With(slog.Any("error", err)).ErrorContext(ctx, "error video channel")
			return err
		}
	}

	return nil
}
