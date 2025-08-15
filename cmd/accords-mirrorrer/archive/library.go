package archive

import (
	"context"
	"fmt"
	"log/slog"

	accords_mirrorrer "git.vs49688.net/zane/accords-mirrorrer"
	"git.vs49688.net/zane/accords-mirrorrer/library"
)

func fetchLibraryEntry(ctx context.Context, state *accords_mirrorrer.State, client library.Client, slug string) (*library.LibraryProps, error) {
	if lib, exists := state.Library[slug]; exists {
		return lib, nil
	}

	lib, err := client.GetLibraryItem(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("error fetching library entry %v: %w", slug, err)
	}

	state.RawLibrary[slug] = lib.Raw
	state.Library[slug] = lib

	return lib, nil
}

func fetchReaderEntry(ctx context.Context, state *accords_mirrorrer.State, client library.Client, slug string) (*library.ReaderProps, error) {
	if r, exists := state.Reader[slug]; exists {
		return r, nil
	}

	r, err := client.GetReader(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("error fetching reader entry %v: %w", slug, err)
	}

	state.RawReader[slug] = r.Raw
	state.Reader[slug] = r
	return r, nil
}

func searchLibrary(ctx context.Context, state *accords_mirrorrer.State, client library.Client, log *slog.Logger) error {
	numPages := 1

	slugs := map[string]struct{}{}

	for i := 0; i < numPages; i += 1 {
		log.With(slog.Int("page", i+1), slog.Int("hits_per_page", 25)).InfoContext(ctx, "searching library")
		sr, err := client.SearchLibrary(ctx, i+1, 25, []string{"slug"})
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
		l.InfoContext(ctx, "fetching library item")

		lib, err := fetchLibraryEntry(ctx, state, client, slug)
		if err != nil {
			l.With(slog.Any("error", err)).ErrorContext(ctx, "error fetching library item")
			return err
		}

		if lib.HasContentScans {
			if _, err := fetchReaderEntry(ctx, state, client, slug); err != nil {
				l.With(slog.Any("error", err)).ErrorContext(ctx, "error fetching scans")
				return err
			}
		}
	}

	return nil
}

func gatherTracks(item *library.LibraryItem, state *accords_mirrorrer.State, lc library.Client, l *slog.Logger) {
	for _, md := range item.Metadata {
		audioMeta, ok := md.Value.(*library.ComponentMetadataAudio)
		if !ok {
			continue
		}

		for _, t := range audioMeta.Tracks {
			trackURL := lc.BuildTrackURL(item.Slug, t.Slug)
			if _, err := state.AddDownload(trackURL.String()); err != nil {
				l.With(slog.Any("error", err)).Warn("unable to add track")
			}
		}
	}
}

func extractLibraryItem(item *library.LibraryItem, client library.Client, state *accords_mirrorrer.State, l *slog.Logger) {
	l = l.With(slog.String("slug", item.Slug))

	// Grab the thumbnail image.
	addUploadFileEntityResponse(item.Thumbnail, client, state, l)

	// Grab the scan archive.
	if item.Download_available {
		archiveURL := client.BuildScanArchiveURL(item.Slug)

		if _, err := state.AddDownload(archiveURL.String()); err != nil {
			l.With(slog.Any("error", err)).Warn("error scan url")
		}
	}

	// Grab the tracks.
	gatherTracks(item, state, client, l)

	// Now drill down into the contents. Only Reader should have these.
	if item.Contents != nil {
		for _, ct := range item.Contents.Data {
			if ct.Attributes == nil {
				continue
			}

			for _, ss := range ct.Attributes.Scan_set {
				if ss.Pages == nil {
					continue
				}

				for _, page := range ss.Pages.Data {
					addUploadFileImage(page.Attributes, client, state, l)
				}
			}
		}
	}
}
