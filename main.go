package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func isFileOpen(filename string) (bool, error) {
	info, err := os.Stat(filename)
	if err != nil {
		return false, err
	}
	return info.ModTime().Unix() != lastModTime, nil
}

var lastModTime int64

func main() {
	app := pocketbase.New()

	// fires only for the "dubbing" collection
	app.OnRecordAfterCreateRequest("dubbing").Add(func(e *core.RecordCreateEvent) error {
		// err := ffmpeg.Input(originalVideoUrl).
		// 	Output(
		// 		outputPath,
		// 		// ffmpeg.KwArgs{"c:v": "libx264"},
		// 		// ffmpeg.KwArgs{"c:a": "aac"},
		// 		// ffmpeg.KwArgs{"strict": "experimental"},
		// 	).
		// 	OverWriteOutput().ErrorToStdOut().Run()
		// go func(e *core.RecordCreateEvent) error {
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
			"-c:v", "libx264",
			"-c:a", "aac",
			"-preset", "ultrafast",
			"-crf", "28",
			// "-c:v", "copy",
			// "-strict", "experimental",
			outputPath,
		)
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf
		err := cmd.Run()
		if err != nil {
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

		// 	return nil
		// }(e)

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
