//nolint:gocyclo // Graceful shutdown flows stay explicit for readability.
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
	errCh := make(chan error, 1)

	go func() {
		errCh <- run()
	}()

	stopCh := make(chan os.Signal, 1)

	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(stopCh)

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("http server %s stopped with error: %w", name, err)
		}

		return nil
	case sig := <-stopCh:
		runLog.WithField("signal", sig.String()).Info("shutting down http server")
	case <-ctx.Done():
		runLog.Info("shutting down http server")
	}

	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), DefaultShutdownTimeout)
	defer cancel()

	if err := shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown http server %s: %w", name, err)
	}

	err := <-errCh
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("http server %s stopped with error: %w", name, err)
	}

	return nil
}

//nolint:cyclop // Graceful shutdown flow intentionally stays explicit.
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
	errCh := make(chan error, 1)

	go func() {
		errCh <- serve()
	}()

	stopCh := make(chan os.Signal, 1)

	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(stopCh)

	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("grpc server %s stopped with error: %w", name, err)
		}

		return nil
	case sig := <-stopCh:
		runLog.WithField("signal", sig.String()).Info("shutting down grpc server")
	case <-ctx.Done():
		runLog.Info("shutting down grpc server")
	}

	gracefulDone := make(chan struct{})

	go func() {
		gracefulStop()
		close(gracefulDone)
	}()

	timer := time.NewTimer(DefaultShutdownTimeout)
	defer timer.Stop()

	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("grpc server %s stopped with error: %w", name, err)
		}

		return nil
	case <-gracefulDone:
		err := <-errCh
		if err != nil {
			return fmt.Errorf("grpc server %s stopped with error: %w", name, err)
		}

		return nil
	case <-timer.C:
		runLog.Warn("grpc graceful shutdown timed out, forcing stop")
		stop()

		err := <-errCh
		if err != nil {
			return fmt.Errorf("grpc server %s stopped with error: %w", name, err)
		}

		return nil
	}
}

func runnerLogger(ctx context.Context, log *logger.Logger, name string) *logger.Logger {
	runLog := log
	if runLog == nil {
		runLog = logger.FromContext(ctx)
	}

	return runLog.WithField("service", name)
}
