package app

import (
	"context"

	"github.com/openconfig/gnoi/file"
)

func (a *App) isDir(ctx context.Context, fileClient file.FileClient, path string) (bool, error) {
	r, err := fileClient.Stat(ctx, &file.StatRequest{
		Path: path,
	})
	if err != nil {
		return false, err
	}
	numStats := len(r.Stats)
	// if number of stats is 0 or more than one, it's a directory
	if numStats != 1 {
		return true, nil
	}
	// else if number of stats is 1, and the returned stat has a different path,
	// it's a directory with a single file/sub directory
	if r.Stats[0].Path != path {
		return true, nil
	}
	// otherwise it's a file
	return false, nil
}
