package upgrade

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const Repo = "ZhangWillThink/zfetch"

var CurrentVersion = "v0.4.0"

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

	exeDir := filepath.Dir(exe)
	tmpDir := exeDir

	if _, err := os.Stat(exe); err == nil {
		if f, err := os.OpenFile(exe, os.O_WRONLY, 0); err != nil {
			_ = f
			return fmt.Errorf("no write permission for %s (try sudo)", exe)
		}
	}

	url := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", Repo, latest, asset)
	tmp := filepath.Join(tmpDir, filepath.Base(exe)+".new")

	if err := downloadFile(url, tmp); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("download failed: %w", err)
	}

	if err := os.Chmod(tmp, 0755); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("chmod failed: %w", err)
	}

	if err := os.Rename(tmp, exe); err != nil {
		if err := copyFile(tmp, exe); err != nil {
			os.Remove(tmp)
			return fmt.Errorf("replace failed: %w", err)
		}
		os.Remove(tmp)
	}

	fmt.Printf("Successfully upgraded to %s\n", latest)
	return nil
}

func getLatestVersion() (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", Repo)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
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
	client := &http.Client{Timeout: 300 * time.Second}
	resp, err := client.Get(url)
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

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		os.Remove(dst)
		return err
	}

	if st, err := in.Stat(); err == nil {
		_ = os.Chmod(dst, st.Mode())
	}
	return nil
}
