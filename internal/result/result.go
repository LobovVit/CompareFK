package result

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/LobovVit/CompareFK/internal/config"
)

var Res *Result

type Result struct {
	DateTimeFolder string
	StatRows       map[string]ScriptStat
	mu             sync.Mutex
}

type ScriptStat struct {
	StartTime time.Time
	EndTime   time.Time
	Count     int
}

func Initialize(_ string) error {
	base := config.Cfg.OutputDir
	if err := os.MkdirAll(base, os.ModePerm); err != nil {
		return fmt.Errorf("mkdir %v: %w", base, err)
	}

	Res = &Result{
		DateTimeFolder: filepath.Join(base, time.Now().Format("2006_01_02_15_04_05")),
		StatRows:       make(map[string]ScriptStat),
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

func (r *Result) AddStat(name string, stat ScriptStat) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.StatRows[name] = stat
}

func (r *Result) GetResult() []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	res := make([]string, 0, len(r.StatRows)+1)
	res = append(res,
		"|"+fmt.Sprintf("%-*v|", 30, "!Выполняемое действие")+
			fmt.Sprintf("%*v|", 20, "Старт")+
			fmt.Sprintf("%*v|", 20, "Стоп")+
			fmt.Sprintf("%*v|", 20, "Длительность")+
			fmt.Sprintf("%*v|", 15, "Кол-во"),
	)
	for k, v := range r.StatRows {
		res = append(res, "|"+fmt.Sprintf("%-*v|", 30, k)+
			fmt.Sprintf("%*v|", 20, v.StartTime.Format("02-01 15:04:05"))+
			fmt.Sprintf("%*v|", 20, v.EndTime.Format("02-01 15:04:05"))+
			fmt.Sprintf("%*v|", 20, v.EndTime.Sub(v.StartTime).Round(time.Second).String())+
			fmt.Sprintf("%*v|", 15, v.Count))
	}
	slices.Sort(res[1:])
	return res
}

func (r *Result) GetResultString(step string) string {
	res := strings.Join(r.GetResult(), "\r\n")
	return fmt.Sprintf("step results: %v \r\n %v", step, res)
}
