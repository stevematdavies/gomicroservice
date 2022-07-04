package main

import (
	"broker/event"
	"broker/logs"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"log"
	"net/http"
	"net/rpc"
	"time"
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

type RPCPayload struct {
	Name string
	Data string
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
		app.rpcLog(w, requestPayload.Log)
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

// Log direct to logger service
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

// Log via RabbitMQ
func (app *Config) logEvent(w http.ResponseWriter, l LogPayload) {
	if err := app.push2Q(l.Name, l.Data); err != nil {
		app.errorJSON(w, err)
		return
	}
	app.writeJSON(w, http.StatusAccepted, jsonResponse{
		Error:   false,
		Message: "logged via Queue",
	})
}

// Log via RPC
func (app *Config) rpcLog(w http.ResponseWriter, l LogPayload) {
	client, err := rpc.Dial("tcp", "logger:5001")
	if err != nil {
		log.Println("Error dialing RPC", err)
		app.errorJSON(w, err)
		return
	}

	var result string
	if err = client.Call("RPCServer.LogInfo", RPCPayload{
		Name: l.Name,
		Data: l.Data,
	}, &result); err != nil {
		app.errorJSON(w, err)
		return
	}

	app.writeJSON(w, http.StatusAccepted, jsonResponse{
		Error:   false,
		Message: result,
	})
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

func (app *Config) push2Q(name, msg string) error {
	e, err := event.NewEventEmitter(app.Rmq)
	if err != nil {
		return err
	}
	p := LogPayload{name, msg}
	j, _ := json.MarshalIndent(&p, "", "\t")
	if err = e.Push(string(j), "log.INFO"); err != nil {
		return err
	}
	return nil
}

func (app *Config) GRPCLog(w http.ResponseWriter, r *http.Request) {
	var p RequestPayload
	if err := app.readJSON(w, r, &p); err != nil {
		app.errorJSON(w, err)
		return
	}

	cn, err := grpc.Dial("logger:5002", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer cn.Close()

	cl := logs.NewLogServiceClient(cn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	_, err = cl.WriteLog(ctx, &logs.LogRequest{
		LogEntry: &logs.Log{
			Name: p.Log.Name,
			Data: p.Log.Data,
		},
	})
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	app.writeJSON(w, http.StatusAccepted, jsonResponse{
		Error:   false,
		Message: "logged via gRPC!",
	})

}
