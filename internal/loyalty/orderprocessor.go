package loyalty

import (
	"context"
	"errors"
	"log"
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

func (processor *OrderProcessor) ProcessOrder(ctx context.Context) {
	numWorkers := 5
	workerCtx, cancel := context.WithTimeout(ctx, 5*time.Second)

	select {
	case <-ctx.Done():
		log.Printf("stop processing order gorutine")
		cancel()
		return
	default:
		for {
			orderInProc, _ := processor.OrderManager.GetAllUnprocOrders(ctx)
			numJobs := len(orderInProc)
			jobs := make(chan string, numJobs)
			result := make(chan ProcessResult, numJobs)

			for i := 0; i < numWorkers; i++ {
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
		select {
		case <-ctx.Done():
			log.Printf("worker end work because close main ctx")
			return
		default:
			for {
				accrualResult, err := client.GetAccural(j)
				procResult := ProcessResult{}
				if err != nil {
					procResult.ProcessingError = err
					procResult.Accrual.Order = j
					res <- procResult
					break
				}
				if accrualResult.RetryInterval != 0 {
					time.Sleep(time.Second * time.Duration(accrualResult.RetryInterval))
					continue
				}
				if accrualResult.Status == "PROCESSING" || accrualResult.Status == "REGISTERED" {
					time.Sleep(time.Second * 1)
					continue
				}
				if accrualResult.Status == "INVALID" || accrualResult.Status == "PROCESSED" {
					procResult.Accrual = *accrualResult
					res <- procResult
					break
				}
			}
		}
	}
}
