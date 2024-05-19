package camb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/models"
	"github.com/resend/resend-go/v2"
)

const EMAIL_TEMPLATE = `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html dir="ltr" lang="en">

  <head>
    <meta content="text/html; charset=UTF-8" http-equiv="Content-Type" />
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Urbanist:ital,wght@0,100..900;1,100..900&display=swap" rel="stylesheet">
  </head>

  <body style="background-color:rgb(255,255,255);margin-top:auto;margin-bottom:auto;margin-left:auto;margin-right:auto;font-family:Urbanist, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, &quot;Segoe UI&quot;, Roboto, &quot;Helvetica Neue&quot;, Arial, &quot;Noto Sans&quot;, sans-serif, &quot;Apple Color Emoji&quot;, &quot;Segoe UI Emoji&quot;, &quot;Segoe UI Symbol&quot;, &quot;Noto Color Emoji&quot;;padding-left:0.5rem;padding-right:0.5rem">
    <table align="center" width="100%" border="0" cellPadding="0" cellSpacing="0" role="presentation" style="max-width:465px;border-width:1px;border-style:solid;border-color:rgb(234,234,234);border-radius:0.25rem;margin-top:40px;margin-bottom:40px;margin-left:auto;margin-right:auto;padding:20px">
      <tbody>
        <tr style="width:100%">
          <td>
            <table align="center" width="100%" border="0" cellPadding="0" cellSpacing="0" role="presentation" style="margin-top:32px">
              <tbody>
                <tr>
                  <td><img alt="Camb AI" src="https://gpec.gauravgosain.dev/assets/camb_logo.png" style="display:block;outline:none;border:none;text-decoration:none;margin-top:0px;margin-bottom:0px;margin-left:auto;margin-right:auto;background-color:white;padding:4px" width="180" /></td>
                </tr>
              </tbody>
            </table>
            <p style="font-size:14px;line-height:24px;margin:16px 0;color:rgb(0,0,0)">Hello <!-- -->{{ .UserName }}<!-- -->,</p>
            <p style="font-size:14px;line-height:24px;margin:16px 0;color:rgb(0,0,0)">Here is your dubbed video powered by <strong>Camb AI</strong>.</p>
            <table align="center" width="100%" border="0" cellPadding="0" cellSpacing="0" role="presentation">
              <tbody>
                <tr>
                  <td>
                    <table align="center" width="100%" border="0" cellPadding="0" cellSpacing="0" role="presentation">
                      <tbody style="width:100%">
                        <tr style="width:100%">
                          <td align="right" data-id="__react-email-column"><img height="64" alt="Video Thumbnail" src="{{ .VideoThumbnail }}" width="64" /></td>
                          <td align="center" data-id="__react-email-column" style="width:20px"></td>
                          <td align="left" data-id="__react-email-column"><img height="64" alt="Audio Waveform" src="{{ .AudioWaveform }}" width="64" /></td>
                        </tr>
                      </tbody>
                    </table>
                  </td>
                </tr>
              </tbody>
            </table>

            <table align="center" width="100%" border="0" cellPadding="0" cellSpacing="0" role="presentation" style="text-align:center;margin-top:32px;margin-bottom:32px">
              <tbody>
                <tr>
                  <td><a href="{{ .VideoDownloadLink }}" style="line-height:100%;text-decoration:none;display:inline-block;max-width:100%;background-color:rgb(0,0,0);border-radius:0.25rem;color:rgb(255,255,255);font-size:12px;font-weight:600;text-decoration-line:none;text-align:center;padding-left:1.25rem;padding-right:1.25rem;padding-top:0.75rem;padding-bottom:0.75rem;padding:12px 20px 12px 20px" target="_blank"><span><!--[if mso]><i style="letter-spacing: 20px;mso-font-width:-100%;mso-text-raise:18" hidden>&nbsp;</i><![endif]--></span><span style="max-width:100%;display:inline-block;line-height:120%;mso-padding-alt:0px;mso-text-raise:9px">Download Video</span><span><!--[if mso]><i style="letter-spacing: 20px;mso-font-width:-100%" hidden>&nbsp;</i><![endif]--></span></a></td>
                </tr>
                <tr style="height:5px;"></tr>
                <tr>
                  <td><a href="{{ .AudioDownloadLink }}" style="line-height:100%;text-decoration:none;display:inline-block;max-width:100%;background-color:rgb(0,0,0);border-radius:0.25rem;color:rgb(255,255,255);font-size:12px;font-weight:600;text-decoration-line:none;text-align:center;padding-left:1.25rem;padding-right:1.25rem;padding-top:0.75rem;padding-bottom:0.75rem;padding:12px 20px 12px 20px" target="_blank"><span><!--[if mso]><i style="letter-spacing: 20px;mso-font-width:-100%;mso-text-raise:18" hidden>&nbsp;</i><![endif]--></span><span style="max-width:100%;display:inline-block;line-height:120%;mso-padding-alt:0px;mso-text-raise:9px">Download Audio</span><span><!--[if mso]><i style="letter-spacing: 20px;mso-font-width:-100%" hidden>&nbsp;</i><![endif]--></span></a></td>
                </tr>
              </tbody>
            </table>
          </td>
        </tr>
      </tbody>
    </table>
  </body>

</html>`

// The HTTP response from the dubbed_run_info endpoint to get the video and audio download links
type RunInfoResponse struct {
	VideoURL string `json:"video_url"` // The URL to download the video
	AudioURL string `json:"audio_url"` // The URL to download the audio
}

func GenerateAudioWaveform(videoPath string) (string, error) {
	outputPath := videoPath + "_waveform.png"

	// Build the FFmpeg command
	cmd := exec.Command("ffmpeg", "-y", "-i", videoPath, "-filter_complex", "showwavespic=s=500x500,scale=500:500:force_original_aspect_ratio=decrease,pad=500:500:(ow-iw)/2:(oh-ih)/2", "-frames:v", "1", outputPath)

	// Run the FFmpeg command
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Fatalf("FFmpeg command failed: %v, %s", err, stderr.String())
	}

	parts := strings.Split(outputPath, "/")

	return parts[len(parts)-1], nil
}

func GenerateVideoThumbnail(videoPath string) (string, error) {
	outputPath := videoPath + "_thumbnail.png"

	// Build the FFmpeg command
	cmd := exec.Command("ffmpeg", "-y", "-i", videoPath, "-vf", "crop='min(iw,ih)':'min(iw,ih)',scale=500:500", "-frames:v", "1", outputPath)

	// Run the FFmpeg command
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Fatalf("FFmpeg command failed: %v, %s", err, stderr.String())
	}

	parts := strings.Split(outputPath, "/")

	return parts[len(parts)-1], nil
}

// Sends an email to the user with the download links for the dubbed video
func (c *Camb) SendEmail(
	app *pocketbase.PocketBase,
	email string,
	run StatusResponse,
	record *models.Record,
	userName string,
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

	resendClient := resend.NewClient(c.ResendAPIKey)

	originalVideoURL := fmt.Sprintf(
		"pb_data/storage/%s/%s/%s",
		record.Collection().Id,
		record.Id,
		record.GetString("original_video"),
	)

	thumbnail, _ := GenerateVideoThumbnail(originalVideoURL)
	audioWaveform, _ := GenerateAudioWaveform(originalVideoURL)

	downloadFileURL := fmt.Sprintf(
		"%s/api/files/%s/",
		app.Settings().Meta.AppUrl,
		record.BaseFilesPath(),
	)

	record.Set("thumbnail", thumbnail)
	record.Set("waveform", audioWaveform)

	app.Dao().SaveRecord(record)

	htmlString := strings.Replace(EMAIL_TEMPLATE, "{{ .UserName }}", userName, 1)
	htmlString = strings.Replace(htmlString, "{{ .VideoThumbnail }}", downloadFileURL+thumbnail, 1)
	htmlString = strings.Replace(htmlString, "{{ .AudioWaveform }}", downloadFileURL+audioWaveform, 1)
	htmlString = strings.Replace(htmlString, "{{ .VideoDownloadLink }}", apiResponse.VideoURL, 1)
	htmlString = strings.Replace(htmlString, "{{ .AudioDownloadLink }}", apiResponse.AudioURL, 1)

	params := &resend.SendEmailRequest{
		From:    "camb.ai <help@camb.ai>",
		To:      []string{email},
		Html:    htmlString,
		Subject: "Download your Dubbed Video! (CAMB.AI x GPEC)",
		ReplyTo: "help@camb.ai",
	}

	// https://demo.react.email/preview/notifications/vercel-invite-user
	// https://stackoverflow.com/questions/32254818/generating-a-waveform-using-ffmpeg

	// ffmpeg -i gauravgosain01@gmail.com.webm -filter_complex "showwavespic=s=500x500,scale=500:500:force_original_aspect_ratio=decrease,pad=500:500:(ow-iw)/2:(oh-ih)/2" -frames:v 1 output.png

	// ffmpeg -i ack@camb.ai.webm -vf "crop='min(iw,ih)':'min(iw,ih)',scale=500:500" -frames:v 1 output.png

	_, err = resendClient.Emails.Send(params)
	if err != nil {
		fmt.Println(err.Error())
		record.Set("status", "Failed to send download email to "+email)
		app.Dao().SaveRecord(record)
		return
	}

	record.Set("status", "Download links sent to "+email)
	app.Dao().SaveRecord(record)

	return
}

func (c *Camb) SendEmailTest(
	email string,
	UserName string,
	VideoThumbnail string,
	AudioWaveform string,
	VideoDownloadLink string,
	AudioDownloadLink string,
) {
	resendClient := resend.NewClient(c.ResendAPIKey)

	htmlString := strings.Replace(EMAIL_TEMPLATE, "{{ .UserName }}", UserName, 1)
	htmlString = strings.Replace(htmlString, "{{ .VideoThumbnail }}", VideoThumbnail, 1)
	htmlString = strings.Replace(htmlString, "{{ .AudioWaveform }}", AudioWaveform, 1)
	htmlString = strings.Replace(htmlString, "{{ .VideoDownloadLink }}", VideoDownloadLink, 1)
	htmlString = strings.Replace(htmlString, "{{ .AudioDownloadLink }}", AudioDownloadLink, 1)

	fmt.Println(htmlString)

	params := &resend.SendEmailRequest{
		From:    "camb.ai <help@camb.ai>",
		To:      []string{email},
		Html:    htmlString,
		Subject: "Download your Dubbed Video! (CAMB.AI x GPEC)",
		ReplyTo: "help@camb.ai",
	}

	// https://demo.react.email/preview/notifications/vercel-invite-user
	// https://stackoverflow.com/questions/32254818/generating-a-waveform-using-ffmpeg

	_, err := resendClient.Emails.Send(params)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}
