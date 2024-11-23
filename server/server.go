package server

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httplog "github.com/mhyllander/go-chi-httplog/v2"

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
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	rcutils.CopyToSettings(dbRc, &theSettings.Settings)

	for _, m := range []uint{codecbase.C_Mode_Auto, codecbase.C_Mode_Heat, codecbase.C_Mode_Cool, codecbase.C_Mode_Dry} {
		temp, fan, err := db.GetModeSettings(m)
		if err != nil {
			slog.Error("apiGetSettings get mode settings failed", "mode", m, "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		ms := codecbase.ModeSettings{}
		rcutils.CopyToModeSettings(temp, fan, &ms)
		theSettings.ModeSettings[codecbase.Mode2String(m)] = ms
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(&theSettings)
	if err != nil {
		slog.Error("apiGetSettings JSON encode settings failed", "err", err)
	}
}

func apiGetSettings(w http.ResponseWriter, r *http.Request) {
	returnCurrentSettings(w)
}

func apiPostSettings(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		slog.Error("apiPostSettings: expecting JSON data", "Content-Type", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("expecting JSON in request"))
		return
	}

	var settings = new(codecbase.Settings)
	err := json.NewDecoder(r.Body).Decode(settings)
	if err != nil {
		slog.Error("apiPostSettings: decode body failed", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	dbRc, err := db.CurrentConfig()
	if err != nil {
		slog.Error("apiPostSettings: save config failed", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	sendRc := rcutils.ComposeSendConfig(settings, dbRc)
	g_irSender.SendConfig(sendRc)

	err = db.SaveConfig(sendRc, dbRc)
	if err != nil {
		slog.Error("apiPostSettings: failed to save config", "err", err)
		w.Write([]byte(err.Error()))
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
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	for _, js := range *jss {
		allJS = append(allJS, JobSet{Name: js.Name, Active: js.Active})
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(&allJS)
	if err != nil {
		slog.Error("apiGetJobsets JSON encode jobsets failed", "err", err)
	}
}

func apiGetJobsets(w http.ResponseWriter, r *http.Request) {
	returnJobSets(w)
}

func apiPostJobsets(w http.ResponseWriter, r *http.Request) {
	var allJS []JobSet

	if r.Header.Get("Content-Type") != "application/json" {
		slog.Error("apiPostJobsets: expecting JSON data", "Content-Type", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("expecting JSON in request"))
		return
	}

	err := json.NewDecoder(r.Body).Decode(&allJS)
	if err != nil {
		slog.Error("apiPostJobsets: decode body failed", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	for _, js := range allJS {
		db.UpdateJobSet(js.Name, js.Active)
		sched.ScheduleJobsForJobset(js.Name, js.Active)
	}

	returnJobSets(w)
}

func StartServer(logLevel string, irSender *codec.IrSender) {
	var err error

	g_irSender = irSender

	r := chi.NewRouter()

	// Logger
	logger := httplog.NewLogger("paninv-controller", httplog.Options{
		JSON:             true,
		LogLevel:         logs.SetLoggerOpts(logLevel).Level.Level(),
		Concise:          true,
		RequestHeaders:   false,
		SourceFieldName:  "",
		TimeFieldName:    "time",
		MessageFieldName: "msg",
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

	err = http.ListenAndServe(":3333", r)
	if errors.Is(err, http.ErrServerClosed) {
		slog.Info("server closed")
	} else if err != nil {
		slog.Error("error starting server", "err", err)
		os.Exit(1)
	}
}
