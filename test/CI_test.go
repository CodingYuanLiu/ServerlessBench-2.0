package test

import (
	"os/exec"
	"regexp"
	util "serverlessbench2/internal/util"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBuild(t *testing.T) {
	Convey("Create cli.", t, func() {
		createCmd := exec.Command("make")
		createCmd.Dir = ".."
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
