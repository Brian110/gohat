package main

import (
	"fmt"
	"github.com/rubyist/gohat/pkg/heapfile"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

type gohatServer struct {
	heapFile *heapfile.HeapFile
	address  string
}

func newGohatServer(address string, heapFile *heapfile.HeapFile) *gohatServer {

	return &gohatServer{heapFile, address}
}

func (s *gohatServer) Run() {
	http.HandleFunc("/", s.mainPage)
	http.HandleFunc("/objects", s.objectsPage)
	http.HandleFunc("/object", s.objectPage)
	http.HandleFunc("/reachable", s.reachablePage)
	http.HandleFunc("/garbage", s.garbagePage)

	log.Printf("Serving %s on %s", s.heapFile.Name, s.address)
	log.Fatal(http.ListenAndServe(s.address, nil))
}

func (s *gohatServer) mainPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		log.Printf("[404] %s", r.URL)
		http.NotFound(w, r)
		return
	}

	render(w, mainTemplate, s.heapFile)
	log.Printf("[200] %s", r.URL)
}

func (s *gohatServer) objectsPage(w http.ResponseWriter, r *http.Request) {
	render(w, objectsTemplate, s.heapFile)
}

func (s *gohatServer) objectPage(w http.ResponseWriter, r *http.Request) {
	objectId := r.URL.Query().Get("id")
	addr, err := strconv.ParseInt(objectId, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	object := s.heapFile.Object(addr)
	if object == nil {
		http.NotFound(w, r)
		return
	}

	data := map[string]interface{}{
		"Name":   s.heapFile.Name,
		"Object": object,
	}

	render(w, objectTemplate, data)
}

func (s *gohatServer) reachablePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "reachable")
}

func (s *gohatServer) garbagePage(w http.ResponseWriter, r *http.Request) {
	render(w, garbageTemplate, s.heapFile)
}

func render(w http.ResponseWriter, templateString string, data interface{}) {
	funcMap := template.FuncMap{
		"hexdump": hexDump,
	}

	t := template.Must(template.New("main").Parse(bodyTemplate))
	t.New("body").Funcs(funcMap).Parse(templateString)
	t.Execute(w, data)
}

var bodyTemplate = `<html>
<head>
	<title>GoHat {{.Name}}</title>
	<style type="text/css">
		body { font-family: monospace; }
	</style>
</head>
<body>
<h1>GoHat</h1>
<a href="/">Main</a>
<a href="/objects">All Objects</a>
<a href="/reachable">Reachable Objects</a>
<a href="/garbage">Garbage Objects</a>
{{template "body" .}}
</body>
</html>
`

var mainTemplate = `
<h2>Heap Parameters</h2>
<table>
<tr><td>Endianness</td><td>{{if .DumpParams.BigEndian}}Big{{else}}Little{{end}} Endian</td></tr>
<tr><td>Pointer Size</td><td>{{.DumpParams.PtrSize}}</td></tr>
<tr><td>Heap Start Address</td><td>{{printf "0x%x" .DumpParams.StartAddress}}</td></tr>
<tr><td>End Addres</td><td>{{printf "0x%x" .DumpParams.EndAddress}}</td></tr>
<tr><td>Arch</td><td>{{.DumpParams.Arch}}</td></tr>
<tr><td>GOEXPERIMENT</td><td>{{.DumpParams.GoExperiment}}</td></tr>
<tr><td>Num CPU</td><td>{{.DumpParams.NCPU}}</td></tr>
</table>

<h2>MemStats</h2>
<table>
<tr><th colspan="2">General Statistics</th></tr>
<tr><td>Alloc</td><td>{{.MemStats.Alloc}}</td></tr>
<tr><td>TotalAlloc</td><td>{{.MemStats.TotalAlloc}}</td></tr>
<tr><td>Sys</td><td>{{.MemStats.Sys}}</td></tr>
<tr><td>Lookups</td><td>{{.MemStats.Lookups}}</td></tr>
<tr><td>Mallocs</td><td>{{.MemStats.Mallocs}}</td></tr>
<tr><td>Frees</td><td>{{.MemStats.Frees}}</td></tr>

<tr><th colspan="2">Main Allocation Heap Statistics</th></tr>
<tr><td>HeapAlloc</td><td>{{.MemStats.HeapAlloc}}</td></tr>
<tr><td>HeapSys</td><td>{{.MemStats.HeapSys}}</td></tr>
<tr><td>HeapIdle</td><td>{{.MemStats.HeapIdle}}</td></tr>
<tr><td>HeapInuse</td><td>{{.MemStats.HeapInuse}}</td></tr>
<tr><td>HeapReleased</td><td>{{.MemStats.HeapReleased}}</td></tr>
<tr><td>HeapObjects</td><td>{{.MemStats.HeapObjects}}</td></tr>

<tr><th colspan="2">Low-level fixed-size structure allocator stats</th></tr>
<tr><td>StackInuse</td><td>{{.MemStats.StackInuse}}</td></tr>
<tr><td>StackSys</td><td>{{.MemStats.StackSys}}</td></tr>
<tr><td>MSpanInuse</td><td>{{.MemStats.MSpanInuse}}</td></tr>
<tr><td>MSpanSys</td><td>{{.MemStats.MSpanSys}}</td></tr>
<tr><td>MCacheInuse</td><td>{{.MemStats.MCacheInuse}}</td></tr>
<tr><td>MCacheSys</td><td>{{.MemStats.MCacheSys}}</td></tr>
<tr><td>BuckHashSys</td><td>{{.MemStats.BuckHashSys}}</td></tr>
<tr><td>GCSys</td><td>{{.MemStats.GCSys}}</td></tr>
<tr><td>OtherSys</td><td>{{.MemStats.OtherSys}}</td></tr>

<tr><th colspan="2">GC Statistics</th></tr>
<tr><td>NextGC</td><td>{{.MemStats.NextGC}}</td></tr>
<tr><td>LastGC</td><td>{{.MemStats.LastGC}}</td></tr>
<tr><td>PauseTotalNs</td><td>{{.MemStats.PauseTotalNs}}</td></tr>
<tr><td>NumGC</td><td>{{.MemStats.NumGC}}</td></tr>
</table>
`

var objectsTemplate = `
<h2>Objects</h2>
{{range .Objects}}
<div><a href="/object?id={{.Address}}">{{printf "0x%x" .Address}} {{.Name}}</a></div>
{{end}}
`

var garbageTemplate = `
<h2>Unreachable Objects</h2>
{{range .Garbage}}
<div><a href="/object?id={{.Address}}">{{printf "0x%x" .Address}} {{.Name}}</a></div>
{{end}}
`

var objectTemplate = `
<h2>{{printf "0x%x" .Object.Address}} {{.Object.Name}}</h2>
<div>Kind: {{.Object.Kind}}</div>
<div>Size: {{.Object.Size}}</div>

<h3>Fields</h3>
{{range .Object.Fields}}
<div>{{.Kind}} {{printf "0x%.4x" .Offset}}</div>
{{end}}

<h3>Content</h3>
<pre>
{{hexdump .Object.Content}}
</pre>

<h3>Children</h3>
{{range .Object.Children}}
<div><a href="/object?id={{.Address}}">{{printf "0x%x" .Address}} {{.Name}}</a></div>
{{end}}
`
