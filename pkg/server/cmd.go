package server

import (
	"context"
	"net"
	"net/http"
	"strconv"

	"github.com/caner-cetin/strafe/internal"
	"github.com/caner-cetin/strafe/pkg/db"
	"github.com/caner-cetin/strafe/pkg/server/endpoints"

	"github.com/rs/zerolog/log"

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
		Run: runServer,
	}
)

func GetRunCmd() *cobra.Command {
	runCmd.PersistentFlags().IntVarP(&port, "port", "p", 0, "port to run the server on")
	runCmd.PersistentFlags().StringVar(&host, "host", "0.0.0.0", "")
	return runCmd
}

func runServer(command *cobra.Command, args []string) {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	app := internal.AppCtx{}
	if err := app.InitializeDB(); err != nil {
		log.Error().Err(err).Msg("failed to initialize database")
		return
	}
	defer app.Cleanup()
	if err := db.Migrate(app.StdDB); err != nil {
		log.Error().Err(err).Msg("failed to migrate database")
		return
	}
	r.Use(WithAppContext(app))

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
			log.Error().
				Str("addr", "localhost:0").
				Err(err).
				Msg("failed to resolve tcp address")
			return
		}

		l, err := net.ListenTCP("tcp", addr)
		if err != nil {
			log.Error().
				Str("addr", addr.String()).
				Err(err).
				Msg("failed to listen on tcp")
			return
		}
		port = l.Addr().(*net.TCPAddr).Port
		l.Close()
	}
	log.Info().
		Str("host", host).
		Int("port", port).
		Msg("server is starting")
	http.ListenAndServe(net.JoinHostPort(host, strconv.Itoa(port)), r)
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
