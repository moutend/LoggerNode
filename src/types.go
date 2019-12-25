package main

type LogMessage struct {
	Level             string `json:"level"`
	Source            string `json:"source"`
	Version           string `json:"version"`
	Message           string `json:"message"`
	ThreadId          int64  `json:"threadId"`
	UnixTimestampSec  int64  `json:"unixTimestampSec"`
	UnixTimestampNano int64  `json:"unixTimestampNano"`
	Path              string `json:"path"`
}
