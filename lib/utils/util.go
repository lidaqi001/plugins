package utils

import (
	"github.com/pkg/errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GetProgramName 获取程序名称
func GetProgramName() (string, error) {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	path, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}
	i := strings.LastIndex(path, "/")
	if i < 0 {
		i = strings.LastIndex(path, "\\")
	}
	if i < 0 {
		return "", errors.New(`error: Can't find "/" or "\".`)
	}

	strProgram := string(path[i+1:])
	j := strings.LastIndex(strProgram, ".")
	if j > 0 {
		return string(strProgram[0:j]), nil
	}

	return string(strProgram[:]), nil
}
