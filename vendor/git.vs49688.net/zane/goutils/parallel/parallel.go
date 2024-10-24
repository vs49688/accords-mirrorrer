package parallel

import (
	"context"
	"runtime"
	"sync"

	"golang.org/x/sync/errgroup"
)

type Proc[T any] func(ctx context.Context, val T) error

func Parallel[T any](ctx context.Context, numWorkers int, input []T, proc Proc[T]) []error {
	if numWorkers <= 0 {
		numWorkers = runtime.NumCPU()
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(numWorkers)

	errors := make([]error, len(input))

	for i, val := range input {
		if err := ctx.Err(); err != nil {
			errors[i] = err
			continue
		}

		g.Go(func() error {
			errors[i] = proc(ctx, val)
			return nil
		})
	}

	_ = g.Wait()
	return errors
}

// ParallelLegacy is the old implementation of Parallel before it was switched to use errgroup.Group
func ParallelLegacy[T any](ctx context.Context, numWorkers int, input []T, proc Proc[T]) []error {
	if numWorkers <= 0 {
		numWorkers = runtime.NumCPU()
	}

	inCh := make(chan int, len(input))
	for i := range input {
		inCh <- i
	}
	close(inCh)

	errors := make([]error, len(input))

	wg := sync.WaitGroup{}
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for idx := range inCh {
				if ctx.Err() != nil {
					return
				}

				errors[idx] = proc(ctx, input[idx])
			}
		}()
	}

	wg.Wait()

	return errors
}
