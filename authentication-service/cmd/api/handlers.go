package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
)

func (app *Config) HealthCheck(w http.ResponseWriter, r *http.Request) {
	app.writeJSON(w, http.StatusAccepted, "healthy")
}

func (app *Config) Authenticate(w http.ResponseWriter, r *http.Request) {

	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		log.Println(err)
		return
	}

	user, err := app.Models.User.GetByEmail(requestPayload.Email)
	if err != nil {
		log.Println(err)
		app.errorJSON(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	log.Println(user)

	valid, err := user.PasswordMatches(requestPayload.Password)
	if !valid || err != nil {
		log.Println(err)
		app.errorJSON(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	// Log Authentication request

	if err = app.logRequest("Authentication attempt", fmt.Sprintf("%s attempted to authenticate at: %s", user.Email, time.Now())); err != nil {
		app.errorJSON(w, err)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Logged in user %s", user.Email),
		Data:    user,
	}

	app.writeJSON(w, http.StatusAccepted, payload)

}

func (app *Config) logRequest(name, data string) error {
	var e struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}

	e.Name = name
	e.Data = data

	j, _ := json.MarshalIndent(e, "", "\t")
	r, err := http.NewRequest("POST", "http://logger:8083/log", bytes.NewBuffer(j))
	if err != nil {
		return err
	}
	client := &http.Client{}
	_, err = client.Do(r)
	if err != nil {
		return err
	}
	return nil
	// mongodb://admin:password:localhost:2707/logs/?authSource=admin&readPreference=primary&appname=MongoDB%20Compass&directConnection=true&ssl=false
}
