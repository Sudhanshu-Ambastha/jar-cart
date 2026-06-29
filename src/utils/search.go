package utils

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type SearchResult struct {
	G             string
	A             string
	LatestVersion string
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
	apiURL := fmt.Sprintf("https://search.maven.org/solrsearch/select?q=a:%s*&rows=20&wt=json", url.QueryEscape(query))

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		fmt.Printf("❌ Network error: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("❌ Search API returned status: %d\n", resp.StatusCode)
		return nil
	}

	var data MavenResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		fmt.Printf("❌ Failed to parse JSON: %v\n", err)
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
	fmt.Printf("🔍 Searching Maven Central for: \033[36m%s\033[0m...\n\n", query)

	boostedQuery := fmt.Sprintf("a:%s^10 OR g:%s^2 OR %s", query, query, query)
	apiURL := fmt.Sprintf("https://search.maven.org/solrsearch/select?q=%s&rows=10&wt=json", url.QueryEscape(boostedQuery))

	resp, err := http.Get(apiURL)
	if err != nil {
		fmt.Printf("❌ Failed to reach search endpoint: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var data MavenResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		fmt.Printf("❌ Failed to parse response: %v\n", err)
		return
	}

	if len(data.Response.Docs) == 0 {
		fmt.Println("ℹ️ No matching dependencies found.")
		return
	}

	fmt.Printf("\033[1m%-55s %-20s\033[0m\n", "Coordinate Notation (Group:Artifact)", "Latest Version")
	fmt.Println("-----------------------------------------------------------------------------")
	for _, doc := range data.Response.Docs {
		fmt.Printf("\033[32m%-55s\033[0m %-20s\n", doc.G+":"+doc.A, doc.LatestVersion)
	}
}