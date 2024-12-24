package search

import (
	"bufio"
	"io"
	"os"
	"strings"
)

func Search(path, query string) {
}

type Finder interface {
	Find(query string) error
}

func NewFzFinder(path string) Finder {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	fz := &fzFinder{
		src: f,
	}

	return fz
}

type fzFinder struct {
	src io.ReadWriteCloser
}

func (fz *fzFinder) Find(query string) error {
	scanner := bufio.NewScanner(fz.src)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		if strings.Contains(scanner.Text(), strings.TrimSpace(query)) {
			ln := lineNum
		}
	}
	return nil
}
