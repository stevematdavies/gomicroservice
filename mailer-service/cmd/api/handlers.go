package main

import (
	"fmt"
	"net/http"
)

type MailMessage struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

func (app *Config) SendMail(w http.ResponseWriter, r *http.Request) {
	var payload MailMessage

	if err := app.readJSON(w, r, payload); err != nil {
		_ = app.errorJSON(w, err)
		return
	}

	msg := Message{
		From:    payload.From,
		To:      payload.To,
		Subject: payload.Subject,
		Data:    payload.Message,
	}

	if err := app.Mailer.Send(msg); err != nil {
		_ = app.errorJSON(w, err)
		return
	}

	rp := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Sent to: %s", payload.To),
	}

	_ = app.writeJSON(w, http.StatusAccepted, rp)
}
