package utils

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"github.com/Sudhanshu-Ambastha/jar-cart/src/ui/components"
)

type SearchResult struct {
	G             string
	A             string
	LatestVersion string
}
type model struct {
	spinner spinner.Model
	loading bool
	results []SearchResult
	query   string
	err     error
}

type searchMsg []SearchResult
type errMsg struct{ err error }

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.performSearchCmd)
}

func (m model) performSearchCmd() tea.Msg {
	results := GetSearchSuggestions(m.query)
	if results == nil {
		return errMsg{fmt.Errorf("no results found")}
	}
	return searchMsg(results)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case searchMsg:
		m.loading = false
		m.results = msg
		return m, tea.Quit
	case errMsg:
		m.err = msg.err
		m.loading = false
		return m, tea.Quit
	}
	return m, nil
}

func (m model) View() tea.View {
	if m.loading {
		return tea.NewView(fmt.Sprintf("\n 🔍 Searching Maven Central for: %s %s\n", 
			m.query, m.spinner.View()))
	}
	return tea.NewView("")
}

type MavenResponse struct {
	Response struct {
		Docs []struct {
			ID            string `json:"id"`
			G             string `json:"g"`
			A             string `json:"a"`
			LatestVersion string `json:"latestVersion"`
		} `json:"docs"`
	} `json:"response"`
}

type MavenMetadata struct {
    Versioning struct {
        Latest string `xml:"latest"`
    } `xml:"versioning"`
}

func FetchLatestVersions(coords []string) map[string]string {
	results := make(map[string]string)
	var mu sync.Mutex
	tasks := make(chan string, len(coords))
	resultsChan := make(chan struct {
		coord   string
		version string
	}, len(coords))

	numWorkers := 5
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for c := range tasks {
				parts := strings.Split(c, ":")
				if len(parts) >= 2 {
					version := GetLatestVersionFromMaven(parts[0], parts[1])
					resultsChan <- struct{ coord, version string }{c, version}
				}
			}
		}()
	}

	for _, c := range coords {
		tasks <- c
	}
	close(tasks)

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	for res := range resultsChan {
		mu.Lock()
		results[res.coord] = res.version
		mu.Unlock()
	}

	return results
}

func GetBestMatch(query string) (group string, artifact string, latest string, err error) {
    results := GetSearchSuggestions(query)
    if len(results) == 0 {
        return "", "", "", fmt.Errorf("no results found")
    }
	
    return results[0].G, results[0].A, results[0].LatestVersion, nil
}

func ScanLocalCache(query string) []string {
    home, _ := os.UserHomeDir()
    cacheDir := filepath.Join(home, ".jar-cart", "cache")
    var matches []string

    if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
        return matches 
    }

    filepath.WalkDir(cacheDir, func(path string, d os.DirEntry, err error) error {
        if err != nil { return nil }
        
        if !d.IsDir() && strings.Contains(d.Name(), query) {
            rel, _ := filepath.Rel(cacheDir, path)
            matches = append(matches, rel)
        }
        return nil
    })
    return matches
}

func GetSearchSuggestions(query string) []SearchResult {
	apiURL := fmt.Sprintf("https://search.maven.org/solrsearch/select?q=%s&rows=20&wt=json", url.QueryEscape(query))
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second, 
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ResponseHeaderTimeout: 15 * time.Second, 
		TLSHandshakeTimeout:   10 * time.Second,
	}

	client := &http.Client{
		Timeout:   30 * time.Second, 
		Transport: transport,
	}

	resp, err := client.Get(apiURL)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var data MavenResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil
	}

	var results []SearchResult
	for _, doc := range data.Response.Docs {
		results = append(results, SearchResult{
			G:             doc.G,
			A:             doc.A,
			LatestVersion: doc.LatestVersion,
		})
	}
	return results
}

func GetLatestVersionFromMaven(group, lib string) string {
    groupPath := strings.ReplaceAll(group, ".", "/")
    url := fmt.Sprintf("https://repo1.maven.org/maven2/%s/%s/maven-metadata.xml", groupPath, lib)

    resp, err := http.Get(url)
    if err != nil || resp.StatusCode != http.StatusOK {
        return ""
    }
    defer resp.Body.Close()

    var meta MavenMetadata
    if err := xml.NewDecoder(resp.Body).Decode(&meta); err == nil {
        return meta.Versioning.Latest
    }
    return ""
}

func SearchMavenCentral(query string) {
	s := components.NewStyledSpinner(spinner.Points, "86") 
	m := model{
		spinner: s,
		loading: true,
		query:   query,
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("❌ TUI Error: %v\n", err)
		os.Exit(1)
	}

	m = finalModel.(model)
	fmt.Printf("DEBUG: Search complete. Found %d results for '%s'\n", len(m.results), query)
	if len(m.results) > 0 {
        columns := []table.Column{
            {Title: "Coordinate (Group:Artifact)", Width: 55},
            {Title: "Latest", Width: 15},
        }

        rows := []table.Row{}
        for _, doc := range m.results {
            rows = append(rows, table.Row{doc.G + ":" + doc.A, doc.LatestVersion})
        }

        t := components.NewDependencyTable(columns, rows)
        fmt.Println(t.View())
    } else {
        fmt.Println("ℹ️ No results found. (Maven Central may be rate-limiting. Try again in a moment.)")
    }
}