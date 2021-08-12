package xenv

import (
	"github.com/xkgo/xkit/xcontext"
	"github.com/xkgo/xkit/xlog"
	"regexp"
	"sync"
)

/**
基于 Map 实现的 env/PropertySource 接口
*/
type MapPropertySource struct {
	name       string    // 给这个命个名
	properties *sync.Map // 配置map
	/**
	配置key变更订阅列表
	*/
	propertyChangeListeners []*PropertyChangeListener
}

func NewMapPropertySource(name string, properties map[string]string) *MapPropertySource {
	source := &MapPropertySource{
		name:       name,
		properties: &sync.Map{},
	}

	if len(properties) > 0 {
		for key, value := range properties {
			source.properties.Store(key, value)
		}
	}

	return source
}

func (m *MapPropertySource) GetName() string {
	return m.name
}

func (m *MapPropertySource) GetProperty(key string) (value string, exists bool) {
	if nil == m.properties {
		m.properties = &sync.Map{}
		return "", false
	}
	val, exists := m.properties.Load(key)
	if !exists {
		return "", exists
	}
	return val.(string), true
}

func (m *MapPropertySource) GetPropertyWithDef(key string, def string) string {
	if value, exists := m.GetProperty(key); exists {
		return value
	}
	return def
}

func (m *MapPropertySource) Each(consumer func(key string, value string) (stop bool)) {
	if consumer == nil || nil == m.properties {
		return
	}
	m.properties.Range(func(key, value interface{}) bool {
		if consumer(key.(string), value.(string)) {
			return false
		}
		return true
	})
}

func (m *MapPropertySource) onKeyChangeEvent(event *KeyChangeEvent) {
	xlog.Info("["+m.name+"]配置发生了变更：key:["+event.Key+"], ov:["+event.Ov+"], nv:["+event.Nv+"], changeType:[", event.ChangeType+"]")
	// 执行监听器
	if len(m.propertyChangeListeners) > 0 {
		for _, listener := range m.propertyChangeListeners {
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
				xcontext.Run(func() {
					handler(event)
				}, func(r interface{}, hadPanic bool) {
					if hadPanic {
						xlog.Warn("配置源["+m.name+"]执行配置变更[", listener, "]发生panic： ", r)
					}
				})
			}
		}
	}
}

/**
设置
*/
func (m *MapPropertySource) Put(key string, value string) {
	changeType := PropertyAdd
	ov, exists := m.properties.Load(key)
	if exists {
		changeType = PropertyUpdate
	}
	sov := ""
	if nil != ov {
		sov = ov.(string)
	}
	event := &KeyChangeEvent{
		Key:        key,
		Ov:         sov,
		Nv:         value,
		ChangeType: changeType,
	}
	m.properties.Store(key, value)

	xcontext.RunByGoroutine(func() {
		m.onKeyChangeEvent(event)
	}, nil)
}

/**
设置
*/
func (m *MapPropertySource) PutAll(kvs map[string]string) {
	if len(kvs) < 1 {
		return
	}
	for key, value := range kvs {
		m.Put(key, value)
	}
}

/**
删除
*/
func (m *MapPropertySource) Remove(keys ...string) {
	if len(keys) < 1 {
		return
	}
	for _, key := range keys {
		ov, exists := m.properties.Load(key)
		if !exists {
			continue
		}
		sov := ""
		if nil != ov {
			sov = ov.(string)
		}
		event := &KeyChangeEvent{
			Key:        key,
			Ov:         sov,
			Nv:         "",
			ChangeType: PropertyDel,
		}
		// 删除 key
		m.properties.Delete(key)

		xcontext.RunByGoroutine(func() {
			m.onKeyChangeEvent(event)
		})
	}
}

func (m *MapPropertySource) Subscribe(keyPattern string, handler func(event *KeyChangeEvent)) {
	if m.propertyChangeListeners == nil {
		m.propertyChangeListeners = make([]*PropertyChangeListener, 0)
	}
	m.propertyChangeListeners = append(m.propertyChangeListeners, NewPropertyChangeListener(keyPattern, handler))
}
