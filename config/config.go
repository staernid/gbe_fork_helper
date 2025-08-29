package config

import "time"

// Global Configuration
const (
	GbeDir         = ".local/share/gbe_fork"
	SteamStoreAPI  = "https://store.steampowered.com/api"
	GithubAPIURL   = "https://api.github.com/repos/Detanup01/gbe_fork/releases/latest"
	SevenZCommand  = "7z"
	StringsCommand = "strings"
)

// PlatformConfig maps platform names to their configuration.
var PlatformConfig = map[string]struct {
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
