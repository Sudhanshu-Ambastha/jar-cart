package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

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

func FetchLatestVersions(coords []string) map[string]string {
	results := make(map[string]string)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, coord := range coords {
		wg.Add(1)
		go func(c string) {
			defer wg.Done()
			parts := strings.Split(c, ":")
			if len(parts) < 2 {
				return
			}
			
			latest := GetLatestVersionFromMaven(parts[0], parts[1])
			
			mu.Lock()
			results[c] = latest
			mu.Unlock()
		}(coord)
	}
	wg.Wait()
	return results
}

func GetSearchSuggestions(query string) []struct{ G, A, LatestVersion string } {
    boostedQuery := fmt.Sprintf("a:%s^10 OR g:%s^2", query, query)
    apiURL := fmt.Sprintf("https://search.maven.org/solrsearch/select?q=%s&rows=10&wt=json", url.QueryEscape(boostedQuery))

    resp, err := http.Get(apiURL)
    if err != nil { return nil }
    defer resp.Body.Close()

    var data MavenResponse
    json.NewDecoder(resp.Body).Decode(&data)
    
    var results []struct{ G, A, LatestVersion string }
    for _, doc := range data.Response.Docs {
        results = append(results, struct{ G, A, LatestVersion string }{doc.G, doc.A, doc.LatestVersion})
    }
    return results
}

func GetLatestVersionFromMaven(group, lib string) string {
	apiURL := fmt.Sprintf("https://search.maven.org/solrsearch/select?q=g:%s+AND+a:%s&rows=1&wt=json", 
		url.QueryEscape(group), url.QueryEscape(lib)) 

	resp, err := http.Get(apiURL)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	var data MavenResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err == nil && len(data.Response.Docs) > 0 {
		return data.Response.Docs[0].LatestVersion
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