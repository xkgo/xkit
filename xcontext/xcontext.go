package xcontext

import (
	"context"
	"fmt"
	"github.com/xkgo/xkit/xrand"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

/**
获取当前 Goroutine ID, 有不少方案，可以参考：https://blog.csdn.net/weiyuefei/article/details/77500653

Golang的开发者故意不提供 gid 的方法，避免开发者滥用Goroutine Id实现Goroutine Local Storage(类似java的Thread Local Storage)，
因为Goroutine Local Storage很难进行垃圾回收。因此尽管Go1.4之前暴露出了相应的方法，现在已经把它隐藏了。

*/
func GetGoroutineId() int {
	defer func() {
		if r := recover(); r != nil {
			_ = fmt.Errorf("获取Goroutine ID 发生 panic recover:panic info:%v", r)
		}
	}()
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		return -1
	}
	return id
}

/**
协程上下文
*/
type GContext struct {
	Gid     int    // 协程ID
	TraceId string // traceId
	Context context.Context
	Data    sync.Map // 数据
}

func newGContext(gid int, traceId string, context context.Context, data map[interface{}]interface{}) *GContext {
	if len(traceId) < 1 {
		traceId = xrand.RandomLetterAndNumberString(10)
	}
	ctx := &GContext{
		Gid:     gid,
		TraceId: traceId,
		Context: context,
		Data:    sync.Map{},
	}
	if nil != data && len(data) > 0 {
		for k, v := range data {
			ctx.Data.Store(k, v)
		}
	}
	return ctx
}

func (c *GContext) ToDataMap() map[interface{}]interface{} {
	m := make(map[interface{}]interface{})
	c.Data.Range(func(key, value interface{}) bool {
		m[key] = value
		return true
	})

	return m
}

/**
Goroutine ID to GContext

注意这个将 GID 和 GContext 绑定后，结束某个请求或者业务的时候，应该解绑，否则这个 map 会一直持有 GContext 引用导致内存泄漏
*/
var gContexts = sync.Map{}

func BindContext(traceId string, context context.Context, data map[interface{}]interface{}) *GContext {
	gid := GetGoroutineId()
	ctx, ok := gContexts.Load(gid)
	if ok {
		gctx := ctx.(*GContext)
		if data != nil && len(data) > 0 {
			for k, v := range data {
				gctx.Data.Store(k, v)
			}
		}
		return gctx
	}
	gctx := newGContext(gid, traceId, context, data)
	gContexts.Store(gid, gctx)
	return gctx
}

/**
将 当前 Goroutine ID 和 Context 解绑, BindContext 和 UnBindContext 应当成对出现，否则就会出现内存泄漏问题
*/
func UnBindContext() {
	defer func() {
		if r := recover(); r != nil {
			_ = fmt.Errorf("解绑定Goid和Context异常(panic): %v", r)
		}
	}()
	gid := GetGoroutineId()
	gContexts.Delete(gid)
}

func GetContext() *GContext {
	gid := GetGoroutineId()
	ctx, ok := gContexts.Load(gid)
	if ok {
		return ctx.(*GContext)
	}
	return nil
}

func GetTraceId() string {
	ctx := GetContext()
	if nil != ctx {
		return ctx.TraceId
	}
	return ""
}

func GetData(key interface{}) (val interface{}, exists bool) {
	ctx := GetContext()
	if nil == ctx {
		return nil, false
	}
	return ctx.Data.Load(key)
}

/**
绑定之后才有效设置
*/
func SetData(key interface{}, value interface{}) bool {
	ctx := GetContext()
	if nil == ctx {
		return false
	}
	ctx.Data.Store(key, value)
	return true
}

/**
通过Goroutine 运行具体业务逻辑
*/
func RunByGoroutine(handler func(), afterHandlers ...func(err interface{}, hadPanic bool)) {
	pctx := GetContext()
	go func() {
		traceId := ""
		var ctx context.Context
		var data map[interface{}]interface{}

		if nil != pctx {
			traceId = pctx.TraceId
			ctx = pctx.Context
			data = pctx.ToDataMap()
		}

		BindContext(traceId, ctx, data)
		defer UnBindContext()
		Run(handler, afterHandlers...)
	}()
}

func Run(handler func(), afterHandlers ...func(err interface{}, hadPanic bool)) {
	defer func() {
		var err interface{}
		if r := recover(); r != nil {
			err = r
			_ = fmt.Errorf("程序执行任务 Panic：%v", r)
		}

		if nil != afterHandlers && len(afterHandlers) > 0 {
			for _, afterHandler := range afterHandlers {
				func() {
					defer func() {
						if r := recover(); r != nil {
							_ = fmt.Errorf("执行AfterHandler异常, err：%v", r)
						}
					}()
					afterHandler(err, err != nil)
				}()
			}
		}
	}()
	// 指定具体业务
	handler()
}
