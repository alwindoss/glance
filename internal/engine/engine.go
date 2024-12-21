package engine

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FileData struct {
	Path         string
	MatchSnippet string
}

type model struct {
	files       []FileData
	filtered    []FileData
	textInput   textinput.Model
	cursorIndex int
}

// Initialize the model with asynchronously processed files
func initialModel(directory string) (model, error) {
	files, err := loadFilesWithSnippetsAsync(directory)
	if err != nil {
		return model{}, err
	}

	ti := textinput.New()
	ti.Placeholder = "Start typing to search files..."
	ti.Focus()

	return model{
		files:    files,
		filtered: make([]FileData, 0),
		// filtered:    files,
		textInput:   ti,
		cursorIndex: 0,
	}, nil
}

func (m model) Init() tea.Cmd {
	return nil
}

// Update logic
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			// Exit on Ctrl+C or Esc
			return m, tea.Quit
		case "enter":
			// Confirm selection (print selected file and quit)
			if len(m.filtered) > 0 && m.cursorIndex < len(m.filtered) {
				fmt.Println("Selected file:", m.filtered[m.cursorIndex].Path)
				fmt.Println("Match snippet:", m.filtered[m.cursorIndex].MatchSnippet)
			}
			return m, tea.Quit
		case "up":
			// Move cursor up
			if m.cursorIndex > 0 {
				m.cursorIndex--
			}
		case "down":
			// Move cursor down
			if m.cursorIndex < len(m.filtered)-1 {
				m.cursorIndex++
			}
		}
	}

	// Update text input and filter results
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	m.filtered = filterFiles(m.files, m.textInput.Value())
	m.cursorIndex = clamp(m.cursorIndex, 0, len(m.filtered)-1)

	return m, cmd
}

// View logic
func (m model) View() string {
	var b strings.Builder

	// Render the text input
	b.WriteString("Search files:\n")
	b.WriteString(m.textInput.View())
	b.WriteString("\n\n")

	// Render the filtered list
	for i, file := range m.filtered {
		cursor := " " // no cursor
		if i == m.cursorIndex {
			cursor = ">" // show cursor
		}
		b.WriteString(cursor + " " + file.Path + "\n")
		b.WriteString("    " + file.MatchSnippet + "\n")
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}

// Helper function to filter files based on a query
func filterFiles(files []FileData, query string) []FileData {
	var results []FileData
	query = strings.ToLower(query)

	if query == "" {
		return results
	}
	for _, file := range files {
		if strings.Contains(strings.ToLower(file.Path), query) || strings.Contains(strings.ToLower(file.MatchSnippet), query) {
			results = append(results, file)
		}
	}
	return results
}

// clamp keeps the index within bounds
func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// loadFilesWithSnippetsAsync loads file paths and snippets asynchronously
func loadFilesWithSnippetsAsync(directory string) ([]FileData, error) {
	var files []FileData
	paths := make(chan string)     // Channel for file paths
	results := make(chan FileData) // Channel for processed files
	var wg sync.WaitGroup          // WaitGroup to wait for workers

	// Start worker pool
	numWorkers := 4
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range paths {
				// Process each file
				snippet := getSnippetFromFile(path)
				results <- FileData{Path: path, MatchSnippet: snippet}
			}
		}()
	}

	// Walk directory and send file paths to workers
	go func() {
		defer close(paths)
		filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() {
				paths <- path
			}
			return nil
		})
	}()

	// Collect results asynchronously
	go func() {
		wg.Wait()
		close(results)
	}()

	// Gather results
	for result := range results {
		files = append(files, result)
	}

	return files, nil
}

// getSnippetFromFile reads the file and returns a snippet
func getSnippetFromFile(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return "(could not read file)"
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var snippet string
	for scanner.Scan() {
		line := scanner.Text()
		snippet += line + " "
		if len(snippet) > 200 {
			snippet = snippet[:200] + "..."
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return "(error while reading file)"
	}
	return snippet
}

func Run(dir string) error {
	// Initialize the program model
	m, err := initialModel(dir)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	// Start the Bubble Tea program
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
	return nil
}
