package library

import (
	"context"
	"encoding/json"
	"net/url"
)

type UploadImageFragment struct {
	Name            string `json:"name"`
	AlternativeText string `json:"alternativeText,omitempty"`
	Caption         string `json:"caption,omitempty"`
	Width           int64  `json:"width,omitempty"`
	Height          int64  `json:"height,omitempty"`
	URL             string `json:"url"`
}

type OgImage struct {
	Image  string `json:"image"`
	Width  int64  `json:"width"`
	Height int64  `json:"height"`
	Alt    string `json:"alt"`
}

type OpenGraph struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Thumbnail   OgImage `json:"thumbnail"`
	Audio       string  `json:"audio,omitempty"`
	Video       string  `json:"video,omitempty"`
}

type Props struct {
	OpenGraph OpenGraph `json:"openGraph"`
}

type Track struct {
	ID    string `json:"id"`
	Slug  string `json:"slug"`
	Title string `json:"title"`
}

type LibraryProps struct {
	Props

	// src/pages/library/[slug]/index.tsx
	Item              *LibraryItem `json:"item"`
	ItemID            string       `json:"itemId"`
	Tracks            []Track      `json:"tracks"`
	IsVariantSet      bool         `json:"isVariantSet"`
	HasContentScans   bool         `json:"hasContentScans"`
	HasContentSection bool         `json:"hasContentSection"`

	Raw json.RawMessage `json:"-"`
}

type Entity interface {
	GetOpenGraph() *OpenGraph
}

func (p *LibraryProps) GetOpenGraph() *OpenGraph {
	return &p.OpenGraph
}

func (p *ReaderProps) GetOpenGraph() *OpenGraph {
	return &p.OpenGraph
}

func (p *FolderProps) GetOpenGraph() *OpenGraph {
	return &p.OpenGraph
}

func (p *ContentProps) GetOpenGraph() *OpenGraph {
	return &p.OpenGraph
}

func (p *WikiProps) GetOpenGraph() *OpenGraph {
	return &p.OpenGraph
}

func (p *WeaponStoryProps) GetOpenGraph() *OpenGraph {
	return &p.OpenGraph
}

func (p *PostProps) GetOpenGraph() *OpenGraph {
	return &p.OpenGraph
}

func (p *ChronologyProps) GetOpenGraph() *OpenGraph {
	return &p.OpenGraph
}

func (p *ChroniclesProps) GetOpenGraph() *OpenGraph {
	return &p.OpenGraph
}

func (p *ChronicleProps) GetOpenGraph() *OpenGraph {
	return &p.OpenGraph
}

func (p *VideoProps) GetOpenGraph() *OpenGraph {
	return &p.OpenGraph
}

func (p *VideoChannelProps) GetOpenGraph() *OpenGraph {
	return &p.OpenGraph
}

type ReaderProps struct {
	Props

	// src/pages/library/[slug]/reader.tsx
	Item      *LibraryItem `json:"item"`
	Pages     []UploadImageFragment
	PageOrder string `json:"pageOrder"`
	BookType  string `json:"bookType"`
	PageWidth int64  `json:"pageWidth"`
	PageRatio string `json:"pageRatio"`
	ItemSlug  string `json:"itemSlug"`

	Raw json.RawMessage `json:"-"`
}

type Folder struct {
	Slug         string `json:"slug"`
	Href         string `json:"href"`
	Translations []struct {
		Title    string `json:"title"`
		Language string `json:"language"`
	} `json:"translations"`
	Fallback struct {
		Title string `json:"title"`
	} `json:"fallback"`
}

type FolderContents struct {
	Slug         string `json:"slug"`
	Href         string `json:"href"`
	Translations []struct {
		Language string `json:"language"`
		PreTitle string `json:"pre_title,omitempty"`
		Title    string `json:"title"`
		Subtitle string `json:"subtitle,omitempty"`
	} `json:"translations"`
	Fallback struct {
		Title string `json:"title"`
	} `json:"fallback"`
	Thumbnail   *UploadImageFragment `json:"thumbnail,omitempty"`
	TopChips    []string             `json:"topChips,omitempty"`
	BottomChips []string             `json:"bottomChips,omitempty"`
}

type FolderProps struct {
	Props

	// src/pages/contents/folder/[slug].tsx
	Subfolders []Folder          `json:"subfolders"`
	Contents   []FolderContents  `json:"contents"`
	Path       []json.RawMessage `json:"path"` // ParentFolderPreviewFragment

	Raw json.RawMessage `json:"-"`
}

type ContentProps struct {
	Props

	// src/pages/contents/[slug].tsx
	Content *Content `json:"content"`

	Raw json.RawMessage `json:"-"`
}

type SearchResult[T any] struct {
	Hits             []*T   `json:"hits"`
	Query            string `json:"query"`
	ProcessingTimeMs int    `json:"processingTimeMs"`
	HitsPerPage      int    `json:"hitsPerPage"`
	Page             int    `json:"page"`
	TotalPages       int    `json:"totalPages"`
	TotalHits        int    `json:"totalHits"`
}

type WikiProps struct {
	Props

	// src/pages/wiki/[slug]/index.tsx
	Page WikiPage `json:"page"`

	Raw json.RawMessage `json:"-"`
}

type WeaponStoryProps struct {
	Props

	PrimaryName string       `json:"primaryName"`
	Aliases     []string     `json:"aliases"`
	Weapon      *WeaponStory `json:"weapon"`

	Raw json.RawMessage `json:"-"`
}

type ChronologyProps struct {
	Props

	ChronologyItems []ChronologyItem `json:"chronologyItems"`
	ChronologyEras  []ChronologyEra  `json:"chronologyEras"`

	Raw json.RawMessage `json:"-"`
}

type PostProps struct {
	Props

	// src/graphql/getPostStaticProps.ts
	Post Post

	Raw json.RawMessage `json:"-"`
}

type ChronicleProps struct {
	Props

	// src/pages/chronicles/[slug]/index.tsx
	Chronicle Chronicle           `json:"chronicle"`
	Chapters  []ChroniclesChapter `json:"chapters"`

	Raw json.RawMessage `json:"-"`
}

type ChroniclesProps struct {
	Props

	// src/pages/chronicles/index.tsx
	Chapters []ChroniclesChapterEntity `json:"chapters"`

	Raw json.RawMessage `json:"-"`
}

type VideoProps struct {
	Props

	// src/pages/archives/videos/v/[uid].tsx
	Channel struct {
		Href        string `json:"href"`
		Subscribers int32  `json:"subscribers"`
		Title       string `json:"title"`
	} `json:"channel"`

	Video struct {
		Uid           string             `json:"uid"`
		IsGone        bool               `json:"isGone"`
		Description   string             `json:"description"`
		Likes         int32              `json:"likes"`
		Source        *ENUM_VIDEO_SOURCE `json:"source,omitempty"`
		PublishedDate string             `json:"publishedDate"`
		Title         string             `json:"title"`
		Views         int32              `json:"views"`
	} `json:"video"`

	Raw json.RawMessage `json:"-"`
}

type VideoChannelProps struct {
	Props

	// src/pages/archives/videos/c/[uid].tsx
	Channel VideoChannel `json:"channel"`

	Raw json.RawMessage `json:"-"`
}

type Client interface {
	GetCMSUrl() *url.URL

	BuildScanArchiveURL(slug string) *url.URL
	BuildTrackURL(itemSlug, trackSlug string) *url.URL
	BuildVideoThumbnailURL(uid string) *url.URL
	BuildVideoFileURL(uid string) *url.URL
	BuildAudioURL(slug, langCode string) *url.URL
	BuildVideoURL(slug, langCode string) *url.URL
	BuildVTTURL(slug, langCode string) *url.URL

	GetLibraryItem(ctx context.Context, slug string) (*LibraryProps, error)
	SearchLibrary(ctx context.Context, page, hitsPerPage int, attributes []string) (*SearchResult[LibraryItem], error)

	GetReader(ctx context.Context, slug string) (*ReaderProps, error)

	ListFolder(ctx context.Context, slug string) (*FolderProps, error)

	GetContents(ctx context.Context, slug string) (*ContentProps, error)

	GetWikiPage(ctx context.Context, slug string) (*WikiProps, error)
	SearchWiki(ctx context.Context, page, hitsPerPage int, attributes []string) (*SearchResult[WikiPage], error)

	GetWeaponStory(ctx context.Context, slug string) (*WeaponStoryProps, error)
	SearchWeaponStories(ctx context.Context, page, hitsPerPage int, attributes []string) (*SearchResult[WeaponStory], error)

	GetPost(ctx context.Context, slug string) (*PostProps, error)
	SearchPosts(ctx context.Context, page, hitsPerPage int, attributes []string) (*SearchResult[Post], error)

	GetChronology(ctx context.Context) (*ChronologyProps, error)

	GetChronicle(ctx context.Context, slug string) (*ChronicleProps, error)

	GetChronicles(ctx context.Context) (*ChroniclesProps, error)

	GetVideo(ctx context.Context, uid string) (*VideoProps, error)
	SearchVideos(ctx context.Context, page, hitsPerPage int, attributes []string) (*SearchResult[Video], error)

	GetVideoChannel(ctx context.Context, uid string) (*VideoChannelProps, error)
}
