package search

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type MatchType uint
type Matches []Match

var results = make(map[string]Matches)

const (
	ExactMatch MatchType = iota
	PartialMatch
)

type Match struct {
	FilePath    string
	LineNumber  uint
	MatchString string
	MatchType   MatchType
}

func Search(path, query string, parallel int) {
	wg := &sync.WaitGroup{}
	filePathCh := make(chan string, 100)
	wg.Add(1)
	go func(path string) {
		defer wg.Done()
		getAllFiles(path, filePathCh)
	}(path)
	matchCh := make(chan Match, 100)
	wg.Add(1)
	go func() {
		defer wg.Done()
		numOfWorkers := parallel
		for i := 0; i < numOfWorkers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				findInFile(query, filePathCh, matchCh)
			}()
		}
	}()

	for m := range matchCh {
		fmt.Println("###############################")
		fmt.Printf("File Path: %s\n", m.FilePath)
		fmt.Printf("Line Num: %d\n", m.LineNumber)
		fmt.Printf("Matching text: %s\n", m.MatchString)
		fmt.Println("###############################")
		fmt.Printf("\n\n")
	}
	wg.Wait()
}

func findInFile(query string, filePathCh chan string, matchCh chan Match) {
	for path := range filePathCh {
		query = strings.ToLower(query)
		file, err := os.Open(path)
		if err != nil {
			log.Printf("error opening file %s", path)
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineNum := 0
		for scanner.Scan() {
			line := scanner.Text()
			lineNum++
			if strings.Contains(strings.ToLower(line), query) {
				m := Match{
					FilePath:    path,
					LineNumber:  uint(lineNum),
					MatchString: line,
					MatchType:   ExactMatch,
				}
				matchCh <- m
			}
		}
	}
}

func getAllFiles(directory string, filePathCh chan string) {
	filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			filePathCh <- path
		}
		return nil
	})

}
