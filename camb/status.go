package camb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type StatusResponse struct {
	Status string `json:"status"`
	RunID  int64  `json:"run_id"`
}

func (c *Camb) DubbingStatus(task ApiResponse) (apiResponse StatusResponse, err error) {
	req, err := http.NewRequest(
		"GET",
		c.API(fmt.Sprintf("/end_to_end_dubbing/%s", task.TaskID)),
		nil,
	)
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
