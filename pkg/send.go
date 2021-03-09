package pkg

import (
	"encoding/json"
	"github.com/IBM/go-sdk-core/core"
	"github.com/avast/retry-go"
	"github.com/cheggaaa/pb/v3"
	"github.com/watson-developer-cloud/go-sdk/v2/speechtotextv1"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func Send(apiKey string, serviceUrl string, fileDir string, outputDir string, overwrite bool) (err error) {
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		err = os.Mkdir(outputDir, 0777)
	}

	authenticator := &core.IamAuthenticator{
		ApiKey: apiKey,
	}

	options := &speechtotextv1.SpeechToTextV1Options{
		Authenticator: authenticator,
	}

	speechToText, speechToTextErr := speechtotextv1.NewSpeechToTextV1(options)

	if speechToTextErr != nil {
		panic(speechToTextErr)
	}

	err = speechToText.SetServiceURL(serviceUrl)
	if err != nil {
		return
	}

	pattern := strings.Join([]string{fileDir, "/*.wav"}, "")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return
	}

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

	wg := &sync.WaitGroup{}
	bar := pb.StartNew(barCount)

	for _, file := range files {
		wg.Add(1)
		go func(filePath string) {

			fileNameWithoutExt := filepath.Base(filePath[:len(filePath)-len(filepath.Ext(filePath))])
			outputFileName := strings.Join([]string{fileNameWithoutExt, ".json"}, "")
			outputFile := filepath.Join(outputDir, outputFileName)

			if _, fileErr := os.Stat(outputFile); !os.IsNotExist(fileErr) && !overwrite {
				wg.Done()
				return
			}

			var audioFile io.ReadCloser
			var audioFileErr error
			audioFile, audioFileErr = os.Open(filePath)
			if audioFileErr != nil {
				panic(audioFileErr)
			}
			err = retry.Do(
				func() error {

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

			bar.Increment()
			wg.Done()
		}(file)
	}
	wg.Wait()
	bar.Finish()
	return nil
}
