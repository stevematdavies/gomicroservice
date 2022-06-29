package main

import (
	"logging/data"
	"net/http"
)

type JSONPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *Config) WriteLog(w http.ResponseWriter, r *http.Request) {
	var payload JSONPayload
	_ = app.readJSON(w, r, &payload)

	evt := data.LogEntry{
		Name: payload.Name,
		Data: payload.Data,
	}

	if err := app.Models.LogEntry.Insert(evt); err != nil {
		app.errorJSON(w, err)
		return
	}

	rs := jsonResponse{
		Error:   false,
		Message: " <> LOGGED <> ",
	}

	app.writeJSON(w, http.StatusAccepted, rs)
}
