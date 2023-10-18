package cmd

import (
	"errors"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cli",
	Short: "Serverless cli",
	Long:  "This serverless cli now support 2 platforms -- HuaweiCloud and AliCloud",
	Run: func(cmd *cobra.Command, args []string) {
		Error(cmd, args, errors.New("unrecognized command"))
	},
}

func Execute() {
	rootCmd.Execute()
}
