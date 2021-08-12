package xenv

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/xkgo/xkit/xcontext"
	"github.com/xkgo/xkit/xjson"
	"github.com/xkgo/xkit/xlog"
	"testing"
	"time"
)

type FixedPublishEventMapPropertySource struct {
	MapPropertySource
}

func NewFixedPublishEventMapPropertySource(name string, properties map[string]string) *FixedPublishEventMapPropertySource {
	source := &FixedPublishEventMapPropertySource{
		MapPropertySource: *NewMapPropertySource(
			name,
			properties,
		),
	}

	source.init()

	return source
}

func (m *FixedPublishEventMapPropertySource) Subscribe(keyPattern string, handler func(event *KeyChangeEvent)) {
	go func() {
		looptimes := 0
		for {
			if looptimes >= 3 {
				break
			}
			event := &KeyChangeEvent{
				Key:        "my.var",
				Ov:         "",
				Nv:         "aaaaa",
				ChangeType: PropertyUpdate,
			}
			fmt.Println("发布事件")
			handler(event)
			looptimes++
			time.Sleep(time.Duration(2) * time.Second)
		}
	}()
}

func (m *FixedPublishEventMapPropertySource) init() {
}

func TestStandardEnvironment_New(t *testing.T) {

	additionalPropertySources := NewMutablePropertySources(
		NewMapPropertySource("test", map[string]string{
			"my.var": "additional",
		}),
		NewFixedPublishEventMapPropertySource("Test1", map[string]string{
			"my.var1": "additional1",
		}),
		NewFixedPublishEventMapPropertySource("Test2", map[string]string{
			"my.var2": "additional2",
		}),
	)

	env := New(ConfigDirs(map[Env]string{Dev: "./testdata"}), AdditionalPropertySources(additionalPropertySources))

	fmt.Println("----------------------------------------------------------------------------------------------------------------------------")

	v, _ := env.GetProperty("redis.server")
	fmt.Println("redis.server: ", v)

	v, _ = env.GetProperty("test.name")
	fmt.Println("test.name: ", v)

	v, _ = env.GetProperty("my.var")
	fmt.Println("my.var: ", v)

	fmt.Println("----------------------------------------------------------------------------------------------------------------------------")

}

type Addr struct {
	Country  string `ck:"country"`
	Province string `ck:"province"`
}

type UserInfo struct {
	Names    []string         `ck:"names"`
	Id       int              `ck:"id"`
	username string           `ck:"username"`
	Addr     *Addr            `ck:"addr" expand:"true"`
	addr2    *Addr            `ck:"addr2" expand:"true"`
	Addr3    *Addr            `ck:"addr3" expand:"true"`
	Addr4    Addr             `ck:"addr4" expand:"true"`
	Addrs    map[string]*Addr `ck:"addrs" expand:"true"`
	Addrs2   map[int]*Addr    `ck:"addrs2" expand:"true"`
	Addrs3   map[int]Addr     `ck:"addrs3" expand:"true"`
	Ints     []int            `ck:"ints"`
}

func (u UserInfo) String() string {
	return fmt.Sprintf("UserInfo{Id:%d, username:%s}", u.Id, u.username)
}

type ArrayField struct {
	Names []string `ck:"names"`
}

func TestStandardEnvironment_BindProperties(t *testing.T) {
	additionalPropertySources := NewMutablePropertySources(
		NewMapPropertySource("test", map[string]string{
			"user.id":       "1",
			"user.username": "Hello_${user.id}",
		}),
	)

	env := New(AdditionalPropertySources(additionalPropertySources))
	user := &UserInfo{}

	_, _ = env.BindProperties("user.", user, false)

	assert.Equal(t, 1, user.Id)
	assert.Equal(t, "Hello_1", user.username)
}

func TestStandardEnvironment_BindSubStructProperties(t *testing.T) {

	source := NewMapPropertySource("test", map[string]string{
		"user.id":                "1",
		"user.username":          "Hello_${user.id}",
		"user.addr.country":      "CN",
		"user.addr2.country":     "US",
		"user.addr3.country":     "ID",
		"user.addr4.country":     "CD",
		"user.addrs.cn.country":  "CN",
		"user.addrs.cn.province": "GZ",
		"user.addrs2.2.country":  "CN",
		"user.addrs2.2.province": "GZ",
		"user.names":             "a,b",
		"user.ints":              "[1,2]",
	})

	additionalPropertySources := NewMutablePropertySources(
		source,
	)

	env := New(AdditionalPropertySources(additionalPropertySources))
	user := &UserInfo{}

	_, _ = env.BindProperties("user.", user, false)

	assert.Equal(t, 1, user.Id)
	assert.Equal(t, "Hello_1", user.username)
	assert.Equal(t, "CN", user.Addr.Country)
	assert.Equal(t, "US", user.addr2.Country)
	assert.Equal(t, "ID", user.Addr3.Country)
	assert.Equal(t, "CD", user.Addr4.Country)

	// 更改线程
	xcontext.RunByGoroutine(func() {
		for {
			time.Sleep(time.Duration(3) * time.Second)

			//source.Put("user.addrs3.2.province", RandomUtils.RandomLetterString(2))
			//source.Put("user.addrs3.2.province", RandomUtils.RandomLetterString(2))
			source.Put("user.names", "v,s")
			//source.Put("user.addrs3.2.province", RandomUtils.RandomLetterString(2))

		}
	}, nil)
	// 打印线程
	xcontext.RunByGoroutine(func() {
		for {
			fmt.Printf("用户信息：%+v\n", xjson.ToJsonStringWithoutError(user))
			time.Sleep(time.Duration(2) * time.Second)
		}
	}, nil)

	time.Sleep(time.Duration(10) * time.Hour)
}

func TestStandardEnvironment_BindSubStructPropertiesArrayField(t *testing.T) {

	source := NewMapPropertySource("test", map[string]string{
		"user.id":                "1",
		"user.username":          "Hello_${user.id}",
		"user.addr.country":      "CN",
		"user.addr2.country":     "US",
		"user.addr3.country":     "ID",
		"user.addr4.country":     "CD",
		"user.addrs.cn.country":  "CN",
		"user.addrs.cn.province": "GZ",
		"user.addrs2.2.country":  "CN",
		"user.addrs2.2.province": "GZ",
		"user.names":             "a,b",
		"user.ints":              "[1,2]",
	})

	additionalPropertySources := NewMutablePropertySources(
		source,
	)

	env := New(AdditionalPropertySources(additionalPropertySources))
	user := &ArrayField{}

	_, _ = env.BindProperties("user.", user, true)

	// 更改线程
	xcontext.RunByGoroutine(func() {
		for {
			time.Sleep(time.Duration(3) * time.Second)

			//source.Put("user.addrs3.2.province", RandomUtils.RandomLetterString(2))
			//source.Put("user.addrs3.2.province", RandomUtils.RandomLetterString(2))
			source.Put("user.names", "v,s")
			//source.Put("user.addrs3.2.province", RandomUtils.RandomLetterString(2))

		}
	}, nil)
	// 打印线程
	xcontext.RunByGoroutine(func() {
		for {
			fmt.Printf("用户信息：%+v\n", xjson.ToJsonStringWithoutError(user))
			time.Sleep(time.Duration(2) * time.Second)
		}
	}, nil)

	time.Sleep(time.Duration(10) * time.Hour)
}

func TestStandardEnvironment_BindxlogProperties(t *testing.T) {
	additionalPropertySources := NewMutablePropertySources(
		NewMapPropertySource("test", map[string]string{
			"xlog.level": "INFO",
		}),
	)

	env := New(AdditionalPropertySources(additionalPropertySources))
	props := &xlog.Properties{}

	_, _ = env.BindProperties("xlog.", props, false)

	fmt.Println(props)
}

func TestStandardEnvironment_MultiInclude(t *testing.T) {
	env := New(
		ConfigDirs(map[Env]string{Dev: "./testdata"}),
		IgnoreUnresolvableNestedPlaceholders(true),
	)

	fmt.Println(env.activeProfiles)
	fmt.Println(env.GetProperty("test.name"))

}

func TestStandardEnvironment_BindPropertiesListen(t *testing.T) {
	env := New(
		ConfigDirs(map[Env]string{Dev: "./testdata"}),
		IgnoreUnresolvableNestedPlaceholders(true),
	)

	type Config struct {
		PageSize int64 `sk:"page-size"`
	}

	config := &Config{}

	bean, err := env.BindProperties("config.", config, false)
	assert.Nil(t, err)

	config = bean.(*Config)

	assert.Equal(t, int64(0), config.PageSize)

}
