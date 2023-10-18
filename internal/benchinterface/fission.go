package benchinterface

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"

	util "serverlessbench2/internal/util"
)

func (fission ProviderFission) CreateWorkflow(flow_name string, func_info ...map[string](map[string]string)) error {
	return nil
}

func (fission ProviderFission) CreateFunction(FunctionName string, MemorySize int, SourceCodeFiles ...string) error {
	dstZipFile := "fission.zip"
	registryName := "serverlessbench-registry"
	tmpWrapper := "./main.py"
	binaryWrapper, err := util.Asset("resources/main.py") //fetch binary file in res.go
	binaryZip, err := util.Asset("resources/python.zip")

	// remove old registry
	rmCmd := exec.Command("docker", "rm", "-f", registryName)
	rmCmd.Run()

	// start new registry
	registryCmd := exec.Command("docker", "run", "-d", "-p", "5000:5000", "--restart", "always", "--name", registryName, "registry:2")
	out, err := registryCmd.CombinedOutput()
	if err != nil {
		err := errors.New(fmt.Sprintf("start docker local registry failed, message: %s try to run \"docker run -d -p 5000:5000 --restart always --name %s registry:2\"", string(out), registryName))
		return err
	}

	// mutiplex output
	var stdBuffer bytes.Buffer
	mw := io.MultiWriter(os.Stdout, &stdBuffer)
	// create temporary directory
	err = os.Mkdir(util.FissionTempDir, os.ModePerm)
	if err != nil {
		return err
	}

	// copy files
	z, _ := os.Create("python.zip")
	z.Write(binaryZip)
	zpcmd := exec.Command("unzip", "-o", "-d", util.FissionTempDir, "python.zip")
	err = zpcmd.Run()
	if err != nil {
		os.Remove("python.zip")
		return err
	}
	exec.Command("mv", util.FissionTempDir+"/resources/python", util.FissionTempDir).Run()
	exec.Command("rm", "-rf", util.FissionTempDir+"/resources").Run()
	err = z.Close()
	os.Remove("python.zip")

	for _, file := range SourceCodeFiles {
		util.Cp(file, util.FissionTempDir)
	}
	content, err := ioutil.ReadFile(util.FissionTempDir + "/requirements.txt")
	if err != nil {
		return err
	}
	file, err := os.OpenFile(util.FissionTempDir+"/python/requirements.txt", os.O_APPEND|os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	file.WriteString(string(content))
	defer file.Close()
	// create docker image
	fmt.Println("Start creating docker image...")
	cmd := exec.Command("docker", "build", "-t", "localhost:5000/"+FunctionName,
		"--build-arg", "PY_BASE_IMG=3.10-alpine", util.FissionTempDir+"/python")
	cmd.Stdout = mw
	cmd.Stderr = mw
	err = cmd.Run()
	if err != nil {
		return err
	}
	// push image
	exec.Command("docker", "push", "localhost:5000/"+FunctionName).CombinedOutput()

	f, _ := os.Create(tmpWrapper)
	f.Write(binaryWrapper)

	givenZip := false
	if strings.HasSuffix(SourceCodeFiles[0], ".zip") {
		dstZipFile = SourceCodeFiles[0]
		givenZip = true
	}
	files := []string{"-j", dstZipFile}
	files = append(files, tmpWrapper)
	var index int = 0
	var file1 string = ""
	if givenZip {
		for index, file1 = range SourceCodeFiles {
			if index > 1 {
				files = append(files, file1)
			}
		}
	} else {
		files = append(files, SourceCodeFiles...)
	}
	util.Zip(files)

	file, _ = os.Open(dstZipFile)
	if file == nil {
		return errors.New(dstZipFile)
	}
	defer file.Close()

	err = f.Close()
	os.Remove(tmpWrapper)

	// create python env
	// with dependencies:fission env create --name python --image fission/python-env --builder fission/python-builder
	exec.Command("fission", "env", "delete", "--name", "serverlessbench-python-"+FunctionName).CombinedOutput()
	exec.Command("fission", "env", "create", "--name", "serverlessbench-python-"+FunctionName,
		"--image", "localhost:5000/"+FunctionName+":latest", "--version", "2", "--poolsize=4",
		"--builder", "fission/python-builder").CombinedOutput()

	// create python function
	out, err = exec.Command("fission", "fn", "create", "--name", FunctionName, "--env", "serverlessbench-python-"+FunctionName,
		"--code", dstZipFile, "--entrypoint", "main.main", fmt.Sprintf("--maxcpu=%d", int(MemorySize*1000/128)),
		fmt.Sprintf("--maxmemory=%d", MemorySize), "--concurrency=1").CombinedOutput()
	if err != nil {
		fmt.Println("Error: " + string(out))
		return err
	} else {
		fmt.Println("Create function:", FunctionName)
		os.Remove(dstZipFile)
	}
	return nil
}

func (fission ProviderFission) InvokeWorkflow(function_name string, params ...string) error {
	return nil
}

func (fission ProviderFission) InvokeFunction(FunctionName string, Params ...string) error {
	// construct the input string
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
	// invoke
	out, err := exec.Command("fission", "fn", "test", "--name", FunctionName, "-b", input).CombinedOutput()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("error occured when invoke function %s.", FunctionName))
	} else {
		fmt.Println(string(out))
	}
	return nil
}

func (fission ProviderFission) DeleteFunction(FunctionName string) error {
	cmd := exec.Command("fission", "fn", "delete", "--name", FunctionName)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error: " + string(out))
		// return err
	} else {
		fmt.Println("Delete function:", FunctionName)
	}
	return nil
}

func (fission ProviderFission) DeleteWorkflow(FlowName string) error {
	cmd := exec.Command("fission", "fn", "delete", "--name", FlowName)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error: " + string(out))
		// return err
	} else {
		fmt.Println("Delete function:", FlowName)
	}
	return nil
}

func (fission ProviderFission) ListFunction() error {
	fmt.Println("List functions:")
	cmd := exec.Command("fission", "fn", "list")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrap(err, "error occured when listing functions on fission.")
	} else {
		fmt.Print(string(out))
	}
	return nil
}
