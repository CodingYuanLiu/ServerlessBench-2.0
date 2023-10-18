package cmd

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/cobra"

	benchinterface "serverlessbench2/internal/benchinterface"
	config "serverlessbench2/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config [flags]",
	Short: "Specify the Provider, Python Interpreter Path, DockerUserName and Directory of tesecases.",
	Run: func(cmd *cobra.Command, args []string) {
		var provider, pythonPath, DockerUserName, testResultDir string
		provider, _ = cmd.Flags().GetString("Provider")
		provider = strings.ToLower(provider)
		if provider == "" {
			provider = benchinterface.CheckProvider()
		}
		pythonPath, _ = cmd.Flags().GetString("PythonPath")
		if pythonPath == "" {
			pythonPath = config.GetPythonPath()
		}
		DockerUserName, _ = cmd.Flags().GetString("DockerUserName")
		if DockerUserName == "" {
			DockerUserName = config.GetDockerUserName()
		}
		testResultDir, _ = cmd.Flags().GetString("TestResultDir")
		if testResultDir == "" {
			testResultDir = config.GetTestResultDir()
		}

		var benchconfig = &config.BaseConfig{
			Provider:       provider,
			PythonPath:     pythonPath,
			DockerUserName: DockerUserName,
			TestResultDir:  testResultDir,
		}
		err := config.WriteConfig(benchconfig)
		if err != nil {
			fmt.Printf("%+v\n\n", err)
			panic(err)
		}
		checkFlag, _ := cmd.Flags().GetBool("check")
		if checkFlag {
			conf := config.GetBaseConfig()
			confType := reflect.TypeOf(conf)
			confValue := reflect.ValueOf(conf)
			fmt.Println("configField\tValue")
			for i := 0; i < confType.NumField(); i++ {
				fieldName := confType.Field(i).Name
				fieldValue := confValue.Field(i).String()
				fmt.Printf("%s\t%s\n", fieldName, fieldValue)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.PersistentFlags().BoolP("check", "c", false, "view config or not")
	configCmd.PersistentFlags().String("PythonPath", "python", "Path of Python intepreter")
	configCmd.PersistentFlags().String("Provider", "", "Cloud provider")
	configCmd.PersistentFlags().String("DockerUserName", "", "User name of docker")
	configCmd.PersistentFlags().String("TestResultDir", "", "Directory for storing test results.")
}
