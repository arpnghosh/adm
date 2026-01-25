package cmd

import (
	"fmt"
	"os"

	"github.com/arpnghosh/adm/internal/parser"
	"github.com/spf13/cobra"
)

var (
	segmentCount int
	outputFile   string
	proxyAddr    string
)

var rootCmd = &cobra.Command{
	Use:  "adm [url]",
	Long: "Adm is a CLI tool for downloading files from the internet. Currently, it only supports HTTP and HTTPS.",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if segmentCount < 1 || segmentCount > 16 {
			return fmt.Errorf("the number of segments must be between 1 and 16")
		}

		url := args[0]
		if err := parser.ParseProtocol(url, segmentCount, outputFile, proxyAddr); err != nil {
			return fmt.Errorf("%w", err)
		}

		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().IntVarP(&segmentCount, "segment", "s", 4, "Number of segments for parallel downloads")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output filename for the downloaded file")
	rootCmd.Flags().StringVarP(&proxyAddr, "proxy", "p", "", `Proxy Server address. Can be:
	 	- http://host:port
	 	- https://host:port
	 	- socks5://host:port`)
}
