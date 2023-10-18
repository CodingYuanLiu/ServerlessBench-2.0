package util

import (
	"fmt"
	"os/exec"

	"github.com/pkg/errors"
)

func KubectlApplyFile(filePath string) error {
	cmd := exec.Command("kubectl", "apply", "-f", filePath)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func DockerBuildPush(dstName string, option string) error {
	var cmd *(exec.Cmd)
	switch option {
	case "build":
		cmd = exec.Command("docker", "build", "-t", dstName, ".")
	case "push":
		cmd = exec.Command("docker", "push", dstName)
	default:
		return errors.New("command option must be push or build.")
	}
	res, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("docker build failed. raw output:%s", res))
	}
	return nil
}

func RunCommand(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	stdout, err := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout

	if err != nil {
		return err
	}

	if err = cmd.Start(); err != nil {
		return err
	}

	for {
		tmp := make([]byte, 1024)
		_, err := stdout.Read(tmp)
		fmt.Print(string(tmp))
		if err != nil {
			break
		}
	}

	if err = cmd.Wait(); err != nil {
		return err
	}
	return nil
}
