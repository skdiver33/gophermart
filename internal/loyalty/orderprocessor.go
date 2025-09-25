package loyalty

import (
	"context"
	"errors"
	"log"
	"sync/atomic"
	"time"

	bm "github.com/skdiver33/gophermart/internal/balance"
	om "github.com/skdiver33/gophermart/internal/order"
)

type OrderProcessor struct {
	BalanceManager *bm.BalanceManager
	OrderManager   *om.OrderManager
	client         *AccrualClient
}

type ProcessResult struct {
	Accrual         AccrualForOrder
	ProcessingError error
}

func NewOrderProcessor(balance *bm.BalanceManager, order *om.OrderManager, clientConfig *AccrualClientConfig) *OrderProcessor {
	newProcessor := OrderProcessor{BalanceManager: balance, OrderManager: order}
	accrClient := NewAccuralClient(clientConfig)
	newProcessor.client = accrClient
	return &newProcessor
}

func (processor *OrderProcessor) Start(ctx context.Context) {
	go processor.ProcessOrder(ctx)
}

const defaultPause = 1

var workerPause = int64(defaultPause)
var workerCounter = int64(0)
var numWorkers int64

func (processor *OrderProcessor) ProcessOrder(ctx context.Context) {

	workerCtx, cancel := context.WithTimeout(ctx, 5*time.Second)

	select {
	case <-ctx.Done():
		log.Printf("stop processing order gorutine")
		cancel()
		return
	default:
		for {
			orderInProc, err := processor.OrderManager.GetAllUnprocOrders(ctx)
			if err != nil {
				log.Printf("error get unproc orders %s", err.Error())
				cancel()
				return
			}
			numJobs := len(orderInProc)
			jobs := make(chan string, numJobs)
			result := make(chan ProcessResult, numJobs)

			numWorkers = int64(numJobs)/5 + 1

			for i := 0; i < int(numWorkers); i++ {
				go worker(workerCtx, processor.client, jobs, result)
			}
			for _, order := range orderInProc {
				jobs <- order.Number
			}
			close(jobs)
			for j := 0; j < numJobs; j++ {
				res := <-result
				if res.ProcessingError != nil {
					switch {
					case errors.Is(res.ProcessingError, ErrAccruallErrorDecodeAnswer):
						log.Printf("error decode accrual system answer %s", res.ProcessingError.Error())

					case errors.Is(res.ProcessingError, ErrAccruallNetworkError):
						log.Printf("%s", res.ProcessingError.Error())

					case errors.Is(res.ProcessingError, ErrAccruallInternalServError):
						log.Printf("accrual internal error %s", res.ProcessingError.Error())

					case errors.Is(res.ProcessingError, ErrAccruallOrderNotExist):
						log.Printf("accrual order %s not found in accrual system. Set status NOT FOUND. error %s", res.Accrual.Order, res.ProcessingError.Error())
						processor.OrderManager.UpdateOrderStatus(ctx, res.Accrual.Order, "NOT FOUND", 0)
					default:
						log.Printf("unknown error in processing order")
					}
				} else {
					processor.BalanceManager.AddAmount(ctx, orderInProc[res.Accrual.Order].UserID, res.Accrual.Accrual)
					processor.OrderManager.UpdateOrderStatus(ctx, res.Accrual.Order, res.Accrual.Status, res.Accrual.Accrual)
				}
			}
			close(result)
		}
	}
}

func worker(ctx context.Context, client *AccrualClient, jobs <-chan string, res chan<- ProcessResult) {
	for j := range jobs {

	worker_loop:
		for {
			timer := time.NewTimer(time.Duration(workerPause) * time.Second)
			if atomic.LoadInt64(&workerPause) != defaultPause {
				if atomic.AddInt64(&workerCounter, 1) == numWorkers {
					atomic.StoreInt64(&workerPause, defaultPause)
					atomic.StoreInt64(&workerCounter, 1)
				}
			}
			select {
			case <-ctx.Done():
				timer.Stop()
				log.Printf("worker end work because close main ctx")
				res <- ProcessResult{}
				return
			case <-timer.C:
				accrualResult, err := client.GetAccural(j)
				procResult := ProcessResult{}
				if err != nil {
					procResult.ProcessingError = err
					procResult.Accrual.Order = j
					res <- procResult
					break worker_loop
				}
				if accrualResult.RetryInterval != 0 {
					atomic.CompareAndSwapInt64(&workerPause, defaultPause, int64(accrualResult.Accrual))
					break
				}
				if accrualResult.Status == "PROCESSING" || accrualResult.Status == "REGISTERED" {
					break
				}
				if accrualResult.Status == "INVALID" || accrualResult.Status == "PROCESSED" {
					procResult.Accrual = *accrualResult
					res <- procResult
					break worker_loop
				}
			}
			timer.Stop()
		}
	}
}
