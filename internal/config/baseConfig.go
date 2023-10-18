package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"serverlessbench2/internal/util"
)

type BaseConfig struct {
	Provider       string
	PythonPath     string
	DockerUserName string
	CliBaseDir     string
	TestResultDir  string
}

func writeBaseConfig(conf *BaseConfig) error {
	u, _ := user.Current()
	homeDir := u.HomeDir
	filename := homeDir + baseConfigPath
	(*conf).CliBaseDir, _ = os.Getwd()
	err := util.EnsureFileExists(filename)
	if err != nil {
		return err
	}
	file, _ := json.MarshalIndent(*conf, "", " ")
	err = ioutil.WriteFile(filename, file, 0644)
	if err != nil {
		return err
	}
	return nil
}

func GetCliBaseDir() string {
	conf := GetBaseConfig()
	return conf.CliBaseDir
}

func GetTestResultDir() string {
	conf := GetBaseConfig()
	return conf.TestResultDir
}
func ChangeTestResultDir(TestResultDir string) error {
	newConfig := GetBaseConfig()
	newConfig.TestResultDir = TestResultDir
	err := writeBaseConfig(&newConfig)
	if err != nil {
		return err
	}
	return nil
}

func GetPythonPath() string {
	conf := GetBaseConfig()
	return conf.PythonPath
}

func GetProvider() string {
	conf := GetBaseConfig()
	return conf.Provider
}
func ChangeProvider(provider string) error {
	newConfig := GetBaseConfig()
	newConfig.Provider = provider
	err := writeBaseConfig(&newConfig)
	if err != nil {
		return err
	}
	return nil
}

func GetDockerUserName() string {
	conf := GetBaseConfig()
	return conf.DockerUserName
}

func GetBaseConfig() BaseConfig {
	u, _ := user.Current()
	homeDir := u.HomeDir
	file, _ := os.Open(homeDir + baseConfigPath)
	if file == nil {
		return BaseConfig{}
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	conf := BaseConfig{}
	err := decoder.Decode(&conf)
	if err != nil {
		fmt.Println("error:", err)
	}
	return conf
}
