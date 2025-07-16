package handlers

import (
	"encoding/json"
	"email-queue-service/models"
	"email-queue-service/queue"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// Simple UUID generation without external dependency
func generateUUID() string {
	return time.Now().Format("20060102150405") + "-" + string(rune(time.Now().UnixNano()%1000))
}

type EmailHandler struct {
	queue  *queue.Queue
	logger *logrus.Logger
}

func NewEmailHandler(q *queue.Queue, logger *logrus.Logger) *EmailHandler {
	return &EmailHandler{
		queue:  q,
		logger: logger,
	}
}

func (h *EmailHandler) SendEmail(w http.ResponseWriter, r *http.Request) {
	var req models.EmailRequest
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid JSON", err.Error())
		return
	}
	
	if err := req.Validate(); err != nil {
		h.respondError(w, http.StatusUnprocessableEntity, "Validation failed", err.Error())
		return
	}
	
	job := &models.EmailJob{
		ID:        generateUUID(),
		To:        req.To,
		Subject:   req.Subject,
		Body:      req.Body,
		Retries:   0,
		CreatedAt: time.Now(),
	}
	
	if err := h.queue.Enqueue(job); err != nil {
		if err == queue.ErrQueueFull {
			h.respondError(w, http.StatusServiceUnavailable, "Queue is full", "Please try again later")
			return
		}
		if err == queue.ErrQueueClosed {
			h.respondError(w, http.StatusServiceUnavailable, "Service is shutting down", "Please try again later")
			return
		}
		
		h.logger.WithError(err).Error("Failed to enqueue job")
		h.respondError(w, http.StatusInternalServerError, "Internal server error", "Failed to process request")
		return
	}
	
	response := models.EmailResponse{
		ID:      job.ID,
		Status:  "accepted",
		Message: "Email queued for processing",
	}
	
	h.respondJSON(w, http.StatusAccepted, response)
}

func (h *EmailHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats := h.queue.GetStats()
	h.respondJSON(w, http.StatusOK, stats)
}

func (h *EmailHandler) GetDeadLetterJobs(w http.ResponseWriter, r *http.Request) {
	jobs := h.queue.GetDeadLetterJobs()
	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"dead_letter_jobs": jobs,
		"count":           len(jobs),
	})
}

func (h *EmailHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *EmailHandler) respondError(w http.ResponseWriter, status int, error, message string) {
	response := models.ErrorResponse{
		Error:   error,
		Message: message,
	}
	h.respondJSON(w, status, response)
}