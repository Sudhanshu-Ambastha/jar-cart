package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Sudhanshu-Ambastha/jar-cart/src/models"
)

type OSVQuery struct {
	Package struct {
		Name      string `json:"name"`
		Ecosystem string `json:"ecosystem"`
	} `json:"package"`
	Version string `json:"version"`
}

type OSVBatchQuery struct {
	Queries []OSVQuery `json:"queries"`
}

type OSVResponse struct {
	Results []struct {
		Vulns []struct {
			ID      string `json:"id"`
			Summary string `json:"summary"`
			Details string `json:"details"`
		} `json:"vulns"`
	} `json:"results"`
}

func CheckVulnerabilities(dependencies []models.Dependency) (*OSVResponse, error) {
	batch := OSVBatchQuery{}

	for _, dep := range dependencies {
		batch.Queries = append(batch.Queries, OSVQuery{
			Package: struct {
				Name      string `json:"name"`
				Ecosystem string `json:"ecosystem"`
			}{
				Name:      dep.Library, 
				Ecosystem: "Maven",
			},
			Version: dep.Version,
		})
	}

	if len(batch.Queries) == 0 {
		return &OSVResponse{Results: []struct {
			Vulns []struct {
				ID      string `json:"id"`
				Summary string `json:"summary"`
				Details string `json:"details"`
			} `json:"vulns"`
		}{}}, nil
	}

	jsonData, err := json.Marshal(batch)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Post("https://api.osv.dev/v1/querybatch", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OSV API returned status: %d", resp.StatusCode)
	}

	var result OSVResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func SplitCoord(coord string) []string {
	return strings.Split(coord, ":")
}