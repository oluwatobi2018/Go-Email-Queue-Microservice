package models

import (
	"net/mail"
	"time"
)

type EmailJob struct {
	ID        string    `json:"id"`
	To        string    `json:"to"`
	Subject   string    `json:"subject"`
	Body      string    `json:"body"`
	Retries   int       `json:"retries"`
	CreatedAt time.Time `json:"created_at"`
}

type EmailRequest struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type EmailResponse struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func (e *EmailRequest) Validate() error {
	if e.To == "" {
		return &ValidationError{Field: "to", Message: "email address is required"}
	}
	if e.Subject == "" {
		return &ValidationError{Field: "subject", Message: "subject is required"}
	}
	if e.Body == "" {
		return &ValidationError{Field: "body", Message: "body is required"}
	}
	
	// Validate email format
	if _, err := mail.ParseAddress(e.To); err != nil {
		return &ValidationError{Field: "to", Message: "invalid email format"}
	}
	
	return nil
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}