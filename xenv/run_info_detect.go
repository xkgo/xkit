package xenv

import "os"

/**
运行环境识别&检测，识别逻辑如下：
*/
type Detector struct {
	Name   string                               // 名称
	Detect func() (info *RunInfo, matched bool) // 环境信息获取函数, 是否 匹配，匹配的话表示能够识别成功
}

/**
标准环境识别：
默认部署环境，使用命令行参数进行部署， --env=dev|test|prod, --set=xxx
*/
var StandardDetector = &Detector{
	Name: "StandardDetector",
	Detect: func() (info *RunInfo, matched bool) {
		wd, _ := os.Getwd()
		properties := GetCommandLineProperties("")
		if envStr, ok := properties["env"]; ok {
			env := ParseEnv(envStr)
			set := properties["set"]
			return &RunInfo{
				Env:     env,
				Set:     set,
				WorkDir: wd,
			}, true
		}

		// 从环境变量中获取
		envStr := os.Getenv("env")
		if len(envStr) > 0 {
			env := ParseEnv(envStr)
			set := os.Getenv("set")
			return &RunInfo{
				Env:     env,
				Set:     set,
				WorkDir: wd,
			}, true
		}

		return &RunInfo{
			Env:     Dev,
			WorkDir: wd,
		}, true
	},
}

// 自定义环境检测器列表，检测顺序按照 detectorNameOrders 执行
var customDetectors = make([]*Detector, 0)

// 检测顺序
var detectorNameOrders = make([]string, 0)

// 自定义运行信息
var customRunInfo *RunInfo

func UseCustomRunInfo(info *RunInfo) {
	customRunInfo = info
}

/**
清空当前的自定义环境检测器
*/
func ResetCustomDetectors(detectors ...*Detector) {
	customDetectors = make([]*Detector, 0)

	if nil != detectors && len(detectors) > 0 {
		customDetectors = append(customDetectors, detectors...)
	}
}

/**
添加到最高优先级的列表
*/
func AddCustomDetectorToFirst(name string, handler func() (info *RunInfo, matched bool)) {
	if len(name) < 1 || nil == handler {
		return
	}
	detector := &Detector{
		Name:   name,
		Detect: handler,
	}
	list := make([]*Detector, 0)
	list = append(list, detector)
	if nil != customDetectors && len(customDetectors) > 0 {
		list = append(list, customDetectors...)
	}
	customDetectors = list
}

func AddCustomDetectorToLast(name string, handler func() (info *RunInfo, matched bool)) {
	if len(name) < 1 || nil == handler {
		return
	}
	detector := &Detector{
		Name:   name,
		Detect: handler,
	}
	if customDetectors == nil {
		customDetectors = make([]*Detector, 0)
	}
	customDetectors = append(customDetectors, detector)
}

func DefineDetectorOrders(nameOrders []string) {
	detectorNameOrders = nameOrders
}

/**
检测当前运行环境
*/
func DetectEnvInfo() *RunInfo {
	if nil != customRunInfo {
		if len(customRunInfo.WorkDir) < 1 {
			customRunInfo.WorkDir, _ = os.Getwd()
		}
		return customRunInfo
	}

	detectors := make([]*Detector, 0)
	existsNames := make(map[string]bool)
	for _, name := range detectorNameOrders {
		existsNames[name] = true
		for _, item := range customDetectors {
			if name == item.Name {
				detectors = append(detectors)
			}
		}
	}
	// 把剩余的加进来
	for _, item := range customDetectors {
		if _, ok := existsNames[item.Name]; !ok {
			detectors = append(detectors)
		}
	}
	// 把默认的加进来
	detectors = append(detectors, StandardDetector)

	// 遍历
	for _, detector := range detectors {
		if info, matched := detector.Detect(); matched && nil != info {
			if len(info.WorkDir) < 1 {
				info.WorkDir, _ = os.Getwd()
			}
			return info
		}
	}
	panic("无法识别当前运行环境！")
}
