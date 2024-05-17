package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/gaurav-gosain/cambai-gpec-backend/camb"
	_ "github.com/joho/godotenv/autoload"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func main() {
	app := pocketbase.New()

	cambApi := camb.Init()

	// fires only for the "dubbing" collection
	app.OnRecordAfterCreateRequest("dubbing").Add(func(e *core.RecordCreateEvent) error {
		originalVideos := e.UploadedFiles["original_video"]

		if len(originalVideos) == 0 {
			return errors.New("video not uploaded")
		}

		originalVideo := originalVideos[0]

		originalVideoUrl := fmt.Sprintf("pb_data/storage/%s/%s/%s", e.Collection.Id, e.Record.Id, originalVideo.Name)

		if _, err := os.Stat(originalVideoUrl); err != nil {
			return err
		}

		var stdoutBuf, stderrBuf bytes.Buffer

		outputPath := originalVideoUrl + ".mp4"
		outputFileName := originalVideo.Name + ".mp4"
		cmd := exec.Command(
			"ffmpeg",
			"-i", originalVideoUrl,

			/*Failed attempts*/

			// "-r", "30",

			// "-map", "0",
			// "-map", "0:a",
			// "-map", "0:v",
			// "-c", "copy",

			/*****************/

			"-c:v", "libx264",
			"-c:a", "aac",
			"-preset", "ultrafast",
			"-crf", "28",

			outputPath,
		)
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf
		err := cmd.Run()
		if err != nil {
			fmt.Println(stdoutBuf.String())
			fmt.Println(stderrBuf.String())
			return err
		}

		record := e.Record
		record.Set("original_video", outputFileName)

		err = app.Dao().SaveRecord(record)
		if err != nil {
			return err
		}

		// delete the originalVideoUrl
		err = os.Remove(originalVideoUrl)
		if err != nil {
			return err
		}

		// expand the "user" relation
		if errs := app.Dao().ExpandRecord(record, []string{"user"}, nil); len(errs) > 0 {
			return fmt.Errorf("failed to expand: %v", errs)
		}

		userRecord := record.ExpandedOne("user")

		// fileDownloadToken, err := tokens.NewRecordFileToken(app, userRecord)
		// if err != nil {
		// 	return err
		// }

		downloadFileURL := fmt.Sprintf(
			"%s/api/files/%s/%s",
			// "%s/api/files/%s/%s?token=%s",
			app.Settings().Meta.AppUrl,
			record.BaseFilesPath(),
			outputFileName,
			// fileDownloadToken,
		)

		fmt.Println(downloadFileURL)

		go cambApi.StartDubbingPipeline(
			app,
			record,
			userRecord.GetString("email"),
			downloadFileURL,
		)

		return nil
	})

	app.OnRecordAfterDeleteRequest("dubbing").Add(func(e *core.RecordDeleteEvent) error {
		recordData := fmt.Sprintf("pb_data/storage/%s/%s", e.Collection.Id, e.Record.Id)

		err := os.RemoveAll(recordData)
		return err
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
