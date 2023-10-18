package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/user"
	"serverlessbench2/internal/util"
)

type openwhiskConfig struct {
	EntrypointFile string
	ZipFileName    string
	TempDirPrefix  string
}

var entrypointFile = "__main__.py"
var zipFileName = "openwhisk.zip"
var tempDirPrefix = "/tmp/openwhisk-"

func (ow openwhiskProvider) writeConfig() error {
	u, _ := user.Current()
	homeDir := u.HomeDir
	filename := homeDir + openwhiskConfigPath
	err := util.EnsureFileExists(filename)
	if err != nil {
		return err
	}
	var knconfig = openwhiskConfig{
		EntrypointFile: entrypointFile,
		ZipFileName:    zipFileName,
		TempDirPrefix:  tempDirPrefix,
	}
	file, _ := json.MarshalIndent(knconfig, "", " ")
	err = ioutil.WriteFile(filename, file, 0644)
	if err != nil {
		return err
	}
	return nil
}

func getOpenwhiskConfig() openwhiskConfig {
	u, _ := user.Current()
	homeDir := u.HomeDir
	file, _ := os.Open(homeDir + openwhiskConfigPath)
	if file == nil {
		return openwhiskConfig{}
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	conf := openwhiskConfig{}
	err := decoder.Decode(&conf)
	if err != nil {
		return openwhiskConfig{}
	}
	return conf
}

func GetOpenwhiskEntrypoint() string {
	conf := getOpenwhiskConfig()
	return conf.EntrypointFile
}

func GetOpenwhiskZipFileName() string {
	conf := getOpenwhiskConfig()
	return conf.ZipFileName
}

func GetOpenwhiskTempDirPrefix() string {
	conf := getOpenwhiskConfig()
	return conf.TempDirPrefix
}

func (ow openwhiskProvider) checkConfig() error {

	return nil
}
