package camb

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type StartDubbingRequestBody struct {
	VideoURL       string `json:"video_url"`
	SourceLanguage int    `json:"source_language"`
	TargetLanguage int    `json:"target_language"`
}

type ApiResponse struct {
	// Define the structure according to the response from the external API
	TaskID string `json:"task_id"`
}

func (c *Camb) StartDubbing(reqBody StartDubbingRequestBody) (apiResponse ApiResponse, err error) {
	payloadBytes, err := json.Marshal(reqBody)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", c.API("/end_to_end_dubbing"), bytes.NewReader(payloadBytes))
	if err != nil {
		return
	}

	req.Header.Set("x-api-key", c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(respBody, &apiResponse)
	if err != nil {
		return
	}
	return
}
