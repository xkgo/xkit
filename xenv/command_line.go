package xenv

import (
	"github.com/xkgo/xkit/xstr"
	"os"
	"strings"
)

/**
获取命令行参数, 命令行参数格式：
--{key}={value}
如：
./app.exe --env=test --set=sg --cluster=asia
*/
func GetCommandLineProperties(appendCommandLine string) map[string]string {
	var args = os.Args
	if len(appendCommandLine) > 4 {
		args = append(args, xstr.SplitByRegex(appendCommandLine, "\\s+")...)
	}

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
	return properties
}
