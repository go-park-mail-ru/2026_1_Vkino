package serverrunner

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
)

const DefaultShutdownTimeout = 5 * time.Second

var errRunnerContextRequired = errors.New("runner context is required")

func RunHTTP(
	ctx context.Context,
	log *logger.Logger,
	name string,
	run func() error,
	shutdown func(context.Context) error,
) error {
	if ctx == nil {
		return errRunnerContextRequired
	}

	runLog := runnerLogger(ctx, log, name)
	errCh := startRunner(run)

	stopCh := notifyStopSignals()
	defer signal.Stop(stopCh)

	if done, err := waitHTTPStop(ctx, runLog, errCh, stopCh, name); done {
		return err
	}

	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), DefaultShutdownTimeout)
	defer cancel()

	if err := shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown http server %s: %w", name, err)
	}

	return wrapHTTPServeErr(name, <-errCh)
}

func RunGRPC(
	ctx context.Context,
	log *logger.Logger,
	name string,
	serve func() error,
	gracefulStop func(),
	stop func(),
) error {
	if ctx == nil {
		return errRunnerContextRequired
	}

	runLog := runnerLogger(ctx, log, name)
	errCh := startRunner(serve)

	stopCh := notifyStopSignals()
	defer signal.Stop(stopCh)

	if done, err := waitGRPCStop(ctx, runLog, errCh, stopCh, name); done {
		return err
	}

	gracefulDone := make(chan struct{})

	go func() {
		gracefulStop()
		close(gracefulDone)
	}()

	timer := time.NewTimer(DefaultShutdownTimeout)
	defer timer.Stop()

	return awaitGRPCShutdown(name, runLog, errCh, gracefulDone, timer.C, stop)
}

func runnerLogger(ctx context.Context, log *logger.Logger, name string) *logger.Logger {
	runLog := log
	if runLog == nil {
		runLog = logger.FromContext(ctx)
	}

	return runLog.WithField("service", name)
}

func startRunner(run func() error) chan error {
	errCh := make(chan error, 1)

	go func() {
		errCh <- run()
	}()

	return errCh
}

func notifyStopSignals() chan os.Signal {
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)

	return stopCh
}

func waitHTTPStop(
	ctx context.Context,
	log *logger.Logger,
	errCh <-chan error,
	stopCh <-chan os.Signal,
	name string,
) (bool, error) {
	select {
	case err := <-errCh:
		return true, wrapHTTPServeErr(name, err)
	case sig := <-stopCh:
		log.WithField("signal", sig.String()).Info("shutting down http server")
	case <-ctx.Done():
		log.Info("shutting down http server")
	}

	return false, nil
}

func wrapHTTPServeErr(name string, err error) error {
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("http server %s stopped with error: %w", name, err)
	}

	return nil
}

func waitGRPCStop(
	ctx context.Context,
	log *logger.Logger,
	errCh <-chan error,
	stopCh <-chan os.Signal,
	name string,
) (bool, error) {
	select {
	case err := <-errCh:
		return true, wrapGRPCServeErr(name, err)
	case sig := <-stopCh:
		log.WithField("signal", sig.String()).Info("shutting down grpc server")
	case <-ctx.Done():
		log.Info("shutting down grpc server")
	}

	return false, nil
}

func awaitGRPCShutdown(
	name string,
	log *logger.Logger,
	errCh <-chan error,
	gracefulDone <-chan struct{},
	timerC <-chan time.Time,
	stop func(),
) error {
	select {
	case err := <-errCh:
		return wrapGRPCServeErr(name, err)
	case <-gracefulDone:
		return wrapGRPCServeErr(name, <-errCh)
	case <-timerC:
		log.Warn("grpc graceful shutdown timed out, forcing stop")
		stop()

		return wrapGRPCServeErr(name, <-errCh)
	}
}

func wrapGRPCServeErr(name string, err error) error {
	if err != nil {
		return fmt.Errorf("grpc server %s stopped with error: %w", name, err)
	}

	return nil
}
