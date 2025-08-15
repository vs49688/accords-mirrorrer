package archive

import (
	"context"
	"fmt"
	"log/slog"

	accords_mirrorrer "git.vs49688.net/zane/accords-mirrorrer"
	"git.vs49688.net/zane/accords-mirrorrer/library"
)

func fetchContents(ctx context.Context, state *accords_mirrorrer.State, client library.Client, slug string) (*library.ContentProps, error) {
	if ct, exists := state.Content[slug]; exists {
		return ct, nil
	}

	ct, err := client.GetContents(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("error fetching content: %v: %w", slug, err)
	}

	state.RawContent[slug] = ct.Raw
	state.Content[slug] = ct
	return ct, err
}

func gatherContents(ctx context.Context, state *accords_mirrorrer.State, client library.Client, log *slog.Logger) error {
	contents := map[string]struct{}{}

	for _, f := range state.Folders {
		for _, ct := range f.Contents {
			contents[ct.Slug] = struct{}{}
		}
	}

	for slug := range contents {
		l := log.With(slog.String("slug", slug))
		l.InfoContext(ctx, "fetching content item")
		if _, err := fetchContents(ctx, state, client, slug); err != nil {
			l.With(slog.Any("error", err)).ErrorContext(ctx, "error fetching content item")
			return err
		}
	}

	return nil
}

func extractContents(item *library.Content, client library.Client, state *accords_mirrorrer.State, l *slog.Logger) {
	if item == nil {
		return
	}

	slug := item.Slug
	l = l.With(slog.String("slug", slug))

	addUploadFileEntityResponse(item.Thumbnail, client, state, l)

	for _, tl := range item.Translations {
		if tl.Language == nil || tl.Language.Data == nil || tl.Language.Data.Attributes == nil || tl.Language.Data.Attributes.Code == "" {
			continue
		}
		langCode := tl.Language.Data.Attributes.Code

		ll := l.With(slog.Any("lang_code", langCode))

		// src/pages/contents/[slug].tsx
		if tl.Audio_set != nil {
			audioURL := client.BuildAudioURL(slug, langCode)

			if _, err := state.AddDownload(audioURL.String()); err != nil {
				ll.With(slog.Any("error", err)).Warn("error adding audio url")
			}
		}

		if tl.Video_set != nil {
			// NB: the frontend always attempts this, even if it doesn't exist
			subURL := client.BuildVTTURL(slug, langCode)
			if _, err := state.AddDownload(subURL.String()); err != nil {
				ll.With(slog.Any("error", err)).Warn("error adding subtitle url")
			}

			videoURL := client.BuildVideoURL(slug, langCode)
			if _, err := state.AddDownload(videoURL.String()); err != nil {
				ll.With(slog.Any("error", err)).Warn("error adding video url")
			}
		}
	}
}
