package runner

import (
	"fmt"
	"os"
	"os/exec"
	"serverlessbench2/internal/config"
	"serverlessbench2/internal/util"
	"strings"
)

func (flow WorkflowRunner) TestcaseRunner(testName string, testcaseDir string, appName string, appInfo interface{}, param string) error {
	var srcDir, reqPath, memSize string
	var infoSlice []string

	pythonPath := config.GetPythonPath()
	CliBaseDir := config.GetCliBaseDir()
	testResultDir := config.GetTestResultDir() + "/" + testName
	provider := config.GetProvider()

	appInfoMap := appInfo.([]map[string](map[string]string))
	_, err := os.Stat(testResultDir)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(testResultDir, 0777)
			if err != nil {
				return err
			}

		}
	}

	srcDirList := []string{}
	memSizeList := []string{}
	reqPathList := []string{}
	stageNameList := []string{}

	for _, stageInfo := range appInfoMap {
		for name, dirMap := range stageInfo {
			infoSlice = util.GetFuncInfoSlice(dirMap)
			tempDir, _ := os.MkdirTemp("", "")
			defer os.RemoveAll(tempDir)
			srcDir, reqPath, memSize = infoSlice[0], infoSlice[1], infoSlice[2]
			cmd := exec.Command("cp", "-r", srcDir+"/.", tempDir)
			_, err = cmd.CombinedOutput()
			if err != nil {
				return err
			}
			srcDir = tempDir
			srcDirList = append(srcDirList, srcDir)
			memSizeList = append(memSizeList, memSize)
			reqPathList = append(reqPathList, reqPath)
			stageNameList = append(stageNameList, name)
			err := os.Rename(srcDir+"/index.py", srcDir+"/indexUserFunc.py")
			if err != nil {
				return err
			}
			util.Cp(testcaseDir+"/index.py", srcDir+"/index.py")
			util.Cp(reqPath, srcDir+"/requirements.txt")
		}
	}

	fmt.Println("Running python test file.")

	err = util.RunCommand(pythonPath, "-u", testcaseDir+"/test.py", "--srcPathList="+strings.Join(srcDirList, ","), "--memSizeList="+strings.Join(memSizeList, ","), "--reqPathList="+strings.Join(reqPathList, ","), "--stageNameList="+strings.Join(stageNameList, ","), "--param="+param, "--appName="+appName, "--cliBase="+CliBaseDir, "--testCaseDir="+testcaseDir, "--provider="+provider, "--resultDir="+testResultDir)
	if err != nil {
		for _, srcDir := range srcDirList {
			_ = os.Remove(srcDir + "/index.py")
			_ = os.Rename(srcDir+"/indexUserFunc.py", srcDir+"/index.py")
		}
		return err
	}
	fmt.Printf("Test result collected in %s\n", testResultDir)
	for _, srcDir := range srcDirList {
		err = os.Remove(srcDir + "/index.py")
		if err != nil {
			return err
		}
		err = os.Rename(srcDir+"/indexUserFunc.py", srcDir+"/index.py")
		if err != nil {
			return err
		}
	}
	return nil
}
