package console

import (
	"html/template"
	"net/http"
)

type Page struct {
	Path        string
	Title       string
	API         string
	Description string
}

var pages = []Page{
	{Path: "/ventilation", Title: "Ventilation control", API: "/api/ventilation", Description: "Fan groups, door isolation, airflow topology and operating commands."},
	{Path: "/gas-zones", Title: "Gas zones", API: "/api/gas-zones", Description: "Current methane readings, calibration epochs and safety alarms."},
	{Path: "/extraction", Title: "Extraction sessions", API: "/api/extraction", Description: "Borehole groups, negative-pressure sessions and valve ownership."},
	{Path: "/incidents", Title: "Incident coordination", API: "/api/incidents", Description: "Evacuation rounds, door states and air-clearance decisions."},
}

var pageTemplate = template.Must(template.New("page").Parse(`<!doctype html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>{{.Title}} | MineAir</title>
<style>body{font-family:system-ui,sans-serif;margin:0;background:#f1f5f9;color:#132238}header{background:#0b3b4f;color:#fff;padding:20px 28px}main{max-width:1100px;margin:28px auto;padding:0 20px}.nav{display:flex;gap:16px;flex-wrap:wrap;margin:20px 0}.nav a{color:#0b526b}.panel{background:#fff;border:1px solid #cbd5e1;border-radius:6px;padding:22px;margin-top:20px}pre{white-space:pre-wrap;word-break:break-word;background:#0f172a;color:#dbeafe;padding:16px;border-radius:4px;min-height:220px}</style></head>
<body><header><strong>MineAir safety console</strong><div>{{.Title}}</div></header><main><nav class="nav">{{range .Links}}<a href="{{.Path}}">{{.Title}}</a>{{end}}</nav><section class="panel"><h1>{{.Title}}</h1><p>{{.Description}}</p><pre id="data">Loading live state...</pre></section></main><script>fetch('{{.API}}').then(r=>r.json()).then(v=>document.getElementById('data').textContent=JSON.stringify(v,null,2)).catch(e=>document.getElementById('data').textContent=e.toString())</script></body></html>`))

func Pages() []Page {
	result := make([]Page, len(pages))
	copy(result, pages)
	return result
}

func Handler(page Page) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = pageTemplate.Execute(w, struct {
			Page
			Links []Page
		}{Page: page, Links: pages})
	})
}
