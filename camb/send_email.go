package camb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/mail"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/tools/mailer"
)

type RunInfoResponse struct {
	VideoURL string `json:"video_url"`
	AudioURL string `json:"audio_url"`
}

func (c *Camb) SendEmail(
	app *pocketbase.PocketBase,
	email string,
	run StatusResponse,
	record *models.Record,
) (apiResponse RunInfoResponse, err error) {
	req, err := http.NewRequest(
		"GET",
		c.API(fmt.Sprintf("/dubbed_run_info/%d", run.RunID)),
		nil,
	)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	req.Header.Set("x-api-key", c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())

		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())

		return
	}

	err = json.Unmarshal(respBody, &apiResponse)
	if err != nil {
		fmt.Println(err.Error())

		return
	}

	message := &mailer.Message{
		From: mail.Address{
			Address: app.Settings().Meta.SenderAddress,
			Name:    app.Settings().Meta.SenderName,
		},
		To:      []mail.Address{{Address: email}},
		Subject: "Download your Dubbed Video! (CAMB.AI x GPEC)",
		HTML: fmt.Sprintf(`
    <h3>
      CAMB.AI x GPEC
    </h3>
    <br>
    <br>
    <p>
      <a class="btn" href="%s" target="_blank" rel="noopener">Download Video</a>
    </p>
    <p>
      <a class="btn" href="%s" target="_blank" rel="noopener">Download Audio</a>
    </p>`,
			apiResponse.VideoURL,
			apiResponse.AudioURL,
		),
	}

	err = app.NewMailClient().Send(message)
	if err != nil {
		fmt.Println(err.Error())

		return
	}

	record.Set("status", "Download links sent to "+email)
	app.Dao().SaveRecord(record)

	return
}

func (c *Camb) SendEmailTest(
	app *pocketbase.PocketBase,
	email string,
	run StatusResponse,
) (apiResponse RunInfoResponse, err error) {
	req, err := http.NewRequest(
		"GET",
		c.API(fmt.Sprintf("/dubbed_run_info/%d", run.RunID)),
		nil,
	)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	req.Header.Set("x-api-key", c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())

		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())

		return
	}

	// err = json.Unmarshal(respBody, &apiResponse)
	// if err != nil {
	// 	fmt.Println(err.Error())
	//
	// 	return
	// }
	//
	// html, err := json.MarshalIndent(apiResponse, "", " ")

	message := &mailer.Message{
		From: mail.Address{
			Address: app.Settings().Meta.SenderAddress,
			Name:    app.Settings().Meta.SenderName,
		},
		To:      []mail.Address{{Address: email}},
		Subject: "Here is your Dubbed Video!",
		HTML:    "<h1>CAMB.AI x GPEC</h1><br><br>" + string(respBody),
	}

	err = app.NewMailClient().Send(message)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	return
}
