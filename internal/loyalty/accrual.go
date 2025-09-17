package loyalty

import (
	"context"
	"fmt"
	"strings"
	"time"

	bm "github.com/skdiver33/gophermart/internal/balance"
	om "github.com/skdiver33/gophermart/internal/order"
)

type OrderProcessor struct {
	BalanceManager *bm.BalanceManager
	OrderManager   *om.OrderManager
	client         *AccrualClient
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
	numWorkers := 10
	select {
	case <-ctx.Done():
		return
	default:
		//main cycle
		for {
			orderInProc, _ := processor.OrderManager.GetAllUnprocOrders(ctx)
			numJobs := len(orderInProc)
			jobs := make(chan string, numJobs)
			result := make(chan AccrualForOrder, numJobs)

			for i := 0; i < numWorkers; i++ {
				go worker(ctx, processor.client, jobs, result)
			}
			for _, order := range orderInProc {
				jobs <- order.Number
			}
			close(jobs)
			for j := 0; j < numJobs; j++ {
				res := <-result
				if strings.Contains(res.Status, "error") {
					continue
				}
				processor.BalanceManager.AddAmount(ctx, orderInProc[res.Order].UserId, res.Accural)
				processor.OrderManager.UpdateOrderStatus(ctx, res.Order, res.Status, res.Accural)
			}
			close(result)

		}

	}
}

func worker(ctx context.Context, client *AccrualClient, jobs <-chan string, res chan<- AccrualForOrder) {
	for j := range jobs {
		select {
		case <-ctx.Done():
			return
		default:
			for {
				accrualResult, err := client.GetAccural(j)
				if err != nil {
					//log.Default("Error get accrual")
					res <- AccrualForOrder{Status: fmt.Sprintf("error %s", err.Error())}
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
					res <- *accrualResult
					break
				}
			}

		}
	}
}
