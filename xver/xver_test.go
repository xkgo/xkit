package xver

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func TestCompare(t *testing.T) {
	assert.Equal(t, EQUALS, compare("1.0.0", "1.0.0"))
	assert.Equal(t, GreaterThan, compare("1.1.0", "1.0.0"))
	assert.Equal(t, LessThan, compare("1.1.0", "2.0.0"))
	assert.Equal(t, BothInvalid, compare("1.d1.0", "2.d0.0"))
	assert.Equal(t, Ver1Invalid, compare("1.1d.0", "2.0.0"))
	assert.Equal(t, Ver2Invalid, compare("1.1.0", "2.d0.0"))

	assert.Equal(t, true, BetweenVersion("1.0.5", "", "", true, true))
	assert.Equal(t, true, BetweenVersion("1.0.5", "", "1.0.8", true, true))
	assert.Equal(t, true, BetweenVersion("0.0.5", "", "0.0.8", true, true))
	assert.Equal(t, true, BetweenVersion("0.0.5", "0.0.1", "0.0.8", true, true))
	assert.Equal(t, false, BetweenVersion("0.0.1", "0.0.1", "0.0.8", false, false))

	assert.Equal(t, true, InRange("0.0.1", VersionRange{Min: "", Max: ""}, false, false))
	assert.Equal(t, true, InRange("0.0.1", VersionRange{Min: "", Max: "1.0.0"}, false, false))
	assert.Equal(t, false, InRange("0.0.1", VersionRange{Min: "", Max: "0.0.1"}, false, false))
	assert.Equal(t, false, InRange("1.0.0", VersionRange{Min: "11.0.0", Max: "1000.0.1"}, true, true))
	assert.Equal(t, false, InRange("0.0.1", VersionRange{Min: "", Max: "", Excludes: []string{"0.0.1"}}, false, false))

	iOSRange := VersionRange{
		Min:      "1.0.0",
		Max:      "1000.0.0",
		Excludes: nil,
	}

	androidRange := VersionRange{
		Min:      "1.0.0",
		Max:      "1000.0.0",
		Excludes: nil,
	}

	rangMap := VersionLimit{
		"iOS":     iOSRange,
		"Android": androidRange,
	}

	vJson, err := json.Marshal(rangMap)

	fmt.Println(vJson, err)
}

func TestVersionRange(t *testing.T) {

	vl := "{\"Android\":{\"min\":\"11.0.0\",\"max\":\"1000.0.0\",\"excludes\":\"\"},\"iOS\":{\"min\":\"1.0.0\",\"max\":\"1000.0.0\",\"excludes\":null}}"

	versionLimit := &VersionLimit{}
	_ = json.Unmarshal([]byte(vl), versionLimit)

	fmt.Println(versionLimit)

}

func TestUserVersionResolve(t *testing.T) {
	regex, _ := regexp.Compile("^(?i)[^\\d]*([\\d]+\\.[\\d]+.[\\d]+).*$")

	fmt.Println(regex.ReplaceAllString("markiOversea&2.0.1-689&adr&official", "$1"))
	fmt.Println(regex.ReplaceAllString("markiOversea&2.0.1&adr&official", "$1"))
	fmt.Println(regex.ReplaceAllString("2.0.1&adr&official", "$1"))
	fmt.Println(regex.ReplaceAllString("2.0.1-1&adr&official", "$1"))
	fmt.Println(regex.ReplaceAllString("2.0.1-1", "$1"))
	fmt.Println(regex.ReplaceAllString("2.0.1", "$1"))
}
