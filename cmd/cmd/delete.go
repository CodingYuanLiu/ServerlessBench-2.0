package cmd

import (
	"fmt"
	benchinterface "serverlessbench2/internal/benchinterface"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete func_name",
	Short: "delete function",
	Run: func(cmd *cobra.Command, args []string) {
		isFlow, _ := cmd.Flags().GetBool("flow")
		if len(args) < 1 {
			panic("Usage: cli delete function_name")
		} else if len(args) == 1 {
			func_name := args[0]
			if !isFlow {
				err := benchinterface.DeleteFunction(func_name)
				if err != nil {
					fmt.Printf("%+v\n\n", err)
					panic(err)
				}
			} else {
				err := benchinterface.DeleteWorkflow(func_name)
				if err != nil {
					fmt.Printf("%+v\n\n", err)
					panic(err)
				}
			}
		} else {
			panic("cli delete: too much args")
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.PersistentFlags().BoolP("flow", "f", false, "delete workflow or not")
}
