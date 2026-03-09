package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var markCmd = &cobra.Command{
	Use:   "mark <id>",
	Short: "Mark an article as read",
	Args:  cobra.ExactArgs(1),
	RunE:  runMark,
}

func runMark(cmd *cobra.Command, args []string) error {
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

	if article.ReadAt != nil {
		fmt.Printf("Already read: \"%s\"\n", article.Title)
		return nil
	}

	if err := d.MarkRead(id); err != nil {
		return fmt.Errorf("marking as read: %w", err)
	}

	fmt.Printf("Marked as read: \"%s\"\n", article.Title)
	return nil
}
