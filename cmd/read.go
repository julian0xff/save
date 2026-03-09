package cmd

import (
	"fmt"
	"strconv"

	"github.com/julian0xff/save/internal/render"
	"github.com/spf13/cobra"
)

var readCmd = &cobra.Command{
	Use:   "read <id>",
	Short: "Read a saved article in the terminal",
	Args:  cobra.ExactArgs(1),
	RunE:  runRead,
}

func runRead(cmd *cobra.Command, args []string) error {
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

	// Mark as read
	_ = d.MarkRead(id)

	// Render header
	fmt.Print(render.ArticleHeader(article))

	// Render content
	rendered, err := render.ArticleContent(article.Content)
	if err != nil {
		fmt.Print(article.Content)
	} else {
		fmt.Print(rendered)
	}

	return nil
}
