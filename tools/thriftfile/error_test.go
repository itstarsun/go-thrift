package thriftfile

import (
	gotoken "go/token"
	"os"
	"regexp"
	"sort"
	"testing"
)

func (p *errors_) removeMultiples() {
	sort.Sort(p)
	var last gotoken.Position
	i := 0
	for _, e := range *p {
		if e.pos.Filename != last.Filename || e.pos.Line != last.Line {
			last = e.pos
			(*p)[i] = e
			i++
		}
	}
	*p = (*p)[0:i]
}

func getFile(fset *gotoken.FileSet, filename string) (file *gotoken.File) {
	fset.Iterate(func(f *gotoken.File) bool {
		if f.Name() == filename {
			if file != nil {
				panic(filename + " used multiple times")
			}
			file = f
		}
		return true
	})
	return file
}

func getPos(fset *gotoken.FileSet, filename string, offset int) gotoken.Pos {
	if f := getFile(fset, filename); f != nil {
		return f.Pos(offset)
	}
	return gotoken.NoPos
}

var errRx = regexp.MustCompile(`^/\* *ERROR *(HERE|AFTER)? *"([^"]*)" *\*/$`)

func expectedErrors(fset *gotoken.FileSet, filename string, src []byte) map[gotoken.Pos]string {
	errors := make(map[gotoken.Pos]string)

	var s scanner
	s.init(getFile(fset, filename), src, nil)
	var prev gotoken.Pos
	var here gotoken.Pos

	for {
		s.next()
		switch s.tok {
		case _EOF:
			return errors
		case _COMMENT:
			x := errRx.FindStringSubmatch(s.lit())
			if len(x) == 3 {
				pos := s.pos
				if x[1] == "HERE" {
					pos = here // start of comment
				} else if x[1] == "AFTER" {
					pos += gotoken.Pos(len(s.lit())) // end of comment
				} else {
					pos = prev // token prior to comment
				}
				errors[pos] = x[2]
			}
		default:
			prev = s.pos
			var l int
			if s.tok.isLiteral() {
				l = len(s.lit())
			} else {
				l = len(s.tok.String())
			}
			here = prev + gotoken.Pos(l)
		}
	}
}

func compareErrors(t *testing.T, fset *gotoken.FileSet, expected map[gotoken.Pos]string, found errors_) {
	t.Helper()
	for _, error := range found {
		pos := getPos(fset, error.pos.Filename, error.pos.Offset)
		if msg, found := expected[pos]; found {
			rx, err := regexp.Compile(msg)
			if err != nil {
				t.Errorf("%s: %v", error.pos, err)
				continue
			}
			if match := rx.MatchString(error.msg); !match {
				t.Errorf("%s: %q does not match %q", error.pos, error.msg, msg)
				continue
			}
			delete(expected, pos)
		} else {
			t.Errorf("%s: unexpected error: %s", error.pos, error.msg)
		}
	}

	if len(expected) > 0 {
		t.Errorf("%d errors not reported:", len(expected))
		for pos, msg := range expected {
			t.Errorf("%s: %s\n", fset.Position(pos), msg)
		}
	}
}

func checkErrors(t *testing.T, filename string, expectErrors bool) {
	t.Helper()
	src, err := os.ReadFile(filename)
	if err != nil {
		t.Error(err)
		return
	}

	fset := gotoken.NewFileSet()
	_, err = Parse(fset, filename, src)
	found, ok := err.(errors_)
	if err != nil && !ok {
		t.Error(err)
		return
	}
	found.removeMultiples()

	expected := map[gotoken.Pos]string{}
	if expectErrors {
		expected = expectedErrors(fset, filename, src)
	}

	compareErrors(t, fset, expected, found)
}
