package runner

import (
	"fmt"
	"os"
	"os/exec"
	"serverlessbench2/internal/config"
	"serverlessbench2/internal/util"
)

func (single SingleRunner) TestcaseRunner(testName string, testcaseDir string, appName string, appInfo interface{}, param string) error {
	var srcDir, reqPath, memSize string
	var infoSlice []string

	pythonPath := config.GetPythonPath()
	CliBaseDir := config.GetCliBaseDir()
	testResultDir := config.GetTestResultDir() + "/" + testName
	provider := config.GetProvider()

	infoSlice = util.GetFuncInfoSlice(appInfo.(map[string]string))
	_, err := os.Stat(testResultDir)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(testResultDir, 0777)
			if err != nil {
				return err
			}

		}
	}
	tempDir, _ := os.MkdirTemp("", "")
	defer os.RemoveAll(tempDir)
	srcDir, reqPath, memSize = infoSlice[0], infoSlice[1], infoSlice[2]
	cmd := exec.Command("cp", "-r", srcDir+"/.", tempDir)
	_, err = cmd.CombinedOutput()
	if err != nil {
		return err
	}
	srcDir = tempDir
	err = os.Rename(srcDir+"/index.py", srcDir+"/indexUserFunc.py")
	if err != nil {
		return err
	}
	util.Cp(testcaseDir+"/index.py", srcDir+"/index.py")
	util.Cp(reqPath, srcDir+"/requirements.txt")
	fmt.Println("Running python test file.")
	err = util.RunCommand(pythonPath, "-u", testcaseDir+"/test.py", "--srcPath="+srcDir, "--memory="+memSize, "--param="+param, "--appName="+appName, "--cliBase="+CliBaseDir, "--testCaseDir="+testcaseDir, "--provider="+provider, "--resultDir="+testResultDir)
	if err != nil {
		_ = os.Remove(srcDir + "/index.py")
		_ = os.Rename(srcDir+"/indexUserFunc.py", srcDir+"/index.py")
		return err
	}
	fmt.Printf("Test result collected in %s\n", testResultDir)
	err = os.Remove(srcDir + "/index.py")
	if err != nil {
		return err
	}
	err = os.Rename(srcDir+"/indexUserFunc.py", srcDir+"/index.py")
	if err != nil {
		return err
	}

	return nil
}
