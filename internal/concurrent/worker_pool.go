package concurrent

import (
	"context"
	"encoding/json"
	"github.com/m3lifaro/gophermart/internal/model"
	"github.com/m3lifaro/gophermart/internal/service"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Job struct {
	OrderID string
	UserID  int32
}

// WorkerPool отвечает за запуск воркеров, передачу задач и координацию паузы и завершения
type WorkerPool struct {
	jobs         chan Job
	pauseAll     chan time.Duration
	workers      int
	shutdownCtx  context.Context
	shutdownFunc context.CancelFunc
	wg           sync.WaitGroup
}

func NewWorkerPool(workers int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerPool{
		jobs:         make(chan Job, 100),
		pauseAll:     make(chan time.Duration, 1),
		workers:      workers,
		shutdownCtx:  ctx,
		shutdownFunc: cancel,
	}
}

func (wp *WorkerPool) Start(orderService *service.OrderService, logger *zap.Logger) {
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(orderService, logger)
	}
}

func (wp *WorkerPool) worker(orderService *service.OrderService, logger *zap.Logger) {
	defer wp.wg.Done()
	for {
		select {
		case <-wp.shutdownCtx.Done():
			return
		case job := <-wp.jobs:
			logger.Debug("worker received job", zap.String("order_id", job.OrderID))
			err := orderService.UpdateOrder(job.OrderID, "PROCESSING", 0, job.UserID)
			if err != nil {
				logger.Error("error updating order", zap.Error(err))
				continue
			}
			for {
				req, err := orderService.ProcessAccrual(wp.shutdownCtx, job.OrderID)
				if err != nil {
					logger.Error("error forming accrual request", zap.Error(err))
					break
				}
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					logger.Error("error getting accrual status", zap.Error(err))
					timer := time.NewTimer(5 * time.Second)
					select {
					case <-wp.shutdownCtx.Done():
						timer.Stop()
						return
					case <-timer.C:
					}
					continue
				}
				status := resp.StatusCode
				defer resp.Body.Close()
				body, _ := io.ReadAll(resp.Body)
				var orderResp model.ExternalOrderResponse
				_ = json.Unmarshal(body, &orderResp)
				if status == http.StatusTooManyRequests {
					retryAfter := 30 * time.Second
					if hdr := resp.Header.Get("Retry-After"); hdr != "" {
						if secs, err := strconv.Atoi(hdr); err == nil {
							retryAfter = time.Duration(secs) * time.Second
						}
					}
					wp.pauseAll <- retryAfter
					logger.Warn("429 received, pausing all workers", zap.Duration("retryAfter", retryAfter))
				}
				if service.AccrualFinalStatuses[orderResp.Status] {
					err := orderService.UpdateOrder(job.OrderID, orderResp.Status, orderResp.Accrual, job.UserID)
					if err != nil {
						logger.Error("error updating order", zap.Error(err))
					}
					break
				}
				timer := time.NewTimer(10 * time.Second)
				select {
				case <-wp.shutdownCtx.Done():
					timer.Stop()
					return
				case <-timer.C:
				}
			}
		case pause := <-wp.pauseAll:
			timer := time.NewTimer(pause)
			select {
			case <-wp.shutdownCtx.Done():
				timer.Stop()
				return
			case <-timer.C:
				continue
			}
		}
	}
}

func (wp *WorkerPool) Shutdown() {
	wp.shutdownFunc()
	wp.wg.Wait()
}

func (wp *WorkerPool) AddJob(orderID string, userID int32) {
	wp.jobs <- Job{OrderID: orderID, UserID: userID}
}
