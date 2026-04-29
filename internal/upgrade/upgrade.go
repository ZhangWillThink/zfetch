package upgrade

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

const (
	Repo           = "ZhangWillThink/zfetch"
	CurrentVersion = "v0.2.0"
)

type release struct {
	TagName string `json:"tag_name"`
}

func Run() error {
	asset := assetName()
	fmt.Printf("Current version: %s\n", CurrentVersion)
	fmt.Printf("Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)

	latest, err := getLatestVersion()
	if err != nil {
		return fmt.Errorf("failed to check latest version: %w", err)
	}

	if latest == CurrentVersion {
		fmt.Println("Already up to date.")
		return nil
	}

	fmt.Printf("Upgrading to %s...\n", latest)

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	url := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", Repo, latest, asset)
	tmp := exe + ".new"

	if err := downloadFile(url, tmp); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("download failed: %w", err)
	}

	if err := os.Chmod(tmp, 0755); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("chmod failed: %w", err)
	}

	if err := os.Rename(tmp, exe); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("replace failed: %w", err)
	}

	fmt.Printf("Successfully upgraded to %s\n", latest)
	return nil
}

func getLatestVersion() (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", Repo)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var r release
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return "", err
	}
	return r.TagName, nil
}

func assetName() string {
	return fmt.Sprintf("zfetch-%s-%s", runtime.GOOS, runtime.GOARCH)
}

func downloadFile(url, path string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}
