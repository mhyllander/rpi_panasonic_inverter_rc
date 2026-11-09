package db

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"

	"rpi_panasonic_inverter_rc/codecbase"
)

var (
	ErrNotFound = errors.New("not found")
	ErrExists   = errors.New("already exists")
	ErrInvalid  = errors.New("invalid")
)

const maxNameLength = 100

type CronJobDefinition struct {
	Name     string
	Schedule string
	Settings codecbase.Settings
}

type JobSetDefinition struct {
	Name     string
	Active   bool
	CronJobs []CronJobDefinition
}

// GetJobSetDefinitions returns all job sets ordered by name, with cron jobs ordered as described by SortCronJobs.
func GetJobSetDefinitions() ([]JobSetDefinition, error) {
	var jobsets []JobSet
	if err := myDb.Order("name").Find(&jobsets).Error; err != nil {
		return nil, err
	}

	var cronjobs []CronJob
	if err := myDb.Order("id").Find(&cronjobs).Error; err != nil {
		return nil, err
	}
	cronjobsByJobSet := make(map[string][]CronJob)
	for _, cronjob := range cronjobs {
		cronjobsByJobSet[cronjob.JobSet] = append(cronjobsByJobSet[cronjob.JobSet], cronjob)
	}

	definitions := make([]JobSetDefinition, 0, len(jobsets))
	for _, jobset := range jobsets {
		definition, err := newJobSetDefinition(jobset, cronjobsByJobSet[jobset.Name])
		if err != nil {
			return nil, err
		}
		definitions = append(definitions, definition)
	}
	return definitions, nil
}

// GetJobSetDefinition returns the named job set, or ErrNotFound.
func GetJobSetDefinition(name string) (JobSetDefinition, error) {
	jobset, err := findJobSet(myDb, name)
	if err != nil {
		return JobSetDefinition{}, err
	}

	var cronjobs []CronJob
	if err := myDb.Where("job_set = ?", name).Order("id").Find(&cronjobs).Error; err != nil {
		return JobSetDefinition{}, err
	}
	return newJobSetDefinition(*jobset, cronjobs)
}

func newJobSetDefinition(jobset JobSet, cronjobs []CronJob) (JobSetDefinition, error) {
	definition := JobSetDefinition{Name: jobset.Name, Active: jobset.Active, CronJobs: make([]CronJobDefinition, 0, len(cronjobs))}
	for _, cronjob := range cronjobs {
		cronJobDefinition := CronJobDefinition{Name: cronjob.Name, Schedule: cronjob.Schedule}
		if err := json.Unmarshal(cronjob.Settings, &cronJobDefinition.Settings); err != nil {
			return JobSetDefinition{}, fmt.Errorf("cronjob %d in jobset %q has invalid settings: %w", cronjob.ID, jobset.Name, err)
		}
		definition.CronJobs = append(definition.CronJobs, cronJobDefinition)
	}
	SortCronJobs(definition.CronJobs)
	return definition, nil
}

func findJobSet(conn *gorm.DB, name string) (*JobSet, error) {
	// Find rather than First, because gorm logs an error when First finds nothing
	var jobsets []JobSet
	if err := conn.Where("name = ?", name).Limit(1).Find(&jobsets).Error; err != nil {
		return nil, err
	}
	if len(jobsets) == 0 {
		return nil, fmt.Errorf("jobset %q: %w", name, ErrNotFound)
	}
	return &jobsets[0], nil
}

func jobSetExists(conn *gorm.DB, name string) (bool, error) {
	var count int64
	if err := conn.Model(&JobSet{}).Where("name = ?", name).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func saveCronJobs(conn *gorm.DB, jobset string, cronJobs []CronJobDefinition) error {
	for _, cronJob := range cronJobs {
		if err := saveCronJob(conn, jobset, cronJob.Name, cronJob.Schedule, &cronJob.Settings); err != nil {
			return fmt.Errorf("failed to save cronjob %q in jobset %q: %w", cronJob.Schedule, jobset, err)
		}
	}
	return nil
}

// CreateJobSet creates an inactive job set. Active in the definition is ignored.
func CreateJobSet(definition JobSetDefinition) error {
	if err := ValidateJobSetDefinition(definition); err != nil {
		return err
	}

	return myDb.Transaction(func(tx *gorm.DB) error {
		exists, err := jobSetExists(tx, definition.Name)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("jobset %q: %w", definition.Name, ErrExists)
		}

		if err := saveJobSet(tx, definition.Name, false); err != nil {
			return fmt.Errorf("failed to save jobset %q: %w", definition.Name, err)
		}
		return saveCronJobs(tx, definition.Name, definition.CronJobs)
	})
}

// ReplaceJobSet renames the job set oldName to definition.Name and replaces its cron jobs. The active state is kept,
// and returned so that the caller can reschedule the jobs.
func ReplaceJobSet(oldName string, definition JobSetDefinition) (active bool, err error) {
	if err := ValidateJobSetDefinition(definition); err != nil {
		return false, err
	}

	err = myDb.Transaction(func(tx *gorm.DB) error {
		jobset, err := findJobSet(tx, oldName)
		if err != nil {
			return err
		}
		active = jobset.Active

		if definition.Name != oldName {
			exists, err := jobSetExists(tx, definition.Name)
			if err != nil {
				return err
			}
			if exists {
				return fmt.Errorf("jobset %q: %w", definition.Name, ErrExists)
			}
			if err := tx.Model(jobset).Update("name", definition.Name).Error; err != nil {
				return err
			}
		}

		if err := tx.Unscoped().Where("job_set = ?", oldName).Delete(&CronJob{}).Error; err != nil {
			return fmt.Errorf("failed to delete cronjobs in jobset %q: %w", oldName, err)
		}
		return saveCronJobs(tx, definition.Name, definition.CronJobs)
	})
	return active, err
}

// DeleteJobSet permanently deletes a job set and its cron jobs, or returns ErrNotFound.
func DeleteJobSet(name string) error {
	return myDb.Transaction(func(tx *gorm.DB) error {
		result := tx.Unscoped().Where("name = ?", name).Delete(&JobSet{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("jobset %q: %w", name, ErrNotFound)
		}
		return tx.Unscoped().Where("job_set = ?", name).Delete(&CronJob{}).Error
	})
}

// Validation

var (
	onOffValues     = []string{"on", "off"}
	modeValues      = []string{"auto", "heat", "cool", "dry"}
	fanValues       = []string{"auto", "lowest", "low", "middle", "high", "highest"}
	ventVertValues  = []string{"auto", "lowest", "low", "middle", "high", "highest"}
	ventHorizValues = []string{"auto", "farleft", "left", "middle", "right", "farright"}
)

// Matches what gocron accepts for a cron job without seconds.
var scheduleParser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)

// ValidateJobSetDefinition checks that the definition can be saved and scheduled. Returned errors wrap ErrInvalid.
func ValidateJobSetDefinition(definition JobSetDefinition) error {
	if err := validateName(definition.Name, "jobset name", true); err != nil {
		return err
	}
	for i, cronJob := range definition.CronJobs {
		if err := validateCronJob(cronJob); err != nil {
			return fmt.Errorf("cronjob %d: %w", i+1, err)
		}
	}
	return nil
}

func validateName(name, what string, required bool) error {
	if required && strings.TrimSpace(name) == "" {
		return fmt.Errorf("%w: %s must not be empty", ErrInvalid, what)
	}
	if name != strings.TrimSpace(name) {
		return fmt.Errorf("%w: %s must not start or end with whitespace", ErrInvalid, what)
	}
	if len([]rune(name)) > maxNameLength {
		return fmt.Errorf("%w: %s must be at most %d characters", ErrInvalid, what, maxNameLength)
	}
	return nil
}

func validateCronJob(cronJob CronJobDefinition) error {
	if err := validateName(cronJob.Name, "name", false); err != nil {
		return err
	}
	if _, err := scheduleParser.Parse(cronJob.Schedule); err != nil {
		return fmt.Errorf("%w: schedule %q: %v", ErrInvalid, cronJob.Schedule, err)
	}
	return validateSettings(cronJob.Settings)
}

func validateSettings(settings codecbase.Settings) error {
	enumFields := []struct {
		name    string
		value   string
		allowed []string
	}{
		{"power", settings.Power, onOffValues},
		{"mode", settings.Mode, modeValues},
		{"powerful", settings.Powerful, onOffValues},
		{"quiet", settings.Quiet, onOffValues},
		{"fan", settings.FanSpeed, fanValues},
		{"vert", settings.VentVertical, ventVertValues},
		{"horiz", settings.VentHorizontal, ventHorizValues},
		{"ton", settings.TimerOn, onOffValues},
		{"toff", settings.TimerOff, onOffValues},
	}
	for _, field := range enumFields {
		if field.value != "" && !contains(field.allowed, field.value) {
			return fmt.Errorf("%w: setting %s has unknown value %q", ErrInvalid, field.name, field.value)
		}
	}

	if settings.Temperature != "" {
		temperature, err := strconv.Atoi(settings.Temperature)
		if err != nil || temperature < codecbase.C_Temp_Min || temperature > codecbase.C_Temp_Max {
			return fmt.Errorf("%w: setting temp must be a number from %d to %d", ErrInvalid, codecbase.C_Temp_Min, codecbase.C_Temp_Max)
		}
	}

	for _, field := range []struct{ name, value string }{{"tont", settings.TimerOnTime}, {"tofft", settings.TimerOffTime}} {
		if field.value != "" && !isValidTime(field.value) {
			return fmt.Errorf("%w: setting %s must be a time as HH:MM", ErrInvalid, field.name)
		}
	}
	return nil
}

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func isValidTime(value string) bool {
	hourText, minuteText, found := strings.Cut(value, ":")
	if !found {
		return false
	}
	hour, hourErr := strconv.Atoi(hourText)
	minute, minuteErr := strconv.Atoi(minuteText)
	return hourErr == nil && minuteErr == nil && hour >= 0 && hour <= 23 && minute >= 0 && minute <= 59
}

// Ordering

var cronDescriptors = map[string]string{
	"@yearly":   "0 0 1 1 *",
	"@annually": "0 0 1 1 *",
	"@monthly":  "0 0 1 * *",
	"@weekly":   "0 0 * * 0",
	"@daily":    "0 0 * * *",
	"@midnight": "0 0 * * *",
	"@hourly":   "0 * * * *",
}

var cronNames = map[string]int{
	"jan": 1, "feb": 2, "mar": 3, "apr": 4, "may": 5, "jun": 6,
	"jul": 7, "aug": 8, "sep": 9, "oct": 10, "nov": 11, "dec": 12,
	"sun": 0, "mon": 1, "tue": 2, "wed": 3, "thu": 4, "fri": 5, "sat": 6,
}

// cronSortKey is the first value of each field of a cron schedule. A wildcard is -1, so that it sorts first.
type cronSortKey struct {
	month, dayOfMonth, hour, minute, dayOfWeek int
}

func (key cronSortKey) less(other cronSortKey) bool {
	for _, pair := range [][2]int{
		{key.month, other.month},
		{key.dayOfMonth, other.dayOfMonth},
		{key.hour, other.hour},
		{key.minute, other.minute},
		{key.dayOfWeek, other.dayOfWeek},
	} {
		if pair[0] != pair[1] {
			return pair[0] < pair[1]
		}
	}
	return false
}

func parseCronSortKey(schedule string) (cronSortKey, bool) {
	fields := strings.Fields(schedule)
	if len(fields) > 0 && (strings.HasPrefix(fields[0], "CRON_TZ=") || strings.HasPrefix(fields[0], "TZ=")) {
		fields = fields[1:]
	}
	if len(fields) == 1 {
		if expanded, found := cronDescriptors[fields[0]]; found {
			fields = strings.Fields(expanded)
		}
	}
	if len(fields) != 5 {
		return cronSortKey{}, false
	}

	values := make([]int, 5)
	for i, field := range fields {
		value, ok := firstCronValue(field)
		if !ok {
			return cronSortKey{}, false
		}
		values[i] = value
	}
	// fields are minute, hour, day of month, month, day of week
	return cronSortKey{month: values[3], dayOfMonth: values[2], hour: values[1], minute: values[0], dayOfWeek: values[4]}, true
}

func firstCronValue(field string) (int, bool) {
	first, _, _ := strings.Cut(field, ",")
	first, _, _ = strings.Cut(first, "/")
	first, _, _ = strings.Cut(first, "-")
	if first == "*" || first == "?" {
		return -1, true
	}
	if value, found := cronNames[strings.ToLower(first)]; found {
		return value, true
	}
	value, err := strconv.Atoi(first)
	return value, err == nil
}

// SortCronJobs orders cron jobs by month, day of month, hour, minute and day of week, using the first value of each
// schedule field. Wildcards sort first, so jobs that run every day or on weekdays come before jobs on specific dates,
// ordered by hour and minute. Jobs with a schedule that can't be interpreted come last. Equal jobs keep their order.
func SortCronJobs(cronJobs []CronJobDefinition) {
	type sortable struct {
		key   cronSortKey
		valid bool
	}
	keys := make(map[string]sortable)
	for _, cronJob := range cronJobs {
		if _, seen := keys[cronJob.Schedule]; !seen {
			key, valid := parseCronSortKey(cronJob.Schedule)
			keys[cronJob.Schedule] = sortable{key, valid}
		}
	}

	sort.SliceStable(cronJobs, func(i, j int) bool {
		a, b := keys[cronJobs[i].Schedule], keys[cronJobs[j].Schedule]
		if a.valid != b.valid {
			return a.valid
		}
		return a.key.less(b.key)
	})
}
