package mail

type EmailService struct {
	SMTPHost    string
	SMTPPort    string
	Username    string
	Password    string
	SendersAddr string
}

// IEmailService defines the interface for email services.
type IEmailService interface {
	SendGenericEmail(to, subject, body string) error
	SendGenericHTMLEmail(to, subject, body string) error
}

type GenericEmailRPC struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type GenericHTMLEmailRPC struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}
