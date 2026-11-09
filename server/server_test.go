package server

import (
	"encoding/json/v2"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"rpi_panasonic_inverter_rc/db"
)

type scheduleCall struct {
	jobset string
	active bool
}

// newTestServer uses a temporary database and records scheduling calls instead of running the scheduler.
func newTestServer(t *testing.T) (*httptest.Server, *[]scheduleCall) {
	t.Helper()

	if err := db.Initialize(filepath.Join(t.TempDir(), "test.db")); err != nil {
		t.Fatalf("db.Initialize failed: %v", err)
	}

	var calls []scheduleCall
	originalSchedule := scheduleJobsForJobset
	scheduleJobsForJobset = func(jobset string, active bool) error {
		calls = append(calls, scheduleCall{jobset, active})
		return nil
	}
	t.Cleanup(func() { scheduleJobsForJobset = originalSchedule })

	router := chi.NewRouter()
	registerRoutes(router)
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)
	return server, &calls
}

func send(t *testing.T, method, url, body string) (*http.Response, string) {
	t.Helper()

	request, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	text, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	return response, string(text)
}

func expectStatus(t *testing.T, response *http.Response, body string, want int) {
	t.Helper()
	if response.StatusCode != want {
		t.Fatalf("status %d, want %d (body: %s)", response.StatusCode, want, strings.TrimSpace(body))
	}
}

func TestJobsetLifecycle(t *testing.T) {
	server, calls := newTestServer(t)
	base := server.URL + "/api/v1/jobsets"

	// create
	createBody := `{"name":"Home","cronjobs":[
		{"name":"Evening","schedule":"0 18 * * *","settings":{"power":"on","temp":"22"}},
		{"schedule":"0 7 * * *","settings":{"power":"off"}}]}`
	response, body := send(t, http.MethodPost, base+"/Home", createBody)
	expectStatus(t, response, body, http.StatusCreated)

	var created JobSetBody
	if err := json.Unmarshal([]byte(body), &created); err != nil {
		t.Fatal(err)
	}
	if created.Active {
		t.Error("created jobset must be inactive")
	}
	if len(created.CronJobs) != 2 || created.CronJobs[0].Schedule != "0 7 * * *" || created.CronJobs[1].Name != "Evening" {
		t.Errorf("cronjobs not returned in time order: %+v", created.CronJobs)
	}
	if len(*calls) != 0 {
		t.Errorf("creating an inactive jobset must not schedule: %v", *calls)
	}

	// duplicate
	response, body = send(t, http.MethodPost, base+"/Home", createBody)
	expectStatus(t, response, body, http.StatusConflict)

	// activate with the bulk endpoint, then edit
	response, body = send(t, http.MethodPost, base, `[{"name":"Home","active":true}]`)
	expectStatus(t, response, body, http.StatusOK)

	response, body = send(t, http.MethodPut, base+"/Home", `{"name":"House","cronjobs":[{"schedule":"0 9 * * *","settings":{}}]}`)
	expectStatus(t, response, body, http.StatusOK)
	want := []scheduleCall{{"Home", true}, {"Home", false}, {"House", true}}
	if len(*calls) != len(want) {
		t.Fatalf("schedule calls %v, want %v", *calls, want)
	}
	for i := range want {
		if (*calls)[i] != want[i] {
			t.Errorf("schedule call %d: got %v, want %v", i, (*calls)[i], want[i])
		}
	}

	response, body = send(t, http.MethodGet, base+"/House", "")
	expectStatus(t, response, body, http.StatusOK)
	var edited JobSetBody
	if err := json.Unmarshal([]byte(body), &edited); err != nil {
		t.Fatal(err)
	}
	if !edited.Active || len(edited.CronJobs) != 1 || edited.CronJobs[0].Schedule != "0 9 * * *" {
		t.Errorf("unexpected jobset after edit: %+v", edited)
	}

	response, body = send(t, http.MethodGet, base+"/Home", "")
	expectStatus(t, response, body, http.StatusNotFound)

	// delete
	response, body = send(t, http.MethodDelete, base+"/House", "")
	expectStatus(t, response, body, http.StatusNoContent)
	if last := (*calls)[len(*calls)-1]; last != (scheduleCall{"House", false}) {
		t.Errorf("delete must unschedule, last call %v", last)
	}
	response, body = send(t, http.MethodDelete, base+"/House", "")
	expectStatus(t, response, body, http.StatusNotFound)
}

func TestJobsetValidationErrors(t *testing.T) {
	server, _ := newTestServer(t)
	base := server.URL + "/api/v1/jobsets"

	response, body := send(t, http.MethodPost, base+"/Bad", `{"name":"Bad","cronjobs":[{"schedule":"not a cron","settings":{}}]}`)
	expectStatus(t, response, body, http.StatusBadRequest)
	if !strings.Contains(body, "not a cron") {
		t.Errorf("error should mention the schedule: %s", body)
	}

	response, body = send(t, http.MethodPost, base+"/Bad", `{"name":"Other","cronjobs":[]}`)
	expectStatus(t, response, body, http.StatusBadRequest)

	response, body = send(t, http.MethodPost, base+"/Bad", `not json`)
	expectStatus(t, response, body, http.StatusBadRequest)

	response, body = send(t, http.MethodPut, base+"/Missing", `{"name":"Missing","cronjobs":[]}`)
	expectStatus(t, response, body, http.StatusNotFound)

	// nothing was created by the failed requests
	response, body = send(t, http.MethodGet, base, "")
	expectStatus(t, response, body, http.StatusOK)
	if strings.TrimSpace(body) != "[]" {
		t.Errorf("jobsets: %s, want []", body)
	}
}

func TestJobsetNameWithSpecialCharacters(t *testing.T) {
	server, _ := newTestServer(t)
	name := `Home & "Away"/100%`
	path := server.URL + "/api/v1/jobsets/" + url.PathEscape(name)

	response, body := send(t, http.MethodPost, path, `{"cronjobs":[]}`)
	expectStatus(t, response, body, http.StatusCreated)

	response, body = send(t, http.MethodGet, path, "")
	expectStatus(t, response, body, http.StatusOK)
	var got JobSetBody
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatal(err)
	}
	if got.Name != name {
		t.Errorf("name %q, want %q", got.Name, name)
	}
}
