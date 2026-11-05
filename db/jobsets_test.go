package db

import (
	"errors"
	"path/filepath"
	"reflect"
	"testing"

	"rpi_panasonic_inverter_rc/codecbase"
)

func initializeTestDb(t *testing.T) {
	t.Helper()
	if err := Initialize(filepath.Join(t.TempDir(), "test.db")); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
}

func cronJobSchedules(cronJobs []CronJobDefinition) []string {
	schedules := make([]string, 0, len(cronJobs))
	for _, cronJob := range cronJobs {
		schedules = append(schedules, cronJob.Schedule)
	}
	return schedules
}

func TestSortCronJobs(t *testing.T) {
	tests := []struct {
		name      string
		schedules []string
		want      []string
	}{
		{
			name:      "hour and minute",
			schedules: []string{"30 18 * * *", "0 7 * * *", "15 7 * * *"},
			want:      []string{"0 7 * * *", "15 7 * * *", "30 18 * * *"},
		},
		{
			name:      "weekday jobs are ordered by hour and minute, weekday breaks ties",
			schedules: []string{"0 22 * * 1-5", "30 6 * * 6", "30 6 * * 1", "0 7 * * mon,wed"},
			want:      []string{"30 6 * * 1", "30 6 * * 6", "0 7 * * mon,wed", "0 22 * * 1-5"},
		},
		{
			name:      "specific dates come after daily and weekday jobs",
			schedules: []string{"0 8 24 12 *", "0 9 1 6 *", "0 23 * * *", "0 8 1 6 *"},
			want:      []string{"0 23 * * *", "0 8 1 6 *", "0 9 1 6 *", "0 8 24 12 *"},
		},
		{
			name:      "steps, ranges, lists and names use their first value",
			schedules: []string{"*/15 12 * * *", "5-10 3 * * *", "0,30 4 * * *", "0 5 1 jan *"},
			want:      []string{"5-10 3 * * *", "0,30 4 * * *", "*/15 12 * * *", "0 5 1 jan *"},
		},
		{
			name:      "descriptors",
			schedules: []string{"@monthly", "@daily", "@hourly"},
			want:      []string{"@hourly", "@daily", "@monthly"},
		},
		{
			name:      "uninterpretable schedules last, in original order",
			schedules: []string{"@every 1h", "0 7 * * *", "bad"},
			want:      []string{"0 7 * * *", "@every 1h", "bad"},
		},
		{
			name:      "equal schedules keep their order",
			schedules: []string{"0 7 * * *", "0 7 * * *"},
			want:      []string{"0 7 * * *", "0 7 * * *"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cronJobs := make([]CronJobDefinition, 0, len(test.schedules))
			for i, schedule := range test.schedules {
				cronJobs = append(cronJobs, CronJobDefinition{Name: string(rune('a' + i)), Schedule: schedule})
			}

			SortCronJobs(cronJobs)

			if got := cronJobSchedules(cronJobs); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("got %v, want %v", got, test.want)
			}
		})
	}
}

func TestSortCronJobsKeepsOrderOfEqualJobs(t *testing.T) {
	cronJobs := []CronJobDefinition{
		{Name: "first", Schedule: "0 7 * * *"},
		{Name: "second", Schedule: "0 7 * * *"},
		{Name: "early", Schedule: "0 6 * * *"},
	}

	SortCronJobs(cronJobs)

	want := []string{"early", "first", "second"}
	for i, cronJob := range cronJobs {
		if cronJob.Name != want[i] {
			t.Fatalf("position %d: got %q, want %q", i, cronJob.Name, want[i])
		}
	}
}

func TestValidateJobSetDefinition(t *testing.T) {
	valid := JobSetDefinition{
		Name: "Home",
		CronJobs: []CronJobDefinition{{
			Name:     "Morning",
			Schedule: "30 6 * * 1-5",
			Settings: codecbase.Settings{Power: "on", Mode: "heat", Temperature: "21", FanSpeed: "auto", TimerOnTime: "06:30"},
		}},
	}
	if err := ValidateJobSetDefinition(valid); err != nil {
		t.Fatalf("valid definition rejected: %v", err)
	}

	invalid := map[string]func(*JobSetDefinition){
		"empty name":         func(d *JobSetDefinition) { d.Name = " " },
		"padded name":        func(d *JobSetDefinition) { d.Name = " Home" },
		"too long name":      func(d *JobSetDefinition) { d.Name = string(make([]rune, maxNameLength+1)) },
		"bad schedule":       func(d *JobSetDefinition) { d.CronJobs[0].Schedule = "61 * * * *" },
		"empty schedule":     func(d *JobSetDefinition) { d.CronJobs[0].Schedule = "" },
		"too long job name":  func(d *JobSetDefinition) { d.CronJobs[0].Name = string(make([]rune, maxNameLength+1)) },
		"unknown mode":       func(d *JobSetDefinition) { d.CronJobs[0].Settings.Mode = "turbo" },
		"temperature range":  func(d *JobSetDefinition) { d.CronJobs[0].Settings.Temperature = "35" },
		"temperature number": func(d *JobSetDefinition) { d.CronJobs[0].Settings.Temperature = "warm" },
		"bad time":           func(d *JobSetDefinition) { d.CronJobs[0].Settings.TimerOffTime = "25:00" },
	}
	for name, change := range invalid {
		t.Run(name, func(t *testing.T) {
			definition := valid
			definition.CronJobs = append([]CronJobDefinition(nil), valid.CronJobs...)
			change(&definition)

			if err := ValidateJobSetDefinition(definition); !errors.Is(err, ErrInvalid) {
				t.Fatalf("got %v, want ErrInvalid", err)
			}
		})
	}
}

func TestCreateAndGetJobSet(t *testing.T) {
	initializeTestDb(t)

	definition := JobSetDefinition{
		Name: "Home",
		CronJobs: []CronJobDefinition{
			{Name: "Evening", Schedule: "0 18 * * *", Settings: codecbase.Settings{Power: "on", Temperature: "22"}},
			{Schedule: "0 7 * * *", Settings: codecbase.Settings{Power: "off"}},
		},
	}
	if err := CreateJobSet(definition); err != nil {
		t.Fatalf("CreateJobSet failed: %v", err)
	}

	got, err := GetJobSetDefinition("Home")
	if err != nil {
		t.Fatalf("GetJobSetDefinition failed: %v", err)
	}
	if got.Active {
		t.Error("a new jobset must be inactive")
	}
	if schedules := cronJobSchedules(got.CronJobs); !reflect.DeepEqual(schedules, []string{"0 7 * * *", "0 18 * * *"}) {
		t.Errorf("cronjobs not ordered by time: %v", schedules)
	}
	if got.CronJobs[1].Name != "Evening" || got.CronJobs[1].Settings.Temperature != "22" {
		t.Errorf("name or settings not round-tripped: %+v", got.CronJobs[1])
	}

	if err := CreateJobSet(definition); !errors.Is(err, ErrExists) {
		t.Errorf("duplicate: got %v, want ErrExists", err)
	}
	if _, err := GetJobSetDefinition("Missing"); !errors.Is(err, ErrNotFound) {
		t.Errorf("missing: got %v, want ErrNotFound", err)
	}
}

func TestReplaceJobSetKeepsActiveAndRenames(t *testing.T) {
	initializeTestDb(t)

	if err := CreateJobSet(JobSetDefinition{Name: "Home", CronJobs: []CronJobDefinition{{Schedule: "0 7 * * *"}, {Schedule: "0 8 * * *"}}}); err != nil {
		t.Fatal(err)
	}
	if err := CreateJobSet(JobSetDefinition{Name: "Away"}); err != nil {
		t.Fatal(err)
	}
	if err := UpdateJobSet("Home", true); err != nil {
		t.Fatal(err)
	}

	active, err := ReplaceJobSet("Home", JobSetDefinition{Name: "House", Active: false, CronJobs: []CronJobDefinition{{Name: "Only", Schedule: "0 9 * * *"}}})
	if err != nil {
		t.Fatalf("ReplaceJobSet failed: %v", err)
	}
	if !active {
		t.Error("active state must be kept and returned")
	}

	if _, err := GetJobSetDefinition("Home"); !errors.Is(err, ErrNotFound) {
		t.Errorf("old name still exists: %v", err)
	}
	got, err := GetJobSetDefinition("House")
	if err != nil {
		t.Fatal(err)
	}
	if !got.Active || len(got.CronJobs) != 1 || got.CronJobs[0].Name != "Only" {
		t.Errorf("unexpected jobset after replace: %+v", got)
	}

	cronJobs, err := GetCronJobs("Home")
	if err != nil {
		t.Fatal(err)
	}
	if len(*cronJobs) != 0 {
		t.Errorf("cronjobs left under the old name: %d", len(*cronJobs))
	}

	if _, err := ReplaceJobSet("House", JobSetDefinition{Name: "Away"}); !errors.Is(err, ErrExists) {
		t.Errorf("rename collision: got %v, want ErrExists", err)
	}
	if _, err := ReplaceJobSet("Missing", JobSetDefinition{Name: "Missing"}); !errors.Is(err, ErrNotFound) {
		t.Errorf("missing: got %v, want ErrNotFound", err)
	}
	if _, err := ReplaceJobSet("House", JobSetDefinition{Name: "House", CronJobs: []CronJobDefinition{{Schedule: "nonsense"}}}); !errors.Is(err, ErrInvalid) {
		t.Errorf("invalid: got %v, want ErrInvalid", err)
	}
	// a failed replace leaves the jobset untouched
	if got, _ := GetJobSetDefinition("House"); len(got.CronJobs) != 1 {
		t.Errorf("failed replace changed the jobset: %+v", got)
	}
}

func TestDeleteJobSet(t *testing.T) {
	initializeTestDb(t)

	if err := CreateJobSet(JobSetDefinition{Name: "Home", CronJobs: []CronJobDefinition{{Schedule: "0 7 * * *"}}}); err != nil {
		t.Fatal(err)
	}
	if err := CreateJobSet(JobSetDefinition{Name: "Away", CronJobs: []CronJobDefinition{{Schedule: "0 8 * * *"}}}); err != nil {
		t.Fatal(err)
	}

	if err := DeleteJobSet("Home"); err != nil {
		t.Fatalf("DeleteJobSet failed: %v", err)
	}
	if err := DeleteJobSet("Home"); !errors.Is(err, ErrNotFound) {
		t.Errorf("second delete: got %v, want ErrNotFound", err)
	}

	cronJobs, _ := GetCronJobs("Home")
	if len(*cronJobs) != 0 {
		t.Error("cronjobs of the deleted jobset remain")
	}
	if away, err := GetJobSetDefinition("Away"); err != nil || len(away.CronJobs) != 1 {
		t.Errorf("other jobset affected: %+v, %v", away, err)
	}

	// the name can be reused
	if err := CreateJobSet(JobSetDefinition{Name: "Home"}); err != nil {
		t.Errorf("recreate failed: %v", err)
	}
}

func TestReplaceJobSetsRoundTripsThroughDefinitions(t *testing.T) {
	initializeTestDb(t)

	imported := []JobSetDefinition{
		{Name: "A", Active: true, CronJobs: []CronJobDefinition{{Name: "n", Schedule: "0 7 * * *", Settings: codecbase.Settings{Mode: "cool"}}}},
		{Name: "B", Active: false},
	}
	if err := ReplaceJobSets(imported); err != nil {
		t.Fatal(err)
	}

	exported, err := GetJobSetDefinitions()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(exported[0], imported[0]) {
		t.Errorf("A: got %+v, want %+v", exported[0], imported[0])
	}
	if exported[1].Name != "B" || exported[1].Active || len(exported[1].CronJobs) != 0 {
		t.Errorf("B: unexpected %+v", exported[1])
	}
}
