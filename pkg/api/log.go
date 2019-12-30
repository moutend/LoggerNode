package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/moutend/LoggerNode/pkg/types"
)

type postLogRequest struct {
	Messages []types.LogMessage `json:"messages"`
}

type logEndpoint struct {
	message chan types.LogMessage
}

func (l *logEndpoint) PostLog(w http.ResponseWriter, r *http.Request) error {
	b := &bytes.Buffer{}

	if _, err := io.Copy(b, r.Body); err != nil {
		return fmt.Errorf("Requested JSON is broken")
	}
	var req postLogRequest

	if err := json.Unmarshal(b.Bytes(), &req); err != nil {
		return fmt.Errorf("Requested JSON is invalid")
	}

	for _, v := range req.Messages {
		l.message <- v
	}

	if _, err := io.WriteString(w, "{}"); err != nil {
		return fmt.Errorf("Internal error")
	}

	return nil
}

func NewLogEndpoint(message chan types.LogMessage) *logEndpoint {
	return &logEndpoint{
		message: message,
	}
}
