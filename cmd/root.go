package cmd

import (
	"fmt"
	"os"

	"github.com/arpnghosh/adm/internal/parser"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "adm [url]",
	Long:  "Adm is a CLI tool for downloading files from the internet. Currently, it only supports HTTP and HTTPS.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		segmentNum, _ := cmd.Flags().GetInt("segment")
		fileName, _ := cmd.Flags().GetString("output")

		if segmentNum < 1 || segmentNum > 16 {
			return fmt.Errorf("the number of segments must be between 1 and 16")
		}

		url := args[0]
		if err := parser.ParseProtocol(url, segmentNum, fileName); err != nil {
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

var name string

func init() {
	rootCmd.Flags().IntP("segment", "s", 4, "Number of segments for parallel downloads")
	rootCmd.Flags().StringVarP(&name, "output", "o", "", "Output filename for the downloaded file")
}
