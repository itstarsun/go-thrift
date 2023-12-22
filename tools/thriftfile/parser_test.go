package thriftfile

import (
	"io/fs"
	"path/filepath"
	"testing"
)

func TestParser(t *testing.T) {
	const testdata = "testdata"
	err := filepath.WalkDir(testdata, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".thrift" {
			return err
		}
		name, _ := filepath.Rel(testdata, path)
		t.Run(name, func(t *testing.T) {
			checkErrors(t, path, true)
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
