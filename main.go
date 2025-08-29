// main.go
package main

import (
	"archive/tar"
	"bytes"
	"compress/bzip2"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
	"golang.org/x/crypto/md4"
)

// Global Configuration
const (
	gbeDir         = ".local/share/gbe_fork"
	steamStoreAPI  = "https://store.steampowered.com/api"
	githubAPIURL   = "https://api.github.com/repos/Detanup01/gbe_fork/releases/latest"
	sevenZCommand  = "7z"
	stringsCommand = "strings"
)

// platformConfig maps platform names to their configuration.
var platformConfig = map[string]struct {
	Subdir, Target, Additional, Generator, Arch string
}{
	"linux": {
		Subdir:     "linux_release",
		Target:     "libsteam_api.so",
		Additional: "steamclient.so",
		Generator:  "generate_interfaces_x64",
		Arch:       "64",
	},
	"win64": {
		Subdir:     "win_release",
		Target:     "steam_api64.dll",
		Additional: "steamclient64.dll",
		Generator:  "generate_interfaces_x64.exe",
		Arch:       "64",
	},
	"win32": {
		Subdir:     "win_release",
		Target:     "steam_api.dll",
		Additional: "steamclient.dll",
		Generator:  "generate_interfaces_x32.exe",
		Arch:       "32",
	},
}

// Release represents a GitHub release.
type Release struct {
	Assets []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
}

// runCmd executes a command and returns its output or an error.
func runCmd(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("command failed: %s %v\nstdout: %s\nstderr: %s", name, args, stdout.String(), stderr.String())
	}
	return stdout.Bytes(), nil
}

// getHash returns the MD5 hash of a file.
func getHash(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	h := md4.New()
	if _, err := h.Write(data); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// backupAndReplace backs up a file and replaces it.
func backupAndReplace(src, dest string) error {
	timestamp := time.Now().Format("20060102-150405")
	backupPath := fmt.Sprintf("%s.%s.ORIGINAL", dest, timestamp)

	if err := os.Rename(dest, backupPath); err != nil {
		return fmt.Errorf("failed to backup %s: %w", dest, err)
	}
	log.Printf("INFO: Backed up '%s' to '%s'", dest, backupPath)

	if err := os.Link(src, dest); err != nil {
		// Fallback to copy if hard link fails
		if err := copyFile(src, dest); err != nil {
			return fmt.Errorf("failed to copy %s to %s: %w", src, dest, err)
		}
	}
	log.Printf("INFO: Replaced with '%s'", src)

	return nil
}

// copyFile is a helper function to copy a file.
func copyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// fetchAppName gets the app name for a Steam AppID.
func fetchAppName(appID string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("%s/appdetails?appids=%s&filters=basic", steamStoreAPI, appID))
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
func fetchDLCs(appID string) {
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
		name, err := fetchAppName(dlcID)
		if err != nil {
			log.Printf("WARN: Failed to get name for DLC %s: %v", dlcID, err)
			continue
		}
		fmt.Printf("%s=%s\n", dlcID, name)
	}
}

// applyGBE applies the GBE patch to a specified platform.
func applyGBE(platform string) error {
	config, ok := platformConfig[platform]
	if !ok {
		var validPlatforms []string
		for p := range platformConfig {
			validPlatforms = append(validPlatforms, p)
		}
		return fmt.Errorf("invalid platform: '%s'. Valid platforms: %s", platform, strings.Join(validPlatforms, ", "))
	}

	gbePath := filepath.Join(os.Getenv("HOME"), gbeDir, config.Subdir, "experimental", "x"+config.Arch)

	targetFiles, err := filepath.Glob(filepath.Join(".", "**", config.Target))
	if err != nil {
		return fmt.Errorf("failed to search for files: %w", err)
	}

	if len(targetFiles) == 0 {
		log.Printf("WARN: No target files found for platform '%s'.", platform)
	}

	sourceFile := filepath.Join(gbePath, config.Target)
	if _, err := os.Stat(sourceFile); os.IsNotExist(err) {
		return fmt.Errorf("source file not found: '%s'", sourceFile)
	}

	sourceHash, err := getHash(sourceFile)
	if err != nil {
		return fmt.Errorf("failed to get hash of source file: %w", err)
	}

	for _, file := range targetFiles {
		log.Printf("INFO: Found potential target: '%s'", file)

		targetHash, err := getHash(file)
		if err != nil {
			log.Printf("ERROR: Failed to get hash of '%s': %v. Skipping.", file, err)
			continue
		}

		if targetHash == sourceHash {
			log.Println("SUCCESS: File is already up-to-date. Skipping.")
			continue
		}

		// Check for GBE fork strings
		output, err := runCmd(stringsCommand, file)
		if err == nil && bytes.Contains(output, []byte("gbe_fork")) {
			log.Printf("WARN: File '%s' appears to be an existing GBE fork. Skipping.", file)
			continue
		}

		if err := backupAndReplace(sourceFile, file); err != nil {
			log.Printf("ERROR: Failed to replace file '%s': %v. Skipping.", file, err)
			continue
		}

		if config.Additional != "" {
			additionalSource := filepath.Join(gbePath, config.Additional)
			additionalDest := filepath.Join(filepath.Dir(file), config.Additional)
			if _, err := os.Stat(additionalSource); err == nil {
				if err := backupAndReplace(additionalSource, additionalDest); err != nil {
					log.Printf("WARN: Failed to replace additional file '%s': %v", additionalDest, err)
				}
			}
		}

		generatorPath := filepath.Join(os.Getenv("HOME"), gbeDir, config.Subdir, "tools", "generate_interfaces", config.Generator)
		if _, err := os.Stat(generatorPath); err == nil {
			log.Printf("INFO: Running generator '%s'...", config.Generator)
			if runtime.GOOS != "windows" {
				if err := os.Chmod(generatorPath, 0755); err != nil {
					log.Printf("WARN: Failed to set executable permissions on '%s': %v", generatorPath, err)
				}
			}
			cmd := exec.Command(generatorPath, filepath.Base(file))
			cmd.Dir = filepath.Dir(file)
			if out, err := cmd.CombinedOutput(); err != nil {
				log.Printf("ERROR: Generator failed: %v\nOutput: %s", err, string(out))
			}
		}
	}

	log.Println("SUCCESS: GBE application process completed.")
	return nil
}

// updateGBE fetches and extracts the latest GBE fork.
func updateGBE() error {
	log.Println("INFO: Fetching latest GBE fork from GitHub...")
	gbeHome := filepath.Join(os.Getenv("HOME"), gbeDir)
	timestampFile := filepath.Join(gbeHome, ".gbe_timestamp")

	resp, err := http.Get(githubAPIURL)
	if err != nil {
		return fmt.Errorf("failed to fetch release information: %w", err)
	}
	defer resp.Body.Close()

	var release Release
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
	if err := downloadAndExtract(linuxURL, filepath.Join(gbeHome, "linux_release"), "tar.bz2"); err != nil {
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
	if err := downloadAndExtract(winURL, filepath.Join(gbeHome, "win_release"), "7z"); err != nil {
		return fmt.Errorf("failed to update Windows release: %w", err)
	}
	log.Println("SUCCESS: Windows release extracted.")

	if err := os.WriteFile(timestampFile, []byte(release.UpdatedAt.String()), 0644); err != nil {
		return fmt.Errorf("failed to write timestamp file: %w", err)
	}

	log.Println("SUCCESS: GBE fork updated successfully.")
	return nil
}

// downloadAndExtract downloads a file and extracts it.
func downloadAndExtract(url, destDir, format string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := os.RemoveAll(destDir); err != nil {
		return err
	}
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	switch format {
	case "tar.bz2":
		bzip2Reader := bzip2.NewReader(resp.Body)
		tarReader := tar.NewReader(bzip2Reader)
		for {
			header, err := tarReader.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
			targetPath := filepath.Join(destDir, header.Name)
			if header.FileInfo().IsDir() {
				if err := os.MkdirAll(targetPath, header.FileInfo().Mode()); err != nil {
					return err
				}
				continue
			}
			file, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, header.FileInfo().Mode())
			if err != nil {
				return err
			}
			if _, err := io.Copy(file, tarReader); err != nil {
				file.Close()
				return err
			}
			file.Close()
		}
	case "7z":
		tempFile := filepath.Join(os.TempDir(), "temp.7z")
		outFile, err := os.Create(tempFile)
		if err != nil {
			return err
		}
		if _, err := io.Copy(outFile, resp.Body); err != nil {
			outFile.Close()
			return err
		}
		outFile.Close()

		if _, err := runCmd(sevenZCommand, "x", tempFile, fmt.Sprintf("-o%s", destDir), "-y"); err != nil {
			os.Remove(tempFile)
			return err
		}

		// Move contents of 'release' subdirectory up
		releasePath := filepath.Join(destDir, "release")
		if _, err := os.Stat(releasePath); err == nil {
			entries, err := os.ReadDir(releasePath)
			if err != nil {
				return err
			}
			for _, entry := range entries {
				if err := os.Rename(filepath.Join(releasePath, entry.Name()), filepath.Join(destDir, entry.Name())); err != nil {
					return err
				}
			}
			if err := os.Remove(releasePath); err != nil {
				return err
			}
		}

		os.Remove(tempFile)
	default:
		return fmt.Errorf("unsupported archive format: %s", format)
	}

	return nil
}

// main handles command-line arguments and dispatches commands.
func main() {
	log.SetFlags(0)
	log.SetPrefix("")

	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Println("Usage: gbe_tool <command> [options]")
		fmt.Println("Commands:")
		fmt.Println("  apply <platform> - Apply GBE to Steam API files")
		fmt.Println("  update           - Update the GBE fork repository")
		fmt.Println("  dlc <appid>      - Fetch DLCs for a given AppID")
		os.Exit(1)
	}

	command := args[0]
	switch command {
	case "apply":
		if len(args) < 2 {
			log.Fatalf("ERROR: Usage: %s apply <platform>", os.Args[0])
		}
		if err := applyGBE(args[1]); err != nil {
			log.Fatalf("ERROR: %v", err)
		}
	case "update":
		if err := updateGBE(); err != nil {
			log.Fatalf("ERROR: %v", err)
		}
	case "dlc":
		if len(args) < 2 {
			log.Fatalf("ERROR: Usage: %s dlc <appid>", os.Args[0])
		}
		fetchDLCs(args[1])
	default:
		log.Fatalf("ERROR: Invalid command: '%s'", command)
	}
}
