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
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v3"

	"rpi_panasonic_inverter_rc/codec"
	"rpi_panasonic_inverter_rc/codecbase"
	"rpi_panasonic_inverter_rc/db"
	"rpi_panasonic_inverter_rc/rcutils"
	"rpi_panasonic_inverter_rc/sched"
	"rpi_panasonic_inverter_rc/web"
)

var g_irSender *codec.IrSender
var rootTemplate = template.Must(template.New("root.gohtml").ParseFS(web.FS, "root.gohtml"))

// Replaceable in tests.
var scheduleJobsForJobset = sched.ScheduleJobsForJobset

// Serializes jobset changes together with rescheduling, so that concurrent requests can't leave the schedule
// out of step with the database.
var jobsetEditMu sync.Mutex

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
	return writeJSONStatus(w, http.StatusOK, value)
}

func writeJSONStatus(w http.ResponseWriter, status int, value any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.MarshalWrite(w, value)
}

// Accepts media type parameters, e.g. "application/json; charset=utf-8".
func hasJSONContentType(r *http.Request) bool {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	return err == nil && mediaType == "application/json"
}

func getRoot(w http.ResponseWriter, r *http.Request) {
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

	if err := sched.CondRestartTimerJobs(settings); err != nil {
		slog.Error("apiPostSettings: failed to restart timer jobs", "err", err)
	}

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

	jobsetEditMu.Lock()
	defer jobsetEditMu.Unlock()

	var scheduleErrors []error
	for _, js := range allJS {
		if err := db.UpdateJobSet(js.Name, js.Active); err != nil {
			slog.Error("apiPostJobsets: failed to update jobset", "jobset", js.Name, "err", err)
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		if err := scheduleJobsForJobset(js.Name, js.Active); err != nil {
			scheduleErrors = append(scheduleErrors, err)
		}
	}
	if len(scheduleErrors) > 0 {
		err := errors.Join(scheduleErrors...)
		slog.Error("apiPostJobsets: failed to schedule all jobs", "err", err)
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	returnJobSets(w)
}

type CronJobBody struct {
	Name     string             `json:"name,omitempty"`
	Schedule string             `json:"schedule"`
	Settings codecbase.Settings `json:"settings"`
}

// JobSetBody is a job set with its cron jobs. Active is only returned, it is changed with apiPostJobsets.
type JobSetBody struct {
	Name     string        `json:"name"`
	Active   bool          `json:"active"`
	CronJobs []CronJobBody `json:"cronjobs"`
}

func newJobSetBody(definition db.JobSetDefinition) JobSetBody {
	body := JobSetBody{Name: definition.Name, Active: definition.Active, CronJobs: make([]CronJobBody, 0, len(definition.CronJobs))}
	for _, cronJob := range definition.CronJobs {
		body.CronJobs = append(body.CronJobs, CronJobBody{Name: cronJob.Name, Schedule: cronJob.Schedule, Settings: cronJob.Settings})
	}
	return body
}

func (body JobSetBody) definition() db.JobSetDefinition {
	definition := db.JobSetDefinition{Name: body.Name, CronJobs: make([]db.CronJobDefinition, 0, len(body.CronJobs))}
	for _, cronJob := range body.CronJobs {
		definition.CronJobs = append(definition.CronJobs, db.CronJobDefinition{Name: cronJob.Name, Schedule: cronJob.Schedule, Settings: cronJob.Settings})
	}
	return definition
}

func statusForDbError(err error) int {
	switch {
	case errors.Is(err, db.ErrInvalid):
		return http.StatusBadRequest
	case errors.Is(err, db.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, db.ErrExists):
		return http.StatusConflict
	}
	return http.StatusInternalServerError
}

// The jobset name from the URL path. chi leaves the parameter escaped if the request has a raw path.
func jobsetNameParam(r *http.Request) (string, error) {
	name := chi.URLParam(r, "name")
	if r.URL.RawPath != "" {
		return url.PathUnescape(name)
	}
	return name, nil
}

func readJobSetBody(w http.ResponseWriter, r *http.Request, handler string) (JobSetBody, bool) {
	var body JobSetBody
	if !hasJSONContentType(r) {
		slog.Error(handler+": expecting JSON data", "Content-Type", r.Header.Get("Content-Type"))
		writeError(w, http.StatusBadRequest, errors.New("expecting JSON in request"))
		return body, false
	}
	if err := json.UnmarshalRead(r.Body, &body); err != nil {
		slog.Error(handler+": decode body failed", "err", err)
		writeError(w, http.StatusBadRequest, err)
		return body, false
	}
	return body, true
}

func apiGetJobset(w http.ResponseWriter, r *http.Request) {
	name, err := jobsetNameParam(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	definition, err := db.GetJobSetDefinition(name)
	if err != nil {
		slog.Error("apiGetJobset: failed to get jobset", "jobset", name, "err", err)
		writeError(w, statusForDbError(err), err)
		return
	}

	if err := writeJSON(w, newJobSetBody(definition)); err != nil {
		slog.Error("apiGetJobset: JSON encode jobset failed", "err", err)
	}
}

// Creates an inactive jobset.
func apiCreateJobset(w http.ResponseWriter, r *http.Request) {
	name, err := jobsetNameParam(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	body, ok := readJobSetBody(w, r, "apiCreateJobset")
	if !ok {
		return
	}
	if body.Name == "" {
		body.Name = name
	}
	if body.Name != name {
		writeError(w, http.StatusBadRequest, errors.New("jobset name in body differs from the name in the URL"))
		return
	}

	jobsetEditMu.Lock()
	defer jobsetEditMu.Unlock()

	if err := db.CreateJobSet(body.definition()); err != nil {
		slog.Error("apiCreateJobset: failed to create jobset", "jobset", name, "err", err)
		writeError(w, statusForDbError(err), err)
		return
	}

	definition, err := db.GetJobSetDefinition(name)
	if err != nil {
		slog.Error("apiCreateJobset: failed to get jobset", "jobset", name, "err", err)
		writeError(w, statusForDbError(err), err)
		return
	}
	if err := writeJSONStatus(w, http.StatusCreated, newJobSetBody(definition)); err != nil {
		slog.Error("apiCreateJobset: JSON encode jobset failed", "err", err)
	}
}

// Replaces the cronjobs of a jobset and renames it if the name in the body differs from the name in the URL.
// The active state is kept.
func apiReplaceJobset(w http.ResponseWriter, r *http.Request) {
	oldName, err := jobsetNameParam(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	body, ok := readJobSetBody(w, r, "apiReplaceJobset")
	if !ok {
		return
	}
	if body.Name == "" {
		body.Name = oldName
	}

	jobsetEditMu.Lock()
	defer jobsetEditMu.Unlock()

	active, err := db.ReplaceJobSet(oldName, body.definition())
	if err != nil {
		slog.Error("apiReplaceJobset: failed to replace jobset", "jobset", oldName, "err", err)
		writeError(w, statusForDbError(err), err)
		return
	}

	var scheduleErrors []error
	if body.Name != oldName {
		if err := scheduleJobsForJobset(oldName, false); err != nil {
			scheduleErrors = append(scheduleErrors, err)
		}
	}
	if err := scheduleJobsForJobset(body.Name, active); err != nil {
		scheduleErrors = append(scheduleErrors, err)
	}
	if len(scheduleErrors) > 0 {
		err := errors.Join(scheduleErrors...)
		slog.Error("apiReplaceJobset: failed to schedule all jobs", "jobset", body.Name, "err", err)
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	definition, err := db.GetJobSetDefinition(body.Name)
	if err != nil {
		slog.Error("apiReplaceJobset: failed to get jobset", "jobset", body.Name, "err", err)
		writeError(w, statusForDbError(err), err)
		return
	}
	if err := writeJSON(w, newJobSetBody(definition)); err != nil {
		slog.Error("apiReplaceJobset: JSON encode jobset failed", "err", err)
	}
}

func apiDeleteJobset(w http.ResponseWriter, r *http.Request) {
	name, err := jobsetNameParam(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	jobsetEditMu.Lock()
	defer jobsetEditMu.Unlock()

	if err := db.DeleteJobSet(name); err != nil {
		slog.Error("apiDeleteJobset: failed to delete jobset", "jobset", name, "err", err)
		writeError(w, statusForDbError(err), err)
		return
	}
	if err := scheduleJobsForJobset(name, false); err != nil {
		slog.Error("apiDeleteJobset: failed to unschedule jobset", "jobset", name, "err", err)
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func registerRoutes(r chi.Router) {
	// status page
	r.Get("/", getRoot)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/settings", apiGetSettings)
		r.Post("/settings", apiPostSettings)
		r.Get("/jobsets", apiGetJobsets)
		r.Post("/jobsets", apiPostJobsets)
		r.Get("/jobsets/{name}", apiGetJobset)
		r.Post("/jobsets/{name}", apiCreateJobset)
		r.Put("/jobsets/{name}", apiReplaceJobset)
		r.Delete("/jobsets/{name}", apiDeleteJobset)
	})
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

	registerRoutes(r)

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
