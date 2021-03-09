package pkg

import (
	"errors"
	"os"
)

func LoadEnv() (apiKey string, serviceUrl string, err error) {
	apiKey = os.Getenv("API_KEY")
	if apiKey == "" {
		err = errors.New("cannot read API_KEY")
		return
	}
	serviceUrl = os.Getenv("SERVICE_URL")
	if serviceUrl == "" {
		err = errors.New("cannot read SERVICE_URL")
	}
	return
}
