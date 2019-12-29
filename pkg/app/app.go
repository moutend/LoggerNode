package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"sync"

	"github.com/moutend/LoggerNode/pkg/api"
	"github.com/moutend/LoggerNode/pkg/mux"
)

type app struct {
	m         *sync.Mutex
	wg        *sync.WaitGroup
	server    *http.Server
	isRunning bool
}

func (a *app) setup() error {
	u, err := user.Current()

	if err != nil {
		return err
	}

	logBaseDir := filepath.Join(u.HomeDir, "AppData", "Roaming", "ScreenReaderX", "EventLog")
	os.MkdirAll(logBaseDir, 0755)

	le := api.NewLogEndpoint(logBaseDir)

	mux := mux.New()

	mux.Post("/v1/log", le.PostLog)

	a.server = &http.Server{
		Addr:    ":4000",
		Handler: mux,
	}

	a.wg.Add(1)

	go func() {
		if err := a.server.ListenAndServe(); err != http.ErrServerClosed {
			panic(err)
		}
		a.wg.Done()
	}()

	return nil
}

func (a *app) Setup() error {
	a.m.Lock()
	defer a.m.Unlock()

	if a.isRunning {
		fmt.Errorf("Setup is already done")
	}
	if err := a.setup(); err != nil {
		return err
	}

	a.isRunning = true

	return nil
}

func (a *app) teardown() error {
	if err := a.server.Shutdown(context.TODO()); err != nil {
		return err
	}

	a.wg.Wait()

	return nil
}

func (a *app) Teardown() error {
	a.m.Lock()
	defer a.m.Unlock()

	if !a.isRunning {
		return fmt.Errorf("Teardown is already done")
	}
	if err := a.teardown(); err != nil {
		return err
	}

	a.isRunning = false

	return nil
}

func New() *app {
	return &app{
		m:  &sync.Mutex{},
		wg: &sync.WaitGroup{},
	}
}
