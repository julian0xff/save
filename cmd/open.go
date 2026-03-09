package cmd

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"

	"github.com/spf13/cobra"
)

var openCmd = &cobra.Command{
	Use:   "open <id>",
	Short: "Open article URL in your browser",
	Args:  cobra.ExactArgs(1),
	RunE:  runOpen,
}

func runOpen(cmd *cobra.Command, args []string) error {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid article ID: %s", args[0])
	}

	d, err := getDB()
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer d.Close()

	article, err := d.GetArticle(id)
	if err != nil {
		return fmt.Errorf("fetching article: %w", err)
	}
	if article == nil {
		return fmt.Errorf("no article with ID %d. Use `save list` to see available articles", id)
	}

	var openCmdName string
	switch runtime.GOOS {
	case "darwin":
		openCmdName = "open"
	case "linux":
		openCmdName = "xdg-open"
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	if err := exec.Command(openCmdName, article.URL).Start(); err != nil {
		return fmt.Errorf("opening browser: %w", err)
	}

	fmt.Printf("Opened: %s\n", article.URL)
	return nil
}
