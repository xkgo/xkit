package xcontext

import (
	"fmt"
	"sync"
	"testing"
)

func TestRunByGoroutineWithContext(t *testing.T) {

	ctx := BindContext("", nil, nil)
	defer UnBindContext()

	fmt.Println("父协程 GContext: ", ctx.Gid, ctx.TraceId)

	wg := sync.WaitGroup{}
	wg.Add(1)

	RunByGoroutine(func() {
		gctx := GetContext()
		fmt.Println("子协程：", gctx.Gid, gctx.TraceId)
		wg.Done()
	})

	wg.Wait()
	fmt.Println("测试结束")
}
