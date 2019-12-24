package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"sync"
)

type postLogRequest struct {
	Messages []LogMessage `json:"messages"`
}

type LogEndpoint struct {
	logBaseDir  string
	fileWriteWG *sync.WaitGroup
}

func writeLogFile(baseDir string, v LogMessage) error {
	data, err := json.Marshal(v)

	if err != nil {
		return err
	}

	p := make([]byte, 32)
	if _, err := rand.Read(p); err != nil {
		return err
	}

	fileName := fmt.Sprintf("%v.json", hex.EncodeToString(p))
	filePath := filepath.Join(baseDir, fileName)
	log.Println(filePath)
	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		return err
	}

	return nil
}

func (le *LogEndpoint) postLog(w http.ResponseWriter, r *http.Request) error {
	b := &bytes.Buffer{}

	if _, err := io.Copy(b, r.Body); err != nil {
		return fmt.Errorf("Requested JSON is broken")
	}

	var req postLogRequest

	if err := json.Unmarshal(b.Bytes(), &req); err != nil {
		return fmt.Errorf("Requested JSON is invalid")
	}

	for _, v := range req.Messages {
		le.fileWriteWG.Add(1)
		go func(v LogMessage) {
			err := writeLogFile(le.logBaseDir, v)
			log.Println(err)
			le.fileWriteWG.Done()
		}(v)
	}
	if _, err := io.WriteString(w, "{}"); err != nil {
		return fmt.Errorf("Internal error")
	}

	return nil
}

func (le *LogEndpoint) Wait() {
	le.fileWriteWG.Wait()
}

func NewLogEndpoint(logBaseDir string) *LogEndpoint {
	return &LogEndpoint{
		logBaseDir:  logBaseDir,
		fileWriteWG: &sync.WaitGroup{},
	}
}
