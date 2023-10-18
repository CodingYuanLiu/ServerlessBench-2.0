package benchinterface

import (
	"fmt"
	"os"
	"os/exec"
	config "serverlessbench2/internal/config"
	"serverlessbench2/internal/util"
	"strconv"
	"strings"
	"syscall"

	"github.com/pkg/errors"

	"github.com/google/uuid"
)

func (ow ProviderOpenwhisk) CreateWorkflow(flowName string, functionInfo ...map[string](map[string]string)) error {
	stages := ""
	for _, finfo := range functionInfo {
		for name := range finfo {
			stageName := flowName + "-stage-" + name
			stages += stageName
			stages += ","
			fmt.Printf("Flush the existing function stage %s.\n", name)
			_ = DeleteFunction(stageName)
			src_path := finfo[name]["src"]
			reqPath := finfo[name]["req"]
			memory, err := strconv.Atoi(finfo[name]["memory"])
			if err != nil {
				return err
			}
			// from src_path get paths of all source code files.
			srcFiles, err := util.ParseFilesFromDir(src_path)
			if err != nil {
				return err
			}
			srcFiles = append(srcFiles, reqPath)
			err = CreateFunction(stageName, memory, srcFiles[:]...)
			if err != nil {
				return err
			}
			fmt.Printf("function %s for workflow %s is created.\n", stageName, flowName)
		}
	}
	cmd := exec.Command("wsk", "-i", "action", "create", flowName, "--sequence", stages[:len(stages)-1])
	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	fmt.Println("Create workflow:", flowName)
	return nil
}

func (ow ProviderOpenwhisk) CreateFunction(FunctionName string, memory_size int, SourceCodeFiles ...string) error {
	dstZipFile := config.GetOpenwhiskZipFileName()
	tmp_wrapper := config.GetOpenwhiskEntrypoint()
	wskTempDir := config.GetOpenwhiskTempDirPrefix() + uuid.New().String()

	// create temporary directory
	mask := syscall.Umask(0)
	defer syscall.Umask(mask)
	err := os.Mkdir(wskTempDir, 0777)
	if err != nil {
		return err
	}

	// copy wrapperFile to temporary directory
	util.LoadFile(tmp_wrapper, wskTempDir)
	util.LoadFile(tmp_wrapper, "/tmp")
	tmp_wrapper = "/tmp/" + tmp_wrapper

	// copy source files to temporary directory
	curdir, _ := os.Getwd()
	for _, file := range SourceCodeFiles {
		err := util.Cp(file, wskTempDir)
		if err != nil {
			return err
		}
	}

	// Check if zip dile name is given in SourceCodeFiles.
	givenZip := false
	if strings.HasSuffix(SourceCodeFiles[0], ".zip") {
		dstZipFile = SourceCodeFiles[0]
		givenZip = true
	}
	files := []string{"-j", dstZipFile}
	files = append(files, tmp_wrapper)

	// skip zip file name(if given)
	if givenZip {
		for index, file := range SourceCodeFiles {
			if index > 1 {
				files = append(files, file)
			}
		}
	} else {
		files = append(files, SourceCodeFiles...)
	}
	err = util.Zip(files)
	if err != nil {
		return err
	}

	_, err = exec.Command("mv", dstZipFile, wskTempDir).CombinedOutput()
	if err != nil {
		return err
	}

	err = os.Chdir(wskTempDir)
	if err != nil {
		return err
	}

	// config the env
	// docker run --rm -v "$PWD:/tmp" openwhisk/python3action bash -c "cd tmp && virtualenv virtualenv && source virtualenv/bin/activate && pip install -r requirements.txt && chmod 777 -R virtualenv"
	fmt.Println("config virtualenv.")
	cmd := exec.Command("docker", "run", "--rm", "-v", wskTempDir+":/tmp", "openwhisk/python3action", "bash", "-c",
		"cd tmp && virtualenv virtualenv && source virtualenv/bin/activate && pip install -r requirements.txt && chmod 777 -R virtualenv")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("fail to config virtualenv. raw result:%s", string(out)))
	}
	// zip env
	_, err = exec.Command("zip", "-r", dstZipFile, "virtualenv").CombinedOutput()
	if err != nil {
		return err
	}

	// deploy function
	fmt.Println("create openwhisk function.")
	cmd = exec.Command("wsk", "-i", "action", "create", FunctionName, dstZipFile, "--kind", "python:3", "--main", "main",
		"-m", fmt.Sprintf("%d", memory_size))
	_, err = cmd.CombinedOutput()
	if err != nil {
		return err
	} else {
		fmt.Println("Create function:", FunctionName)
	}
	err = os.Chdir(curdir)
	if err != nil {
		return err
	}
	os.Remove(tmp_wrapper)
	err = os.RemoveAll(wskTempDir)
	if err != nil {
		return err
	}
	return nil
}

func (ow ProviderOpenwhisk) InvokeWorkflow(function_name string, params ...string) error {
	return InvokeFunction(function_name, params...)
}

func (ow ProviderOpenwhisk) InvokeFunction(function_name string, params ...string) error {
	cmd := exec.Command("wsk", "-i", "action", "invoke", "-r", function_name)
	if len(params)%2 != 0 {
		err := fmt.Sprintf("params must be key-value pairs")
		return errors.New(err)
	}
	for i := 0; i < len(params); i += 2 {
		cmd.Args = append(cmd.Args, "-p", params[i], params[i+1])
	}
	out, _ := cmd.CombinedOutput()

	fmt.Println(string(out))
	return nil
}

func (ow ProviderOpenwhisk) DeleteFunction(function_name string) error {
	cmd := exec.Command("wsk", "-i", "action", "delete", function_name)
	_, _ = cmd.CombinedOutput()
	return nil
}

func (ow ProviderOpenwhisk) DeleteWorkflow(function_name string) error {
	cmd := exec.Command("wsk", "-i", "action", "delete", function_name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error: " + string(out))
	}
	fmt.Println("Delete workflow:", function_name)
	return nil
}

func (ow ProviderOpenwhisk) ListFunction() error {
	fmt.Println("List functions:")
	cmd := exec.Command("wsk", "-i", "action", "list")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	fmt.Print(string(out))
	return nil
}
