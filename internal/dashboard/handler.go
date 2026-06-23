package dashboard

import (
	"context"
	"encoding/json"
	"html/template"
	"net/http"
	"time"

	"github.com/mayanksekhar/GoLang-EKS-Monitor-app/internal/metrics"
)

type Handler struct {
	collector *metrics.Collector
	tmpl      *template.Template
}

func NewHandler(collector *metrics.Collector) *Handler {
	tmpl := template.Must(template.New("dashboard").Parse(dashboardHTML))
	return &Handler{collector: collector, tmpl: tmpl}
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	data, err := h.collector.Collect(ctx)
	if err != nil {
		http.Error(w, "failed to collect metrics: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	h.tmpl.Execute(w, data)
}

func (h *Handler) APIMetrics(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	data, err := h.collector.Collect(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta http-equiv="refresh" content="30">
  <title>EKS Monitor — Thinkwerke</title>
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body { background: #1F2123; color: #F0F0F0; font-family: 'Courier New', monospace; padding: 2rem; }
    header { border-bottom: 2px solid #E6890A; padding-bottom: 1rem; margin-bottom: 2rem; }
    header h1 { color: #E6890A; font-size: 1.6rem; letter-spacing: 2px; }
    header p { color: #5B6066; font-size: 0.85rem; margin-top: 0.3rem; }
    .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap: 1.5rem; margin-bottom: 2rem; }
    .card { background: #2B2D30; border-radius: 8px; padding: 1.5rem; border-left: 4px solid #E6890A; }
    .card h2 { color: #E6890A; font-size: 0.8rem; letter-spacing: 2px; margin-bottom: 1rem; text-transform: uppercase; }
    .stat { display: flex; justify-content: space-between; margin-bottom: 0.6rem; font-size: 0.9rem; }
    .stat .label { color: #9DA1A5; }
    .stat .value { color: #F0F0F0; font-weight: bold; }
    .bar-wrap { background: #1F2123; border-radius: 4px; height: 6px; margin-top: 4px; margin-bottom: 12px; }
    .bar { height: 6px; border-radius: 4px; background: #E6890A; }
    .bar.warn { background: #C9A227; }
    .bar.crit { background: #B3261E; }
    .node-card { background: #2B2D30; border-radius: 8px; padding: 1.2rem; margin-bottom: 1rem; }
    .node-name { color: #E6890A; font-size: 0.85rem; margin-bottom: 0.8rem; }
    footer { color: #3A3D40; font-size: 0.75rem; margin-top: 2rem; border-top: 1px solid #2B2D30; padding-top: 1rem; }
  </style>
</head>
<body>
  <header>
    <h1>⬡ THINKWERKE // EKS MONITOR</h1>
    <p>Cluster: {{ .ClusterName }} &nbsp;|&nbsp; Auto-refresh: 30s</p>
  </header>

  <div class="grid">
    <div class="card">
      <h2>Cluster Overview</h2>
      <div class="stat"><span class="label">Nodes</span><span class="value">{{ len .Nodes }}</span></div>
      <div class="stat"><span class="label">Avg CPU</span><span class="value">{{ printf "%.1f" .TotalCPU }}%</span></div>
      <div class="bar-wrap"><div class="bar {{ if gt .TotalCPU 80.0 }}crit{{ else if gt .TotalCPU 60.0 }}warn{{ end }}" style="width:{{ printf "%.0f" .TotalCPU }}%"></div></div>
      <div class="stat"><span class="label">Avg Memory</span><span class="value">{{ printf "%.1f" .TotalMem }}%</span></div>
      <div class="bar-wrap"><div class="bar {{ if gt .TotalMem 80.0 }}crit{{ else if gt .TotalMem 60.0 }}warn{{ end }}" style="width:{{ printf "%.0f" .TotalMem }}%"></div></div>
    </div>

    <div class="card">
      <h2>Workloads</h2>
      <div class="stat"><span class="label">Total Pods</span><span class="value">{{ .TotalPods }}</span></div>
      <div class="stat"><span class="label">Running</span><span class="value" style="color:#4CAF50">{{ .RunningPods }}</span></div>
      <div class="stat"><span class="label">Pending</span><span class="value" style="color:#C9A227">{{ .PendingPods }}</span></div>
      <div class="stat"><span class="label">Namespaces</span><span class="value">{{ .Namespaces }}</span></div>
    </div>
  </div>

  <h2 style="color:#E6890A; font-size:0.8rem; letter-spacing:2px; margin-bottom:1rem;">NODE DETAILS</h2>
  {{ range .Nodes }}
  <div class="node-card">
    <div class="node-name">▸ {{ .Name }} &nbsp;<span style="color:#5B6066;font-size:0.75rem">[{{ .Role }}]</span></div>
    <div class="stat"><span class="label">CPU</span><span class="value">{{ printf "%.1f" .CPUPercent }}% &nbsp;({{ .CPUUsage }}m / {{ .CPUCapacity }}m)</span></div>
    <div class="bar-wrap"><div class="bar {{ if gt .CPUPercent 80.0 }}crit{{ else if gt .CPUPercent 60.0 }}warn{{ end }}" style="width:{{ printf "%.0f" .CPUPercent }}%"></div></div>
    <div class="stat"><span class="label">Memory</span><span class="value">{{ printf "%.1f" .MemPercent }}% &nbsp;({{ .MemUsage }}MB / {{ .MemCapacity }}MB)</span></div>
    <div class="bar-wrap"><div class="bar {{ if gt .MemPercent 80.0 }}crit{{ else if gt .MemPercent 60.0 }}warn{{ end }}" style="width:{{ printf "%.0f" .MemPercent }}%"></div></div>
  </div>
  {{ end }}

  <footer>Thinkwerke DevSecOps Portfolio &nbsp;|&nbsp; Runtime security powered by Falco &nbsp;|&nbsp; Supply chain secured by Cosign</footer>
</body>
</html>`
