package endpoints

import (
	"net/http"
	"runtime"
	"strconv"
)

// mem (/mem) returns a list of memory related information of the program.
func (s server) mem(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// First run the GC so that we get a more accurate view.
	runtime.GC()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	writeJSON(w, map[string]interface{}{
		"heap_alloc":    m.HeapAlloc,
		"heap_alloc_mb": strconv.Itoa(int(m.HeapAlloc/1024/1024)) + "MB",
		"sys_mem":       m.Sys,
		"sys_mem_mb":    strconv.Itoa(int(m.Sys/1024/1024)) + "MB",
	})
}

// uptime (/uptime) returns the time passed since the server started.
func (s server) uptime(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	uptime := s.s.Uptime()
	writeJSON(w, map[string]interface{}{
		"nanoseconds": uptime.Nanoseconds(),
		"seconds":     uptime.Seconds(),
		"minutes":     uptime.Minutes(),
		"hours":       uptime.Hours(),
		"string":      uptime.String(),
	})
}
