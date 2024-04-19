package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

// TODO: add a 'topN' int flag, condense others into "Other" bucket.

type Score struct {
	Name   string
	Points int
}

type Model struct {
	total     int
	scores    []Score
	scoreChan chan Score
	width     int
}

var fetchUrl = flag.String("url", "", "url to fetch content from")
var jsonField = flag.String("json", "", "top-level json field to read from body")
var headerField = flag.String("header", "", "HTTP header to read for bucket")
var reqInterval = flag.Duration("interval", 3*time.Second, "time between requests")
var useStatus = flag.Bool("status", false, "uses http status as bucket")

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		sendEvent(m.scoreChan),
		readEvent(m.scoreChan),
		tickUrlfetch(),
	)
}
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

			// any other key scores points
			// default:
			// 	return m, func() tea.Msg {
			// 		return Score{Name: msg.String(), Points: 1}
			// 	}
		}
	case Score:
		m.total += msg.Points
		for i, s := range m.scores {
			if s.Name == msg.Name {
				m.scores[i].Points += msg.Points
				// return m, readEvent(m.scoreChan)
				return m, tickUrlfetch()
			}
		}
		m.scores = append(m.scores, msg)
		// return m, readEvent(m.scoreChan)
		return m, tickUrlfetch()

	default:
		fmt.Printf("not sure about: %#v\n", msg)
	}
	return m, nil

}
func readEvent(ch chan Score) tea.Cmd {
	return func() tea.Msg {
		return <-ch
	}
}
func sendEvent(ch chan Score) tea.Cmd {
	return func() tea.Msg {
		for {
			if inDecoder.More() {
				var m map[string]interface{}
				err := inDecoder.Decode(&m)
				if err != nil {
					slog.Error("Failed to decode message", "error", err)
				}
				time.Sleep(600 * time.Millisecond)
				ch <- Score{Name: fmt.Sprint(m["name"]), Points: 1}
			}
		}
	}
}
func (m Model) View() string {
	rval := "\n"

	if m.total == 0 {
		return "no data yet\n"
	}

	sort.Slice(m.scores, func(i, j int) bool {
		return m.scores[i].Points > m.scores[j].Points
	})

	for _, s := range m.scores {
		rval += fmt.Sprintf("Bucket: %s\n", s.Name)
		p := float64(s.Points) / float64(m.total)
		rval += progress.New(progress.WithWidth(m.width)).ViewAs(p)
		rval += "\n\n"
	}
	return rval
}

func initialModel() Model {

	m := Model{
		total:     0,
		scores:    make([]Score, 0),
		scoreChan: make(chan Score),
	}
	return m
}

var inDecoder *json.Decoder

func main() {
	flag.Parse()
	inDecoder = json.NewDecoder(os.Stdin)
	p := tea.NewProgram(initialModel(), tea.WithInput(nil) /* tea.WithAltScreen() */)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

func tickUrlfetch() tea.Cmd {
	return tea.Tick(*reqInterval, func(t time.Time) tea.Msg {
		resp, err := http.Get(*fetchUrl)
		if err != nil {
			log.Print(err)
		}
		defer resp.Body.Close()
		var m = make(map[string]interface{})
		b, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(b, &m)
		if err != nil {
			log.Print(err)
		}
		var name string
		if len(*jsonField) > 0 {
			name = fmt.Sprint(m[*jsonField])
		}
		if len(*headerField) > 0 {
			name = resp.Header.Get(*headerField)
		}
		if *useStatus {
			name = resp.Status
		}
		return Score{Name: name, Points: 1}
	})

}
