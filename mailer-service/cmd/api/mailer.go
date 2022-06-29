package main

import (
	"bytes"
	"github.com/vanng822/go-premailer/premailer"
	mail "github.com/xhit/go-simple-mail/v2"
	"html/template"
	"time"
)

type Mail struct {
	Domain      string
	Host        string
	Port        int
	Username    string
	Password    string
	Encryption  string
	FromAddress string
	FromName    string
}

type Message struct {
	From        string
	FromName    string
	To          string
	Subject     string
	Attachments []string
	Data        any
	DataMap     map[string]any
}

func (m *Mail) Send(msg Message) error {

	if msg.From == "" {
		msg.From = m.FromAddress
	}

	if msg.FromName == "" {
		msg.FromName = m.FromName
	}

	data := map[string]any{
		"message": msg.Data,
	}

	msg.DataMap = data

	html, err := m.htmlFromMessage(msg)
	if err != nil {
		return err
	}

	txt, err := m.plainTextFromMessage(msg)
	if err != nil {
		return err
	}

	mailServer := mail.NewSMTPClient()
	mailServer.Host = m.Host
	mailServer.Port = m.Port
	mailServer.Username = m.Username
	mailServer.Password = m.Password
	mailServer.Encryption = m.Encrypt()
	mailServer.KeepAlive = false
	mailServer.ConnectTimeout = 10 * time.Second
	mailServer.SendTimeout = 10 * time.Second

	mailClient, err := mailServer.Connect()
	if err != nil {
		return err
	}

	e := mail.NewMSG()
	e.SetFrom(msg.From).
		AddTo(msg.To).
		SetSubject(msg.Subject).
		SetBody(mail.TextPlain, txt).
		AddAlternative(mail.TextHTML, html)
	if len(msg.Attachments) > 0 {
		for _, a := range msg.Attachments {
			e.AddAttachment(a)
		}
	}
	if err = e.Send(mailClient); err != nil {
		return err
	}

	return nil
}

func (m *Mail) htmlFromMessage(msg Message) (string, error) {
	tmpl := "./templates/mail.html.gohtml"
	t, err := template.New("email-html").ParseFiles(tmpl)
	if err != nil {
		return "", err
	}
	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", msg.DataMap); err != nil {
		return "", err
	}

	fmtd := tpl.String()
	fmtd, err = m.withInlineStyling(fmtd)
	if err != nil {
		return "", err
	}
	return fmtd, nil
}

func (m *Mail) withInlineStyling(msg string) (string, error) {
	var opts = premailer.Options{
		RemoveClasses:     false,
		CssToAttributes:   false,
		KeepBangImportant: true,
	}
	p, err := premailer.NewPremailerFromString(msg, &opts)
	if err != nil {
		return "", err
	}
	html, err := p.Transform()
	if err != nil {
		return "", err
	}
	return html, nil
}

func (m *Mail) plainTextFromMessage(msg Message) (string, error) {
	tmpl := "./templates/mail.plain.gohtml"
	t, err := template.New("email-plain").ParseFiles(tmpl)
	if err != nil {
		return "", err
	}
	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", msg.DataMap); err != nil {
		return "", err
	}

	return tpl.String(), nil
}

func (m *Mail) Encrypt() mail.Encryption {
	switch m.Encryption {
	case "tls":
		return mail.EncryptionSTARTTLS
	case "ssl":
		return mail.EncryptionSSLTLS
	case "none":
		return mail.EncryptionNone
	default:
		return mail.EncryptionSTARTTLS
	}
}
