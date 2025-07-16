package service

import (
	"context"
	"email-queue-service/models"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

type EmailService struct {
	logger *logrus.Logger
}

func NewEmailService(logger *logrus.Logger) *EmailService {
	return &EmailService{
		logger: logger,
	}
}

func (s *EmailService) SendEmail(ctx context.Context, job *models.EmailJob) error {
	// Simulate email sending with sleep
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(1 * time.Second):
		// Simulate potential failure (10% chance)
		if time.Now().UnixNano()%10 == 0 {
			s.logger.WithFields(logrus.Fields{
				"job_id": job.ID,
				"to":     job.To,
				"retry":  job.Retries,
			}).Error("Failed to send email")
			return fmt.Errorf("failed to send email to %s", job.To)
		}
		
		s.logger.WithFields(logrus.Fields{
			"job_id":  job.ID,
			"to":      job.To,
			"subject": job.Subject,
			"retry":   job.Retries,
		}).Info("Email sent successfully")
		
		return nil
	}
}