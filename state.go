package accords_mirrorrer

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"

	"git.vs49688.net/zane/goutils/download"

	"github.com/vs49688/accords-mirrorrer/library"
)

const (
	CurrentStateVersion = "2"
)

type State struct {
	Version    string                     `json:"version,omitempty"`
	RawFolders map[string]json.RawMessage `json:"folders"`
	RawContent map[string]json.RawMessage `json:"content"`

	RawLibrary        map[string]json.RawMessage `json:"library"`
	RawReader         map[string]json.RawMessage `json:"reader"`
	RawWiki           map[string]json.RawMessage `json:"wiki"`
	RawWeaponsStories map[string]json.RawMessage `json:"weapon_stories"`
	RawChronology     json.RawMessage            `json:"chronology"`
	RawPosts          map[string]json.RawMessage `json:"posts"`
	RawVideos         map[string]json.RawMessage `json:"videos"`
	RawVideoChannels  map[string]json.RawMessage `json:"video_channels"`

	Chronicles struct {
		RawIndex json.RawMessage          `json:"index"`
		Index    *library.ChroniclesProps `json:"-"`

		RawEntries map[string]json.RawMessage         `json:"entries"`
		Entries    map[string]*library.ChronicleProps `json:"-"`
	} `json:"chronicles"`

	Downloads map[string]*download.DownloadInfo `json:"downloads"`

	Folders       map[string]*library.FolderProps       `json:"-"`
	Content       map[string]*library.ContentProps      `json:"-"`
	Library       map[string]*library.LibraryProps      `json:"-"`
	Reader        map[string]*library.ReaderProps       `json:"-"`
	Wiki          map[string]*library.WikiProps         `json:"-"`
	WeaponStories map[string]*library.WeaponStoryProps  `json:"-"`
	Chronology    *library.ChronologyProps              `json:"-"`
	Posts         map[string]*library.PostProps         `json:"-"`
	Videos        map[string]*library.VideoProps        `json:"-"`
	VideoChannels map[string]*library.VideoChannelProps `json:"-"`
}

func (s *State) AddDownload(uu string) (*download.DownloadInfo, error) {
	if di, exists := s.Downloads[uu]; exists {
		return di, nil
	}

	u, err := url.Parse(uu)
	if err != nil {
		return nil, err
	}

	di := &download.DownloadInfo{
		URL:     u.String(),
		OutPath: path.Join(u.Host, u.Path),
	}

	s.Downloads[uu] = di

	return di, nil
}

func migrateV1(data []byte) ([]byte, error) {
	type v1Data struct {
		PageProps json.RawMessage `json:"pageProps"`
	}

	type v1State struct {
		Version    string                            `json:"version,omitempty"`
		RawIndex   map[string]json.RawMessage        `json:"index"`
		RawLibrary map[string]json.RawMessage        `json:"library"`
		Content    map[string]*download.DownloadInfo `json:"content"`
		Assets     map[string]*download.DownloadInfo `json:"assets"`
	}

	s := &v1State{}
	if err := json.Unmarshal(data, s); err != nil {
		return nil, err
	}

	unnest := func(m map[string]json.RawMessage) (map[string]json.RawMessage, error) {
		out := make(map[string]json.RawMessage, len(m))
		for k, v := range m {
			d := v1Data{}
			if err := json.Unmarshal(v, &d); err != nil {
				return nil, err
			}

			out[k] = d.PageProps
		}

		return out, nil
	}

	rawLibrary, err := unnest(s.RawLibrary)
	if err != nil {
		return nil, err
	}

	rawIndex, err := unnest(s.RawIndex)
	if err != nil {
		return nil, err
	}

	fixedIndex := make(map[string]json.RawMessage, len(rawIndex))
	for k, v := range rawIndex {
		_, slug := path.Split(k)
		fixedIndex[slug] = v
	}

	downloads := make(map[string]*download.DownloadInfo, len(s.Content)+len(s.Assets))
	for k, ct := range s.Content {
		downloads[k] = ct
	}

	for k, asset := range s.Assets {
		downloads[k] = asset
	}

	ss := &State{
		Version:    CurrentStateVersion,
		RawFolders: fixedIndex,
		RawLibrary: rawLibrary,
		Downloads:  downloads,
	}
	return json.Marshal(ss)
}

func updateState(data []byte) (*State, error) {
	type stateVersion struct {
		Version string `json:"version"`
	}
	var err error

	s := &stateVersion{}
again:

	if err := json.Unmarshal(data, s); err != nil {
		return nil, err
	}

	switch s.Version {
	case "", "1":
		if data, err = migrateV1(data); err != nil {
			return nil, err
		}
		goto again
	case CurrentStateVersion:
		break
	default:
		return nil, fmt.Errorf("unknown state version: %s", s.Version)
	}

	ss := &State{}
	if err := json.Unmarshal(data, ss); err != nil {
		return nil, err
	}

	return ss, nil
}

func unraw[T any](raw json.RawMessage) (*T, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	val := new(T)
	if err := json.Unmarshal(raw, val); err != nil {
		return nil, err
	}

	reflect.ValueOf(val).Elem().FieldByName("Raw").SetBytes(raw)
	return val, nil
}

func unrawMap[T any](raw map[string]json.RawMessage) (map[string]*T, error) {
	out := make(map[string]*T, len(raw))

	for k, v := range raw {
		val, err := unraw[T](v)
		if err != nil {
			return nil, err
		}

		out[k] = val
	}

	return out, nil
}

func LoadState(path string) (*State, error) {
	b, err := os.ReadFile(filepath.Clean(path))
	if errors.Is(err, os.ErrNotExist) {
		b = []byte(`{}`)
	}

	state, err := updateState(b)
	if err != nil {
		return nil, err
	}

	if state.RawFolders == nil {
		state.RawFolders = map[string]json.RawMessage{}
	}

	if state.Folders, err = unrawMap[library.FolderProps](state.RawFolders); err != nil {
		return nil, err
	}

	if state.RawContent == nil {
		state.RawContent = map[string]json.RawMessage{}
	}

	if state.Content, err = unrawMap[library.ContentProps](state.RawContent); err != nil {
		return nil, err
	}

	if state.RawLibrary == nil {
		state.RawLibrary = map[string]json.RawMessage{}
	}

	if state.Library, err = unrawMap[library.LibraryProps](state.RawLibrary); err != nil {
		return nil, err
	}

	if state.RawReader == nil {
		state.RawReader = map[string]json.RawMessage{}
	}

	if state.Reader, err = unrawMap[library.ReaderProps](state.RawReader); err != nil {
		return nil, err
	}

	if state.RawWiki == nil {
		state.RawWiki = map[string]json.RawMessage{}
	}

	if state.Wiki, err = unrawMap[library.WikiProps](state.RawWiki); err != nil {
		return nil, err
	}

	if state.RawWeaponsStories == nil {
		state.RawWeaponsStories = map[string]json.RawMessage{}
	}

	if state.WeaponStories, err = unrawMap[library.WeaponStoryProps](state.RawWeaponsStories); err != nil {
		return nil, err
	}

	if state.RawPosts == nil {
		state.RawPosts = map[string]json.RawMessage{}
	}

	if state.Posts, err = unrawMap[library.PostProps](state.RawPosts); err != nil {
		return nil, err
	}

	if state.Chronology, err = unraw[library.ChronologyProps](state.RawChronology); err != nil {
		return nil, err
	}

	if state.Chronicles.Index, err = unraw[library.ChroniclesProps](state.Chronicles.RawIndex); err != nil {
		return nil, err
	}

	if state.Chronicles.RawEntries == nil {
		state.Chronicles.RawEntries = map[string]json.RawMessage{}
	}

	if state.Chronicles.Entries, err = unrawMap[library.ChronicleProps](state.Chronicles.RawEntries); err != nil {
		return nil, err
	}

	if state.RawVideos == nil {
		state.RawVideos = map[string]json.RawMessage{}
	}

	if state.Videos, err = unrawMap[library.VideoProps](state.RawVideos); err != nil {
		return nil, err
	}

	if state.RawVideoChannels == nil {
		state.RawVideoChannels = map[string]json.RawMessage{}
	}

	if state.VideoChannels, err = unrawMap[library.VideoChannelProps](state.RawVideoChannels); err != nil {
		return nil, err
	}

	if state.Downloads == nil {
		state.Downloads = map[string]*download.DownloadInfo{}
	}

	return state, nil
}

func SaveState(state *State, path string) error {
	state.Version = CurrentStateVersion

	b, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, b, 0600); err != nil {
		return err
	}

	return nil
}
