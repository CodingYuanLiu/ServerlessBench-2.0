package util

import (
	"os/exec"
	"strings"
)

func Zip(files []string) error {
	c := exec.Command("zip", files...)
	_, err := c.Output()
	if err != nil {
		return err
	}
	return nil
}

func GetStringInBetween(str string, start string, end string) (result string) {
	s := strings.Index(str, start)
	s += len(start)
	e := strings.Index(str[s:], end)
	return str[s : e+s]
}
