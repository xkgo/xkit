package xenv

import (
	"encoding/json"
	"errors"
	"github.com/xkgo/xkit/xfile"
	"github.com/xkgo/xkit/xjson"
	"github.com/xkgo/xkit/xlog"
	"github.com/xkgo/xkit/xreflect"
	"github.com/xkgo/xkit/xstr"
	"os"
	"reflect"
	"regexp"
	"strings"
)

const (
	RunInfoEnvironmentPropertySourceName            = "runInfoEnvironment"
	DefaultApplicationEnvironmentPropertySourceName = "defaultApplicationEnvironment"
	RunInfoSetKey                                   = "runInfo.set"
	RunInfoEnvKey                                   = "runInfo.env"
	RunInfoWorkDirKey                               = "runInfo.workDir"
)

func init() {
	xlog.InitLogger(&xlog.Properties{})
}

/**
标准环境实现, 实现接口 Environment
*/
type StandardEnvironment struct {
	// 选项
	options *Options

	/**
	当前运行环境信息，部署环境&配置信息
	*/
	runInfo *RunInfo

	/**
	配置文件所在目录
	*/
	configDir string

	/**
	激活的 profile
	*/
	activeProfiles []string

	/**
	是否忽略无法处理的占位符，如果忽略则不处理，不忽略的话，那么遇到不能解析的占位符直接 panic
	*/
	ignoreUnresolvableNestedPlaceholders bool

	/**
	配置来源，默认是从 profileDirs 中进行检索实例化的，当然，也是可以定义外部配置中心的，比如 apollo、nacos、consul、zookeeper等等
	默认配置处理逻辑：
	> 将命令行参数 作为优先级最高的 propertySource， --------- 之后每次向 propertySources 添加元素，都要重新进行日志配置，这样子才能每次应用最新配置
	> 添加系统环境变量
	> 自动解析当前运行环境相关属性： 环境(dev,test,fat,prod), set(分组：可能以全球大区、机房等来区分部署集群等等，将这个抽象即可)，将环境相关组成 propertySource ，然后添加进去 propertySources
	> 读取默认配置文件 application.properties|yml|toml, 然后添加到 propertySources 的 命令行之后，从 propertySources 中读取 xenv-profile-include，作为 activeProfiles
	> 获取 profileDirs 下的所有配置文件，按照profile 分组，然后按照顺序依次加载配置文件，最后按顺序添加到 propertySources 的默认配置文件之后
	> profileDirs 都加载完成后， 将 additionalPropertySources 添加到 propertySources 之后
	> 添加系统环境变量到 propertySources 最后面
	*/
	propertySources *MutablePropertySources

	/**
	配置解析器，读取配置、处理占位符
	*/
	propertyResolver PropertyResolver

	/**
	配置key变更订阅列表
	*/
	propertyChangeListeners []*PropertyChangeListener

	/**
	Beans
	*/
	bindBeans map[reflect.Type]interface{}
}

func (s *StandardEnvironment) IsDev() bool {
	return s.runInfo.IsDev()
}

func (s *StandardEnvironment) IsTest() bool {
	return s.runInfo.IsTest()
}

func (s *StandardEnvironment) IsFat() bool {
	return s.runInfo.IsFat()
}

func (s *StandardEnvironment) IsProd() bool {
	return s.runInfo.IsProd()
}

func (s *StandardEnvironment) GetRunInfo() *RunInfo {
	return s.runInfo
}

func (s *StandardEnvironment) GetSet() string {
	return s.runInfo.Set
}

/**
新建环境
*/
func New(options ...Option) *StandardEnvironment {
	env := &StandardEnvironment{
		options:                 &Options{},
		propertyChangeListeners: make([]*PropertyChangeListener, 0),
		bindBeans:               make(map[reflect.Type]interface{}),
	}

	// 设置选项
	if options != nil && len(options) > 0 {
		for _, option := range options {
			option(env)
		}
	}

	env.propertySources = NewMutablePropertySources()
	// 订阅并更新日志信息
	env.subscribeAndOverrideXlogProperties()

	// 订阅数据源变更，然后循环检查 xenv.profile.include, 然后导入数据源
	env.subscribeAndAddIncludeProfiles()

	// 添加运行时信息
	addRunInfo(env)

	// 计算 configDir
	env.configDir = env.resolveConfigDir()
	xlog.Infof("配置文件目录为：%v", env.configDir)

	// 追加默认配置 application.properties|yaml|yml
	env.addDefaultApplicationPropertySource()

	// 将 additionalPropertySources 添加到 propertySources 之后
	additionalPropertySources := env.options.additionalPropertySources
	if nil != additionalPropertySources && len(additionalPropertySources.propertySourceList) > 0 {
		additionalPropertySources.Each(func(index int, source PropertySource) (stop bool) {
			if !env.propertySources.Contains(source.GetName()) {
				env.propertySources.AddLast(source)
			}
			return false
		})
	}

	// 刷新、初始化
	env.refresh()

	return env
}

func (s *StandardEnvironment) subscribeAndOverrideXlogProperties() {
	var logProp *xlog.Properties = nil
	s.propertySources.Subscribe(func(self *MutablePropertySources, changeType PropertySourcesChangeType, source PropertySource) {
		prop := &xlog.Properties{}
		_, _ = s.doBindProperties("xlog.", prop, false)
		if s.runInfo != nil && s.runInfo.IsDev() {
			prop.ConsoleLog = true // 开发环境下强制开启 console log
		}
		if prop.Equals(logProp) {
			return
		}
		logProp = prop
		xlog.Info("property source changed, will reset xlog: ", xjson.ToJsonStringWithoutError(prop))
		xlog.InitLogger(prop)
	})
}

/**
订阅数据源变更，然后循环检查 xenv.profile.include, 然后导入数据源
*/
func (s *StandardEnvironment) subscribeAndAddIncludeProfiles() {
	s.propertySources.Subscribe(func(self *MutablePropertySources, changeType PropertySourcesChangeType, source PropertySource) {
		if changeType != PropertySourcesChangeType_Add && changeType != PropertySourcesChangeType_Update {
			return
		}

		sInclude, ok := source.GetProperty("xenv.profile.include")
		if !ok || len(sInclude) < 1 {
			return
		}
		sInclude = xstr.Trim(s.ResolvePlaceholders(sInclude))
		profiles := xstr.SplitByRegex(sInclude, "[,，;；\\s]+")

		size := len(profiles)
		if size < 1 {
			return
		}

		activeProfiles := make([]string, 0)
		for i := size - 1; i >= 0; i-- {
			profile := profiles[i]
			// 搜索配置文件夹下的配置文件，然后加载
			xfile.ListDirFiles(s.configDir, func(pdir string, fileInfo os.FileInfo) bool {
				if fileInfo.IsDir() {
					return false
				}
				if !strings.HasPrefix(fileInfo.Name(), "application-"+profile+".") {
					return false
				}

				sName := fileInfo.Name()

				if self.Contains(sName) {
					return false
				}

				configFile := pdir + "/" + fileInfo.Name()
				kvs, err := xfile.ReadAsMap(configFile)
				if nil != err {
					xlog.Warnf("读取配置文件：%v 异常，err:%v", configFile, err)
					return false
				}

				activeProfiles = append(activeProfiles, profile)
				// 添加
				self.AddFirst(NewMapPropertySource(sName, kvs))
				return true
			}, 1)
		}

		if len(activeProfiles) > 0 {
			if s.activeProfiles == nil {
				s.activeProfiles = make([]string, 0)
			}
			sSize := len(activeProfiles)
			for i := sSize - 1; i >= 0; i-- {
				s.activeProfiles = append(s.activeProfiles, activeProfiles[i])
			}
		}
	})
}

/**
追加默认配置 application.properties|yaml|yml，目前仅支持这两种模式
*/
func (s *StandardEnvironment) addDefaultApplicationPropertySource() {
	r, _ := regexp.Compile("(?i)(app|application)\\.[^\\\\.]+$")

	properties := make(map[string]string)
	// 遍历配置目录下的文
	xfile.ListDirFiles(s.configDir, func(pdir string, fileInfo os.FileInfo) bool {
		if fileInfo.IsDir() {
			return false
		}
		filename := strings.ToLower(fileInfo.Name())
		if !r.MatchString(filename) {
			return false
		}

		kvs, err := xfile.ReadAsMap(pdir + "/" + fileInfo.Name())
		if nil == err && len(kvs) > 0 {
			for k, v := range kvs {
				properties[k] = v
			}
		}
		return true
	}, 1)

	// 添加一个空的默认配置来源
	s.propertySources.AddFirst(NewMapPropertySource(DefaultApplicationEnvironmentPropertySourceName, properties))
}

/**
添加激活的 属性来源，逻辑：
1.
*/
func (s *StandardEnvironment) addActiveProfilePropertySources() {

}

func addRunInfo(env *StandardEnvironment) {
	if env.options.customRunInfo != nil {
		env.runInfo = env.options.customRunInfo
	}

	// 将命令行参数作为最高优先级的属性来源
	env.propertySources.AddFirst(NewCommandLinePropertySource(env.options.appendCommandLine))
	// 添加系统环境变量
	env.propertySources.AddLast(NewSystemEnvironmentPropertySource())

	if env.runInfo == nil {
		// 添加部署信息到配置来源
		env.runInfo = DetectEnvInfo()
		deployProperties := env.runInfo.Properties
		if deployProperties == nil {
			deployProperties = make(map[string]string)
		}
	}
	if env.runInfo.Properties == nil {
		env.runInfo.Properties = make(map[string]string)
	}
	env.runInfo.Properties[RunInfoEnvKey] = string(env.runInfo.Env)
	env.runInfo.Properties[RunInfoSetKey] = env.runInfo.Set
	env.runInfo.Properties[RunInfoWorkDirKey] = env.runInfo.WorkDir
	env.propertySources.AddAfter(CommandLineEnvironmentPropertySourceName, NewMapPropertySource(RunInfoEnvironmentPropertySourceName, env.runInfo.Properties))
}

/**
默认就是 ./config 目录， 如果 ./config 不存在，那么就直接是 工作目录
*/
func (s *StandardEnvironment) resolveConfigDir() string {
	configDir := s.options.configDirs[s.runInfo.Env]
	if len(configDir) < 1 {
		configDir = "./config"
	}

	if xfile.IsDirExists(configDir) {
		s.configDir = configDir
		return s.configDir
	}

	// 直接就是工作目录
	s.configDir = s.runInfo.WorkDir
	return s.configDir
}

func (s *StandardEnvironment) InitPropertyResolver() {
	if s.propertySources == nil || s.propertyResolver == nil {
		s.propertyResolver = &PropertySourcesPropertyResolver{
			propertySources:                      s.propertySources,
			ignoreUnresolvableNestedPlaceholders: s.ignoreUnresolvableNestedPlaceholders,
		}
	}
}

func (s *StandardEnvironment) ContainsProperty(key string) bool {
	s.InitPropertyResolver()
	return s.propertyResolver.ContainsProperty(key)
}

func (s *StandardEnvironment) GetProperty(key string) (value string, exists bool) {
	s.InitPropertyResolver()
	return s.propertyResolver.GetProperty(key)
}

func (s *StandardEnvironment) GetPropertyWithDef(key string, def string) string {
	s.InitPropertyResolver()
	return s.propertyResolver.GetPropertyWithDef(key, def)
}

func (s *StandardEnvironment) GetRequiredProperty(key string) string {
	s.InitPropertyResolver()
	return s.propertyResolver.GetRequiredProperty(key)
}

func (s *StandardEnvironment) ResolvePlaceholders(text string) string {
	s.InitPropertyResolver()
	return s.propertyResolver.ResolvePlaceholders(text)
}

func (s *StandardEnvironment) ResolveRequiredPlaceholders(text string) string {
	s.InitPropertyResolver()
	return s.propertyResolver.ResolveRequiredPlaceholders(text)
}

func (s *StandardEnvironment) GetActiveProfiles() []string {
	return s.activeProfiles
}

func (s *StandardEnvironment) GetPropertySources() *MutablePropertySources {
	if nil == s.propertySources {
		s.propertySources = &MutablePropertySources{
			propertySourceList: make([]PropertySource, 0),
		}
	}
	return s.propertySources
}

func (s *StandardEnvironment) Merge(parent Environment) {
	if parent == nil {
		return
	}

	parentSources := parent.GetPropertySources()
	if parentSources != nil {
		if s.propertySources == nil {
			s.GetPropertySources()
		}
		parentSources.Each(func(index int, source PropertySource) (stop bool) {
			if !s.propertySources.Contains(source.GetName()) {
				s.propertySources.AddLast(source)
			}
			return false
		})
	}
	// 添加激活的配置文件
	parentActiveProfiles := parent.GetActiveProfiles()
	if len(parentActiveProfiles) > 0 {
		if s.activeProfiles == nil {
			s.activeProfiles = parentActiveProfiles
		} else {
			s.activeProfiles = append(s.activeProfiles, parentActiveProfiles...)
		}
	}
}

func (s *StandardEnvironment) Subscribe(keyPattern string, handler func(event *KeyChangeEvent)) {
	if s.propertyChangeListeners == nil {
		s.propertyChangeListeners = make([]*PropertyChangeListener, 0)
	}
	s.propertyChangeListeners = append(s.propertyChangeListeners, NewPropertyChangeListener(keyPattern, handler))
}

func (s *StandardEnvironment) refresh() {
	s.initPropertySourceListen()
}

func (s *StandardEnvironment) initPropertySourceListen() {
	// 执行所有配置来源的监听
	s.propertySources.Each(func(index int, source PropertySource) (stop bool) {
		source.Subscribe("*", func() func(event *KeyChangeEvent) {
			return func(event *KeyChangeEvent) {
				xlog.Info("收到配置来源["+source.GetName()+"]的配置变更事件：", event)
				s.onKeyChangeEvent(source, event)
			}
		}())
		return false
	})
}

/**
Key 变更处理
*/
func (s *StandardEnvironment) onKeyChangeEvent(source PropertySource, event *KeyChangeEvent) {
	// 执行监听器
	if len(s.propertyChangeListeners) > 0 {
		for _, listener := range s.propertyChangeListeners {
			keyPattern := listener.KeyPattern
			handler := listener.Handler
			if handler == nil {
				continue
			}
			if keyPattern == "" || keyPattern == "*" || keyPattern == event.Key {
				handler(event)
				continue
			}
			regex, err := regexp.Compile(keyPattern)
			if err == nil && regex.MatchString(event.Key) {
				handler(event)
			}
		}
	}
}

func (s *StandardEnvironment) BindProperties(keyPrefix string, cfgPtr interface{}, changedListen bool) (beanPtr interface{}, err error) {
	return s.doBindProperties(keyPrefix, cfgPtr, changedListen)
}

func (s *StandardEnvironment) doBindProperties(keyPrefix string, cfgPtr interface{}, listen bool) (beanPtr interface{}, err error) {
	// 反射解析所有属性
	t := reflect.TypeOf(cfgPtr)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	} else {
		return nil, errors.New("注册配置Bean异常，必须是指针类型, 当前注册类型为：[" + t.Name() + "], keyPrefix:" + keyPrefix)
	}

	v := reflect.ValueOf(cfgPtr)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		tfield := t.Field(i)
		vfield := v.Field(i)

		fieldName := tfield.Name

		// 默认是首字母小写
		configKey := keyPrefix + xstr.FirstLetterLower(fieldName)
		subKey := tfield.Tag.Get("ck")
		if len(subKey) < 1 {
			subKey = tfield.Tag.Get("sk")
		}
		if len(subKey) > 0 {
			configKey = keyPrefix + subKey
		}

		// 是否需要展开
		expand := "true" == tfield.Tag.Get("expand")

		if expand {
			// Map
			if tfield.Type.Kind() == reflect.Map || (tfield.Type.Kind() == reflect.Ptr && tfield.Type.Elem().Kind() == reflect.Map) {
				_, err := s.doBindSubMapField(t, keyPrefix, tfield, vfield, subKey, true)
				if err != nil {
					return nil, err
				}
				continue
			}
			// 判断是不是结构体，如果是结构体并且需要继续展开，如果是的话，创建一个新的对象，进行绑定
			if tfield.Type.Kind() == reflect.Struct || (tfield.Type.Kind() == reflect.Ptr && tfield.Type.Elem().Kind() == reflect.Struct) {
				// 循环依赖了，不允许嵌套
				if t == tfield.Type || (tfield.Type.Kind() == reflect.Ptr && t == tfield.Type.Elem()) {
					panic("[" + t.Name() + "." + tfield.Name + "] 属性是expand 类型的，不允许嵌套，不能是[" + t.Name() + "]类型")
				}
				_, err := s.doBindSubStructField(keyPrefix, tfield, vfield, subKey, listen)
				if err != nil {
					return nil, err
				}
				continue
			}

		}

		// 初始值
		initVal := tfield.Tag.Get("def")

		// 获取配置的值
		value, exists := s.GetProperty(configKey)
		if !exists {
			value = s.ResolvePlaceholders(initVal)
		}
		// 反射进行配置回写
		s.applyBeanPropertyValue(t, tfield, vfield, initVal, value, PropertyUpdate)

		if listen {
			// 注册监听器, 占位符问题，每次变更的话，都需要重新检查占位符，当占位符变化这个也要变化
			s.Subscribe(strings.Replace(configKey, ".", "\\.", -1)+".*", func() func(event *KeyChangeEvent) {
				return func(event *KeyChangeEvent) {
					nv, _ := s.GetProperty(event.Key)
					s.applyBeanPropertyValue(t, tfield, vfield, initVal, nv, event.ChangeType)
				}
			}())
		}
	}
	jsonText, err := json.Marshal(cfgPtr)
	if err != nil {
		return nil, err
	}
	xlog.Info("绑定配置Bean["+t.Name()+"] => ", string(jsonText))
	return cfgPtr, nil
}

func (s *StandardEnvironment) doBindSubStructField(keyPrefix string, tfield reflect.StructField, vfield reflect.Value, subKey string, changeListen bool) (interface{}, error) {
	if vfield.Type().Kind() == reflect.Ptr {
		if vfield.IsNil() {
			nValue := reflect.New(vfield.Type().Elem())
			_, err := s.doBindProperties(keyPrefix+subKey+".", nValue.Interface(), true)
			if nil != err {
				return nil, err
			}
			err = xreflect.SetFieldValueByField(tfield, vfield, nValue)
			if nil != err {
				return nil, err
			}
		} else {
			_, err := s.doBindProperties(keyPrefix+subKey+".", vfield.Interface(), changeListen)
			if nil != err {
				return nil, err
			}
		}
	} else {
		_, err := s.doBindProperties(keyPrefix+subKey+".", vfield.Addr().Interface(), changeListen)
		if nil != err {
			return nil, err
		}
	}
	return nil, nil
}

/**
获取属性对象
@param typeTemplate 类型模板，可以提供属性类型的指针类型，也可以直接提供 reflect.Type 类型
*/
func (s *StandardEnvironment) GetProperties(typeTemplate interface{}) (beanPtr interface{}) {
	ptype, ok := typeTemplate.(reflect.Type)
	if !ok {
		ptype = reflect.TypeOf(typeTemplate)
	}
	if ptype.Kind() == reflect.Ptr {
		ptype = ptype.Elem()
	}

	if bean, ok := s.bindBeans[ptype]; ok {
		return bean
	}
	return nil
}

func (s *StandardEnvironment) applyBeanPropertyValue(beanType reflect.Type, tfield reflect.StructField, vfield reflect.Value, initVal string, value string, changeType KeyChangeType) {
	if PropertyDel == changeType {
		// 删除，设置回原来的初始值
		value = initVal
	}

	var cerr error

	defer func() {
		if r := recover(); r != nil {
			xlog.Error("配置转换异常：panic,Property:["+beanType.Name()+"."+tfield.Name+":"+tfield.Type.Name()+"], newVal:["+value+"]", r)
		} else {
			if cerr != nil {
				xlog.Error("配置转换失败,Property:["+beanType.Name()+"."+tfield.Name+":"+tfield.Type.Name()+"], newVal:["+value+"]", cerr)
			}
		}
	}()

	cerr = xreflect.SetFieldValueByField(tfield, vfield, value)
}

func (s *StandardEnvironment) Properties() map[string]string {
	properties := make(map[string]string)
	s.propertySources.Each(func(index int, source PropertySource) (pstop bool) {
		source.Each(func(key, value string) (stop bool) {
			properties[key] = value
			return false
		})
		return false
	})
	return properties
}

/**
遍历所有的配置项&值, consumer 处理过程中如果返回 stop=true则停止遍历
*/
func (s *StandardEnvironment) EachProperty(consumer func(key, value string) (stop bool)) {
	properties := s.Properties()
	if properties == nil || len(properties) < 1 {
		return
	}

	for k, v := range properties {
		if consumer(k, v) {
			return
		}
	}
}

func (s *StandardEnvironment) doBindSubMapField(t reflect.Type, keyPrefix string, tfield reflect.StructField, vfield reflect.Value, subKey string, listen bool) (interface{}, error) {
	// MAP 类型， 要求key必须是 int 或者 string 类型
	keyTypeName := tfield.Type.Key().Name()
	if !strings.HasPrefix(keyTypeName, "int") && keyTypeName != "string" {
		panic("[" + t.Name() + "." + tfield.Name + "] map 的 key 必须是int|string类型")
	}
	configKey := keyPrefix + subKey + "."

	// key 类型
	kType := tfield.Type.Key()
	// 元素类型
	vType := tfield.Type.Elem()

	keys := make(map[string]bool)
	s.EachProperty(func(key, value string) (stop bool) {
		if strings.HasPrefix(key, configKey) { // 前缀
			key = strings.Replace(key, configKey, "", 1)
			index1 := strings.Index(key, ".")
			fieldKey := key[0:index1]
			keys[fieldKey] = true
		}
		return false
	})

	nMap := reflect.MakeMap(tfield.Type)
	// 构造map
	for fieldKey, _ := range keys {
		kValue, err := xreflect.ConvertTo(fieldKey, kType)
		if nil != err {
			panic("Map属性[" + fieldKey + "]无法转换成[" + kType.Name() + "]")
		}

		if vType.Kind() == reflect.Ptr {
			vValue := reflect.New(vType.Elem())
			// 注入
			_, err = s.doBindProperties(configKey+fieldKey+".", vValue.Interface(), false)
			if nil != err {
				panic("Map属性处理失败, keyPrefix: " + configKey + fieldKey + ".")
				return nil, err
			}
			nMap.SetMapIndex(kValue, vValue)
		} else {
			vValue := reflect.New(vType)
			// 注入
			_, err = s.doBindProperties(configKey+fieldKey+".", vValue.Interface(), false)
			if nil != err {
				panic("Map属性处理失败, keyPrefix: " + configKey + fieldKey + ".")
				return nil, err
			}
			nMap.SetMapIndex(kValue, vValue.Elem())
		}
	}

	vfield.Set(nMap)

	if listen {
		s.doListenMapField(configKey, t, keyPrefix, tfield, vfield, subKey)
	}

	return vfield.Interface(), nil
}

func (s *StandardEnvironment) doListenMapField(configKey string, t reflect.Type, keyPrefix string, tfield reflect.StructField, vfield reflect.Value, subKey string) {
	// 注册监听器, 占位符问题，每次变更的话，都需要重新检查占位符，当占位符变化这个也要变化
	s.Subscribe(strings.Replace(configKey, ".", "\\.", -1)+".*", func() func(event *KeyChangeEvent) {
		return func(event *KeyChangeEvent) {
			_, _ = s.doBindSubMapField(t, keyPrefix, tfield, vfield, subKey, false)
		}
	}())
}
