package steam

import (
	"encoding/json"
	"fmt"
	"gbe_fork_helper/config"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings" // Added for strings.Builder
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
func FetchDLCs(appID, libraryPath string) error {
	log.Printf("INFO: Fetching DLCs for AppID %s in library path %s...", appID, libraryPath)

	// Write steam_appid.txt
	appIDFilePath := filepath.Join(libraryPath, "steam_appid.txt")
	if err := os.WriteFile(appIDFilePath, []byte(appID), 0644); err != nil {
		return fmt.Errorf("failed to write steam_appid.txt: %w", err)
	}
	log.Printf("INFO: Wrote steam_appid.txt with AppID %s to %s", appID, appIDFilePath)

	// Prepare for configs.app.ini
	steamSettingsDir := filepath.Join(libraryPath, "steam_settings")
	if err := os.MkdirAll(steamSettingsDir, 0755); err != nil {
		return fmt.Errorf("failed to create steam_settings directory: %w", err)
	}
	configsAppIniPath := filepath.Join(steamSettingsDir, "configs.app.ini")

	var dlcContent strings.Builder
	dlcContent.WriteString("[app::dlcs]\nunlock_all=0\n")

	// Fetch DLCs

	dlcURL := fmt.Sprintf("https://store.steampowered.com/dlc/%s/random/ajaxgetfilteredrecommendations/?query&count=10000", appID)
	resp, err := http.Get(dlcURL)
	if err != nil {
		return fmt.Errorf("failed to fetch DLCs: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	re := regexp.MustCompile(`data-ds-appid=\\"(\d+)`)
	matches := re.FindAllStringSubmatch(string(body), -1)

	if len(matches) == 0 {
		log.Printf("WARN: No DLCs found for AppID %s.", appID)
		return nil
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
		dlcContent.WriteString(fmt.Sprintf("%s=%s\n", dlcID, name))
	}

	if err := os.WriteFile(configsAppIniPath, []byte(dlcContent.String()), 0644); err != nil {
		return fmt.Errorf("failed to write configs.app.ini: %w", err)
	}
	log.Printf("INFO: Wrote DLC configuration to %s", configsAppIniPath)

	return nil
}
