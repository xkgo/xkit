package xenv

import (
	"github.com/xkgo/xkit/xstr"
	"os"
	"strings"
)

const (
	/** 命令行变量 PropertySource GetName */
	CommandLineEnvironmentPropertySourceName = "commandLineEnvironment"
)

/***
命令行参数解析，解析格式：
格式： --key=value
实例：
	--name=arvin ==> key=name, value = arvin
	--name=      ==> key=name, value = ""
*/
type CommandLinePropertySource struct {
	MapPropertySource
}

func NewCommandLinePropertySource(appendCommandLine string) *CommandLinePropertySource {
	var args []string = os.Args
	if len(appendCommandLine) > 4 {
		args = append(args, xstr.SplitByRegex(appendCommandLine, "\\s+")...)
	}

	name := CommandLineEnvironmentPropertySourceName
	properties := make(map[string]string)

	for _, arg := range args {
		if len(arg) < 4 || !strings.HasPrefix(arg, "--") { // 至少四个字符， --k=
			continue
		}
		// 替换一次
		arg = xstr.Trim(strings.Replace(arg, "--", "", 1))
		index := strings.Index(arg, "=")
		if index < 1 {
			continue
		}

		key := xstr.Trim(arg[0:index])
		value := xstr.Trim(arg[index+1:])

		properties[key] = value
	}

	source := &CommandLinePropertySource{*NewMapPropertySource(name, properties)}

	return source
}

/**
获取指定的命令行参数
*/
func GetCommandLineProperty(key string) (value string, exists bool) {
	if len(key) < 1 {
		return "", false
	}

	for _, arg := range os.Args {
		if len(arg) < 4 || !strings.HasPrefix(arg, "--") { // 至少四个字符， --k=
			continue
		}
		// 替换一次
		arg = xstr.Trim(strings.Replace(arg, "--", "", 1))
		index := strings.Index(arg, "=")
		if index < 1 {
			continue
		}

		k := xstr.Trim(arg[0:index])

		if k == key {
			value := xstr.Trim(arg[index+1:])
			return value, true
		}
	}
	return "", false
}
