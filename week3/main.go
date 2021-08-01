package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var once sync.Once

func main() {
	// 信号只监听了SIGHUP（挂起）, SIGINT（中断）或 SIGTERM（终止）默认会使得程序终止退出的信号
	//before()
	after()
}

// before 1.16之前处理signal信号
func before() {
	r := generateHttpHandler()

	ctx := context.Background()

	// 定义withCancel -> cancel()方法去取消下游的Context
	ctx, cancel := context.WithCancel(ctx)

	// 使用errgroup进行goroutine取消
	group, errCtx := errgroup.WithContext(ctx)

	server := &http.Server{Addr: ":8080", Handler: r}
	group.Go(func() error {
		return server.ListenAndServe()
	})

	group.Go(func() error {
		<-errCtx.Done() //阻塞。因为cancel、timeout、deadline都可能导致Done被close
		fmt.Println("http server stop")

		return server.Shutdown(errCtx) // 关闭http server
	})

	// 1.16之前需要单独创建一个channel处理
	channel := make(chan os.Signal, 1) //这里要用buffer为1的chan
	signal.Notify(channel, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	group.Go(func() error {
		for {
			select {
			case <-errCtx.Done(): // 因为cancel、timeout、deadline都可能导致Done被close
				return errCtx.Err()
			case <-channel: // 中止信号
				cancel()
				return nil
			}
		}
	})

	if err := group.Wait(); err != nil {
		fmt.Println("group error: ", err)
	}
	fmt.Println("all group done!")
}

// after 1.16之后处理signal信号
func after() {
	r := generateHttpHandler()

	// 1.16之后新增了NotifyContext方法
	// 监控系统信号和创建Context一步搞定
	// 可以在创建的Context上继续作处理（WithCancel、WithTimeout……）
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	// 定义withCancel -> cancel()方法去取消下游的Context
	ctx, cancel := context.WithCancel(ctx)

	// 使用errgroup进行goroutine取消
	group, errCtx := errgroup.WithContext(ctx)

	server := &http.Server{Addr: ":8080", Handler: r}
	group.Go(func() error {
		return server.ListenAndServe()
	})

	group.Go(func() error {
		defer func() {
			once.Do(func() {
				fmt.Println("defer stop when http shutdown")
				stop() // 在收到信号的时候，会自动触发ctx的Done，这个stop 是不再捕获注册的信号的意思，算是一种释放资源
			})
		}()

		<-errCtx.Done() //阻塞。因为cancel、timeout、deadline都可能导致Done被 close
		fmt.Println("http server stop")

		return server.Shutdown(errCtx) // 关闭http server
	})

	group.Go(func() error {
		defer func() {
			once.Do(func() {
				fmt.Println("defer stop when get program termination exit signal")
				stop() // 在收到信号的时候，会自动触发ctx的Done，这个stop 是不再捕获注册的信号的意思，算是一种释放资源
			})
		}()

		for {
			select {
			case <-errCtx.Done(): // 因为cancel、timeout、deadline都可能导致Done被close
				return errCtx.Err()
			case <-ctx.Done(): // 收到中止信号后会自动调用done
				cancel()
			}
		}
	})

	if err := group.Wait(); err != nil {
		fmt.Println("group error: ", err)
	}
	fmt.Println("all group done!")
}

func generateHttpHandler() *gin.Engine {
	r := gin.Default()
	r.GET("/week3", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "success",
			"data":    "complete",
		})
	})
	return r
}
