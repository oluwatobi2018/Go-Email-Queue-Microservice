package main

import (
	"context"
	"email-queue-service/config"
	"email-queue-service/handlers"
	"email-queue-service/queue"
	"email-queue-service/service"
	"email-queue-service/worker"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)
	
	// Load configuration
	cfg := config.Load()
	
	logger.WithFields(logrus.Fields{
		"port":         cfg.Port,
		"worker_count": cfg.WorkerCount,
		"queue_size":   cfg.QueueSize,
		"max_retries":  cfg.MaxRetries,
	}).Info("Starting email queue service")
	
	// Initialize services
	emailSvc := service.NewEmailService(logger)
	q := queue.NewQueue(cfg.QueueSize, cfg.MaxRetries, logger)
	workerPool := worker.NewPool(cfg.WorkerCount, q, emailSvc, logger)
	
	// Initialize handlers
	emailHandler := handlers.NewEmailHandler(q, logger)
	
	// Setup HTTP routes
	router := mux.NewRouter()
	
	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/send-email", emailHandler.SendEmail).Methods("POST")
	api.HandleFunc("/stats", emailHandler.GetStats).Methods("GET")
	api.HandleFunc("/dead-letter", emailHandler.GetDeadLetterJobs).Methods("GET")
	
	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")
	
	// Setup metrics server
	metricsRouter := mux.NewRouter()
	metricsRouter.Handle("/metrics", promhttp.Handler())
	
	// Start metrics server
	metricsServer := &http.Server{
		Addr:         ":" + cfg.MetricsPort,
		Handler:      metricsRouter,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	
	go func() {
		logger.WithField("port", cfg.MetricsPort).Info("Starting metrics server")
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start metrics server")
		}
	}()
	
	// Start worker pool
	workerPool.Start()
	
	// Setup main HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	
	// Start server in goroutine
	go func() {
		logger.WithField("port", cfg.Port).Info("Starting HTTP server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start HTTP server")
		}
	}()
	
	// Wait for interrupt signal for graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	
	<-c
	logger.Info("Shutting down gracefully...")
	
	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// Shutdown HTTP server
	if err := server.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("Failed to shutdown HTTP server")
	}
	
	// Shutdown metrics server
	if err := metricsServer.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("Failed to shutdown metrics server")
	}
	
	// Close queue to stop accepting new jobs
	q.Close()
	
	// Stop worker pool and wait for workers to finish
	workerPool.Stop()
	
	logger.Info("Service stopped")
}