package download

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"

	"github.com/PuerkitoBio/goquery"
)

func fetchHTML(req *http.Request, client *http.Client) (*html.Node, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %v", resp.Status)
	}

	return html.Parse(resp.Body)
}

// ScrapeIndex recursively scrapes a directory listing for files, returning the URLs.
func ScrapeIndex(ctx context.Context, start *url.URL, client *http.Client, l *slog.Logger) ([]*url.URL, error) {
	queue := []*url.URL{start}
	var files []*url.URL

	startPrefix := start.String()
	checked := map[string]struct{}{}

	l.With(slog.String("start", startPrefix)).InfoContext(ctx, "beginning index scrape")

	for current := 0; current < len(queue); current += 1 {
		u := queue[current]

		us := u.String()

		ll := l.With(slog.String("url", us))

		if _, done := checked[us]; done {
			continue
		}

		// Skip URLs going above us
		if !strings.HasPrefix(us, startPrefix) {
			checked[us] = struct{}{}
			continue
		}

		if !strings.HasSuffix(u.Path, "/") {
			files = append(files, u)
			checked[us] = struct{}{}
			continue
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, us, nil)
		if err != nil {
			ll.With(slog.Any("error", err)).ErrorContext(ctx, "unable to build request")

			return nil, err
		}

		doc, err := fetchHTML(req, client)
		if err != nil {
			ll.With(slog.Any("error", err)).ErrorContext(ctx, "error scraping")
			return nil, err
		}

		gq := goquery.NewDocumentFromNode(doc)

		numTotal := 0
		numAdded := 0

		for _, a := range gq.Find("a").Nodes {
			var href string
			for _, attr := range a.Attr {
				if attr.Key == "href" {
					href = attr.Val
					break
				}
			}

			// Skip index sorting things
			if href == "" || strings.HasPrefix(href, "?C=") {
				checked[us] = struct{}{}
				continue
			}

			hrefUrl, err := url.Parse(href)
			if err != nil {
				ll.With("href", hrefUrl).WarnContext(ctx, "skipping invalid href")
				continue
			}

			next := u.ResolveReference(hrefUrl)

			numTotal += 1
			if _, done := checked[next.String()]; !done {
				queue = append(queue, next)
				numAdded += 1
			}
		}

		ll.With(slog.Int("total_urls", numTotal), slog.Int("new_urls", numAdded)).InfoContext(ctx, "scraped")
	}

	return files, nil
}
