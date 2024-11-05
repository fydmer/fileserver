package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"sync"
	"syscall"
)

type App struct {
	ctx     context.Context
	wg      *sync.WaitGroup
	stopFns []func()
}

func NewApp() *App {
	ctx, cancel := context.WithCancel(context.Background())

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
	slog.SetDefault(logger)

	app := &App{
		ctx:     ctx,
		wg:      &sync.WaitGroup{},
		stopFns: []func(){cancel},
	}

	go func(a *App) {
		a.wg.Add(2)

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

		for {
			sig := <-sigCh
			slog.Info("stopping app...", slog.String("signal", sig.String()))
			for _, fn := range a.stopFns {
				fn()
				a.wg.Done()
			}
			break
		}

		slog.Info("app terminated")
		a.wg.Done()
	}(app)

	slog.Info("starting app...")

	return app
}

func (a *App) Context() context.Context {
	return a.ctx
}

func (a *App) WG() *sync.WaitGroup {
	return a.wg
}

func (a *App) Panic(err error) {
	if err == nil {
		return
	}

	callerStr := ""
	if _, file, line, ok := runtime.Caller(1); ok {
		callerStr = fmt.Sprintf("%s:%d", file, line)
	}

	slog.Log(a.ctx, slog.LevelError+2, err.Error(),
		slog.String("caller", callerStr),
		slog.String("stacktrace", string(debug.Stack())))
	os.Exit(1)
}

func (a *App) AddStopFn(fn func()) {
	a.wg.Add(1)
	a.stopFns = append(a.stopFns, fn)
}

func (a *App) Keep() {
	a.wg.Wait()
}
