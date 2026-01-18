package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"gopkg.in/gomail.v2"
)

func sendSignupNotification(signup Signup, workshopTitle string, workshopDate string) error {
	// Get email config from environment
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	smtpFrom := os.Getenv("SMTP_FROM")
	notificationEmail := os.Getenv("NOTIFICATION_EMAIL")

	// Skip if email not configured
	if smtpHost == "" || smtpUsername == "" || smtpPassword == "" {
		log.Println("⚠️  Email not configured, skipping notification")
		return nil
	}

	port, err := strconv.Atoi(smtpPort)
	if err != nil {
		port = 587 // Default SMTP port
	}

	// Create message
	m := gomail.NewMessage()
	m.SetHeader("From", smtpFrom)
	m.SetHeader("To", notificationEmail)
	m.SetHeader("Subject", fmt.Sprintf("New Signup: %s", workshopTitle))

	body := fmt.Sprintf(`
New workshop signup received!

Workshop: %s
Date: %s

Participant Details:
- Name: %s %s
- Email: %s
- Phone: %s

Signed up at: %s

View all signups at your admin panel.
    `, workshopTitle, workshopDate, signup.FirstName, signup.LastName, signup.Email, signup.Phone, signup.CreatedAt)

	m.SetBody("text/plain", body)

	// Send email
	d := gomail.NewDialer(smtpHost, port, smtpUsername, smtpPassword)

	if err := d.DialAndSend(m); err != nil {
		log.Printf("Failed to send email: %v", err)
		return err
	}

	log.Println("✓ Signup notification email sent")
	return nil
}

func sendConfirmationEmail(signup Signup, workshopTitle string, workshopDate string, workshopLocation string) error {
	// Get email config from environment
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	smtpFrom := os.Getenv("SMTP_FROM")

	// Skip if email not configured
	if smtpHost == "" || smtpUsername == "" || smtpPassword == "" {
		return nil
	}

	port, err := strconv.Atoi(smtpPort)
	if err != nil {
		port = 587
	}

	// Create message
	m := gomail.NewMessage()
	m.SetHeader("From", smtpFrom)
	m.SetHeader("To", signup.Email)
	m.SetHeader("Subject", fmt.Sprintf("Registration Confirmed: %s", workshopTitle))

	body := fmt.Sprintf(`
Dear %s,

Thank you for registering for our workshop!

Workshop Details:
- Title: %s
- Date: %s
- Location: %s

We look forward to seeing you there!

If you have any questions, please reply to this email.

Namaste 🙏
    `, signup.FirstName, workshopTitle, workshopDate, workshopLocation)

	m.SetBody("text/plain", body)

	// Send email
	d := gomail.NewDialer(smtpHost, port, smtpUsername, smtpPassword)

	if err := d.DialAndSend(m); err != nil {
		log.Printf("Failed to send confirmation email: %v", err)
		return err
	}

	log.Println("✓ Confirmation email sent to participant")
	return nil
}
