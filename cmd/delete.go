package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a saved article",
	Args:  cobra.ExactArgs(1),
	RunE:  runDelete,
}

func runDelete(cmd *cobra.Command, args []string) error {
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

	if err := d.DeleteArticle(id); err != nil {
		return fmt.Errorf("deleting article: %w", err)
	}

	fmt.Printf("Deleted: \"%s\"\n", article.Title)
	return nil
}
