package archive

import (
	"log/slog"

	accords_mirrorrer "github.com/vs49688/accords-mirrorrer"
	"github.com/vs49688/accords-mirrorrer/library"
)

func processOpenGraph(og *library.OpenGraph, state *accords_mirrorrer.State, l *slog.Logger) {
	if url := og.Thumbnail.Image; url != "" {
		if _, err := state.AddDownload(url); err != nil {
			l.With(slog.String("url", url)).Warn("error adding thumbnail download")
		}
	}

	if og.Audio != "" {
		if _, err := state.AddDownload(og.Audio); err != nil {
			l.With(slog.String("url", og.Audio)).Warn("error adding audio download")
		}
	}

	if og.Video != "" {
		if _, err := state.AddDownload(og.Video); err != nil {
			l.With(slog.String("url", og.Video)).Warn("error adding video download")
		}
	}
}

func processOpenGraphMap[T library.Entity](state *accords_mirrorrer.State, props map[string]T, l *slog.Logger) {
	for _, v := range props {
		processOpenGraph(v.GetOpenGraph(), state, l)
	}
}

func processStateOpenGraph(state *accords_mirrorrer.State, l *slog.Logger) {
	processOpenGraphMap(state, state.Folders, l)
	processOpenGraphMap(state, state.Content, l)
	processOpenGraphMap(state, state.Library, l)
	processOpenGraphMap(state, state.Reader, l)
	processOpenGraphMap(state, state.Wiki, l)
	processOpenGraphMap(state, state.WeaponStories, l)
	processOpenGraph(state.Chronology.GetOpenGraph(), state, l)
	processOpenGraphMap(state, state.Posts, l)
	processOpenGraph(state.Chronicles.Index.GetOpenGraph(), state, l)
	processOpenGraphMap(state, state.Chronicles.Entries, l)
	processOpenGraphMap(state, state.Videos, l)
	processOpenGraphMap(state, state.VideoChannels, l)
}
