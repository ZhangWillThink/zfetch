package upgrade

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const Repo = "ZhangWillThink/zfetch"

// CurrentVersion is set at link time by scripts/build.sh (see -ldflags -X …).
var CurrentVersion = "v0.0.0-dev"

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

	if err := verifyReleaseBinarySHA256(asset, latest, tmp); err != nil {
		os.Remove(tmp)
		return err
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
	name := fmt.Sprintf("zfetch-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return name
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

func verifyReleaseBinarySHA256(asset, releaseTag, path string) error {
	sumURL := fmt.Sprintf("https://github.com/%s/releases/download/%s/SHA256SUMS", Repo, releaseTag)
	client := &http.Client{Timeout: 45 * time.Second}
	resp, err := client.Get(sumURL)
	if err != nil {
		return fmt.Errorf("could not fetch SHA256SUMS: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		_, _ = fmt.Fprintln(os.Stderr, "Warning: SHA256SUMS missing; skipping integrity verification.")
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("SHA256SUMS download: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read SHA256SUMS: %w", err)
	}
	wantHex, ok := hashForAsset(body, asset)
	if !ok {
		return fmt.Errorf("asset %q not listed in SHA256SUMS for %s", asset, releaseTag)
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("hash downloaded binary: %w", err)
	}
	gotHex := hex.EncodeToString(h.Sum(nil))

	if gotHex != wantHex {
		return fmt.Errorf(
			"checksum mismatch for %s (manifest %s, got %s) — aborted",
			asset, wantHex, gotHex,
		)
	}
	return nil
}

func hashForAsset(sums []byte, asset string) (hexHash string, ok bool) {
	text := strings.ReplaceAll(string(sums), "\r\n", "\n")
	for line := range strings.SplitSeq(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		name := strings.TrimPrefix(parts[len(parts)-1], "*")
		if filepath.Base(name) != asset && name != asset {
			continue
		}
		want := strings.ToLower(strings.TrimSpace(parts[0]))
		if len(want) != 64 || !sha256HexValid(want) {
			continue
		}
		return want, true
	}
	return "", false
}

func sha256HexValid(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, ch := range s {
		switch {
		case ch >= '0' && ch <= '9':
		case ch >= 'a' && ch <= 'f':
		default:
			return false
		}
	}
	return true
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
