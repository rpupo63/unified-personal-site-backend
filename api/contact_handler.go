package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/rpupo63/unified-personal-site-backend/errs"
	"github.com/rpupo63/unified-personal-site-backend/services"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type contactHandler struct {
	responder Responder
	logger    zerolog.Logger
}

func newContactHandler() contactHandler {
	logger := log.With().Str("handlerName", "contactHandler").Logger()

	return contactHandler{
		responder: NewResponder(logger),
		logger:    logger,
	}
}

// ContactFormRequest represents the contact form submission request
type ContactFormRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Subject string `json:"subject,omitempty"`
	Message string `json:"message"`
}

// NewsletterSubscribeRequest represents the newsletter subscription request
type NewsletterSubscribeRequest struct {
	Email string `json:"email"`
}

// ContactResponse represents the response from contact endpoints
type ContactResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// submitContactForm handles contact form submissions
// @Summary Submit contact form
// @Description Handles contact form submissions and sends an email notification
// @Tags Contact
// @Accept json
// @Produce json
// @Param contactForm body ContactFormRequest true "Contact form data"
// @Success 200 {object} ContactResponse "Contact form submitted successfully"
// @Failure 400 {object} api.ErrorResponse "Bad Request - Invalid contact form data"
// @Failure 500 {object} api.ErrorResponse "Internal Server Error - Error sending email"
// @Router /api/contact [post]
func (h contactHandler) submitContactForm() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			h.logger.Error().Err(err).Msg("Failed to read request body")
			h.responder.WriteError(w, errs.NewBadRequestError("failed to read request body"))
			return
		}

		var contactForm ContactFormRequest
		if err := json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&contactForm); err != nil {
			h.logger.Error().Err(err).Str("body", string(bodyBytes)).Msg("Failed to decode contact form request body")
			h.responder.WriteError(w, errs.NewBadRequestError("malformed request body"))
			return
		}

		// Validate required fields
		if strings.TrimSpace(contactForm.Name) == "" {
			h.responder.WriteError(w, errs.NewBadRequestError("name is required"))
			return
		}

		if strings.TrimSpace(contactForm.Email) == "" {
			h.responder.WriteError(w, errs.NewBadRequestError("email is required"))
			return
		}

		// Basic email validation
		if !strings.Contains(contactForm.Email, "@") {
			h.responder.WriteError(w, errs.NewBadRequestError("invalid email address"))
			return
		}

		if strings.TrimSpace(contactForm.Message) == "" {
			h.responder.WriteError(w, errs.NewBadRequestError("message is required"))
			return
		}

		// Check minimum message length (matching frontend validation)
		if len(strings.TrimSpace(contactForm.Message)) < 10 {
			h.responder.WriteError(w, errs.NewBadRequestError("message must be at least 10 characters"))
			return
		}

		// Get recipient email from environment variable or use default
		// This should be set to the site owner's email
		recipientEmail := getContactRecipientEmail()
		if recipientEmail == "" {
			h.logger.Error().Msg("CONTACT_RECIPIENT_EMAIL environment variable is not set")
			h.responder.WriteError(w, errs.NewInternalError("contact form is not configured"))
			return
		}

		// Build email subject
		subject := "Contact Form Submission"
		if contactForm.Subject != "" {
			subject = contactForm.Subject
		}

		// Build email body
		emailBody := buildContactEmailBody(contactForm)

		// Send email
		err = services.SendEmail(subject, emailBody, []string{recipientEmail})
		if err != nil {
			h.logger.Error().Err(err).Msg("Failed to send contact form email")
			h.responder.WriteError(w, errs.NewInternalError("failed to send email"))
			return
		}

		h.logger.Info().
			Str("name", contactForm.Name).
			Str("email", contactForm.Email).
			Msg("Contact form submitted successfully")

		response := ContactResponse{
			Success: true,
			Message: "Thank you for your message! I'll get back to you soon.",
		}

		h.responder.WriteJSON(w, response)
	}
}

// subscribeNewsletter handles newsletter subscription requests
// @Summary Subscribe to newsletter
// @Description Handles newsletter subscription requests and sends a confirmation email
// @Tags Contact
// @Accept json
// @Produce json
// @Param subscription body NewsletterSubscribeRequest true "Newsletter subscription data"
// @Success 200 {object} ContactResponse "Newsletter subscription successful"
// @Failure 400 {object} api.ErrorResponse "Bad Request - Invalid email address"
// @Failure 500 {object} api.ErrorResponse "Internal Server Error - Error processing subscription"
// @Router /api/newsletter/subscribe [post]
func (h contactHandler) subscribeNewsletter() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			h.logger.Error().Err(err).Msg("Failed to read request body")
			h.responder.WriteError(w, errs.NewBadRequestError("failed to read request body"))
			return
		}

		var subscription NewsletterSubscribeRequest
		if err := json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&subscription); err != nil {
			h.logger.Error().Err(err).Str("body", string(bodyBytes)).Msg("Failed to decode newsletter subscription request body")
			h.responder.WriteError(w, errs.NewBadRequestError("malformed request body"))
			return
		}

		// Validate email
		email := strings.TrimSpace(subscription.Email)
		if email == "" {
			h.responder.WriteError(w, errs.NewBadRequestError("email is required"))
			return
		}

		// Basic email validation
		if !strings.Contains(email, "@") {
			h.responder.WriteError(w, errs.NewBadRequestError("invalid email address"))
			return
		}

		// Get recipient email from environment variable or use default
		recipientEmail := getContactRecipientEmail()
		if recipientEmail == "" {
			h.logger.Error().Msg("CONTACT_RECIPIENT_EMAIL environment variable is not set")
			h.responder.WriteError(w, errs.NewInternalError("newsletter subscription is not configured"))
			return
		}

		// Build email body for newsletter subscription notification
		emailBody := buildNewsletterSubscriptionEmailBody(email)

		// Send email notification
		err = services.SendEmail("New Newsletter Subscription", emailBody, []string{recipientEmail})
		if err != nil {
			h.logger.Error().Err(err).Msg("Failed to send newsletter subscription email")
			h.responder.WriteError(w, errs.NewInternalError("failed to process subscription"))
			return
		}

		h.logger.Info().Str("email", email).Msg("Newsletter subscription received")

		response := ContactResponse{
			Success: true,
			Message: "Thank you for subscribing to the newsletter!",
		}

		h.responder.WriteJSON(w, response)
	}
}

// getContactRecipientEmail retrieves the recipient email from environment variable
func getContactRecipientEmail() string {
	// Try CONTACT_RECIPIENT_EMAIL first
	recipientEmail := os.Getenv("CONTACT_RECIPIENT_EMAIL")
	if recipientEmail != "" {
		return recipientEmail
	}

	// Fallback to RESEND_FROM_EMAIL if available (for development)
	return os.Getenv("RESEND_FROM_EMAIL")
}

// buildContactEmailBody builds the HTML email body for contact form submissions
func buildContactEmailBody(contactForm ContactFormRequest) string {
	subjectLine := ""
	if contactForm.Subject != "" {
		subjectLine = "<p><strong>Subject:</strong> " + htmlEscape(contactForm.Subject) + "</p>"
	}

	return `
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
		.container { max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background-color: #f4f4f4; padding: 20px; border-radius: 5px; margin-bottom: 20px; }
		.content { padding: 20px; }
		.message-box { background-color: #f9f9f9; padding: 15px; border-left: 4px solid #007bff; margin: 20px 0; }
		.footer { margin-top: 20px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 12px; color: #666; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h2>New Contact Form Submission</h2>
		</div>
		<div class="content">
			<p><strong>From:</strong> ` + htmlEscape(contactForm.Name) + `</p>
			<p><strong>Email:</strong> ` + htmlEscape(contactForm.Email) + `</p>
			` + subjectLine + `
			<div class="message-box">
				<p><strong>Message:</strong></p>
				<p>` + htmlEscape(contactForm.Message) + `</p>
			</div>
		</div>
		<div class="footer">
			<p>This email was sent from your personal site contact form.</p>
		</div>
	</div>
</body>
</html>
`
}

// buildNewsletterSubscriptionEmailBody builds the HTML email body for newsletter subscriptions
func buildNewsletterSubscriptionEmailBody(email string) string {
	return `
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
		.container { max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background-color: #f4f4f4; padding: 20px; border-radius: 5px; margin-bottom: 20px; }
		.content { padding: 20px; }
		.footer { margin-top: 20px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 12px; color: #666; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h2>New Newsletter Subscription</h2>
		</div>
		<div class="content">
			<p>A new user has subscribed to your newsletter:</p>
			<p><strong>Email:</strong> ` + htmlEscape(email) + `</p>
		</div>
		<div class="footer">
			<p>This email was sent from your personal site newsletter subscription form.</p>
		</div>
	</div>
</body>
</html>
`
}

// htmlEscape escapes HTML special characters
func htmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}
