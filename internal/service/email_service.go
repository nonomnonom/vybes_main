package service

import (
	"fmt"
	"vybes/internal/config"

	"github.com/resend/resend-go/v2"
)

// EmailService defines the interface for sending emails.
type EmailService interface {
	SendOTPEmail(to, otp string) error
}

type resendEmailService struct {
	client      *resend.Client
	senderEmail string
}

// NewResendEmailService creates a new email service using Resend.
func NewResendEmailService(cfg *config.Config) EmailService {
	client := resend.NewClient(cfg.ResendAPIKey)
	return &resendEmailService{
		client:      client,
		senderEmail: cfg.SenderEmail,
	}
}

// SendOTPEmail sends an email with the OTP to the user via Resend.
func (s *resendEmailService) SendOTPEmail(to, otp string) error {
	subject := "Your Vybes OTP Code"
	htmlBody := fmt.Sprintf("<h1>Your OTP Code</h1><p>Your password reset OTP is: <b>%s</b></p>", otp)

	params := &resend.SendEmailRequest{
		From:    s.senderEmail,
		To:      []string{to},
		Subject: subject,
		Html:    htmlBody,
	}

	_, err := s.client.Emails.Send(params)
	if err != nil {
		return err
	}
	return nil
}
