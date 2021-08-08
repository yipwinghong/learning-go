package main

import (
	"context"
	"golang.org/x/sync/errgroup"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// 基于 errgroup 实现一个 http server 的启动和关闭 ，以及 linux signal 信号的注册和处理，要保证能够一个退出，全部注销退出。
func main() {

	// 创建 HTTP Server 运行周期的 Context，封装 errgroup，实现 goroutine 级联 cancel。
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	group, ctx := errgroup.WithContext(ctx)

	// 创建处理操作系统信号的 chan。
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan)

	server := &http.Server{Addr: ":80"}

	// group.Go 内部通过 Once 实现只执行一次 cancel() 操作（并记录第一个出错信息）。

	// 启动、关闭 HTTP Server。
	group.Go(func() error {
		http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
			_, _ = io.WriteString(w, "hello, world!\n")
		})
		log.Println("http server start")
		return server.ListenAndServe()
	})
	group.Go(func() error {
		<-ctx.Done()
		log.Println("http server stop")
		// 如直接使用外部的 ctx 作为 http.Shutdown 的参数，会导致 http srv 不会等待服务协程处理完已连接的请求。\
		// 应使用新的 Context 对象来控制退出周期的运行。
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		return server.Shutdown(ctx) // 关闭 http server。
	})

	// 持续输出计数，被中断后退出。
	group.Go(func() error {
		for i := 0; ; i++ {
			timer := time.NewTimer(1 * time.Second)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-timer.C:
				log.Println(i)
				// 在接收到信号到退出会造成最长 1s 的延迟，应使用 timer 优化。
				//default:
				//    time.Sleep(1 * time.Second)
				//    log.Println(i)
			}
		}
	})

	// 处理 OS 信号以及 Context 的取消。
	group.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case sig := <-sigChan:
				// 只处理 SIGTERM 和 SIGINT。
				switch sig {
				case syscall.SIGTERM, syscall.SIGINT:
					cancel()
				}

			}
		}
	})

	// 通过 errgroup 阻塞主 goroutine。
	if err := group.Wait(); err != nil {
		log.Println("error: ", err)
	}
}