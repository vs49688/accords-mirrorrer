package archive

import (
	"context"
	"fmt"
	"log/slog"

	accords_mirrorrer "git.vs49688.net/zane/accords-mirrorrer"
	"git.vs49688.net/zane/accords-mirrorrer/library"
)

func fetchPost(ctx context.Context, state *accords_mirrorrer.State, client library.Client, slug string) (*library.PostProps, error) {
	if p, exists := state.Posts[slug]; exists {
		return p, nil
	}

	p, err := client.GetPost(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("error fetching post %v: %w", slug, err)
	}

	state.RawPosts[slug] = p.Raw
	state.Posts[slug] = p
	return p, nil
}

func searchPosts(ctx context.Context, state *accords_mirrorrer.State, client library.Client, log *slog.Logger) error {
	slugs, err := searchEntity[library.Post](ctx, log, "slug", client.SearchPosts)
	if err != nil {
		return err
	}

	for _, slug := range slugs {
		l := log.With(slog.String("slug", slug))
		l.InfoContext(ctx, "fetching post")

		if _, err := fetchPost(ctx, state, client, slug); err != nil {
			l.With(slog.Any("error", err)).ErrorContext(ctx, "error fetching post")
			return err
		}
	}

	return nil
}
