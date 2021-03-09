package pkg

import (
	"context"
	"encoding/json"
	"github.com/IBM/go-sdk-core/core"
	"github.com/avast/retry-go"
	"github.com/cheggaaa/pb/v3"
	"github.com/watson-developer-cloud/go-sdk/v2/speechtotextv1"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Send is a function to send request processing to api
func Send(apiKey string, serviceUrl string, fileDir string, outputDir string, overwrite bool, parallel int) (err error) {
	// If there is no outputDir, create one.
	if _, fileErr := os.Stat(outputDir); os.IsNotExist(fileErr) {
		err = os.Mkdir(outputDir, 0777)
	}
	if err != nil {
		return
	}

	// API config
	authenticator := &core.IamAuthenticator{
		ApiKey: apiKey,
	}
	options := &speechtotextv1.SpeechToTextV1Options{
		Authenticator: authenticator,
	}
	speechToText, err := speechtotextv1.NewSpeechToTextV1(options)
	if err != nil {
		return
	}
	err = speechToText.SetServiceURL(serviceUrl)
	if err != nil {
		return
	}

	// Search target files
	pattern := strings.Join([]string{fileDir, "/*.wav"}, "")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return
	}

	// Count bar progress
	fileNameMap := map[string]struct{}{}
	for _, f := range files {
		fileNameWithoutExt := filepath.Base(f[:len(f)-len(filepath.Ext(f))])
		fileNameMap[fileNameWithoutExt] = struct{}{}
	}
	pattern = strings.Join([]string{outputDir, "/*.json"}, "")
	jsons, err := filepath.Glob(pattern)
	if err != nil {
		return
	}
	diffCount := 0
	for _, j := range jsons {
		jsonNameWithoutExt := filepath.Base(j[:len(j)-len(filepath.Ext(j))])
		if _, ok := fileNameMap[jsonNameWithoutExt]; ok {
			diffCount++
		}
	}
	barCount := len(files) - diffCount

	// Managing Concurrency
	s := semaphore.NewWeighted(int64(parallel))
	eg := &errgroup.Group{}

	// Progress bar start
	bar := pb.StartNew(barCount)

	for _, file := range files {
		filePath := file
		eg.Go(
			func() error {
				// Manipulate the upper limit of the number of parallelism
				semaphoreErr := s.Acquire(context.Background(), 1)
				if semaphoreErr != nil {
					return semaphoreErr
				}

				// Overwrite an already existing file or
				fileNameWithoutExt := filepath.Base(filePath[:len(filePath)-len(filepath.Ext(filePath))])
				outputFileName := strings.Join([]string{fileNameWithoutExt, ".json"}, "")
				outputFile := filepath.Join(outputDir, outputFileName)
				if _, fileErr := os.Stat(outputFile); !os.IsNotExist(fileErr) && !overwrite {
					s.Release(1)
					return nil
				}

				// Open an audio file
				var audioFile io.ReadCloser
				var audioFileErr error
				audioFile, audioFileErr = os.Open(filePath)
				if audioFileErr != nil {
					return audioFileErr
				}

				// Retry processing
				sendErr := retry.Do(
					func() error {
						// Send request
						result, _, responseErr := speechToText.Recognize(
							&speechtotextv1.RecognizeOptions{
								Audio:                     audioFile,
								ContentType:               core.StringPtr("audio/wav"),
								Timestamps:                core.BoolPtr(true),
								WordAlternativesThreshold: core.Float32Ptr(0.9),
								Model:                     core.StringPtr("ja-JP_BroadbandModel"),
							},
						)
						if responseErr != nil {
							return responseErr
						}
						b, _ := json.MarshalIndent(result, "", "  ")

						if fileErr := ioutil.WriteFile(outputFile, b, 0755); fileErr != nil {
							return fileErr
						}
						return nil
					},
					retry.DelayType(retry.BackOffDelay),
					retry.Attempts(3),
				)
				if sendErr != nil {
					return sendErr
				}

				bar.Increment()
				s.Release(1)
				return nil
			},
		)
	}

	if waitErr := eg.Wait(); waitErr != nil {
		log.Fatal(waitErr)
	}
	bar.Finish()
	return
}
