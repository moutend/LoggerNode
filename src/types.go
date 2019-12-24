package main

type LogMessage struct {
	Level             string `json:"level"`
	Source            string `json:"source"`
	Version           string `json:"version"`
	Message           string `json:"message"`
	Thread            int64  `json:"thread"`
	UnixTimestampSec  int64  `json:"timestampSec"`
	UnixTimestampNano int64  `json:"timestampNano"`
	Path              string `json:"path"`
}
