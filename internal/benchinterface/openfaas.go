package benchinterface

import (
	"fmt"
	"os"
	"os/exec"
	util "serverlessbench2/internal/util"
	"strings"

	"github.com/pkg/errors"
)

func (of ProviderOpenfaas) CreateWorkflow(flow_name string, func_info ...map[string](map[string]string)) error {
	return nil
}

func (of ProviderOpenfaas) CreateFunction(FunctionName string, MemorySize int, SourceCodeFiles ...string) error {

	registryName := "serverlessbench-registry"

	binary_wrapper, err := util.Asset("resources/handler.py") //fetch binary file in res.go
	binary_zip, err := util.Asset("resources/faastemplate.zip")
	binary_init, err := util.Asset("resources/__init__.py")

	// remove old registry
	rmCmd := exec.Command("docker", "rm", "-f", registryName)
	rmCmd.Run()

	// start new registry
	registryCmd := exec.Command("docker", "run", "-d", "-p", "5000:5000", "--restart", "always", "--name", registryName, "registry:2")
	out, err := registryCmd.CombinedOutput()
	if err != nil {
		return errors.New(fmt.Sprintf(`start docker local registry failed, message: %s 
		try to run \"docker run -d -p 5000:5000 --restart always --name %s registry:2\"`, string(out), registryName))
	}

	// move template
	z, _ := os.Create("faastemplate.zip")
	z.Write(binary_zip)
	zpcmd := exec.Command("unzip", "-o", "-d", util.FaasTempDir, "faastemplate.zip")
	err = zpcmd.Run()
	if err != nil {
		os.Remove("faastemplate.zip")
		return err
	}
	exec.Command("mv", util.FaasTempDir+"/resources/faastemplate/python3", util.FaasTempDir).Run()
	exec.Command("rm", "-rf", util.FaasTempDir+"/resources").Run()
	err = z.Close()
	os.Remove("faastemplate.zip")

	// create function from template
	fmt.Println("faas-cli new function...")
	functionNewCmd := exec.Command("faas-cli", "new", FunctionName, "--lang", "python3", "-p", "localhost:5000")
	out, err = functionNewCmd.CombinedOutput()
	if err != nil {
		// out: Folder: (  ) already exists
		exec.Command("rm", "-rf", "build", FunctionName, FunctionName+".yml", util.FaasTempDir).Run()
		return errors.New("Error creating function. " + string(out))
	}
	// replace init.py
	finit, err := os.OpenFile(FunctionName+"/__init__.py", os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0777)
	finit.Write(binary_init)
	defer finit.Close()

	// replace handler.py
	fhandler, err := os.OpenFile(FunctionName+"/handler.py", os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0777)
	fhandler.Write(binary_wrapper)
	defer fhandler.Close()

	// add source file
	for _, file := range SourceCodeFiles {
		originFile := file
		paths := strings.Split(file, "/")
		file = paths[len(paths)-1]
		targetFlie := FunctionName + "/" + file
		if strings.HasSuffix(file, "requirements.txt") { //copy requirements.txt
			targetFlie = "template/python3/requirements.txt"
		}
		err := util.Cp(originFile, targetFlie)
		if err != nil {
			return err
		}
	}
	// limits:
	//   memory: 256Mi
	//   cpu: 2000m
	file, err := os.OpenFile(FunctionName+".yml", os.O_APPEND|os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		exec.Command("rm", "-rf", "build", FunctionName, FunctionName+".yml", util.FaasTempDir).Run()
		return err
	}
	file.WriteString("    limits:\n")
	file.WriteString("      memory: " + fmt.Sprintf("%d", MemorySize) + "Mi\n")
	file.WriteString("      cpu: " + fmt.Sprintf("%d", int(MemorySize*1000/128)) + "m\n")
	file.WriteString("    requests:\n")
	file.WriteString("      memory: 0Mi\n")
	file.WriteString("      cpu: 0m\n")
	defer file.Close()

	// up function
	fmt.Println("faas-cli up...")
	upCmd := exec.Command("faas-cli", "up", "-f", FunctionName+".yml")
	out, err = upCmd.CombinedOutput()
	if err != nil {
		exec.Command("rm", "-rf", "build", FunctionName, FunctionName+".yml", util.FaasTempDir).Run()
		return errors.Wrap(err, fmt.Sprintf("faas-cli up failed. raw result:%s", string(out)))
	} else {
		fmt.Println("Create function:", FunctionName)
	}

	// remove template
	rmTemplateCmd := exec.Command("rm", "-rf", FunctionName, FunctionName+".yml", util.FaasTempDir)
	out, err = rmTemplateCmd.CombinedOutput()
	if err != nil {
		return err
	}
	return nil
}

func (of ProviderOpenfaas) InvokeWorkflow(function_name string, params ...string) error {
	return nil
}

func (of ProviderOpenfaas) InvokeFunction(FunctionName string, Params ...string) error {
	testCmd := exec.Command("faas-cli", "invoke", FunctionName)
	out, err := testCmd.CombinedOutput()
	if err != nil {
		return err
	}

	input := "{"
	if len(Params)%2 != 0 {
		err := fmt.Sprintf("params must be key-value pairs")
		return errors.New(err)
	}
	for i := 0; i < len(Params); i += 2 {
		input += "\"" + Params[i] + "\": \"" + Params[i+1] + "\""
		if i+2 != len(Params) {
			input += ","
		}
	}
	input += "}"

	describeCmd := exec.Command("faas-cli", "describe", FunctionName)
	out, err = describeCmd.CombinedOutput()
	url := ""
	if err != nil {
		return err
	} else {
		tmp := strings.Split(string(out), "\n")
		if len(tmp) < 8 {
			fmt.Println("Error describe function, can not get url")
		}
		url = tmp[7]
		tmp = strings.Split(url, " ")
		url = tmp[17]
	}

	Cmd := exec.Command("curl", url, "-d", input)
	out, err = Cmd.Output()
	if err != nil {
		return err
	} else {
		fmt.Print(string(out))
	}

	return nil
}

func (of ProviderOpenfaas) DeleteFunction(FunctionName string) error {
	rmcmd := "sudo rm -rf build/" + FunctionName
	exec.Command("/bin/sh", "-c", rmcmd).Run()
	cmd := exec.Command("faas-cli", "remove", FunctionName)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error: " + string(out))
		// return err
	} else {
		fmt.Println("Delete function:", FunctionName)
	}
	return nil
}

func (of ProviderOpenfaas) DeleteWorkflow(FunctionName string) error {
	rmcmd := "sudo rm -rf build/" + FunctionName
	exec.Command("/bin/sh", "-c", rmcmd).Run()
	cmd := exec.Command("faas-cli", "remove", FunctionName)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error: " + string(out))
		// return err
	} else {
		fmt.Println("Delete function:", FunctionName)
	}
	return nil
}

func (of ProviderOpenfaas) ListFunction() error {
	fmt.Println("List functions:")
	cmd := exec.Command("faas-cli", "list")
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error: " + string(out))
		// return err
	} else {
		fmt.Print(string(out))
	}
	return nil
}
