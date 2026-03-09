package cmd

import (
	"fmt"
	"strings"

	"github.com/julian0xff/save/internal/render"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Full-text search across saved articles",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runSearch,
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := strings.Join(args, " ")

	d, err := getDB()
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer d.Close()

	results, err := d.SearchArticles(query)
	if err != nil {
		return fmt.Errorf("searching: %w", err)
	}

	fmt.Print(render.SearchResults(results))
	return nil
}
