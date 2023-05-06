package autoupdate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
)

type UpdaterConfig struct {
	CurrentVersion  string
	CheckInterval   time.Duration
	VersionInfoURL  string
	Shutdown        func()
	ShutdownTimeout time.Duration
}

type VersionInfo struct {
	Version  string `json:"version"`
	URL      string `json:"url"`
	FileHash string `json:"file_hash"`
}

var currentVersion string
var checkInterval time.Duration

func Run(config UpdaterConfig) {
	currentVersion = config.CurrentVersion
	checkInterval = config.CheckInterval

	go func() {
		ticker := time.NewTicker(checkInterval)
		for range ticker.C {
			checkAndUpdate(config)
		}
	}()
}

func checkAndUpdate(config UpdaterConfig) {
	versionInfo, err := checkLatestVersion(config)
	if err != nil {
		log.Println("Error checking latest version:", err)
		return
	}

	if versionInfo.Version != currentVersion {
		fmt.Println("A new version is available:", versionInfo.Version)
		if err := upgrade(versionInfo, config); err != nil {
			log.Println("Error upgrading:", err)
		}
	}
}

func checkLatestVersion(config UpdaterConfig) (*VersionInfo, error) {
	resp, err := http.Get(config.VersionInfoURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	versionInfo := &VersionInfo{}
	if err := json.NewDecoder(resp.Body).Decode(versionInfo); err != nil {
		return nil, err
	}

	return versionInfo, nil
}

func upgrade(versionInfo *VersionInfo, config UpdaterConfig) error {
	// Compare the fetched version with the current version
	if versionInfo.Version <= config.CurrentVersion {
		fmt.Println("Already up-to-date.")
		return nil
	}

	// Download the new version
	tempFile, err := downloadFile(versionInfo.URL)
	if err != nil {
		return err
	}
	defer os.Remove(tempFile)

	// Verify the hash of the downloaded file
	if err := verifyHash(tempFile, versionInfo.FileHash); err != nil {
		return err
	}

	// Replace the current executable and restart
	execPath, err := os.Executable()
	if err != nil {
		return err
	}

	return replaceAndRestart(execPath, tempFile, config)
}

func verifyHash(filePath, expectedHash string) error {
	fileData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	hasher := sha256.New()
	_, err = hasher.Write(fileData)
	if err != nil {
		return fmt.Errorf("failed to compute hash: %v", err)
	}

	actualHash := hex.EncodeToString(hasher.Sum(nil))
	if actualHash != expectedHash {
		return fmt.Errorf("hash mismatch: expected %s, got %s", expectedHash, actualHash)
	}

	return nil
}

func downloadFile(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download file, status code: %d", resp.StatusCode)
	}

	tempFile, err := ioutil.TempFile("", "update-")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return "", err
	}

	return tempFile.Name(), nil
}

func replaceAndRestart(execPath, tempPath string, config UpdaterConfig) error {
	if runtime.GOOS == "windows" {
		return replaceAndRestartWindows(execPath, tempPath, config)
	}
	return replaceAndRestartUnix(execPath, tempPath, config)
}

func replaceAndRestartUnix(execPath, tempPath string, config UpdaterConfig) error {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("mv %s %s && exec %s", tempPath, execPath, execPath))

	// Shutdown the application gracefully with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), config.ShutdownTimeout)
	defer cancel()

	done := make(chan struct{})
	go func() {
		config.Shutdown()
		close(done)
	}()

	select {
	case <-done:
		// Shutdown completed
	case <-ctx.Done():
		// Shutdown timed out
		log.Println("Shutdown timed out")
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	return cmd.Run()
}

func replaceAndRestartWindows(execPath, tempPath string, config UpdaterConfig) error {
	baseExec := filepath.Base(execPath)
	script := fmt.Sprintf(`
	@echo off
	set retries=0
	:wait
	tasklist /FI "IMAGENAME eq %s" 2>NUL | find /I /N "%s">NUL
	if "%%errorlevel%%"=="0" (
		set /a retries+=1
		if %%retries%% gtr 10 (
			echo "Timeout waiting for program to exit, please close it manually and run the script again."
			exit /b 1
		)
		timeout /t 1
		goto wait
	)
	move /Y "%s" "%s"
	start "" "%s"
	`, baseExec, baseExec, tempPath, execPath, execPath)

	scriptPath := execPath + ".bat"
	if err := ioutil.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		return err
	}

	// Shutdown the application gracefully with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), config.ShutdownTimeout)
	defer cancel()

	done := make(chan struct{})
	go func() {
		config.Shutdown()
		close(done)
	}()

	select {
	case <-done:
		// Shutdown completed
	case <-ctx.Done():
		// Shutdown timed out
		log.Println("Shutdown timed out")
	}

	cmd := exec.Command("cmd", "/C", scriptPath)
	if err := cmd.Start(); err != nil {
		return err
	}

	os.Exit(0)
	return nil
}
