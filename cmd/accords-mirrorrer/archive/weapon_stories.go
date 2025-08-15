package archive

import (
	"context"
	"fmt"
	"log/slog"

	accords_mirrorrer "git.vs49688.net/zane/accords-mirrorrer"
	"git.vs49688.net/zane/accords-mirrorrer/library"
)

func fetchWeaponStory(ctx context.Context, state *accords_mirrorrer.State, client library.Client, slug string) (*library.WeaponStoryProps, error) {
	if w, exists := state.WeaponStories[slug]; exists {
		return w, nil
	}

	w, err := client.GetWeaponStory(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("error fetching weapon story %v: %w", slug, err)
	}

	state.RawWeaponsStories[slug] = w.Raw
	state.WeaponStories[slug] = w
	return w, nil
}

func searchWeaponStories(ctx context.Context, state *accords_mirrorrer.State, client library.Client, log *slog.Logger) error {
	numPages := 1

	slugs := map[string]struct{}{}

	for i := 0; i < numPages; i += 1 {
		log.With(slog.Int("page", i+1), slog.Int("hits_per_page", 25)).InfoContext(ctx, "searching wiki")
		sr, err := client.SearchWeaponStories(ctx, i+1, 25, []string{"slug"})
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
		l.InfoContext(ctx, "fetching weapon story")

		if _, err := fetchWeaponStory(ctx, state, client, slug); err != nil {
			l.With(slog.Any("error", err)).ErrorContext(ctx, "error fetching weapon story")
			return err
		}
	}

	return nil
}

func extractWeaponStory(item *library.WeaponStory, client library.Client, state *accords_mirrorrer.State, l *slog.Logger) {
	if item == nil {
		return
	}

	addUploadFileEntityResponse(item.Thumbnail, client, state, l)

	if item.Weapon_group != nil && item.Weapon_group.Data != nil && item.Weapon_group.Data.Attributes != nil && item.Weapon_group.Data.Attributes.Weapons != nil {
		for _, wg := range item.Weapon_group.Data.Attributes.Weapons.Data {
			if wg.Attributes != nil {
				addUploadFileEntityResponse(wg.Attributes.Thumbnail, client, state, l)
			}
		}
	}
}
