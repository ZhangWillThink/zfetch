package upgrade

import (
	"path/filepath"
	"strings"
	"testing"
)

func Test_hashForAsset(t *testing.T) {
	sums := []byte(`
# comment
abcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcdabcd  missingname
`)

	if _, ok := hashForAsset(sums, "zfetch-linux-amd64"); ok {
		t.Fatal("expected no match")
	}

	sumsOK := []byte("e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  zfetch-linux-amd64\n")
	h, ok := hashForAsset(sumsOK, "zfetch-linux-amd64")
	if !ok || h != "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" {
		t.Fatalf("unexpected: ok=%v h=%q", ok, h)
	}

	sumLine := strings.Repeat("a", 64) + "  *zfetch-windows-amd64.exe\n"
	h2, ok2 := hashForAsset([]byte(sumLine), "zfetch-windows-amd64.exe")
	if !ok2 || h2 != strings.Repeat("a", 64) {
		t.Fatalf("star form: ok=%v h=%q", ok2, h2)
	}

	var longPath []byte
	longPath = append(longPath, "aaaabbbbccccddddaaaabbbbccccddddaaaabbbbccccddddaaaabbbbccccdddd  "...)
	longPath = append(longPath, filepath.Join("/some/path", "zfetch-linux-arm64")...)
	longPath = append(longPath, '\n')

	h3, ok3 := hashForAsset(longPath, "zfetch-linux-arm64")
	if !ok3 || h3 != "aaaabbbbccccddddaaaabbbbccccddddaaaabbbbccccddddaaaabbbbccccdddd" {
		t.Fatalf("basename match: ok=%v h=%q", ok3, h3)
	}
}

func Test_sha256HexValid(t *testing.T) {
	if !sha256HexValid("abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789") {
		t.Fatal("expected valid")
	}
	if sha256HexValid("ghijkl0123456789abcdef0123456789abcdef0123456789abcdef0123456789") {
		t.Fatal("expected invalid")
	}
	if sha256HexValid("abc") {
		t.Fatal("length")
	}
}
