Speech-to-text-helper is a tool for requesting audio files in bulk to the IBM Watson™ Speech to Text service.

## Installation
```
go install github.com/hourglasshoro/speech-to-text-helper@latest
```

## Usage
1. Set the API key and URL of IBM Watson™ Speech to Text to environment variables.
```
export API_KEY=<your api key>
export SERVICE_URL=<service url>
```

2. Take the directory containing the wav file as input. Execute by specifying the input and output destinations.
```
speech-to-text-helper -s ./input -o ./output
```