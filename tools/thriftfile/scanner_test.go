package thriftfile

import (
	gotoken "go/token"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestScanner(t *testing.T) {
	const testdata = "testdata"
	err := filepath.WalkDir(testdata, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".thrift" {
			return err
		}
		name, _ := filepath.Rel(testdata, path)
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var s scanner
			s.init(gotoken.NewFileSet().AddFile(path, -1, len(data)), data, func(pos gotoken.Position, msg string) {
				t.Errorf("%s: %s", pos, msg)
			})
			for {
				s.next()
				if s.tok == _EOF {
					break
				}
				t.Log(s.file.Position(s.pos), s.tok, s.lit())
			}
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
