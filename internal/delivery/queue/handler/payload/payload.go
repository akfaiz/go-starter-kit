package payload

import "github.com/akfaiz/go-starter-kit/internal/domain"

type MailPayload struct {
	To      []string `json:"to"`
	Cc      []string `json:"cc"`
	Bcc     []string `json:"bcc"`
	Subject string   `json:"subject"`
	Text    string   `json:"text"`
	HTML    string   `json:"html"`
}

func (p *MailPayload) ToDomain() *domain.Mail {
	return &domain.Mail{
		To:      p.To,
		Cc:      p.Cc,
		Bcc:     p.Bcc,
		Subject: p.Subject,
		Text:    p.Text,
		HTML:    p.HTML,
	}
}
