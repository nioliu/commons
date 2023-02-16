package component

import (
	"github.com/jordan-wright/email"
	"net/smtp"
)

type EmailVerifier struct {
	email *email.Email

	FromEmailName   string
	FromEmailPass   string
	EmailServerHost string
	EmailServerAddr string

	CodeFromUser string
}

func newEmailVerifier(email *email.Email, fromEmailName string, fromEmailPass string,
	emailServerHost string, emailServerAddr string) *EmailVerifier {
	return &EmailVerifier{email: email, FromEmailName: fromEmailName,
		FromEmailPass: fromEmailPass, EmailServerHost: emailServerHost,
		EmailServerAddr: emailServerAddr}
}

func GetNewEmail(to []string, bcc []string, From, FromEmailName, FromEmailPass, emailServerHost, emailServerAddr string) *EmailVerifier {
	return newEmailVerifier(&email.Email{
		From: From, // 发送者
		To:   to,   // 收件人
		Bcc:  bcc,  // 抄送
	}, FromEmailName,
		FromEmailPass,
		emailServerHost,
		emailServerAddr)
}

func (e *EmailVerifier) SendContentEmail(subject string, content []byte, attaches ...string) error {
	e.email.Text = content
	e.email.Subject = subject

	if attaches != nil {
		for _, f := range attaches {
			_, err := e.email.AttachFile(f)
			if err != nil {
				return err
			}
		}
	}

	// send email
	plainAuth := smtp.PlainAuth("", e.FromEmailName, e.FromEmailPass, e.EmailServerHost)

	if err := e.email.Send(e.EmailServerAddr, plainAuth); err != nil {
		return err
	}

	return nil
}
