package util

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCmd(t *testing.T) {
	// Test a successful command
	output, err := RunCmd("echo", "hello", "world")
	if err != nil {
		t.Fatalf("RunCmd failed: %v", err)
	}
	expectedOutput := "hello world\n"
	if string(output) != expectedOutput {
		t.Errorf("Expected output %q, got %q", expectedOutput, string(output))
	}

	// Test a command that fails
	_, err = RunCmd("false")
	if err == nil {
		t.Fatalf("RunCmd was expected to fail but succeeded")
	}
	if !strings.Contains(err.Error(), "command failed") {
		t.Errorf("Expected 'command failed' error, got %v", err)
	}

	// Test a command with stderr output
	_, err = RunCmd("sh", "-c", "echo 'error' >&2 && exit 1")
	if err == nil {
		t.Fatalf("RunCmd was expected to fail but succeeded")
	}
	if !strings.Contains(err.Error(), "stderr: error") {
		t.Errorf("Expected stderr output in error, got %v", err)
	}
}

func TestGetHash(t *testing.T) {
	// Create a temporary file
	tmpfile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	content := "test content for hashing"
	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	expectedHash := "866437cb7a794bce2b727acc0362ee27" // MD4 hash of "hello"

	// Re-create temp file with "hello" content
	tmpfile2, err := os.CreateTemp("", "testfile2")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile2.Name()) // clean up
	if _, err := tmpfile2.WriteString("hello"); err != nil {
		t.Fatal(err)
	}
	tmpfile2.Close()

	hash2, err := GetHash(tmpfile2.Name())
	if err != nil {
		t.Fatalf("GetHash failed for 'hello': %v", err)
	}
	if hash2 != expectedHash {
		t.Errorf("Expected hash %q, got %q", expectedHash, hash2)
	}

	// Test non-existent file
	_, err = GetHash("nonexistentfile.txt")
	if err == nil {
		t.Fatalf("GetHash was expected to fail for non-existent file but succeeded")
	}
	if !os.IsNotExist(err) {
		t.Errorf("Expected 'file not exist' error, got %v", err)
	}
}

func TestCopyFile(t *testing.T) {
	srcContent := "This is the source file content."
	tmpDir, err := os.MkdirTemp("", "testcopyfile")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	srcPath := filepath.Join(tmpDir, "source.txt")
	destPath := filepath.Join(tmpDir, "destination.txt")

	// Create source file
	if err := os.WriteFile(srcPath, []byte(srcContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Copy file
	if err := CopyFile(srcPath, destPath); err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}

	// Verify destination file exists and has correct content
	destContent, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}
	if string(destContent) != srcContent {
		t.Errorf("Expected destination content %q, got %q", srcContent, string(destContent))
	}

	// Test copying to a non-existent directory (should fail)
	nonExistentDirPath := filepath.Join(tmpDir, "nonexistent", "destination.txt")
	err = CopyFile(srcPath, nonExistentDirPath)
	if err == nil {
		t.Fatalf("CopyFile was expected to fail for non-existent directory but succeeded")
	}
}

func TestBackupAndReplace(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "testbackupreplace")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	srcContent := "new content"
	destContent := "original content"

	srcPath := filepath.Join(tmpDir, "source.txt")
	destPath := filepath.Join(tmpDir, "destination.txt")

	// Create source and destination files
	if err := os.WriteFile(srcPath, []byte(srcContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(destPath, []byte(destContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Backup and replace
	if err := BackupAndReplace(srcPath, destPath); err != nil {
		t.Fatalf("BackupAndReplace failed: %v", err)
	}

	// Verify new destination content
	newDestContent, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read new destination file: %v", err)
	}
	if string(newDestContent) != srcContent {
		t.Errorf("Expected new destination content %q, got %q", srcContent, string(newDestContent))
	}

	// Verify backup file exists and has original content
	// The backup file name includes a timestamp, so we need to find it
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	backupFound := false
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "destination.txt.") && strings.HasSuffix(f.Name(), ".ORIGINAL") {
			backupPath := filepath.Join(tmpDir, f.Name())
			backupContent, err := os.ReadFile(backupPath)
			if err != nil {
				t.Fatalf("Failed to read backup file %s: %v", backupPath, err)
			}
			if string(backupContent) != destContent {
				t.Errorf("Expected backup content %q, got %q", destContent, string(backupContent))
			}
			backupFound = true
			break
		}
	}
	if !backupFound {
		t.Errorf("Backup file not found")
	}

	// Test case where destination file does not exist initially (should create it)
	os.Remove(destPath) // Remove the destination file
	if err := BackupAndReplace(srcPath, destPath); err != nil {
		t.Fatalf("BackupAndReplace failed when dest did not exist: %v", err)
	}
	newDestContentAfterNoBackup, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read new destination file after no backup: %v", err)
	}
	if string(newDestContentAfterNoBackup) != srcContent {
		t.Errorf("Expected new destination content %q, got %q", srcContent, string(newDestContentAfterNoBackup))
	}
}
