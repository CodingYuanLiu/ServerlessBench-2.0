package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"serverlessbench2/internal/util"

	"github.com/pkg/errors"
)

type knativeConfig struct {
	WorkflowYaml           string
	DiaplayerAndSenderYaml string
	GoWrapperDir           string
	GoWrapperFile          string
	ServiceYaml            string
	ServiceFile            string
	PythonDir              string
	WorkflowDockerfile     string
}

var workflowYaml = "workflow.yaml"
var diaplayerAndSenderYaml = "diaplayerAndSender.yaml"
var goWrapperDir = "wrapper"
var goWrapperFile = "wrapper.go"
var serviceYaml = "service.yaml"
var serviceFile = "server.py"
var pythonDir = "pyfiles"
var workflowDockerfile = "workflowDockerfile"

func (kn knativeProvider) writeConfig() error {
	u, _ := user.Current()
	homeDir := u.HomeDir
	filename := homeDir + knativeConfigPath
	err := util.EnsureFileExists(filename)
	if err != nil {
		return err
	}
	var knconfig = knativeConfig{
		WorkflowYaml:           workflowYaml,
		DiaplayerAndSenderYaml: diaplayerAndSenderYaml,
		GoWrapperDir:           goWrapperDir,
		GoWrapperFile:          goWrapperFile,
		ServiceYaml:            serviceYaml,
		ServiceFile:            serviceFile,
		PythonDir:              pythonDir,
		WorkflowDockerfile:     workflowDockerfile,
	}
	file, _ := json.MarshalIndent(knconfig, "", " ")
	err = ioutil.WriteFile(filename, file, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (kn knativeProvider) checkConfig() error {
	cmd := exec.Command("which", "kn")
	out, _ := cmd.Output()
	if string(out) == "" {
		return errors.New("Require kn to support knative")
	}
	return nil
}

func getKnativeConfig() knativeConfig {
	u, _ := user.Current()
	homeDir := u.HomeDir
	file, _ := os.Open(homeDir + knativeConfigPath)
	if file == nil {
		return knativeConfig{}
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	conf := knativeConfig{}
	err := decoder.Decode(&conf)
	if err != nil {
		return knativeConfig{}
	}
	return conf
}

func GetKnWorkflowYaml() string {
	conf := getKnativeConfig()
	return conf.WorkflowYaml
}
func GetKnDiaplayerAndSenderYaml() string {
	conf := getKnativeConfig()
	return conf.DiaplayerAndSenderYaml
}
func GetKnGoWrapperDir() string {
	conf := getKnativeConfig()
	return conf.GoWrapperDir
}
func GetKnGoWrapperFile() string {
	conf := getKnativeConfig()
	return conf.GoWrapperFile
}
func GetKnServiceYaml() string {
	conf := getKnativeConfig()
	return conf.ServiceYaml
}
func GetKnServiceFile() string {
	conf := getKnativeConfig()
	return conf.ServiceFile

}
func GetKnPythonDir() string {
	conf := getKnativeConfig()
	return conf.PythonDir

}
func GetKnWorkflowDockerfile() string {
	conf := getKnativeConfig()
	return conf.WorkflowDockerfile
}
