package email

import "gopkg.in/gomail.v2"

const (
	HeaderFrom    = "From"
	HeaderTo      = "To"
	HeaderSubject = "Subject"
)

type (
	Option struct {
		Host     string
		Port     int
		Email    string
		Password string
	}

	MailBuilder struct {
		mailer *gomail.Message
		dialer *gomail.Dialer
	}
)

func New(opt *Option) *MailBuilder {
	mailer := gomail.NewMessage()
	mailer.SetHeader(HeaderFrom, opt.Email)

	dialer := gomail.NewDialer(
		opt.Host,
		opt.Port,
		opt.Email,
		opt.Password,
	)

	return &MailBuilder{
		mailer: mailer,
		dialer: dialer,
	}
}

func (mb *MailBuilder) SetTo(to ...string) *MailBuilder {
	mb.mailer.SetHeader(HeaderTo, to...)
	return mb
}

func (mb *MailBuilder) SetSubject(subject string) *MailBuilder {
	mb.mailer.SetHeader(HeaderSubject, subject)
	return mb
}

func (mb *MailBuilder) SetBody(contentType, body string) *MailBuilder {
	mb.mailer.SetBody(contentType, body)
	return mb
}

func (mb *MailBuilder) Send() error {
	return mb.dialer.DialAndSend(mb.mailer)
}
