package xenv

import (
	"errors"
	"github.com/xkgo/xkit/xlog"
)

type PropertySourcesChangeType string

const (
	PropertySourcesChangeType_Add    = "ADD"
	PropertySourcesChangeType_Update = "UPDATE"
)

type PropertySources interface {

	/**
	是否包含指定名称的配置来源
	*/
	Contains(name string) bool

	/**
	获取指定名称的配置来源
	@return source 如果存在则返回
	@return exists 是否存在
	*/
	Get(name string) (source PropertySource, exists bool)

	/**
	遍历配置来源, index 从 0开始，数字越小，优先级越低, consumer 变量过程中，如果 stop 返回true则停止遍历
	*/
	Each(consumer func(index int, source PropertySource) (stop bool))

	/**
	反向遍历配置来源, index 从 长度 开始，数字越小，优先级越低, consumer 变量过程中，如果 stop 返回true则停止遍历
	*/
	EachRevert(consumer func(index int, source PropertySource) (stop bool))
}

/**
可变配置来源
*/
type MutablePropertySources struct {
	/**
	配置来源列表，左边的优先生效，比如同一个key在 第一第二个元素上面都存在，那么则会优先使用第一个元素上面的值，
	不管第一个是否为空字符串都要以第一个元素为准
	*/
	propertySourceList []PropertySource // 配置来源列表

	// 监听器
	listeners []func(self *MutablePropertySources, changeType PropertySourcesChangeType, source PropertySource)
}

func NewMutablePropertySources(propertySourceList ...PropertySource) *MutablePropertySources {
	if len(propertySourceList) < 1 {
		propertySourceList = make([]PropertySource, 0)
	}
	return &MutablePropertySources{propertySourceList: propertySourceList}
}

func (s *MutablePropertySources) Subscribe(listener func(self *MutablePropertySources, changeType PropertySourcesChangeType, source PropertySource)) {
	if s.listeners == nil {
		s.listeners = make([]func(self *MutablePropertySources, changeType PropertySourcesChangeType, source PropertySource), 0)
	}
	s.listeners = append(s.listeners, listener)
}

func (s *MutablePropertySources) onPropertySourceChanged(changeType PropertySourcesChangeType, source PropertySource) {
	if len(s.listeners) > 0 {
		for _, listener := range s.listeners {
			listener(s, changeType, source)
		}
	}
}

func (s *MutablePropertySources) Contains(name string) bool {
	_, err := s.assertPresentAndGetIndex(name)
	if err != nil {
		return false
	}
	return true
}

func (s *MutablePropertySources) Get(name string) (source PropertySource, exists bool) {
	index, err := s.assertPresentAndGetIndex(name)
	if err != nil {
		return nil, false
	}
	return s.propertySourceList[index], true
}

func (s *MutablePropertySources) Each(consumer func(index int, source PropertySource) (stop bool)) {
	list := s.getPropertySourceList()
	for idx, item := range list {
		if consumer(idx, item) {
			return
		}
	}
}

func (s *MutablePropertySources) EachRevert(consumer func(index int, source PropertySource) (stop bool)) {
	list := s.getPropertySourceList()
	size := len(list)
	for i := size - 1; i >= 0; i-- {
		if consumer(i, list[i]) {
			return
		}
	}
}

/**
添加到第一个， 最优先生效
*/
func (s *MutablePropertySources) AddFirst(propertySource PropertySource) {
	if xlog.IsDebugEnabled() {
		xlog.Debug("Adding PropertySource '" + propertySource.GetName() + "' with highest search precedence")
	}
	// 如果已经存在，那么删除
	s.removeIfPresent(propertySource.GetName())
	// 添加到第一个元素
	list := s.getPropertySourceList()
	newList := make([]PropertySource, 0)
	newList = append(newList, propertySource)
	newList = append(newList, list...)
	s.propertySourceList = newList

	s.onPropertySourceChanged(PropertySourcesChangeType_Add, propertySource)
}

/**
加到最后面，最后生效
*/
func (s *MutablePropertySources) AddLast(propertySource PropertySource) {
	if xlog.IsDebugEnabled() {
		xlog.Debug("Adding PropertySource '" + propertySource.GetName() + "' with lowest search precedence")
	}
	s.removeIfPresent(propertySource.GetName())
	// 添加到最后
	list := s.getPropertySourceList()
	list = append(list, propertySource)
	s.propertySourceList = list

	s.onPropertySourceChanged(PropertySourcesChangeType_Add, propertySource)
}

/**
添加到指定名称之前, 如果指定名称不存在则返回异常
*/
func (s *MutablePropertySources) AddBefore(relativePropertySourceName string, propertySource PropertySource) error {
	if xlog.IsDebugEnabled() {
		xlog.Debug("Adding PropertySource '" + propertySource.GetName() +
			"' with search precedence immediately higher than '" + relativePropertySourceName + "'")
	}

	if relativePropertySourceName == propertySource.GetName() {
		return errors.New("PropertySource named '" + relativePropertySourceName + "' cannot be added relative to itself")
	}
	s.removeIfPresent(propertySource.GetName())

	list := s.getPropertySourceList()
	// 检查要插入到之前的那个配置来源是否存在
	index, err := s.assertPresentAndGetIndex(relativePropertySourceName)
	if err != nil {
		return err
	}

	// 添加到之前
	leftList := list[0:index]
	rightList := list[index:]

	newList := make([]PropertySource, 0)
	if len(leftList) > 0 {
		newList = append(newList, leftList...)
	}
	newList = append(newList, propertySource)
	if len(rightList) > 0 {
		newList = append(newList, rightList...)
	}
	s.propertySourceList = newList
	s.onPropertySourceChanged(PropertySourcesChangeType_Add, propertySource)
	return nil
}

func (s *MutablePropertySources) AddAfter(relativePropertySourceName string, propertySource PropertySource) error {
	if xlog.IsDebugEnabled() {
		xlog.Debug("Adding PropertySource '" + propertySource.GetName() +
			"' with search precedence immediately lower than '" + relativePropertySourceName + "'")
	}
	if relativePropertySourceName == propertySource.GetName() {
		return errors.New("PropertySource named '" + relativePropertySourceName + "' cannot be added relative to itself")
	}

	s.removeIfPresent(propertySource.GetName())

	list := s.getPropertySourceList()
	// 检查要插入到之前的那个配置来源是否存在
	index, err := s.assertPresentAndGetIndex(relativePropertySourceName)
	if err != nil {
		return err
	}
	if index == len(list)-1 {
		// 刚好最后一个
		s.propertySourceList = append(list, propertySource)
		s.onPropertySourceChanged(PropertySourcesChangeType_Add, propertySource)
		return nil
	}

	// 添加到之后
	leftList := list[0 : index+1]
	rightList := list[index+1:]

	newList := make([]PropertySource, 0)
	if len(leftList) > 0 {
		newList = append(newList, leftList...)
	}
	newList = append(newList, propertySource)
	if len(rightList) > 0 {
		newList = append(newList, rightList...)
	}
	s.propertySourceList = newList

	s.onPropertySourceChanged(PropertySourcesChangeType_Add, propertySource)
	return nil
}

func (s *MutablePropertySources) Replace(name string, propertySource PropertySource) error {
	if xlog.IsDebugEnabled() {
		xlog.Debug("Replacing PropertySource '" + name + "' with '" + propertySource.GetName() + "'")
	}

	index, err := s.assertPresentAndGetIndex(name)
	if err != nil {
		return err
	}
	// 直接替换
	s.propertySourceList[index] = propertySource

	s.onPropertySourceChanged(PropertySourcesChangeType_Update, propertySource)
	return nil
}

func (s *MutablePropertySources) Size() int {
	return len(s.getPropertySourceList())
}

func (s *MutablePropertySources) getPropertySourceList() []PropertySource {
	if s.propertySourceList == nil {
		s.propertySourceList = make([]PropertySource, 0)
	}
	return s.propertySourceList
}

func (s *MutablePropertySources) removeIfPresent(name string) PropertySource {
	list := s.getPropertySourceList()
	if len(list) < 1 {
		return nil
	}
	if len(list) == 1 {
		if list[0].GetName() == name {
			s.propertySourceList = make([]PropertySource, 0)
			return list[0]
		}
	} else {
		var deleted PropertySource = nil
		newList := make([]PropertySource, 0)
		for _, item := range list {
			if item.GetName() != name {
				newList = append(newList, item)
			} else{
				deleted = item
			}
		}
		s.propertySourceList = newList
		return deleted
	}
	return nil
}

func (s *MutablePropertySources) assertPresentAndGetIndex(name string) (index int, err error) {
	list := s.getPropertySourceList()
	for index, item := range list {
		if item.GetName() == name {
			return index, nil
		}
	}
	return -1, errors.New("PropertySource named '" + name + "' does not exist")
}
