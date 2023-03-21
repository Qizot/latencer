package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"latencer/hls"
)

var rootCmd = &cobra.Command{
	Use:   "latencer",
	Short: "Latencer is a tool for measuring latency video streamiing protocols",
	Long:  "Latencer is a tool for measuring latency video streamiing protocols",
	Run: func(cmd *cobra.Command, args []string) {
    fmt.Println("Here we go")
		// Do Stuff Here
	},
}

var (
  duration int
  src string
)

func init() {
  rootCmd.AddCommand(hls.HlsCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	Execute()
}
