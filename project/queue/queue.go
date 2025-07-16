package queue

import (
	"context"
	"email-queue-service/models"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
)

var (
	queueLength = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "email_queue_length",
		Help: "Current number of jobs in the queue",
	})
	
	jobsProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "email_jobs_processed_total",
		Help: "Total number of email jobs processed",
	}, []string{"status"})
)

type Queue struct {
	jobs        chan *models.EmailJob
	retryQueue  chan *models.EmailJob
	deadLetter  []*models.EmailJob
	maxRetries  int
	mu          sync.RWMutex
	logger      *logrus.Logger
	closed      bool
}

func NewQueue(size int, maxRetries int, logger *logrus.Logger) *Queue {
	return &Queue{
		jobs:       make(chan *models.EmailJob, size),
		retryQueue: make(chan *models.EmailJob, size),
		deadLetter: make([]*models.EmailJob, 0),
		maxRetries: maxRetries,
		logger:     logger,
	}
}

func (q *Queue) Enqueue(job *models.EmailJob) error {
	q.mu.RLock()
	defer q.mu.RUnlock()
	
	if q.closed {
		return ErrQueueClosed
	}
	
	select {
	case q.jobs <- job:
		queueLength.Inc()
		q.logger.WithFields(logrus.Fields{
			"job_id": job.ID,
			"to":     job.To,
		}).Info("Job enqueued")
		return nil
	default:
		return ErrQueueFull
	}
}

func (q *Queue) Dequeue(ctx context.Context) (*models.EmailJob, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case job := <-q.jobs:
		queueLength.Dec()
		return job, nil
	case job := <-q.retryQueue:
		return job, nil
	}
}

func (q *Queue) Retry(job *models.EmailJob) {
	job.Retries++
	
	if job.Retries >= q.maxRetries {
		q.mu.Lock()
		q.deadLetter = append(q.deadLetter, job)
		q.mu.Unlock()
		
		jobsProcessed.WithLabelValues("failed").Inc()
		q.logger.WithFields(logrus.Fields{
			"job_id":  job.ID,
			"to":      job.To,
			"retries": job.Retries,
		}).Error("Job moved to dead letter queue")
		return
	}
	
	// Exponential backoff
	delay := time.Duration(job.Retries) * time.Second
	go func() {
		time.Sleep(delay)
		select {
		case q.retryQueue <- job:
			q.logger.WithFields(logrus.Fields{
				"job_id":  job.ID,
				"to":      job.To,
				"retries": job.Retries,
			}).Info("Job requeued for retry")
		default:
			q.logger.WithFields(logrus.Fields{
				"job_id": job.ID,
				"to":     job.To,
			}).Error("Failed to requeue job")
		}
	}()
}

func (q *Queue) MarkSuccess(job *models.EmailJob) {
	jobsProcessed.WithLabelValues("success").Inc()
	q.logger.WithFields(logrus.Fields{
		"job_id": job.ID,
		"to":     job.To,
	}).Info("Job completed successfully")
}

func (q *Queue) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()
	
	if !q.closed {
		q.closed = true
		close(q.jobs)
		close(q.retryQueue)
		q.logger.Info("Queue closed")
	}
}

func (q *Queue) GetDeadLetterJobs() []*models.EmailJob {
	q.mu.RLock()
	defer q.mu.RUnlock()
	
	result := make([]*models.EmailJob, len(q.deadLetter))
	copy(result, q.deadLetter)
	return result
}

func (q *Queue) GetStats() map[string]interface{} {
	q.mu.RLock()
	defer q.mu.RUnlock()
	
	return map[string]interface{}{
		"queue_length":      len(q.jobs),
		"retry_queue_length": len(q.retryQueue),
		"dead_letter_count":  len(q.deadLetter),
		"is_closed":         q.closed,
	}
}

var (
	ErrQueueFull   = &QueueError{Message: "queue is full"}
	ErrQueueClosed = &QueueError{Message: "queue is closed"}
)

type QueueError struct {
	Message string
}

func (e *QueueError) Error() string {
	return e.Message
}