package web

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/LobovVit/CompareFK/internal/config"
	"github.com/LobovVit/CompareFK/internal/result"
	"github.com/LobovVit/CompareFK/pkg/logger"
	"go.uber.org/zap"
)

type Server struct {
	httpServer *http.Server
}

type pageData struct {
	RefreshSec int
	Snapshot   result.Snapshot
}

func Start(ctx context.Context) (*Server, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/api/status", handleStatus)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok"))
	})

	srv := &http.Server{
		Addr:         config.Cfg.WebListen,
		Handler:      mux,
		ReadTimeout:  time.Duration(config.Cfg.WebReadTimeout) * time.Second,
		WriteTimeout: time.Duration(config.Cfg.WebWriteTimeout) * time.Second,
	}

	ws := &Server{httpServer: srv}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Log.Warn("web server shutdown", zap.Error(err))
		}
	}()
	go func() {
		logger.Log.Info("web monitor started", zap.String("listen", config.Cfg.WebListen))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Error("web server failed", zap.Error(err))
		}
	}()
	return ws, nil
}

func handleStatus(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(result.Res.Snapshot())
}

func handleIndex(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := indexTemplate.Execute(w, pageData{RefreshSec: config.Cfg.WebRefreshSec, Snapshot: result.Res.Snapshot()}); err != nil {
		http.Error(w, fmt.Sprintf("template error: %v", err), http.StatusInternalServerError)
	}
}

var indexTemplate = template.Must(template.New("index").Parse(`<!doctype html>
<html lang="ru">
<head>
  <meta charset="utf-8">
  <meta http-equiv="refresh" content="{{.RefreshSec}}">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>CompareFK monitor</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; margin: 24px; color: #1f2937; }
    h1,h2 { margin: 0 0 12px 0; }
    .muted { color: #6b7280; }
    .cards { display: grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); gap: 12px; margin: 16px 0 24px; }
    .card { border: 1px solid #e5e7eb; border-radius: 12px; padding: 14px; background: #fff; box-shadow: 0 1px 2px rgba(0,0,0,.04); }
    .label { color: #6b7280; font-size: 12px; margin-bottom: 6px; }
    .value { font-size: 24px; font-weight: 600; }
    table { width: 100%; border-collapse: collapse; font-size: 14px; }
    th, td { padding: 10px 8px; border-bottom: 1px solid #e5e7eb; vertical-align: top; text-align: left; }
    th { position: sticky; top: 0; background: #f9fafb; }
    .status-running { color: #1d4ed8; font-weight: 600; }
    .status-done { color: #15803d; font-weight: 600; }
    .status-failed { color: #b91c1c; font-weight: 600; }
    .status-pending { color: #6b7280; font-weight: 600; }
    .mono { font-family: ui-monospace, SFMono-Regular, Menlo, monospace; }
    .toolbar { margin-bottom: 14px; display: flex; gap: 16px; align-items: center; flex-wrap: wrap; }
    .pill { padding: 6px 10px; border-radius: 999px; background: #eef2ff; color: #3730a3; font-size: 13px; }
  </style>
</head>
<body>
  <h1>CompareFK monitor</h1>
  <div class="toolbar">
    <div class="pill">Статус: <strong>{{.Snapshot.RunStatus}}</strong></div>
    <div class="pill">Режим: <strong>{{.Snapshot.Mode}}</strong></div>
    <div class="pill">Обновление: каждые {{.RefreshSec}} сек</div>
    <div class="muted">JSON: <a href="/api/status">/api/status</a></div>
  </div>

  <div class="cards">
    <div class="card"><div class="label">Время запуска</div><div class="value" style="font-size:16px">{{.Snapshot.StartedAt.Format "2006-01-02 15:04:05"}}</div></div>
    <div class="card"><div class="label">Прошло</div><div class="value">{{.Snapshot.Elapsed}}</div></div>
    <div class="card"><div class="label">Всего задач</div><div class="value">{{.Snapshot.TotalTasks}}</div></div>
    <div class="card"><div class="label">Выполняются</div><div class="value">{{.Snapshot.RunningTasks}}</div></div>
    <div class="card"><div class="label">Завершены</div><div class="value">{{.Snapshot.CompletedTasks}}</div></div>
    <div class="card"><div class="label">Ошибки</div><div class="value">{{.Snapshot.FailedTasks}}</div></div>
    <div class="card"><div class="label">Обработано строк</div><div class="value">{{.Snapshot.TotalRowsProcessed}}</div></div>
  </div>

  <div class="muted" style="margin-bottom:12px;">Каталог текущего запуска: <span class="mono">{{.Snapshot.CurrentRunDir}}</span></div>

  <h2>Состояние выполнения</h2>
  <table>
    <thead>
      <tr>
        <th>Действие</th>
        <th>Фаза</th>
        <th>Статус</th>
        <th>Строк</th>
        <th>Старт</th>
        <th>Стоп</th>
        <th>Длительность</th>
        <th>Обновлено</th>
        <th>Сообщение</th>
      </tr>
    </thead>
    <tbody>
      {{range .Snapshot.Tasks}}
      <tr>
        <td class="mono">{{.Name}}</td>
        <td>{{.Phase}}</td>
        <td class="status-{{.Status}}">{{.Status}}</td>
        <td>{{.Rows}}</td>
        <td>{{.Started}}</td>
        <td>{{.Ended}}</td>
        <td>{{.Duration}}</td>
        <td>{{.UpdatedAt}}</td>
        <td>{{.Message}}</td>
      </tr>
      {{end}}
    </tbody>
  </table>
</body>
</html>`))
