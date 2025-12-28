package cmd

import (
	"fmt"
	"os"

	"github.com/arpnghosh/adm/internal/parser"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "adm [url]",
	Short: "A CLI tool for downloading files",
	Long:  "A CLI tool that allows parallel downloads using a segment option.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		segmentNum, _ := cmd.Flags().GetInt("segment")

		if segmentNum < 1 || segmentNum > 16 {
			return fmt.Errorf("the number of segments must be between 1 and 16")
		}

		url := args[0]
		if err := parser.ParseProtocol(url, segmentNum); err != nil {
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
	rootCmd.Flags().IntP("segment", "s", 4, "Number of segments for parallel downloads")
}
