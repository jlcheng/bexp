package app

import (
	"os"
	"path"
)

func RelPath(unixPath string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return path.Join(wd, unixPath), nil
}
