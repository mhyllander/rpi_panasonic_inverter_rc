package server

import (
	"compress/gzip"
	"encoding/json/v2"
	"errors"
	"html/template"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v3"

	"rpi_panasonic_inverter_rc/codec"
	"rpi_panasonic_inverter_rc/codecbase"
	"rpi_panasonic_inverter_rc/db"
	"rpi_panasonic_inverter_rc/logs"
	"rpi_panasonic_inverter_rc/rcutils"
	"rpi_panasonic_inverter_rc/sched"
)

var g_irSender *codec.IrSender
var rootTemplate *template.Template

type RootData struct {
	PageTitle string
}

// Gzip Compression
type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func Gzip(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			handler.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		gzw := gzipResponseWriter{Writer: gz, ResponseWriter: w}
		handler.ServeHTTP(gzw, r)
	})
}

func writeError(w http.ResponseWriter, status int, err error) {
	http.Error(w, err.Error(), status)
}

func writeJSON(w http.ResponseWriter, value any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.MarshalWrite(w, value)
}

// Accepts media type parameters, e.g. "application/json; charset=utf-8".
func hasJSONContentType(r *http.Request) bool {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	return err == nil && mediaType == "application/json"
}

func getRoot(w http.ResponseWriter, r *http.Request) {
	if logs.IsLogLevelDebug() {
		slog.Info("reloading template while debugging")
		webFunctions := template.FuncMap{}
		rootTemplate = template.Must(template.New("root.gohtml").Funcs(webFunctions).ParseFiles("web/root.gohtml"))
	}

	slog.Debug("GET /")
	data := RootData{
		PageTitle: "Panasonic Inverter RC",
	}
	err := rootTemplate.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Return all settings as JSON
func returnCurrentSettings(w http.ResponseWriter) {
	var theSettings codecbase.AllSettings
	theSettings.ModeSettings = make(codecbase.ModeSettingsMap)

	dbRc, err := db.CurrentConfig()
	if err != nil {
		slog.Error("apiGetSettings get current config failed", "err", err)
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	rcutils.CopyToSettings(dbRc, &theSettings.Settings)

	for _, m := range []uint{codecbase.C_Mode_Auto, codecbase.C_Mode_Heat, codecbase.C_Mode_Cool, codecbase.C_Mode_Dry} {
		temp, fan, err := db.GetModeSettings(m)
		if err != nil {
			slog.Error("apiGetSettings get mode settings failed", "mode", m, "err", err)
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		ms := codecbase.ModeSettings{}
		rcutils.CopyToModeSettings(temp, fan, &ms)
		theSettings.ModeSettings[codecbase.Mode2String(m)] = ms
	}

	err = writeJSON(w, &theSettings)
	if err != nil {
		slog.Error("apiGetSettings JSON encode settings failed", "err", err)
	}
}

func apiGetSettings(w http.ResponseWriter, r *http.Request) {
	returnCurrentSettings(w)
}

func apiPostSettings(w http.ResponseWriter, r *http.Request) {
	if !hasJSONContentType(r) {
		slog.Error("apiPostSettings: expecting JSON data", "Content-Type", r.Header.Get("Content-Type"))
		writeError(w, http.StatusBadRequest, errors.New("expecting JSON in request"))
		return
	}

	var settings = new(codecbase.Settings)
	err := json.UnmarshalRead(r.Body, settings)
	if err != nil {
		slog.Error("apiPostSettings: decode body failed", "err", err)
		writeError(w, http.StatusBadRequest, err)
		return
	}

	err = db.UpdateConfig(
		func(current *codec.RcConfig) (*codec.RcConfig, error) {
			return rcutils.ComposeSendConfig(settings, current), nil
		},
		g_irSender.SendConfig,
	)
	if err != nil {
		slog.Error("apiPostSettings: failed to update config", "err", err)
		status := http.StatusInternalServerError
		if errors.Is(err, codec.ErrSenderBusy) || errors.Is(err, codec.ErrSenderStopped) {
			status = http.StatusServiceUnavailable
		}
		writeError(w, status, err)
		return
	}

	sched.CondRestartTimerJobs(settings)

	returnCurrentSettings(w)
}

type JobSet struct {
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

func returnJobSets(w http.ResponseWriter) {
	var allJS []JobSet = make([]JobSet, 0)

	jss, err := db.GetJobSets()
	if err != nil {
		slog.Error("apiGetJobsets get jobset failed", "err", err)
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	for _, js := range *jss {
		allJS = append(allJS, JobSet{Name: js.Name, Active: js.Active})
	}

	err = writeJSON(w, &allJS)
	if err != nil {
		slog.Error("apiGetJobsets JSON encode jobsets failed", "err", err)
	}
}

func apiGetJobsets(w http.ResponseWriter, r *http.Request) {
	returnJobSets(w)
}

func apiPostJobsets(w http.ResponseWriter, r *http.Request) {
	var allJS []JobSet

	if !hasJSONContentType(r) {
		slog.Error("apiPostJobsets: expecting JSON data", "Content-Type", r.Header.Get("Content-Type"))
		writeError(w, http.StatusBadRequest, errors.New("expecting JSON in request"))
		return
	}

	err := json.UnmarshalRead(r.Body, &allJS)
	if err != nil {
		slog.Error("apiPostJobsets: decode body failed", "err", err)
		writeError(w, http.StatusBadRequest, err)
		return
	}

	for _, js := range allJS {
		if err := db.UpdateJobSet(js.Name, js.Active); err != nil {
			slog.Error("apiPostJobsets: failed to update jobset", "jobset", js.Name, "err", err)
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		sched.ScheduleJobsForJobset(js.Name, js.Active)
	}

	returnJobSets(w)
}

func StartServer(logLevel string, irSender *codec.IrSender) {
	var err error

	g_irSender = irSender

	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(httplog.RequestLogger(slog.Default(), &httplog.Options{}))
	r.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	// Handler to return compressed responses
	r.Use(Gzip)

	// status page
	r.Get("/", getRoot)
	webFunctions := template.FuncMap{}
	rootTemplate = template.Must(template.New("root.gohtml").Funcs(webFunctions).ParseFiles("web/root.gohtml"))

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/settings", apiGetSettings)
		r.Post("/settings", apiPostSettings)
		r.Get("/jobsets", apiGetJobsets)
		r.Post("/jobsets", apiPostJobsets)
	})

	srv := &http.Server{
		Addr:     ":3333",
		Handler:  r,
		ErrorLog: slog.NewLogLogger(slog.Default().Handler(), slog.LevelError),
	}

	err = srv.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		slog.Info("server closed")
	} else if err != nil {
		slog.Error("error starting server", "err", err)
		os.Exit(1)
	}
}
