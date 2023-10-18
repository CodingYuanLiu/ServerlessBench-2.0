package test

import (
	"os/exec"
	"regexp"
	config "serverlessbench2/internal/config"
	util "serverlessbench2/internal/util"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPrequisite(t *testing.T) {
	Convey("Config file should be filled properly. ", t, func() {

		Convey("Path of Python interpreter shouldn't be empty. ", func() {
			PythonPath := config.GetPythonPath()
			So(PythonPath, ShouldNotBeNil)
		})
		Convey("Name of cloud Provider should be set properly. ", func() {
			Provider := config.GetProvider()
			So(Provider, ShouldBeIn, []string{"knative", "openwhisk", "openfaas", "fission"})
		})
		Convey("Username of Docker shouldn't be empty. ", func() {
			DockerUserName := config.GetDockerUserName()
			So(DockerUserName, ShouldNotBeNil)
		})
		Convey("Path of result directory shouldn't be empty. ", func() {
			TestResultDir := config.GetTestResultDir()
			So(TestResultDir, ShouldNotBeNil)
		})
	})
	Convey("Create cli.", t, func() {
		createCmd := exec.Command("make")
		createCmd.Dir = "../"
		_, err := createCmd.CombinedOutput()

		Convey("Makefile executed.", func() {
			So(err, ShouldBeNil)
		})
		Convey("cli exists.", func() {
			flieList, err := util.ParseFilesFromDir("..")
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
	})
}
