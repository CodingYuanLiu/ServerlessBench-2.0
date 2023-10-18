package cmd

import (
	"fmt"
	benchinterface "serverlessbench2/internal/benchinterface"

	"github.com/spf13/cobra"
)

var invokeCmd = &cobra.Command{
	Use:   "invoke func_name param_name1 param_value1 ...",
	Short: "create function",
	Run: func(cmd *cobra.Command, args []string) {
		isFlow, _ := cmd.Flags().GetBool("flow")
		if len(args) < 1 {
			panic("cli invoke too few args")
		} else if len(args) == 1 {
			funcName := args[0]
			if isFlow {
				err := benchinterface.InvokeWorkflow(funcName)
				if err != nil {
					fmt.Printf("%+v\n\n", err)
					panic(err)
				}
			} else {
				err := benchinterface.InvokeFunction(funcName)
				if err != nil {
					fmt.Printf("%+v\n\n", err)
					panic(err)
				}
			}
		} else {
			funcName := args[0]
			params := args[1:]
			if isFlow {
				err := benchinterface.InvokeWorkflow(funcName, params...)
				if err != nil {
					fmt.Printf("%+v\n\n", err)
					panic(err)
				}
			} else {
				err := benchinterface.InvokeFunction(funcName, params...)
				if err != nil {
					fmt.Printf("%+v\n\n", err)
					panic(err)
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(invokeCmd)
	invokeCmd.PersistentFlags().String("param", "", "parameter")
	invokeCmd.PersistentFlags().BoolP("flow", "f", false, "create workflow or not")
}
