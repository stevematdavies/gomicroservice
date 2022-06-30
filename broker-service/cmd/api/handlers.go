package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type MailPayload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
	Log    LogPayload  `json:"log,omitempty"`
	Mail   MailPayload `json:"mail,omitempty"`
}

func (app *Config) Broker(w http.ResponseWriter, _ *http.Request) {
	payload := jsonResponse{
		Error:   false,
		Message: "Hit the broker",
	}

	_ = app.writeJSON(w, http.StatusOK, payload)
}

func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	if err := app.readJSON(w, r, &requestPayload); err != nil {
		_ = app.errorJSON(w, err)
		return
	}

	switch requestPayload.Action {
	case "auth":
		app.authenticate(w, requestPayload.Auth)
	case "log":
		app.log(w, requestPayload.Log)
	case "mail":
		app.sendMail(w, requestPayload.Mail)
	default:
		_ = app.errorJSON(w, errors.New("unknown action"))
	}
}

func (app *Config) authenticate(w http.ResponseWriter, a AuthPayload) {
	jsonData, _ := json.MarshalIndent(a, "", "\t")
	request, err := http.NewRequest("POST", "http://auth:8082/authenticate", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println(err)
		_ = app.errorJSON(w, err)
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Println(err)
		_ = app.errorJSON(w, err)
		return
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			_ = app.errorJSON(w, err)
			return
		}
	}(response.Body)

	if response.StatusCode == http.StatusUnauthorized {
		_ = app.errorJSON(w, errors.New("Invalid credentials"))
	} else if response.StatusCode != http.StatusAccepted {
		_ = app.errorJSON(w, errors.New("Error calling auth service"))
	}

	var jsonFromService jsonResponse

	if err := json.NewDecoder(response.Body).Decode(&jsonFromService); err != nil {
		_ = app.errorJSON(w, err)
		return
	}

	if jsonFromService.Error {
		_ = app.errorJSON(w, err, http.StatusUnauthorized)
		return
	}

	var payload jsonResponse

	payload.Error = false
	payload.Message = "Authenticated!"
	payload.Data = jsonFromService.Data

	_ = app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) log(w http.ResponseWriter, l LogPayload) {
	jsonData, _ := json.MarshalIndent(l, "", "\t")
	request, err := http.NewRequest("POST", "http://logger:8083/log", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println(err)
		_ = app.errorJSON(w, err)
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		_ = app.errorJSON(w, err)
		return
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			_ = app.errorJSON(w, err)
			return
		}
	}(response.Body)

	if response.StatusCode != http.StatusAccepted {
		_ = app.errorJSON(w, err)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Logged!"

	_ = app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) sendMail(w http.ResponseWriter, m MailPayload) {
	jsonData, _ := json.MarshalIndent(m, "", "\t")
	request, err := http.NewRequest("POST", "http://mailer:8084/send", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println(err)
		_ = app.errorJSON(w, err)
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		_ = app.errorJSON(w, err)
		return
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			_ = app.errorJSON(w, err)
			return
		}
	}(response.Body)

	if response.StatusCode != http.StatusAccepted {
		_ = app.errorJSON(w, err)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = fmt.Sprintf("Email to: < %s > successfully sent from < %s >.", m.To, m.From)

	_ = app.writeJSON(w, http.StatusAccepted, payload)
}
