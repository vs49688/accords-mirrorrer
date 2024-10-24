package archive

import (
	"log/slog"
	"net/url"

	accords_mirrorrer "github.com/vs49688/accords-mirrorrer"
	"github.com/vs49688/accords-mirrorrer/library"
)

func addURL(u string, baseURL *url.URL, state *accords_mirrorrer.State, l *slog.Logger) {
	uu, err := url.Parse(u)
	if err != nil {
		l.With(slog.Any("error", err)).Warn("error parsing url")
		return
	}

	finalUrl := baseURL.ResolveReference(uu)

	if _, err := state.AddDownload(finalUrl.String()); err != nil {
		l.With(slog.Any("error", err)).Warn("error adding url")
	}
}

func addUploadFileImage(uf *library.UploadFile, client library.Client, state *accords_mirrorrer.State, l *slog.Logger) {
	if uf == nil || uf.Url == "" {
		return
	}

	addURL(uf.Url, client.GetCMSUrl(), state, l)
}

func addUploadImageFragment(uif *library.UploadImageFragment, client library.Client, state *accords_mirrorrer.State, l *slog.Logger) {
	if uif == nil || uif.URL == "" {
		return
	}

	addURL(uif.URL, client.GetCMSUrl(), state, l)
}

func addUploadFileEntityResponse(uifr *library.UploadFileEntityResponse, client library.Client, state *accords_mirrorrer.State, l *slog.Logger) {
	if uifr == nil || uifr.Data == nil {
		return
	}

	addUploadFileImage(uifr.Data.Attributes, client, state, l)
}
