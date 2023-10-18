package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	benchinterface "serverlessbench2/internal/benchinterface"
	utility "serverlessbench2/internal/util"
)

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply a test from yaml config file",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			panic("Wrong Parameter. Usage: cli apply [filename].yaml")
		} else {
			yamlContent := utility.TestYaml{}
			yamlFile, err := ioutil.ReadFile(args[0])
			if err != nil {
				fmt.Printf("%+v\n\n", err)
				panic(err)
			}
			if err = yaml.Unmarshal(yamlFile, &yamlContent); err != nil {
				fmt.Printf("%+v\n\n", err)
				panic(err)
			}
			err = benchinterface.ApplyYaml(yamlContent)
			if err != nil {
				fmt.Printf("%+v\n\n", err)
				panic(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)
}
