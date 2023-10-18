package cmd

import (
	"fmt"
	benchinterface "serverlessbench2/internal/benchinterface"
	"serverlessbench2/internal/util"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list function and workflow.",
	Run: func(cmd *cobra.Command, args []string) {
		lsTest, _ := cmd.Flags().GetBool("test")
		if len(args) > 0 {
			panic("cli list too much args")
		} else {
			if lsTest {
				testcaseYamlByte, _ := util.Asset("resources/testcase.yaml")
				resdata_raw := make(map[string]interface{})
				if err := yaml.Unmarshal(testcaseYamlByte, &resdata_raw); err != nil {
					fmt.Printf("%+v\n\n", err)
					panic(err)
				}
				fmt.Println("list all avalible testcases:")
				fmt.Println("Name\t\t\tPath")
				for key, value := range resdata_raw {
					strKey := fmt.Sprintf("%v", key)
					strValue := fmt.Sprintf("%v", value)
					fmt.Printf("%s\t\t%s\n", strKey, strValue)
				}
			} else {
				err := benchinterface.ListFunction()
				if err != nil {
					fmt.Printf("%+v\n\n", err)
					panic(err)
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.PersistentFlags().BoolP("test", "t", false, "list all avalible testcases")
}
