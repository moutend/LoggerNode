package main

import (
	"net/http"
	"os"
	"os/user"
	"path/filepath"
)

func main() {
	u, err := user.Current()

	if err != nil {
		panic(err)
	}

	logBaseDir := filepath.Join(u.HomeDir, "AppData", "Roaming", "ScreenReaderX", "EventLog")
	os.MkdirAll(logBaseDir, 0755)

	le := NewLogEndpoint(logBaseDir)

	mux := NewMux()

	mux.Post("/v1/log", le.postLog)

	server := &http.Server{
		Addr:    ":7901",
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}

	le.Wait()

	defer server.Close()
}
