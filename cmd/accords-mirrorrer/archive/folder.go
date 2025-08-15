package archive

import (
	"context"
	"fmt"
	"log/slog"

	accords_mirrorrer "git.vs49688.net/zane/accords-mirrorrer"
	"git.vs49688.net/zane/accords-mirrorrer/library"
)

func fetchFolder(ctx context.Context, state *accords_mirrorrer.State, slug string, client library.Client) (*library.FolderProps, error) {
	if idx, exists := state.Folders[slug]; exists {
		return idx, nil
	}

	folder, err := client.ListFolder(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("error fetching folder: %v: %w", slug, err)
	}

	state.Folders[slug] = folder
	state.RawFolders[slug] = folder.Raw
	return folder, nil
}

// collectFolders recursively scans the "folders", gathering all the information
func collectFolders(ctx context.Context, state *accords_mirrorrer.State, client library.Client, log *slog.Logger) error {
	queue := []string{"root"}

	for current := 0; current < len(queue); current += 1 {
		slug := queue[current]

		l := log.With(slog.Any("slug", slug))
		l.InfoContext(ctx, "fetching folder")

		idx, err := fetchFolder(ctx, state, slug, client)
		if err != nil {
			l.With(slog.Any("error", err)).ErrorContext(ctx, "error fetching folder")
			return err
		}

		for _, sf := range idx.Subfolders {
			queue = append(queue, sf.Slug)
		}
	}

	return nil
}
