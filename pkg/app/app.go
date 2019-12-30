package app

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net/http"
	"os/user"
	"path/filepath"
	"sync"
	"time"

	"github.com/moutend/LoggerNode/pkg/api"
	"github.com/moutend/LoggerNode/pkg/loops"
	"github.com/moutend/LoggerNode/pkg/mux"
	"github.com/moutend/LoggerNode/pkg/types"
)

type app struct {
	m         *sync.Mutex
	wg        *sync.WaitGroup
	message   chan types.LogMessage
	quit      chan struct{}
	server    *http.Server
	isRunning bool
}

func (a *app) setup() error {
	u, err := user.Current()

	if err != nil {
		return err
	}

	rand.Seed(time.Now().Unix())
	p := make([]byte, 16)

	if _, err := rand.Read(p); err != nil {
		return err
	}

	fileName := fmt.Sprintf("EventLog-%s.txt", hex.EncodeToString(p))
	outputPath := filepath.Join(u.HomeDir, "AppData", "Roaming", "ScreenReaderX", "EventLog", fileName)

	a.message = make(chan types.LogMessage, 1024)
	a.quit = make(chan struct{})

	logLoop := &loops.LogLoop{
		Quit:       a.quit,
		Message:    a.message,
		OutputPath: outputPath,
	}

	a.wg.Add(1)

	go func() {
		if err := logLoop.Run(); err != nil {
			panic(err)
		}

		a.wg.Done()
	}()

	le := api.NewLogEndpoint(a.message)

	mux := mux.New()

	mux.Post("/v1/log", le.PostLog)

	a.server = &http.Server{
		Addr:    ":7901",
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

	a.quit <- struct{}{}

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
