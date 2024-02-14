package server

import (
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"rpi_panasonic_inverter_rc/utils"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v2"
)

var rootTemplate *template.Template

// type contextKey string

type RootData struct {
	PageTitle string
}

func getRoot(w http.ResponseWriter, r *http.Request) {
	// ctx := r.Context()

	slog.Debug("GET /")
	data := RootData{
		PageTitle: "Panasonic Inverter RC",
	}
	err := rootTemplate.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func StartServer(logLevel string) {
	var err error

	r := chi.NewRouter()

	// Logger
	logger := httplog.NewLogger("paninv-controller", httplog.Options{
		JSON:            true,
		LogLevel:        utils.SetLoggerOpts(logLevel).Level.Level(),
		Concise:         true,
		RequestHeaders:  false,
		SourceFieldName: "",
	})

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(httplog.RequestLogger(logger))
	r.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	// status page
	r.Get("/", getRoot)
	webFunctions := template.FuncMap{}
	rootTemplate = template.Must(template.New("root.gohtml").Funcs(webFunctions).ParseFiles("web/root.gohtml"))

	err = http.ListenAndServe(":3333", r)
	if errors.Is(err, http.ErrServerClosed) {
		slog.Info("server closed")
	} else if err != nil {
		slog.Error("error starting server", "err", err)
		os.Exit(1)
	}
}
