package archive

import (
	"context"
	"fmt"
	"log/slog"

	accords_mirrorrer "github.com/vs49688/accords-mirrorrer"
	"github.com/vs49688/accords-mirrorrer/library"
)

func fetchWikiPage(ctx context.Context, state *accords_mirrorrer.State, client library.Client, slug string) (*library.WikiProps, error) {
	if w, exists := state.Wiki[slug]; exists {
		return w, nil
	}

	w, err := client.GetWikiPage(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("error fetching wiki page %v: %w", slug, err)
	}

	state.RawWiki[slug] = w.Raw
	state.Wiki[slug] = w
	return w, nil
}

func searchWiki(ctx context.Context, state *accords_mirrorrer.State, client library.Client, log *slog.Logger) error {
	numPages := 1

	slugs := map[string]struct{}{}

	for i := 0; i < numPages; i += 1 {
		log.With(slog.Int("page", i+1), slog.Int("hits_per_page", 25)).InfoContext(ctx, "searching wiki")
		sr, err := client.SearchWiki(ctx, i+1, 25, []string{"slug"})
		if err != nil {
			return err
		}

		if i == 0 {
			numPages = sr.TotalPages
		}

		for _, hit := range sr.Hits {
			slugs[hit.Slug] = struct{}{}
		}
	}

	for slug := range slugs {
		l := log.With(slog.String("slug", slug))
		l.InfoContext(ctx, "fetching wiki entry")

		if _, err := fetchWikiPage(ctx, state, client, slug); err != nil {
			l.With(slog.Any("error", err)).ErrorContext(ctx, "error fetching wiki entry")
			return err
		}
	}

	return nil
}
