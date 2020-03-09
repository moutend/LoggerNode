package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/moutend/LoggerNode/pkg/types"
)

type postLogRequest struct {
	Messages []types.LogMessage `json:"messages"`
}

type logEndpoint struct {
	output io.Writer
}

func (l *logEndpoint) Post(w http.ResponseWriter, r *http.Request) {
	b := &bytes.Buffer{}

	if _, err := io.Copy(b, r.Body); err != nil {
		log.Println(err)
		http.Error(w, `{"errors":[{"message":"invalid request","path":[]}],"data":null}`, http.StatusBadRequest)
		return
	}

	var req postLogRequest

	if err := json.Unmarshal(b.Bytes(), &req); err != nil {
		log.Println(err)
		http.Error(w, `{"errors":[{"message":"invalid request","path":[]}],"data":null}`, http.StatusBadRequest)
		return
	}

	for _, message := range req.Messages {
		go func(v types.LogMessage) {
			data, err := json.Marshal(v)

			if err != nil {
				return
			}
			if _, err := io.Copy(l.output, bytes.NewBuffer(data)); err != nil {
				return
			}
		}(message)
	}

	if _, err := io.WriteString(w, fmt.Sprintf("{\"data\":{\"saved\":%d}}", len(req.Messages))); err != nil {
		log.Println(err)
	}
}

func NewLogEndpoint(outputPath string) *logEndpoint {
	return &logEndpoint{
		output: types.NewBackgroundWriter(outputPath),
	}
}
