package server

import (
	"encoding/json"
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"rpi_panasonic_inverter_rc/codec"
	"rpi_panasonic_inverter_rc/common"
	"rpi_panasonic_inverter_rc/db"
	"rpi_panasonic_inverter_rc/sched"
	"rpi_panasonic_inverter_rc/utils"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v2"
)

var g_irSender *codec.IrSender
var rootTemplate *template.Template

type RootData struct {
	PageTitle string
}

func getRoot(w http.ResponseWriter, r *http.Request) {
	if common.IsLogLevelDebug() {
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
	var theSettings common.AllSettings
	theSettings.ModeSettings = make(common.ModeSettingsMap)

	dbRc, err := db.CurrentConfig()
	if err != nil {
		slog.Error("apiGetSettings get current config failed", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	utils.CopyToSettings(dbRc, &theSettings.Settings)

	for _, m := range []uint{common.C_Mode_Auto, common.C_Mode_Heat, common.C_Mode_Cool, common.C_Mode_Dry} {
		temp, fan, err := db.GetModeSettings(m)
		if err != nil {
			slog.Error("apiGetSettings get mode settings failed", "mode", m, "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		ms := common.ModeSettings{}
		utils.CopyToModeSettings(temp, fan, &ms)
		theSettings.ModeSettings[common.Mode2String(m)] = ms
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content_Type", "application/json")
	err = json.NewEncoder(w).Encode(&theSettings)
	if err != nil {
		slog.Error("apiGetSettings JSON encode settings failed", "err", err)
	}
}

func apiGetSettings(w http.ResponseWriter, r *http.Request) {
	returnCurrentSettings(w)
}

func apiPostSettings(w http.ResponseWriter, r *http.Request) {
	var settings = new(common.Settings)

	if r.Header.Get("Content-Type") != "application/json" {
		slog.Error("apiPostSettings expecting JSON data", "Content-Type", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("expecting JSON in request"))
		return
	}

	err := json.NewDecoder(r.Body).Decode(settings)
	if err != nil {
		slog.Error("apiPostSettings decode body failed", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	dbRc, err := db.CurrentConfig()
	if err != nil {
		slog.Error("apiPostSettings save config failed", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	sendRc := utils.ComposeSendConfig(settings, dbRc)
	sendRc.LogConfigAndChecksum("")
	g_irSender.SendConfig(sendRc)

	err = db.SaveConfig(sendRc, dbRc)
	if err != nil {
		slog.Error("failed to save config", "err", err)
		w.Write([]byte(err.Error()))
		return
	}

	sched.CondRestartTimerJobs(settings)

	returnCurrentSettings(w)
}

func StartServer(logLevel string, irSender *codec.IrSender) {
	var err error

	g_irSender = irSender

	r := chi.NewRouter()

	// Logger
	logger := httplog.NewLogger("paninv-controller", httplog.Options{
		JSON:            true,
		LogLevel:        common.SetLoggerOpts(logLevel).Level.Level(),
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

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/settings", apiGetSettings)
		r.Post("/settings", apiPostSettings)
	})

	err = http.ListenAndServe(":3333", r)
	if errors.Is(err, http.ErrServerClosed) {
		slog.Info("server closed")
	} else if err != nil {
		slog.Error("error starting server", "err", err)
		os.Exit(1)
	}
}
