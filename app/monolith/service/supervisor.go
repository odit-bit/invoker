package service

import (
	"context"
	"sync"

	"go.uber.org/multierr"
)

type Service interface {
	Run(context.Context) error
}

type Supervised []Service

func (spv Supervised) Run(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	runCtx, cancelFn := context.WithCancel(ctx)
	defer cancelFn()

	var wg sync.WaitGroup

	errCh := make(chan error, len(spv))
	wg.Add(len(spv))
	for _, svc := range spv {
		go func(svc Service) {
			if err := svc.Run(runCtx); err != nil {
				errCh <- err
				cancelFn()
			}
			wg.Done()
		}(svc)
	}

	<-runCtx.Done()
	wg.Wait()

	var err error
	close(errCh)
	for svcErr := range errCh {
		err = multierr.Append(err, svcErr)
	}
	return err
}
