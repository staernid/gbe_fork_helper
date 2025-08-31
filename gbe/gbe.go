package gbe

import (
	"fmt"
	"gbe_fork_helper/config"
	"gbe_fork_helper/util"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// applyGBE applies the GBE patch to a specified platform.
func ApplyGBE(platform string) error {
	platformCfg, ok := config.PlatformConfig[platform]
	if !ok {
		var validPlatforms []string
		for p := range config.PlatformConfig {
			validPlatforms = append(validPlatforms, p)
		}
		return fmt.Errorf("invalid platform: '%s'. Valid platforms: %s", platform, strings.Join(validPlatforms, ", "))
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}
	gbePath := filepath.Join(homeDir, config.GbeDir, platformCfg.Subdir, "experimental", "x"+platformCfg.Arch)

	var targetFiles []string
	walkErr := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && d.Name() == platformCfg.Target {
			targetFiles = append(targetFiles, path)
		}
		return nil
	})
	if walkErr != nil {
		return fmt.Errorf("failed to search for files: %w", walkErr)
	}

	if len(targetFiles) == 0 {
		log.Printf("WARN: No target files found for platform '%s'.", platform)
	}

	sourceFile := filepath.Join(gbePath, platformCfg.Target)
	if _, err := os.Stat(sourceFile); os.IsNotExist(err) {
		return fmt.Errorf("source file not found: '%s'", sourceFile)
	}

	sourceHash, err := util.GetHash(sourceFile)
	if err != nil {
		return fmt.Errorf("failed to get hash of source file: %w", err)
	}

	for _, file := range targetFiles {
		log.Printf("INFO: Found potential target: '%s'", file)

		targetHash, err := util.GetHash(file)
		if err != nil {
			log.Printf("ERROR: Failed to get hash of '%s': %v. Skipping.", file, err)
			continue
		}

		if targetHash == sourceHash {
			log.Println("SUCCESS: File is already up-to-date. Skipping.")
			continue
		}

		if err := util.BackupAndReplace(sourceFile, file); err != nil {
			log.Printf("ERROR: Failed to replace file '%s': %v. Skipping.", file, err)
			continue
		}

		if platformCfg.Additional != "" {
			additionalSource := filepath.Join(gbePath, platformCfg.Additional)
			additionalDest := filepath.Join(filepath.Dir(file), platformCfg.Additional)
			if _, err := os.Stat(additionalSource); err == nil {
				if err := util.BackupAndReplace(additionalSource, additionalDest); err != nil {
					log.Printf("WARN: Failed to replace additional file '%s': %v", additionalDest, err)
				}
			}
		}

		homeDir, err = os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		generatorPath := filepath.Join(homeDir, config.GbeDir, platformCfg.Subdir, "tools", "generate_interfaces", platformCfg.Generator)
		if _, err := os.Stat(generatorPath); err == nil {
			log.Printf("INFO: Running generator '%s'...", platformCfg.Generator)
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
