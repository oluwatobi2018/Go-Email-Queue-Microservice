package worker

import (
	"context"
	"email-queue-service/models"
	"email-queue-service/queue"
	"email-queue-service/service"
	"sync"

	"github.com/sirupsen/logrus"
)

type Pool struct {
	workers     []*Worker
	queue       *queue.Queue
	emailSvc    *service.EmailService
	logger      *logrus.Logger
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

type Worker struct {
	id       int
	pool     *Pool
	logger   *logrus.Logger
}

func NewPool(workerCount int, q *queue.Queue, emailSvc *service.EmailService, logger *logrus.Logger) *Pool {
	ctx, cancel := context.WithCancel(context.Background())
	
	pool := &Pool{
		workers:  make([]*Worker, workerCount),
		queue:    q,
		emailSvc: emailSvc,
		logger:   logger,
		ctx:      ctx,
		cancel:   cancel,
	}
	
	for i := 0; i < workerCount; i++ {
		pool.workers[i] = &Worker{
			id:     i + 1,
			pool:   pool,
			logger: logger,
		}
	}
	
	return pool
}

func (p *Pool) Start() {
	p.logger.WithField("worker_count", len(p.workers)).Info("Starting worker pool")
	
	for _, worker := range p.workers {
		p.wg.Add(1)
		go worker.start()
	}
}

func (p *Pool) Stop() {
	p.logger.Info("Stopping worker pool")
	p.cancel()
	p.wg.Wait()
	p.logger.Info("Worker pool stopped")
}

func (w *Worker) start() {
	defer w.pool.wg.Done()
	
	w.logger.WithField("worker_id", w.id).Info("Worker started")
	
	for {
		select {
		case <-w.pool.ctx.Done():
			w.logger.WithField("worker_id", w.id).Info("Worker stopping")
			return
		default:
			job, err := w.pool.queue.Dequeue(w.pool.ctx)
			if err != nil {
				if err == context.Canceled {
					return
				}
				w.logger.WithError(err).Error("Failed to dequeue job")
				continue
			}
			
			w.processJob(job)
		}
	}
}

func (w *Worker) processJob(job *models.EmailJob) {
	w.logger.WithFields(logrus.Fields{
		"worker_id": w.id,
		"job_id":    job.ID,
		"to":        job.To,
		"retry":     job.Retries,
	}).Info("Processing job")
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // 5 second timeout
	defer cancel()
	
	if err := w.pool.emailSvc.SendEmail(ctx, job); err != nil {
		w.logger.WithFields(logrus.Fields{
			"worker_id": w.id,
			"job_id":    job.ID,
			"error":     err.Error(),
		}).Error("Failed to send email")
		
		w.pool.queue.Retry(job)
		return
	}
	
	w.pool.queue.MarkSuccess(job)
}