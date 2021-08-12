package xenv

import (
	"github.com/xkgo/xkit/xstr"
	"os"
	"sync"
)

const (
	/** 系统环境变量 PropertySource GetName */
	SystemEnvironmentPropertySourceName = "systemEnvironment"
)

/**
系统环境变量 属性来源
*/
type SystemEnvironmentPropertySource struct {
	MapPropertySource
}

func NewSystemEnvironmentPropertySource() *SystemEnvironmentPropertySource {
	source := &SystemEnvironmentPropertySource{}
	source.name = SystemEnvironmentPropertySourceName
	source.properties = &sync.Map{}

	envs := os.Environ()
	for _, kv := range envs {
		kvs := xstr.SplitByRegex(kv, "\\s*=\\s*")
		if len(kvs) == 1 {
			source.properties.Store(xstr.Trim(kvs[0]), "")
		} else if len(kvs) == 2 {
			source.properties.Store(xstr.Trim(kvs[0]), xstr.Trim(kvs[1]))
		}
	}
	return source
}
