package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strafe/internal"
	"strafe/server/endpoints"

	log "github.com/sirupsen/logrus"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/spf13/cobra"
)

var (
	port   int
	host   string
	runCmd = &cobra.Command{
		Use: "server [--port -p]",
		Run: internal.WrapCommandWithResources(runServer, internal.ResourceConfig{Resources: []internal.ResourceType{internal.ResourceDatabase}}),
	}
)

func GetRunCmd() *cobra.Command {
	runCmd.PersistentFlags().IntVarP(&port, "port", "p", 0, "port to run the server on")
	runCmd.PersistentFlags().StringVar(&host, "host", "0.0.0.0", "")
	return runCmd
}

func runServer(command *cobra.Command, args []string) {
	ctx := command.Context()
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(WithAppContext(ctx.Value(internal.APP_CONTEXT_KEY).(internal.AppCtx)))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://cansu.dev", "http://localhost:5173", "https://dj.cansu.dev"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
	}))

	r.Get("/health", endpoints.Health)
	r.Route("/track", func(track chi.Router) {
		track.Post("/random", endpoints.GetRandomTrack)
		track.Get("/{trackId}", endpoints.GetTrack)
	})
	if port == 0 {
		addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
		if err != nil {
			log.Fatal(err)
		}

		l, err := net.ListenTCP("tcp", addr)
		if err != nil {
			log.Fatal(err)
		}
		port = l.Addr().(*net.TCPAddr).Port
		l.Close()
	}
	log.WithFields(log.Fields{
		"port": port,
		"host": host,
	}).Info("server is starting")
	http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), r)
}

func WithAppContext(app internal.AppCtx) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), internal.APP_CONTEXT_KEY, app)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}
