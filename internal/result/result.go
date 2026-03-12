package result

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/LobovVit/CompareFK/internal/config"
	"github.com/LobovVit/CompareFK/pkg/db"
)

var Res *Result

type Result struct {
	DateTimeFolder string
	StartedAt      time.Time
	FinishedAt     time.Time
	RunStatus      string
	Tasks          map[string]*TaskState
	Order          []string
	mu             sync.RWMutex
}

type ScriptStat struct {
	StartTime time.Time
	EndTime   time.Time
	Count     int
}

type TaskState struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Phase     string    `json:"phase"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Rows      int       `json:"rows"`
	TotalRows int       `json:"total_rows"`
	Message   string    `json:"message,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Snapshot struct {
	GeneratedAt        time.Time  `json:"generated_at"`
	RunStatus          string     `json:"run_status"`
	StartedAt          time.Time  `json:"started_at"`
	FinishedAt         time.Time  `json:"finished_at"`
	Mode               string     `json:"mode"`
	OutputDir          string     `json:"output_dir"`
	CurrentRunDir      string     `json:"current_run_dir"`
	TotalTasks         int        `json:"total_tasks"`
	RunningTasks       int        `json:"running_tasks"`
	CompletedTasks     int        `json:"completed_tasks"`
	FailedTasks        int        `json:"failed_tasks"`
	PendingTasks       int        `json:"pending_tasks"`
	TotalRowsProcessed int        `json:"total_rows_processed"`
	KnownTotalRows     int        `json:"known_total_rows"`
	OverallProgressPct float64    `json:"overall_progress_pct"`
	RowsPerSec         float64    `json:"rows_per_sec"`
	ETA                string     `json:"eta"`
	CurrentSQL         []string   `json:"current_sql"`
	Elapsed            string     `json:"elapsed"`
	Tasks              []TaskView `json:"tasks"`
}

type TaskView struct {
	Name        string  `json:"name"`
	Status      string  `json:"status"`
	Phase       string  `json:"phase"`
	Rows        int     `json:"rows"`
	TotalRows   int     `json:"total_rows"`
	ProgressPct float64 `json:"progress_pct"`
	RowsPerSec  float64 `json:"rows_per_sec"`
	ETA         string  `json:"eta"`
	Started     string  `json:"started"`
	Ended       string  `json:"ended"`
	Duration    string  `json:"duration"`
	Message     string  `json:"message,omitempty"`
	UpdatedAt   string  `json:"updated_at"`
}

func Initialize(_ string) error {
	base := config.Cfg.OutputDir
	if err := os.MkdirAll(base, os.ModePerm); err != nil {
		return fmt.Errorf("mkdir %v: %w", base, err)
	}

	startedAt := time.Now()
	Res = &Result{
		DateTimeFolder: filepath.Join(base, startedAt.Format("2006_01_02_15_04_05")),
		StartedAt:      startedAt,
		RunStatus:      "running",
		Tasks:          make(map[string]*TaskState),
		Order:          make([]string, 0, 32),
	}
	if err := os.MkdirAll(Res.DateTimeFolder, os.ModePerm); err != nil {
		return fmt.Errorf("mkdir %v: %w", Res.DateTimeFolder, err)
	}
	for _, sub := range []string{"sql", "work"} {
		path := filepath.Join(Res.DateTimeFolder, sub)
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return fmt.Errorf("mkdir %v: %w", path, err)
		}
	}
	return nil
}

func (r *Result) ensureTask(name string) *TaskState {
	if _, ok := r.Tasks[name]; !ok {
		r.Order = append(r.Order, name)
		r.Tasks[name] = &TaskState{Name: name}
	}
	return r.Tasks[name]
}

func (r *Result) StartTask(name, phase string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	t := r.ensureTask(name)
	t.Name = name
	t.Phase = phase
	t.Status = "running"
	if t.StartTime.IsZero() {
		t.StartTime = now
	}
	t.EndTime = time.Time{}
	t.Rows = 0
	t.Message = ""
	t.UpdatedAt = now
}

func (r *Result) SetTaskTotal(name string, total int) {
	if total <= 0 {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	t := r.ensureTask(name)
	if total > t.TotalRows {
		t.TotalRows = total
	}
	t.UpdatedAt = time.Now()
}

func (r *Result) AddRows(name string, delta int) {
	if delta == 0 {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	t := r.ensureTask(name)
	if t.Status == "" {
		t.Status = "running"
	}
	if t.StartTime.IsZero() {
		t.StartTime = time.Now()
	}
	t.Rows += delta
	t.UpdatedAt = time.Now()
}

func (r *Result) FinishTask(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if t, ok := r.Tasks[name]; ok {
		now := time.Now()
		t.Status = "done"
		t.EndTime = now
		if t.TotalRows == 0 || t.Rows > t.TotalRows {
			t.TotalRows = t.Rows
		}
		t.UpdatedAt = now
	}
}

func (r *Result) FailTask(name string, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if t, ok := r.Tasks[name]; ok {
		now := time.Now()
		t.Status = "failed"
		t.EndTime = now
		t.UpdatedAt = now
		if err != nil {
			t.Message = db.RedactText(err.Error())
		}
	}
}

func (r *Result) AddStat(name string, stat ScriptStat) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t := r.ensureTask(name)
	t.Name = name
	if t.Phase == "" {
		t.Phase = inferPhase(name)
	}
	t.StartTime = stat.StartTime
	t.EndTime = stat.EndTime
	t.Rows = stat.Count
	if t.TotalRows == 0 || stat.Count > t.TotalRows {
		t.TotalRows = stat.Count
	}
	t.Status = "done"
	t.UpdatedAt = time.Now()
}

func (r *Result) FinishRun() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.RunStatus = "done"
	r.FinishedAt = time.Now()
}

func (r *Result) FailRun(err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.RunStatus = "failed"
	r.FinishedAt = time.Now()
	if err != nil {
		t := r.ensureTask("run_error")
		t.Status = "failed"
		t.Phase = "system"
		t.StartTime = r.StartedAt
		t.EndTime = r.FinishedAt
		t.Message = db.RedactText(err.Error())
		t.UpdatedAt = r.FinishedAt
	}
}

func (r *Result) Snapshot() Snapshot {
	r.mu.RLock()
	defer r.mu.RUnlock()

	now := time.Now()
	elapsedUntil := now
	if !r.FinishedAt.IsZero() {
		elapsedUntil = r.FinishedAt
	}
	elapsedSeconds := elapsedUntil.Sub(r.StartedAt).Seconds()
	if elapsedSeconds <= 0 {
		elapsedSeconds = 1
	}

	tasks := make([]TaskView, 0, len(r.Order))
	currentSQL := make([]string, 0, 8)
	var running, done, failed, pending, totalRows, knownTotalRows, knownDoneRows int

	for _, name := range r.Order {
		t := r.Tasks[name]
		if t == nil {
			continue
		}
		totalRows += t.Rows
		if t.TotalRows > 0 {
			knownTotalRows += t.TotalRows
			if t.Rows > t.TotalRows {
				knownDoneRows += t.TotalRows
			} else {
				knownDoneRows += t.Rows
			}
		}
		switch t.Status {
		case "running":
			running++
			currentSQL = append(currentSQL, t.Name)
		case "done":
			done++
		case "failed":
			failed++
		default:
			pending++
		}
		durationSeconds := durationSeconds(t.StartTime, t.EndTime, t.Status)
		rowsPerSec := 0.0
		if durationSeconds > 0 {
			rowsPerSec = float64(t.Rows) / durationSeconds
		}
		progressPct := 0.0
		eta := ""
		if t.TotalRows > 0 {
			progressPct = clampPercent(float64(minInt(t.Rows, t.TotalRows)) / float64(t.TotalRows) * 100)
			if t.Rows < t.TotalRows && rowsPerSec > 0 {
				eta = time.Duration(float64(time.Second) * (float64(t.TotalRows-t.Rows) / rowsPerSec)).Round(time.Second).String()
			}
		} else if t.Status == "done" {
			progressPct = 100
		}
		tasks = append(tasks, TaskView{
			Name:        t.Name,
			Status:      t.Status,
			Phase:       t.Phase,
			Rows:        t.Rows,
			TotalRows:   t.TotalRows,
			ProgressPct: progressPct,
			RowsPerSec:  round2(rowsPerSec),
			ETA:         eta,
			Started:     formatTime(t.StartTime),
			Ended:       formatTime(t.EndTime),
			Duration:    formatDuration(t.StartTime, t.EndTime, t.Status),
			Message:     t.Message,
			UpdatedAt:   formatTime(t.UpdatedAt),
		})
	}

	rowsPerSec := round2(float64(totalRows) / elapsedSeconds)
	overallProgress := 0.0
	eta := ""
	if knownTotalRows > 0 {
		overallProgress = clampPercent(float64(knownDoneRows) / float64(knownTotalRows) * 100)
		if knownDoneRows < knownTotalRows && rowsPerSec > 0 {
			eta = time.Duration(float64(time.Second) * (float64(knownTotalRows-knownDoneRows) / rowsPerSec)).Round(time.Second).String()
		}
	}

	return Snapshot{
		GeneratedAt:        now,
		RunStatus:          r.RunStatus,
		StartedAt:          r.StartedAt,
		FinishedAt:         r.FinishedAt,
		Mode:               config.Cfg.Mode,
		OutputDir:          config.Cfg.OutputDir,
		CurrentRunDir:      r.DateTimeFolder,
		TotalTasks:         len(tasks),
		RunningTasks:       running,
		CompletedTasks:     done,
		FailedTasks:        failed,
		PendingTasks:       pending,
		TotalRowsProcessed: totalRows,
		KnownTotalRows:     knownTotalRows,
		OverallProgressPct: overallProgress,
		RowsPerSec:         rowsPerSec,
		ETA:                eta,
		CurrentSQL:         currentSQL,
		Elapsed:            elapsedUntil.Sub(r.StartedAt).Round(time.Second).String(),
		Tasks:              tasks,
	}
}

func (r *Result) GetResult() []string {
	snap := r.Snapshot()
	res := make([]string, 0, len(snap.Tasks)+1)
	res = append(res,
		"|"+fmt.Sprintf("%-*v|", 30, "!Выполняемое действие")+
			fmt.Sprintf("%*v|", 12, "Статус")+
			fmt.Sprintf("%*v|", 12, "Фаза")+
			fmt.Sprintf("%*v|", 20, "Старт")+
			fmt.Sprintf("%*v|", 20, "Стоп")+
			fmt.Sprintf("%*v|", 20, "Длительность")+
			fmt.Sprintf("%*v|", 15, "Кол-во")+
			fmt.Sprintf("%*v|", 15, "Всего")+
			fmt.Sprintf("%*v|", 10, "%")+
			fmt.Sprintf("%*v|", 12, "rows/s")+
			fmt.Sprintf("%*v|", 12, "ETA"),
	)
	for _, t := range snap.Tasks {
		res = append(res, "|"+fmt.Sprintf("%-*v|", 30, t.Name)+
			fmt.Sprintf("%*v|", 12, t.Status)+
			fmt.Sprintf("%*v|", 12, t.Phase)+
			fmt.Sprintf("%*v|", 20, t.Started)+
			fmt.Sprintf("%*v|", 20, t.Ended)+
			fmt.Sprintf("%*v|", 20, t.Duration)+
			fmt.Sprintf("%*v|", 15, t.Rows)+
			fmt.Sprintf("%*v|", 15, zeroToBlank(t.TotalRows))+
			fmt.Sprintf("%*.1f|", 10, t.ProgressPct)+
			fmt.Sprintf("%*.2f|", 12, t.RowsPerSec)+
			fmt.Sprintf("%*v|", 12, t.ETA))
	}
	slices.Sort(res[1:])
	return res
}

func (r *Result) GetResultString(step string) string {
	res := strings.Join(r.GetResult(), "\r\n")
	return fmt.Sprintf("step results: %v \r\n %v", step, res)
}

func inferPhase(name string) string {
	switch {
	case strings.HasPrefix(name, "slave"):
		return "slave"
	case strings.HasPrefix(name, "z_compute"):
		return "result"
	case strings.HasPrefix(name, "run_"):
		return "system"
	default:
		return "master"
	}
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("02-01 15:04:05")
}

func formatDuration(start, end time.Time, status string) string {
	if start.IsZero() {
		return ""
	}
	if end.IsZero() && status == "running" {
		return time.Since(start).Round(time.Second).String()
	}
	if end.IsZero() {
		return ""
	}
	return end.Sub(start).Round(time.Second).String()
}

func durationSeconds(start, end time.Time, status string) float64 {
	if start.IsZero() {
		return 0
	}
	if end.IsZero() && status == "running" {
		return time.Since(start).Seconds()
	}
	if end.IsZero() {
		return 0
	}
	return end.Sub(start).Seconds()
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func clampPercent(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return round2(v)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func zeroToBlank(v int) any {
	if v == 0 {
		return ""
	}
	return v
}
