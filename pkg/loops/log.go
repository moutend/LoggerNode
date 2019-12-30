package loops

import (
	"bytes"
	"encoding/json"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/moutend/LoggerNode/pkg/types"
)

type LogLoop struct {
	Quit              chan struct{}
	Message           chan types.LogMessage
	OutputPath        string
	HeartbeatDuration time.Duration
	HeartbeatURL      *url.URL
}

func (l *LogLoop) Run() error {
	var v types.LogMessage

	if l.HeartbeatDuration == 0 {
		l.HeartbeatDuration = 5 * time.Minute
	}
	for {
		select {
		case <-time.After(l.HeartbeatDuration):
			// go http.Post(l.HeartbeatURL.String(), `application/json`, nil)

			continue
		case <-l.Quit:
			return nil
		case v = <-l.Message:
		}

		data, err := json.Marshal(v)

		if err != nil {
			return err
		}

		os.MkdirAll(filepath.Dir(l.OutputPath), 0755)

		file, err := os.OpenFile(l.OutputPath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)

		if err != nil {
			return err
		}
		if _, err := io.Copy(file, bytes.NewBuffer(data)); err != nil {
			return err
		}
		if _, err := io.WriteString(file, "\n"); err != nil {
			return err
		}

		file.Close()
	}
}
