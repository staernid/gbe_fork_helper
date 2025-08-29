package github

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"gbe_fork_helper/config"
	"gbe_fork_helper/util"

	"github.com/charmbracelet/glamour"
)

// updateGBE fetches and extracts the latest GBE fork.
func UpdateGBE() error {
	log.Println("INFO: Fetching latest GBE fork from GitHub...")
	gbeHome := filepath.Join(os.Getenv("HOME"), config.GbeDir)
	timestampFile := filepath.Join(gbeHome, ".gbe_timestamp")

	resp, err := http.Get(config.GithubAPIURL)
	if err != nil {
		return fmt.Errorf("failed to fetch release information: %w", err)
	}
	defer resp.Body.Close()

	var release config.Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}

	if _, err := os.Stat(timestampFile); err == nil {
		timestamp, err := os.ReadFile(timestampFile)
		if err == nil && string(timestamp) == release.UpdatedAt.String() {
			log.Println("SUCCESS: GBE fork is already up-to-date.")
			return nil
		}
	}
	// Create a new renderer with the desired style
	// glamour.WithAutoStyle() automatically detects the current terminal's dark/light mode
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
	)

	// Render the markdown text
	renderedText, err := renderer.Render(release.Body)
	if err != nil {
		fmt.Println("Error rendering markdown:", err)
		return nil
	}

	// Print the rendered output to the command line
	fmt.Println(renderedText)

	if err := os.MkdirAll(gbeHome, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", gbeHome, err)
	}

	// Download and extract Linux release
	linuxURL := ""
	for _, asset := range release.Assets {
		if strings.HasSuffix(asset.Name, "linux-release.tar.bz2") {
			linuxURL = asset.BrowserDownloadURL
			break
		}
	}
	if linuxURL == "" {
		return fmt.Errorf("failed to find Linux download URL")
	}

	log.Println("INFO: Downloading Linux release...")
	if err := util.DownloadAndExtract(linuxURL, filepath.Join(gbeHome, "linux_release"), "tar.bz2"); err != nil {
		return fmt.Errorf("failed to update Linux release: %w", err)
	}
	log.Println("SUCCESS: Linux release extracted.")

	// Download and extract Windows release
	winURL := ""
	for _, asset := range release.Assets {
		if strings.HasSuffix(asset.Name, "win-release.7z") {
			winURL = asset.BrowserDownloadURL
			break
		}
	}
	if winURL == "" {
		return fmt.Errorf("failed to find Windows download URL")
	}

	log.Println("INFO: Downloading Windows release...")
	if err := util.DownloadAndExtract(winURL, filepath.Join(gbeHome, "win_release"), "7z"); err != nil {
		return fmt.Errorf("failed to update Windows release: %w", err)
	}
	log.Println("SUCCESS: Windows release extracted.")

	if err := os.WriteFile(timestampFile, []byte(release.UpdatedAt.String()), 0644); err != nil {
		return fmt.Errorf("failed to write timestamp file: %w", err)
	}

	log.Println("SUCCESS: GBE fork updated successfully.")
	return nil
}
