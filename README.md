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

### Improve Platform Compatibility:

        [ ] Abstract external commands (7z). Consider using native Go libraries for archive extraction to remove the dependency on a pre-installed 7z.

### Implement DLC Configuration (steam_settings):

        [ ] Create a new command (gbe_tool dlc configure <appid>) that uses fetchDLCs and allows the user to select and save which DLCs to enable.

### User Interface

        [ ] Research and select a Go GUI library (like Fyne, Wails, or Gio).

        [ ] Create a new application entry point for the GUI (e.g., in cmd/gbe_gui/main.go).

        [ ] Build the UI components (buttons for "Update", "Apply", a list for DLCs, etc.).

        [ ] Connect the UI buttons to call the functions in your refactored gbe, updater, and steam packages. The GUI should be a thin layer that orchestrates calls to your well-tested core logic.