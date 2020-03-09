package app

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"time"

	"github.com/moutend/LoggerNode/pkg/api"
	"github.com/moutend/LoggerNode/pkg/types"

	"github.com/go-chi/chi"
	"github.com/go-chi/valve"
	"github.com/spf13/cobra"
)

var RootCommand = &cobra.Command{
	Use:  "LoggerNode",
	RunE: rootRunE,
}

func rootRunE(cmd *cobra.Command, args []string) error {
	rand.Seed(time.Now().Unix())
	p := make([]byte, 16)

	if _, err := rand.Read(p); err != nil {
		return err
	}

	myself, err := user.Current()

	if err != nil {
		return err
	}

	fileName := fmt.Sprintf("LoggerNode-%s.txt", hex.EncodeToString(p))
	outputPath := filepath.Join(myself.HomeDir, "AppData", "Roaming", "ScreenReaderX", "SystemLog", fileName)

	bw := types.NewBackgroundWriter(outputPath)
	defer bw.Close()

	log.SetFlags(log.Ldate | log.Ltime | log.LUTC | log.Llongfile)
	log.SetOutput(bw)

	valv := valve.New()
	baseCtx := valv.Context()

	logEndpoint := api.NewLogEndpoint(
		filepath.Join(myself.HomeDir, "AppData", "Roaming", "ScreenReaderX", "EventLog",
			fmt.Sprintf("Event-%s.txt", hex.EncodeToString(p)),
		))

	router := chi.NewRouter()
	router.Post("/v1/log", logEndpoint.Post)

	server := http.Server{
		Addr:    "localhost:7901",
		Handler: chi.ServerBaseContext(baseCtx, router),
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for range c {
			log.Println("Shutting down server")
			valv.Shutdown(30 * time.Second)

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			server.Shutdown(ctx)

			select {
			case <-time.After(31 * time.Second):
				log.Println("Failed to complete shutting down the server")
			case <-ctx.Done():
				log.Println("Complete shutting down the server")
			}
		}
	}()

	log.Printf("Listening on %s\n", server.Addr)

	return server.ListenAndServe()
}