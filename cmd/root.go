package cmd

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/julian0xff/save/internal/db"
	"github.com/julian0xff/save/internal/extractor"
	"github.com/spf13/cobra"
)

var (
	dbPath string
	tags   string
)

var rootCmd = &cobra.Command{
	Use:     "save [url]",
	Short:   "Save articles for reading later",
	Long:    "A CLI read-later tool. Save web articles, search them, and read them in your terminal.",
	Version: "dev",
	Args:    cobra.MaximumNArgs(1),
	RunE:           runSave,
	SilenceUsage:   true,
	SilenceErrors:  true,
	CompletionOptions: cobra.CompletionOptions{
		HiddenDefaultCmd: true,
	},
}

func SetVersion(v string) {
	rootCmd.Version = v
}

func init() {
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "", "database path (default ~/.local/share/save/articles.db)")
	rootCmd.Flags().StringVar(&tags, "tag", "", "comma-separated tags")

	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(readCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(openCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(markCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func getDB() (*db.DB, error) {
	if dbPath != "" {
		return db.Open(dbPath)
	}
	return db.OpenDefault()
}

func normalizeURL(rawURL string) (string, error) {
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "https://" + rawURL
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %s", rawURL)
	}
	if u.Host == "" {
		return "", fmt.Errorf("invalid URL: %s", rawURL)
	}
	return u.String(), nil
}

func runSave(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}

	rawURL := args[0]
	normalized, err := normalizeURL(rawURL)
	if err != nil {
		return err
	}

	d, err := getDB()
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer d.Close()

	existing, err := d.GetArticleByURL(normalized)
	if err != nil {
		return fmt.Errorf("checking for duplicate: %w", err)
	}
	if existing != nil {
		fmt.Fprintf(os.Stderr, "Already saved (ID #%d). Use `save read %d` to view.\n", existing.ID, existing.ID)
		return nil
	}

	fmt.Fprintf(os.Stderr, "Fetching %s...\n", normalized)

	result, err := extractor.Extract(normalized)
	if err != nil {
		return fmt.Errorf("extracting article: %w", err)
	}

	wordCount := len(strings.Fields(result.TextContent))
	readTime := wordCount / 238
	if readTime == 0 && wordCount > 0 {
		readTime = 1
	}

	title := result.Title
	if title == "" {
		title = normalized
	}

	article := &db.Article{
		URL:       normalized,
		Title:     title,
		Author:    result.Author,
		Content:   result.Content,
		WordCount: wordCount,
		ReadTime:  readTime,
		Tags:      strings.TrimSpace(tags),
	}

	id, err := d.SaveArticle(article)
	if err != nil {
		return fmt.Errorf("saving article: %w", err)
	}

	fmt.Printf("Saved: \"%s\" (ID #%d, %d words, ~%d min read)\n", title, id, wordCount, readTime)
	if tags != "" {
		fmt.Printf("Tags: %s\n", tags)
	}

	return nil
}
