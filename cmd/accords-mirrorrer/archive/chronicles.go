package archive

import (
	"context"
	"fmt"
	"log/slog"

	accords_mirrorrer "git.vs49688.net/zane/accords-mirrorrer"
	"git.vs49688.net/zane/accords-mirrorrer/library"
)

func fetchChronicle(ctx context.Context, state *accords_mirrorrer.State, client library.Client, slug string) (*library.ChronicleProps, error) {
	if p, exists := state.Chronicles.Entries[slug]; exists {
		return p, nil
	}

	ch, err := client.GetChronicle(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("error fetching chronicle %v: %w", slug, err)
	}

	state.Chronicles.RawEntries[slug] = ch.Raw
	state.Chronicles.Entries[slug] = ch
	return ch, nil
}

func gatherChronicles(ctx context.Context, state *accords_mirrorrer.State, client library.Client, log *slog.Logger) error {
	root, err := client.GetChronicles(ctx)
	if err != nil {
		return err
	}
	state.Chronicles.RawIndex = root.Raw
	state.Chronicles.Index = root

	slugs := map[string]struct{}{}

	for _, chapter := range state.Chronicles.Index.Chapters {
		if chapter.Attributes == nil {
			continue
		}

		if chapter.Attributes.Chronicles == nil {
			continue
		}

		for _, chronicle := range chapter.Attributes.Chronicles.Data {
			if chronicle.Attributes == nil {
				continue
			}

			slugs[chronicle.Attributes.Slug] = struct{}{}
		}
	}

	for slug := range slugs {
		l := log.With(slog.String("slug", slug))
		l.InfoContext(ctx, "fetching chronicle")

		if _, err := fetchChronicle(ctx, state, client, slug); err != nil {
			l.With(slog.Any("error", err)).ErrorContext(ctx, "error fetching chronicle")
			return err
		}
	}

	for slug, chronicle := range state.Chronicles.Entries {
		l := log.With(slog.String("slug", slug))

		if chronicle.Chronicle.Contents == nil {
			continue
		}

		for _, ct := range chronicle.Chronicle.Contents.Data {
			if ct.Attributes == nil {
				continue
			}

			ll := l.With(slog.String("content_slug", ct.Attributes.Slug))
			ll.InfoContext(ctx, "fetching chronicle content")
			if _, err := fetchContents(ctx, state, client, ct.Attributes.Slug); err != nil {
				ll.With(slog.Any("error", err)).ErrorContext(ctx, "error chronicle content")
				return err
			}
		}
	}

	return nil
}

func extractChronicle(item *library.Chronicle, client library.Client, state *accords_mirrorrer.State, l *slog.Logger) {
	if item == nil {
		return
	}

	if item.Contents != nil {
		for _, ct := range item.Contents.Data {
			if ct.Attributes != nil {
				addUploadFileEntityResponse(ct.Attributes.Thumbnail, client, state, l)
			}
		}
	}
}
