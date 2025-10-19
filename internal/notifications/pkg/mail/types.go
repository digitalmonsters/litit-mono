package mail

type EmailService struct {
	Host       string `json:"host"`
	SenderName string `json:"sendername"`
	User       string `json:"user"`
	Password   string `json:"password"`
	SenderMail string `json:"sendermail"`
	Port       string `json:"port"`
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
