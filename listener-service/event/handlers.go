package event

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func LogEvent(p Payload) error {
	jsonData, _ := json.MarshalIndent(p, "", "\t")
	request, err := http.NewRequest("POST", "http://logger:8083/log", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		return err
	}
	return nil

}
