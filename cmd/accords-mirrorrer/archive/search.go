package archive

import (
	"context"
	"log/slog"
	"reflect"

	"github.com/vs49688/accords-mirrorrer/library"
)

func searchEntity[T any](ctx context.Context, log *slog.Logger, idAttrib string, search func(ctx context.Context, page, hitsPerPage int, attributes []string) (*library.SearchResult[T], error)) ([]string, error) {
	numPages := 1

	slugs := map[string]struct{}{}

	for i := 0; i < numPages; i += 1 {
		log.With(slog.Int("page", i+1), slog.Int("hits_per_page", 25)).InfoContext(ctx, "searching wiki")
		sr, err := search(ctx, i+1, 25, []string{idAttrib})
		if err != nil {
			return nil, err
		}

		if i == 0 {
			numPages = sr.TotalPages
		}

		for _, hit := range sr.Hits {
			// FIXME: bit hacky, I don't care anymore
			var fieldName string
			switch idAttrib {
			case "slug":
				fieldName = "Slug"
			case "uid":
				fieldName = "Uid"
			default:
				panic("unknown field name " + idAttrib)
			}

			slug := reflect.ValueOf(hit).Elem().FieldByName(fieldName).String()
			slugs[slug] = struct{}{}
		}
	}

	out := make([]string, 0, len(slugs))
	for slug := range slugs {
		out = append(out, slug)
	}

	return out, nil
}
