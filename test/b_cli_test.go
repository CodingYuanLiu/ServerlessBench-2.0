package test

import (
	"encoding/json"
	"os"
	"os/exec"
	"regexp"
	config "serverlessbench2/internal/config"
	util "serverlessbench2/internal/util"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCli(t *testing.T) {
	testDir, _ := os.Getwd()
	funcSrcPath := "./apps/function/numAddOne/index.py"
	reqPathFirst := "./apps/function/numAddOne/requirements.txt"
	deleteCmd := exec.Command("./cli", "delete", "test-numaddone")
	createCmd := exec.Command("./cli", "create", "test-numaddone", funcSrcPath, reqPathFirst)
	invokeCmd := exec.Command("./cli", "invoke", "test-numaddone", "number", "19")
	Convey("Change to root dir of serverlessbench2.", t, func() {
		err := os.Chdir("..")
		So(err, ShouldBeNil)
		currentDir, err := os.Getwd()
		So(err, ShouldBeNil)
		flieList, err := util.ParseFilesFromDir(currentDir)
		So(err, ShouldBeNil)
		rx := regexp.MustCompile("cli$")
		cliExistFlag := false
		for _, fileName := range flieList {
			cliExistFlag = rx.MatchString(fileName)
			if cliExistFlag {
				break
			}
		}
		So(cliExistFlag, ShouldBeTrue)
	})
	Convey("Test modifying config.", t, func() {

		Convey("Correct command check.", func() {
			DockerUserName := config.GetDockerUserName()
			resultDir := config.GetTestResultDir()
			pythonPath := config.GetPythonPath()
			Provider := config.GetProvider()
			recoverConfigCmd := exec.Command("./cli", "config", "--DockerUserName", DockerUserName, "--Provider", Provider, "--PythonPath", pythonPath, "--TestResultDir", resultDir)
			configCmd := exec.Command("./cli", "config", "--DockerUserName", "user", "--Provider", "openfaas", "--PythonPath", "python", "--TestResultDir", "result")
			_, err := configCmd.CombinedOutput()
			So(err, ShouldBeNil)
			So(config.GetDockerUserName(), ShouldEqual, "user")
			So(config.GetTestResultDir(), ShouldEqual, "result")
			So(config.GetPythonPath(), ShouldEqual, "python")
			So(config.GetProvider(), ShouldEqual, "openfaas")
			_, err = recoverConfigCmd.CombinedOutput()
			So(err, ShouldBeNil)
			So(config.GetDockerUserName(), ShouldEqual, DockerUserName)
			So(config.GetTestResultDir(), ShouldEqual, resultDir)
			So(config.GetPythonPath(), ShouldEqual, pythonPath)
			So(config.GetProvider(), ShouldEqual, Provider)
		})
		Convey("Default pythonPath check.", func() {
			DockerUserName := config.GetDockerUserName()
			resultDir := config.GetTestResultDir()
			pythonPath := config.GetPythonPath()
			Provider := config.GetProvider()
			recoverConfigCmd := exec.Command("./cli", "config", "--DockerUserName", DockerUserName, "--Provider", Provider, "--PythonPath", pythonPath, "--TestResultDir", resultDir)
			configCmd := exec.Command("./cli", "config", "--PythonPath", "p")
			_, err := configCmd.CombinedOutput()
			So(err, ShouldBeNil)
			So(config.GetPythonPath(), ShouldEqual, "p")
			defaultPythonCmd := exec.Command("./cli", "config")
			_, err = defaultPythonCmd.CombinedOutput()
			So(err, ShouldBeNil)
			So(config.GetPythonPath(), ShouldEqual, "python")
			_, err = recoverConfigCmd.CombinedOutput()
			So(err, ShouldBeNil)
			So(config.GetDockerUserName(), ShouldEqual, DockerUserName)
			So(config.GetTestResultDir(), ShouldEqual, resultDir)
			So(config.GetPythonPath(), ShouldEqual, pythonPath)
			So(config.GetProvider(), ShouldEqual, Provider)
		})
		Convey("Special platform setting check.", func() {
			provider := config.GetProvider()
			switch provider {
			case "knative":
				So(config.GetKnGoWrapperFile(), ShouldNotEqual, "")
				So(config.GetKnServiceYaml(), ShouldNotEqual, "")
				So(config.GetKnServiceFile(), ShouldNotEqual, "")
				So(config.GetKnPythonDir(), ShouldNotEqual, "")
				So(config.GetKnWorkflowDockerfile(), ShouldNotEqual, "")
				So(config.GetKnWorkflowYaml(), ShouldNotEqual, "")
				So(config.GetKnDiaplayerAndSenderYaml(), ShouldNotEqual, "")
				So(config.GetKnGoWrapperDir(), ShouldNotEqual, "")
			case "openwhisk":
				So(config.GetOpenwhiskZipFileName(), ShouldNotEqual, "")
				So(config.GetOpenwhiskTempDirPrefix(), ShouldNotEqual, "")
			}
		})
	})
	Convey("Test operations for single function.", t, func() {
		Convey("Create single function.", func() {
			Convey("Correct command check.", func() {
				_, err := deleteCmd.CombinedOutput()
				_, err = createCmd.CombinedOutput()
				So(err, ShouldBeNil)
			})

			Convey("Too few parameter check.", func() {
				wrongCreateCmd := exec.Command("./cli", "create", "wrong-func")
				_, err := wrongCreateCmd.CombinedOutput()
				So(err, ShouldNotBeNil)
			})
			Convey("Wrong function name check.", func() {
				wrongNameCreateCmdOne := exec.Command("./cli", "create", "Wrong-func", "index.py")
				wrongNameCreateCmdTwo := exec.Command("./cli", "create", "wrong_func", "index.py")
				_, err := wrongNameCreateCmdOne.CombinedOutput()
				So(err, ShouldNotBeNil)
				_, err = wrongNameCreateCmdTwo.CombinedOutput()
				So(err, ShouldNotBeNil)
			})
			Convey("No requirement file check.", func() {
				wrongCreateCmd := exec.Command("./cli", "create", "Wrong-func", "index.py")
				_, err := wrongCreateCmd.CombinedOutput()
				So(err, ShouldNotBeNil)
			})
			Convey("Directory without requirement file check.", func() {
				err := os.Mkdir("testFuncDir", os.ModePerm)
				So(err, ShouldBeNil)
				wrongCreateCmd := exec.Command("./cli", "create", "-d", "Wrong-func", "testFuncDir")
				_, err = wrongCreateCmd.CombinedOutput()
				So(err, ShouldNotBeNil)
				err = os.RemoveAll("testFuncDir")
				So(err, ShouldBeNil)
			})
		})
		Convey("Invoke single function.", func() {
			Convey("Correct command check.", func() {

				out, err := invokeCmd.CombinedOutput()
				res := make(map[string]int)
				err = json.Unmarshal([]byte(out), &res)
				So(res["number"], ShouldEqual, 20)
				deleteCmdNew := exec.Command(deleteCmd.Path, deleteCmd.Args[1:]...)
				_, err = deleteCmdNew.CombinedOutput()
				So(err, ShouldBeNil)
				_, err = invokeCmd.CombinedOutput()
				So(err, ShouldNotBeNil)
			})
			Convey("Wrong number of parameters check.", func() {
				wrongInvokeCmd := exec.Command("./cli", "invoke", "test-numaddone", "text")
				_, err := wrongInvokeCmd.CombinedOutput()
				So(err, ShouldNotBeNil)
			})
		})
	})
	Convey("Change to test dir.", t, func() {
		err := os.Chdir(testDir)
		So(err, ShouldBeNil)
		currentDir, err := os.Getwd()
		So(err, ShouldBeNil)
		So(currentDir, ShouldEqual, testDir)
	})
}
