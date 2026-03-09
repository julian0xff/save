package cmd

import (
	"fmt"

	"github.com/julian0xff/save/internal/db"
	"github.com/julian0xff/save/internal/render"
	"github.com/spf13/cobra"
)

var (
	listUnread bool
	listTag    string
	listLimit  int
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List saved articles",
	RunE:  runList,
}

func init() {
	listCmd.Flags().BoolVar(&listUnread, "unread", false, "show only unread articles")
	listCmd.Flags().StringVar(&listTag, "tag", "", "filter by tag")
	listCmd.Flags().IntVar(&listLimit, "limit", 50, "max articles to show")
}

func runList(cmd *cobra.Command, args []string) error {
	d, err := getDB()
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer d.Close()

	articles, err := d.ListArticles(db.ListOpts{
		Unread: listUnread,
		Tag:    listTag,
		Limit:  listLimit,
	})
	if err != nil {
		return fmt.Errorf("listing articles: %w", err)
	}

	fmt.Print(render.ArticleList(articles))
	return nil
}
