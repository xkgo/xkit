package xver

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
)

// 版本限制， key 为 平台代码， value 为版本限制
type VersionLimit map[string]VersionRange

type VersionRange struct {
	Min      string   `json:"min"`      // 最小版本号，含
	Max      string   `json:"max"`      // 最大版本号，含
	Excludes []string `json:"excludes"` // 忽略版本
}

var prefixNotNumRegex, _ = regexp.Compile("^[^\\d]+")
var suffixNotNumRegex, _ = regexp.Compile("[^\\d]+$")
var verRegex, _ = regexp.Compile("^[\\d]+(\\.[\\d]+)*$")

/**
整理版本号，将头尾不是数字的去掉
*/
func resolveVersion(version string) (string, bool) {
	version = prefixNotNumRegex.ReplaceAllString(version, "")
	version = suffixNotNumRegex.ReplaceAllString(version, "")
	if verRegex.MatchString(version) {
		return version, true
	}
	return "", false
}

func CheckVersion(version string) (string, bool) {
	return resolveVersion(version)
}

const (
	BothInvalid = -111 // 两个版本号规则都错误
	Ver1Invalid = -110 // 第一个版本号不合法
	Ver2Invalid = -101 // 第二个版本号不合法
	EQUALS      = 0    // 相等
	GreaterThan = 1    // ver1 > ver2
	LessThan    = -1   // ver1 < ver2
)

func IsValid(version string) bool {
	_, ok := resolveVersion(version)
	return ok
}

/*
版本号比较
返回值说明：
	-111	两个版本号规则都错误
	-110 	第一个版本号不合法
	-101	第二个版本号不合法
	   0    两个版本号相等
       1    ver1 > ver2
      -1    ver1 < ver2
*/
func compare(ver1, ver2 string) int {
	if ver1 == ver2 {
		return EQUALS
	}
	ver1, isValid1 := resolveVersion(ver1)
	ver2, isValid2 := resolveVersion(ver2)

	if !isValid1 && !isValid2 {
		return BothInvalid
	}
	if !isValid1 {
		return Ver1Invalid
	}
	if !isValid2 {
		return Ver2Invalid
	}

	if ver1 == ver2 {
		return EQUALS
	}

	arr1 := strings.Split(ver1, ".")
	arr2 := strings.Split(ver2, ".")

	// 从左到右进行比对
	var i = 0
	for i < len(arr1) && i < len(arr2) {
		v1, _ := strconv.ParseInt(arr1[i], 10, 64)
		v2, _ := strconv.ParseInt(arr2[i], 10, 64)

		if v1 != v2 {
			if v1 > v2 {
				return GreaterThan
			}
			return LessThan
		}
		i++
	}
	return EQUALS
}

/**
 * 第一个版本号是否 大于等于 第二个版本号
 *
 * @param first  第一个版本号
 * @param second 第二个版本号
 * @return true 是， false - 不是
 */
func FirstGteSecond(first, second string) bool {
	ret := compare(first, second)
	return ret == EQUALS || ret == GreaterThan
}

/**
 * 第一个版本号是否 大于 第二个版本号
 *
 * @param first  第一个版本号
 * @param second 第二个版本号
 * @return true 是， false - 不是
 */
func FirstGtSecond(first, second string) bool {
	ret := compare(first, second)
	return ret > GreaterThan
}

/**
 * 第一个版本号是否 小于等于 第二个版本号
 *
 * @param first  第一个版本号
 * @param second 第二个版本号
 * @return true 是， false - 不是
 */
func FirstLteSecond(first, second string) bool {
	ret := compare(first, second)
	return ret == EQUALS || ret == LessThan
}

/**
 * 第一个版本号是否 小于 第二个版本号
 *
 * @param first  第一个版本号
 * @param second 第二个版本号
 * @return true 是， false - 不是
 */
func FirstLtSecond(first, second string) bool {
	ret := compare(first, second)
	return ret == LessThan
}

/**
 * 判断一个版本号是否在最大最小之间， 如果minVersion 为空，表示最小版本号没限制， 如果maxVersion 为空表示最大版本号没限制
 *
 * @param version    要检测的版本号
 * @param minVersion 最小版本号
 * @param maxVersion 最大版本号
 * @param includeMin 是否包含最小版本号
 * @param includeMax 是否包含最大版本号
 * @return true 是， false - 否
 */
func BetweenVersion(version, min, max string, includeMin, includeMax bool) bool {
	version, valid := resolveVersion(version)
	if !valid {
		return false
	}
	var minPassed bool
	if len(min) < 1 {
		minPassed = true
	} else {
		if includeMin {
			minPassed = FirstGteSecond(version, min)
		} else {
			minPassed = FirstGtSecond(version, min)
		}
	}
	if !minPassed {
		return false
	}

	var maxPassed bool
	if len(max) < 1 {
		maxPassed = true
	} else {
		if includeMax {
			maxPassed = FirstLteSecond(version, max)
		} else {
			maxPassed = FirstLtSecond(version, max)
		}
	}
	return maxPassed
}

/**
是否在范围内
*/
func InRange(version string, versionRange VersionRange, includeMin, includeMax bool) bool {
	version, valid := resolveVersion(version)
	if !valid {
		return false
	}
	if len(versionRange.Excludes) > 0 {
		for _, ver := range versionRange.Excludes {
			if ver == version {
				return false
			}
		}
	}
	return BetweenVersion(version, versionRange.Min, versionRange.Max, includeMin, includeMax)
}

func AcceptPlatformVersion(platform, clientVersion, versionLimitConf string) bool {
	if len(versionLimitConf) < 1 || len(clientVersion) < 1 {
		return true
	}
	accept := false
	versionLimit := VersionLimit{}
	err := json.Unmarshal([]byte(versionLimitConf), &versionLimit)
	if err == nil {
		if versionRange, ok := versionLimit[platform]; ok {
			if InRange(clientVersion, versionRange, true, true) {
				accept = true
			}
		} else {
			accept = true
		}
	}
	return accept
}
