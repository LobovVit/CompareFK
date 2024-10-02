package result

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

var Res *Result

type Result struct {
	DateTimeFolder string
	StatRows       map[string]ScriptStat
}

type ScriptStat struct {
	StartTime time.Time
	EndTime   time.Time
	Count     int
}

func Initialize() error {
	Res = &Result{DateTimeFolder: time.Now().Format("2006_01_02_15_04_05"),
		StatRows: make(map[string]ScriptStat)}
	if err := os.Mkdir(Res.DateTimeFolder, os.ModePerm); err != nil {
		return fmt.Errorf("mkdir %v: %w", Res.DateTimeFolder, err)
	}
	path := filepath.Join(Res.DateTimeFolder, "sql")
	if err := os.Mkdir(path, os.ModePerm); err != nil {
		return fmt.Errorf("mkdir %v: %w", Res.DateTimeFolder, err)
	}
	return nil
}

func (r *Result) GetResult() []string {
	res := make([]string, 0, len(r.StatRows))
	res = append(res, "|"+fmt.Sprintf("%-*v|", 30, "!Выполняемое действие")+
		fmt.Sprintf("%*v|", 20, "Старт")+
		fmt.Sprintf("%*v|", 20, "Стоп")+
		fmt.Sprintf("%*v|", 20, "Длительность")+
		fmt.Sprintf("%*v|", 15, "Кол-во"))
	for k, v := range r.StatRows {
		res = append(res, "|"+fmt.Sprintf("%-*v|", 30, k)+
			fmt.Sprintf("%*v|", 20, v.StartTime.Format("02-05 15:04:05"))+
			fmt.Sprintf("%*v|", 20, v.EndTime.Format("02-05 15:04:05"))+
			fmt.Sprintf("%*v|", 20, v.EndTime.Sub(v.StartTime).Round(time.Second).String())+
			fmt.Sprintf("%*v|", 15, v.Count))
	}
	slices.Sort(res)
	return res
}

func (r *Result) GetResultString(step string) string {
	res := strings.Join(r.GetResult()[0:], "\r\n")
	res = fmt.Sprintf("step results: %v \r\n %v", step, res)
	return res
}
