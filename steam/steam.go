package steam

import (
	"encoding/json"
	"fmt"
	"gbe_fork_helper/config"
	"io"
	"log"
	"net/http"
	"regexp"
)

// fetchAppName gets the app name for a Steam AppID.
func FetchAppName(appID string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("%s/appdetails?appids=%s&filters=basic", config.SteamStoreAPI, appID))
	if err != nil {
		return "", fmt.Errorf("failed to fetch app details: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]struct {
		Data struct {
			Name string `json:"name"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode JSON: %w", err)
	}
	if appData, ok := result[appID]; ok {
		return appData.Data.Name, nil
	}
	return "", fmt.Errorf("app details not found for AppID %s", appID)
}

// fetchDLCs fetches DLCs for a given AppID.
func FetchDLCs(appID string) {
	log.Printf("INFO: Fetching DLCs for AppID %s...", appID)

	dlcURL := fmt.Sprintf("https://store.steampowered.com/dlc/%s/random/ajaxgetfilteredrecommendations/?query&count=10000", appID)
	resp, err := http.Get(dlcURL)
	if err != nil {
		log.Fatalf("ERROR: Failed to fetch DLCs: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("ERROR: Failed to read response body: %v", err)
	}

	re := regexp.MustCompile(`data-ds-appid=\\"(\d+)`)
	matches := re.FindAllStringSubmatch(string(body), -1)

	if len(matches) == 0 {
		log.Printf("WARN: No DLCs found for AppID %s.", appID)
		return
	}

	uniqueDLCs := make(map[string]struct{})
	for _, m := range matches {
		uniqueDLCs[m[1]] = struct{}{}
	}

	for dlcID := range uniqueDLCs {
		name, err := FetchAppName(dlcID)
		if err != nil {
			log.Printf("WARN: Failed to get name for DLC %s: %v", dlcID, err)
			continue
		}
		fmt.Printf("%s=%s\n", dlcID, name)
	}
}
