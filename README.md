# gbe_fork_helper

`gbe_fork_helper` is a utility tool designed to streamline the management of your `gbe_fork` (Goldberg Steam Emulator) installation. It helps in updating your `gbe_fork` directory and applying necessary configurations to your Steam API files.

Usage:
```
Usage: gbe_fork_helper <command> [options]

Commands:
            apply <platform> - Apply GBE to Steam API files
            update           - Update the GBE fork repository
            dlc <appid>      - Fetch DLCs for a given AppID
```

## Roadmap

### Standardize Error Handling:

        [ ] Refactor all functions that currently use log.Fatalf to return an error instead. Let the main function decide whether to exit on an error. This is crucial for a GUI, which shouldn't just crash.

### Add Unit Tests:

        [ ] Write a unit test for a simple helper, like getHash.

        [ ] Write tests for the steam package, using mock HTTP responses to avoid hitting the real Steam API during tests.

        [ ] Begin testing the gbe package, mocking the filesystem and command execution where possible.

### Improve Platform Compatibility:

        - [] Replace Linux-specific code (like os.Getenv("HOME")) with cross-platform equivalents (os.UserHomeDir()).

        [ ] Abstract external commands (7z, strings). Consider using native Go libraries for archive extraction to remove the dependency on a pre-installed 7z.

        [ ] Fully implement and test the win64 and win32 logic defined in your platformConfig.

### Implement DLC Configuration (steam_settings):

        [ ] Design the structure for a configuration file (e.g., dlcs.json or settings.ini).

        [ ] Create functions to read the list of enabled DLCs from this file.

        [ ] Create a new command (gbe_tool dlc configure <appid>) that uses fetchDLCs and allows the user to select and save which DLCs to enable.

### User Interface

        [ ] Research and select a Go GUI library (like Fyne, Wails, or Gio).

        [ ] Create a new application entry point for the GUI (e.g., in cmd/gbe_gui/main.go).

        [ ] Build the UI components (buttons for "Update", "Apply", a list for DLCs, etc.).

        [ ] Connect the UI buttons to call the functions in your refactored gbe, updater, and steam packages. The GUI should be a thin layer that orchestrates calls to your well-tested core logic.