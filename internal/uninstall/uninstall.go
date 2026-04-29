package uninstall

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Run() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	fmt.Printf("This will remove: %s\n", exe)
	fmt.Print("Continue? [y/N] ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input != "y" && input != "yes" {
		fmt.Println("Cancelled.")
		return nil
	}

	if err := os.Remove(exe); err != nil {
		return fmt.Errorf("failed to remove: %w", err)
	}

	fmt.Println("zfetch has been uninstalled.")
	fmt.Println()
	fmt.Println("Config files (optional):")
	fmt.Println("  rm -rf ~/.config/zfetch")
	return nil
}
