package util

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	"github.com/pkg/errors"
)

func LoadFile(src string, dst string) {
	f, _ := os.Create(dst + "/" + src)
	bi, _ := Asset("resources/" + src)
	f.Write(bi)
	defer f.Close()
}

func Cp(src string, dst string) error {
	c := exec.Command("cp", "-r", src, dst)
	_, err := c.CombinedOutput()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("error occured when copying file %s to %s.", src, dst))
	}
	return nil
}

func Mv(src string, dst string) error {
	c := exec.Command("mv", "-r", src, dst)
	_, err := c.CombinedOutput()
	if err != nil {
		return err
	}
	return nil
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func EnsureFileExists(filename string) error {
	_, err := os.Stat(filename)

	if os.IsNotExist(err) {
		file, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer file.Close()
		err = os.Chmod(filename, 0644)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}

func ParseFilesFromDir(dirPath string) (ret []string, err error) {
	var src_files []string
	_dir, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}
	for _, _file := range _dir {
		src_files = append(src_files, path.Join(dirPath, _file.Name()))
	}
	return src_files, nil
}

func ReplaceWordInFile(fileName string, key []string, value []string) error {
	for i := 0; i < len(key); i++ {
		input, err := ioutil.ReadFile(fileName)
		if err != nil {
			return err
		}
		output := bytes.Replace(input, []byte(key[i]), []byte(value[i]), -1)
		if err = ioutil.WriteFile(fileName, output, 0666); err != nil {
			return err
		}
	}
	return nil
}
