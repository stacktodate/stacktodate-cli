package cmd

import (
	"fmt"
	"os"

	"github.com/stacktodate/stacktodate-cli/cmd/helpers"
	"github.com/stacktodate/stacktodate-cli/cmd/lib/cache"
	"github.com/spf13/cobra"
)

var fetchCatalogCmd = &cobra.Command{
	Use:   "fetch-catalog",
	Short: "Fetch and cache the product catalog from stacktodate.club",
	Long: `Fetch the complete list of products and their release information from stacktodate.club API
and store it locally for faster version detection and truncation.

The catalog is cached in ~/.stacktodate/products-cache.json and automatically refreshed
once every 24 hours. You can use this command to manually refresh the cache at any time.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(os.Stderr, "Fetching product catalog from stacktodate.club...\n")

		if err := cache.FetchAndCache(); err != nil {
			helpers.ExitOnError(err, "failed to fetch catalog")
		}

		// Load and display info about cached products
		products, err := cache.LoadCache()
		if err != nil {
			helpers.ExitOnError(err, "failed to load cached products")
		}

		cachePath, _ := cache.GetCachePath()
		fmt.Fprintf(os.Stderr, "✓ Successfully cached %d products\n", len(products.Products))
		fmt.Fprintf(os.Stderr, "Cache location: %s\n", cachePath)
	},
}

func init() {
	rootCmd.AddCommand(fetchCatalogCmd)
}
