package app

import (
	"context"
	"log"
	"sync"

	"github.com/ZakirAvrora/exchange-rate/internal/exchangerates"
	"github.com/ZakirAvrora/exchange-rate/pkg/external/exchangeratesapi"
)

const _defaultWorkerNumber = 5

type Backends struct {
	RecordsService    exchangerates.RecodsService
	ExternalAPIClient exchangeratesapi.Client
}

type consumer struct {
	b            Backends
	queue        <-chan exchangerates.Record
	ctx          context.Context
	workerNumber int
	wg           *sync.WaitGroup
	doneChan     chan struct{}
}

func NewConsumer(ctx context.Context, b Backends, ch <-chan exchangerates.Record) (*consumer, <-chan struct{}) {
	doneChan := make(chan struct{})
	var wg sync.WaitGroup

	return &consumer{
		b:            b,
		queue:        ch,
		ctx:          ctx,
		workerNumber: _defaultWorkerNumber,
		wg:           &wg,
		doneChan:     doneChan,
	}, doneChan
}

func (c *consumer) Start() {
	go func() {
		for workerNumber := 0; workerNumber < c.workerNumber; workerNumber++ {
			c.wg.Add(1)
			go func() {
				defer c.wg.Done()
				ctx := c.ctx
				for {
					select {
					case <-ctx.Done():
						return
					case record := <-c.queue:
						rate, err := c.b.ExternalAPIClient.GetLatestRate(ctx, record.Base, record.Secondary)
						if err != nil {
							if err = c.b.RecordsService.ShiftFailed(ctx, record.Identifier); err != nil {
								// NoReturnErr: log error and continue work on other tasks
								log.Println(err, record.Identifier)
							}
						}

						if err = c.b.RecordsService.ShiftUpdated(ctx, record.Identifier, rate.Value); err != nil {
							// NoReturnErr: log error and continue work on other tasks
							log.Println(err, record.Identifier)
						}
					}
				}
			}()
		}
		c.wg.Wait()
		c.doneChan <- struct{}{}
	}()

}
