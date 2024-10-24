package library

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
)

type libraryClient struct {
	client     *http.Client
	libraryURL *url.URL
	cmsURL     *url.URL
	searchURL  *url.URL
	assetsURL  *url.URL

	searchToken string
}

func NewClient(client *http.Client) Client {
	return &libraryClient{
		client:      client,
		libraryURL:  &url.URL{Scheme: "https", Host: "accords-library.com", Path: "/_next/data/Ay_yaf-7EZPNxVxgNhT3_/en"},
		cmsURL:      &url.URL{Scheme: "https", Host: "strapi.accords-library.com"},
		searchURL:   &url.URL{Scheme: "https", Host: "search.accords-library.com"},
		assetsURL:   &url.URL{Scheme: "https", Host: "resha.re", Path: "/accords/"},
		searchToken: "c8774757ecf8122116e3eec4888c98873953c9c75fe18a7165812cc7dda7c0c9",
	}
}

// https://github.com/Accords-Library/accords-library.com/blob/main/src/helpers/libraryItem.ts

func (c *libraryClient) BuildScanArchiveURL(slug string) *url.URL {
	u := *c.assetsURL
	u.Path = path.Join(u.Path, "library", "scans", slug+".zip")
	return &u
}

func (c *libraryClient) BuildTrackURL(itemSlug, trackSlug string) *url.URL {
	u := *c.assetsURL
	u.Path = path.Join(u.Path, "library", "tracks", itemSlug, trackSlug+".mp3")
	return &u
}

func (c *libraryClient) BuildVideoThumbnailURL(uid string) *url.URL {
	return c.assetsURL.ResolveReference(&url.URL{Path: path.Join("videos", uid+".webp")})
}

func (c *libraryClient) BuildVideoFileURL(uid string) *url.URL {
	return c.assetsURL.ResolveReference(&url.URL{Path: path.Join("videos", uid+".mp4")})
}

func (c *libraryClient) BuildAudioURL(slug, langCode string) *url.URL {
	return c.assetsURL.ResolveReference(&url.URL{Path: path.Join("contents", "audios", slug+"_"+langCode+".mp3")})
}

func (c *libraryClient) BuildVideoURL(slug, langCode string) *url.URL {
	return c.assetsURL.ResolveReference(&url.URL{Path: path.Join("contents", "videos", slug+"_"+langCode+".mp4")})
}

func (c *libraryClient) BuildVTTURL(slug, langCode string) *url.URL {
	return c.assetsURL.ResolveReference(&url.URL{Path: path.Join("contents", "videos", slug+"_"+langCode+".vtt")})
}

func (c *libraryClient) GetCMSUrl() *url.URL {
	u := *c.cmsURL
	return &u
}

func fetch200_2(req *http.Request, client *http.Client, out any) ([]byte, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %v", resp.Status)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(raw, out); err != nil {
		return nil, err
	}

	return raw, err
}

func fetchEntity[T any](ctx context.Context, url string, client *http.Client) (*T, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}

	type rawEnvelope struct {
		PageProps json.RawMessage `json:"pageProps"`
	}

	raw := rawEnvelope{}
	if _, err := fetch200_2(req, client, &raw); err != nil {
		return nil, nil, err
	}

	val := new(T)
	if err := json.Unmarshal(raw.PageProps, val); err != nil {
		return nil, nil, err
	}

	return val, raw.PageProps, nil
}

func (c *libraryClient) GetLibraryItem(ctx context.Context, slug string) (*LibraryProps, error) {
	u := *c.libraryURL
	u.Path = path.Join(u.Path, "library", slug+".json")
	u.RawQuery = url.Values{"slug": []string{slug}}.Encode()

	us := u.String()

	val, raw, err := fetchEntity[LibraryProps](ctx, us, c.client)
	if err != nil {
		return nil, err
	}

	val.Raw = raw
	return val, err
}

func (c *libraryClient) GetReader(ctx context.Context, slug string) (*ReaderProps, error) {
	u := *c.libraryURL
	u.Path = path.Join(u.Path, "library", slug, "reader.json")
	u.RawQuery = url.Values{"slug": []string{slug}}.Encode()

	us := u.String()

	val, raw, err := fetchEntity[ReaderProps](ctx, us, c.client)
	if err != nil {
		return nil, err
	}

	val.Raw = raw
	return val, err
}

func (c *libraryClient) ListFolder(ctx context.Context, slug string) (*FolderProps, error) {
	u := *c.libraryURL
	u.Path = path.Join(u.Path, "contents", "folder", slug+".json")
	u.RawQuery = url.Values{"slug": []string{slug}}.Encode()

	us := u.String()

	val, raw, err := fetchEntity[FolderProps](ctx, us, c.client)
	if err != nil {
		return nil, err
	}

	val.Raw = raw
	return val, err
}

func (c *libraryClient) GetContents(ctx context.Context, slug string) (*ContentProps, error) {
	u := *c.libraryURL
	u.Path = path.Join(u.Path, "contents", slug+".json")
	u.RawQuery = url.Values{"slug": []string{slug}}.Encode()

	us := u.String()

	val, raw, err := fetchEntity[ContentProps](ctx, us, c.client)
	if err != nil {
		return nil, err
	}

	val.Raw = raw
	return val, err
}

func search[T any](ctx context.Context, url string, token string, client *http.Client, page, hitsPerPage int, attributes []string) (*SearchResult[T], error) {
	type searchRequest struct {
		Q                     string   `json:"q"`
		HitsPerPage           int      `json:"hitsPerPage"`
		Page                  int      `json:"page"`
		AttributesToRetrieve  []string `json:"attributesToRetrieve"`
		AttributesToHighlight []string `json:"attributesToHighlight"`
		AttributesToCrop      []string `json:"attributesToCrop"`
		Sort                  []string `json:"sort"`
		Filter                []string `json:"filter"`
		HighlightPreTag       string   `json:"highlightPreTag"`
		HighlightPostTag      string   `json:"highlightPostTag"`
		ShowMatchesPosition   bool     `json:"showMatchesPosition"`
		CropLength            int      `json:"cropLength"`
		CropMarker            string   `json:"cropMarker"`
	}

	sreq := searchRequest{
		HitsPerPage:          hitsPerPage,
		Page:                 page,
		AttributesToRetrieve: attributes,
	}

	b, err := json.Marshal(&sreq)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %v", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	sr := &SearchResult[T]{}
	if err := json.Unmarshal(data, sr); err != nil {
		return nil, err
	}

	return sr, nil
}

func (c *libraryClient) SearchLibrary(ctx context.Context, page, hitsPerPage int, attributes []string) (*SearchResult[LibraryItem], error) {
	u := *c.searchURL
	u.Path = path.Join(u.Path, "/indexes/library-item/search")

	return search[LibraryItem](ctx, u.String(), c.searchToken, c.client, page, hitsPerPage, attributes)
}

func (c *libraryClient) GetWikiPage(ctx context.Context, slug string) (*WikiProps, error) {
	u := *c.libraryURL
	u.Path = path.Join(u.Path, "wiki", slug+".json")
	u.RawQuery = url.Values{"slug": []string{slug}}.Encode()

	us := u.String()

	val, raw, err := fetchEntity[WikiProps](ctx, us, c.client)
	if err != nil {
		return nil, err
	}

	val.Raw = raw
	return val, err
}

func (c *libraryClient) SearchWiki(ctx context.Context, page, hitsPerPage int, attributes []string) (*SearchResult[WikiPage], error) {
	u := *c.searchURL
	u.Path = path.Join(u.Path, "/indexes/wiki-page/search")

	return search[WikiPage](ctx, u.String(), c.searchToken, c.client, page, hitsPerPage, attributes)
}

func (c *libraryClient) GetWeaponStory(ctx context.Context, slug string) (*WeaponStoryProps, error) {
	u := *c.libraryURL
	u.Path = path.Join(u.Path, "wiki", "weapons", slug+".json")
	u.RawQuery = url.Values{"slug": []string{slug}}.Encode()

	us := u.String()

	val, raw, err := fetchEntity[WeaponStoryProps](ctx, us, c.client)
	if err != nil {
		return nil, err
	}

	val.Raw = raw
	return val, err
}

func (c *libraryClient) SearchWeaponStories(ctx context.Context, page, hitsPerPage int, attributes []string) (*SearchResult[WeaponStory], error) {
	u := *c.searchURL
	u.Path = path.Join(u.Path, "/indexes/weapon-story/search")

	return search[WeaponStory](ctx, u.String(), c.searchToken, c.client, page, hitsPerPage, attributes)
}

func (c *libraryClient) GetChronology(ctx context.Context) (*ChronologyProps, error) {
	u := *c.libraryURL
	u.Path = path.Join(u.Path, "wiki", "chronology.json")

	us := u.String()

	val, raw, err := fetchEntity[ChronologyProps](ctx, us, c.client)
	if err != nil {
		return nil, err
	}

	val.Raw = raw
	return val, err
}

func (c *libraryClient) GetPost(ctx context.Context, slug string) (*PostProps, error) {
	u := *c.libraryURL
	u.Path = path.Join(u.Path, "news", slug+".json")
	u.RawQuery = url.Values{"slug": []string{slug}}.Encode()

	us := u.String()

	val, raw, err := fetchEntity[PostProps](ctx, us, c.client)
	if err != nil {
		return nil, err
	}

	val.Raw = raw
	return val, err
}

func (c *libraryClient) SearchPosts(ctx context.Context, page, hitsPerPage int, attributes []string) (*SearchResult[Post], error) {
	u := *c.searchURL
	u.Path = path.Join(u.Path, "/indexes/post/search")

	return search[Post](ctx, u.String(), c.searchToken, c.client, page, hitsPerPage, attributes)
}

func (c *libraryClient) GetChronicle(ctx context.Context, slug string) (*ChronicleProps, error) {
	u := *c.libraryURL
	u.Path = path.Join(u.Path, "chronicles", slug+".json")
	u.RawQuery = url.Values{"slug": []string{slug}}.Encode()

	us := u.String()

	val, raw, err := fetchEntity[ChronicleProps](ctx, us, c.client)
	if err != nil {
		return nil, err
	}

	val.Raw = raw
	return val, err
}

func (c *libraryClient) GetChronicles(ctx context.Context) (*ChroniclesProps, error) {
	u := *c.libraryURL
	u.Path = path.Join(u.Path, "chronicles.json")

	us := u.String()

	val, raw, err := fetchEntity[ChroniclesProps](ctx, us, c.client)
	if err != nil {
		return nil, err
	}

	val.Raw = raw
	return val, err
}

func (c *libraryClient) GetVideo(ctx context.Context, uid string) (*VideoProps, error) {
	u := *c.libraryURL
	u.Path = path.Join(u.Path, "archives", "videos", "v", uid+".json")
	u.RawQuery = url.Values{"uid": []string{uid}}.Encode()

	us := u.String()

	val, raw, err := fetchEntity[VideoProps](ctx, us, c.client)
	if err != nil {
		return nil, err
	}

	val.Raw = raw
	return val, err
}

func (c *libraryClient) SearchVideos(ctx context.Context, page, hitsPerPage int, attributes []string) (*SearchResult[Video], error) {
	u := *c.searchURL
	u.Path = path.Join(u.Path, "/indexes/video/search")

	return search[Video](ctx, u.String(), c.searchToken, c.client, page, hitsPerPage, attributes)
}

func (c *libraryClient) GetVideoChannel(ctx context.Context, uid string) (*VideoChannelProps, error) {
	u := *c.libraryURL
	u.Path = path.Join(u.Path, "archives", "videos", "c", uid+".json")
	u.RawQuery = url.Values{"uid": []string{uid}}.Encode()

	us := u.String()

	val, raw, err := fetchEntity[VideoChannelProps](ctx, us, c.client)
	if err != nil {
		return nil, err
	}

	val.Raw = raw
	return val, err
}
