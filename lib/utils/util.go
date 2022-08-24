package utils

import (
	"github.com/pkg/errors"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
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

	strProgram := path[i+1:]
	j := strings.LastIndex(strProgram, ".")
	if j > 0 {
		return strProgram[0:j], nil
	}

	return strProgram[:], nil
}

// RandomMixString 随机生成字符串(英文数字混合)
func RandomMixString(l int) string {
	str := "0123456789AaBbCcDdEeFfGgHhIiJjKkLlMmNnOoPpQqRrSsTtUuVvWwXxYyZz"
	bytes := []byte(str)
	var result []byte = make([]byte, 0, l)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < l; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return BytesToString(result)
}

// RandomNumberString 随机生成字符串
func RandomNumberString(l int) (result string) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var tmp int
	i := 0
	for i < l {
		tmp = r.Intn(10)
		if i == 0 && tmp == 0 {
			// 首字符为0，跳过
			continue
		}
		result += Int2String(tmp)
		i++
	}
	return
}
