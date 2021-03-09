package pkg

import (
	"encoding/json"
	"github.com/IBM/go-sdk-core/core"
	"github.com/watson-developer-cloud/go-sdk/v2/speechtotextv1"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func Send(apiKey string, serviceUrl string, fileDir string, outputDir string) (err error) {
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

	wg := &sync.WaitGroup{}
	for _, file := range files {
		wg.Add(1)
		go func(filePath string) {
			var audioFile io.ReadCloser
			var audioFileErr error
			audioFile, audioFileErr = os.Open(filePath)
			if audioFileErr != nil {
				panic(audioFileErr)
			}
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
				err = responseErr
				return
			}
			b, _ := json.MarshalIndent(result, "", "  ")

			fileNameWithoutExt := filepath.Base(filePath[:len(filePath)-len(filepath.Ext(filePath))])
			outputFileName := strings.Join([]string{fileNameWithoutExt, ".json"}, "")
			outputFileDir := filepath.Join(outputDir, outputFileName)

			if err := ioutil.WriteFile(outputFileDir, b, 0755); err != nil {
				log.Fatalln(err)
			}
			wg.Done()
		}(file)
	}
	wg.Wait()
	return nil
}
