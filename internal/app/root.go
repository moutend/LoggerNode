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
	"strings"
	"time"

	"github.com/spf13/viper"

	"github.com/moutend/LoggerNode/internal/api"
	"github.com/moutend/LoggerNode/internal/types"

	"github.com/go-chi/chi"
	"github.com/go-chi/valve"
	"github.com/spf13/cobra"
)

var RootCommand = &cobra.Command{
	Use:  "LoggerNode",
	RunE: rootRunE,
}

func rootRunE(cmd *cobra.Command, args []string) error {
	if path, _ := cmd.Flags().GetString("config"); path != "" {
		viper.SetConfigFile(path)
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	rand.Seed(time.Now().Unix())
	p := make([]byte, 16)

	if _, err := rand.Read(p); err != nil {
		return err
	}

	myself, err := user.Current()

	if err != nil {
		return err
	}

	fileName := fmt.Sprintf("LogServer-%s.txt", hex.EncodeToString(p))
	outputPath := filepath.Join(myself.HomeDir, "AppData", "Roaming", "ScreenReaderX", "Logs", "SystemLog", fileName)

	bw := types.NewBackgroundWriter(outputPath)
	defer bw.Close()

	log.SetFlags(log.Ldate | log.Ltime | log.LUTC | log.Llongfile)
	log.SetOutput(bw)

	valv := valve.New()
	baseCtx := valv.Context()

	logEndpoint := api.NewLogEndpoint(
		filepath.Join(myself.HomeDir, "AppData", "Roaming", "ScreenReaderX", "Logs", "EventLog",
			fmt.Sprintf("Event-%s.txt", hex.EncodeToString(p)),
		))

	router := chi.NewRouter()
	router.Post("/v1/log", logEndpoint.Post)

	server := http.Server{
		Addr:    viper.GetString("server.address"),
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

func init() {
	RootCommand.PersistentFlags().StringP("config", "c", "", "Path to configuration file")
}
