package xenv

import (
	"fmt"
	"github.com/xkgo/xkit/xstr"
)

type Env string

const (
	Dev  Env = "dev"  // 开发
	Test Env = "test" // 测试
	Fat  Env = "fat"  // 预发布
	Prod Env = "prod" // 生产
)

type RunInfo struct {
	Env        Env               // 当前运行环境
	Set        string            // 当前部署所在部署集，如所在大区，或者说所在部署集群等标识， 默认就是空字符串
	WorkDir    string            // 应用工作所在目录
	Properties map[string]string // 当前运行环境下的属性配置信息， 可能每个部署平台都有自己特殊的一些配置信息
}

func (i *RunInfo) String() string {
	return fmt.Sprintf("Env:%v, Set:%v, props:%v", i.Env, i.Set, i.Properties)
}

func (i *RunInfo) IsDev() bool {
	return Dev == i.Env
}

func (i *RunInfo) IsTest() bool {
	return Test == i.Env
}

func (i *RunInfo) IsFat() bool {
	return Fat == i.Env
}

func (i *RunInfo) IsProd() bool {
	return Prod == i.Env
}

// 自定义环境
var customEnvs = make(map[string]Env)

/**
自定义环境
*/
func DefineEnv(envStr string, env Env) {
	if len(envStr) > 0 && len(env) > 0 {
		customEnvs[envStr] = env
	}
}

func ParseEnv(env string) Env {
	if len(customEnvs) > 0 {
		for key, val := range customEnvs {
			if xstr.EqualsIgnoreCase(env, key) {
				return val
			}
		}
	}
	if xstr.EqualsIgnoreCase(env, "test") {
		return Test
	}
	if xstr.EqualsIgnoreCase(env, "fat") {
		return Fat
	}
	if xstr.EqualsIgnoreCase(env, "prod") {
		return Prod
	}
	return Dev
}
