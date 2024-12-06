package cmd

import (
	"os"

	download "github.com/arpnghosh/adm/src"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "adm [url]",
	Short: "A CLI tool for downloading files",
	Long:  "A CLI tool that allows parallel downloads using a segment option.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cmd.Help()
			os.Exit(1)
		}
		segmentNum, _ := cmd.Flags().GetInt("segment")
		url := args[0]
		download.DownloadUtil(url, segmentNum)
	},
}

func Execute() {
	if len(os.Args) == 1 || len(os.Args) > 4 {
		rootCmd.Help()
		os.Exit(1)
	} else {
		err := rootCmd.Execute()
		if err != nil {
			os.Exit(1)
		}
	}
}

func init() {
	var segment int
	rootCmd.Flags().IntVarP(&segment, "segment", "s", 4, "Number of segments for parallel download")
}
