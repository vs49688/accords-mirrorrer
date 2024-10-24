package migstate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"git.vs49688.net/zane/goutils/download"
	"github.com/urfave/cli/v2"
	accords_mirrorrer "github.com/vs49688/accords-mirrorrer"
	"net/url"
	"os"
	"path"
)

func RegisterCommand(app *cli.App) *cli.App {
	app.Commands = append(app.Commands, &cli.Command{
		Name:  "migstate",
		Usage: "nasty hacky state migration",
		Action: func(c *cli.Context) error {
			return migState(c.Context)
		},
		Hidden: true,
	})
	return app
}

func oops() error {
	goodState, err := accords_mirrorrer.LoadState("state.json")
	if err != nil {
		return err
	}

	for _, di := range goodState.Downloads {
		u, _ := url.Parse(di.URL)
		di.OutPath = path.Join(u.Host, u.Path)
	}

	if err := accords_mirrorrer.SaveState(goodState, "state.json"); err != nil {
		return err
	}

	os.Exit(0)
	return nil
}

func migState(_ context.Context) error {

	//oops()

	goodState, err := accords_mirrorrer.LoadState("state2.json")
	if err != nil {
		return err
	}

	type oldState struct {
		Content map[string]*download.DownloadInfo `json:"content"`
		Assets  map[string]*download.DownloadInfo `json:"assets"`
	}

	ooo := oldState{}
	b, err := os.ReadFile("state.json")
	if err != nil {
		return err
	}

	if err := json.Unmarshal(b, &ooo); err != nil {
		return err
	}

	allFiles := make(map[string]*download.DownloadInfo, len(ooo.Assets)+len(ooo.Content))
	for k, v := range ooo.Assets {
		allFiles[k] = v
	}

	for k, v := range ooo.Content {
		allFiles[k] = v
	}

	dirs := map[string]struct{}{}
	files := make(map[string]string, len(allFiles))

	for _, v := range allFiles {
		di, err := goodState.AddDownload(v.URL)
		if err != nil {
			return err
		}

		oldPath := v.OutPath
		newPath := di.OutPath

		files[oldPath] = newPath

		*di = *v
		di.OutPath = newPath

		dir, _ := path.Split(newPath)
		dirs[dir] = struct{}{}

		//fmt.Printf("mv %v %v", oldPath, newPath)
	}

	for dir, _ := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	for oldPath, newPath := range files {
		fmt.Printf("renaming %v => %v\n", oldPath, newPath)
		if err := os.Rename(oldPath, newPath); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return err
		}
	}

	if err := accords_mirrorrer.SaveState(goodState, "state.json"); err != nil {
		return err
	}

	return nil
}
