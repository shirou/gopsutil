package server

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

func MakeRouter() *mux.Router {
	return mux.NewRouter()
}

func AddRoutes(r *mux.Router) *mux.Router {
	r.HandleFunc("/", homeHandler).Methods("GET")
	r.HandleFunc("/memory", memoryHandler).Methods("GET")
	r.HandleFunc("/cpu", cpuHandler).Methods("GET")
	r.HandleFunc("/load", loadHandler).Methods("GET")
	r.HandleFunc("/host", hostHandler).Methods("GET")
	r.HandleFunc("/process", processHandler).Methods("GET")
	return r
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"name": "gopsserver"})
}

func memoryHandler(w http.ResponseWriter, r *http.Request) {
	if memory, err := mem.VirtualMemory(); err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(memory)
	}
}

func cpuHandler(w http.ResponseWriter, r *http.Request) {
	if c, err := cpu.Info(); err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(c)
	}
}

func loadHandler(w http.ResponseWriter, r *http.Request) {
	if l, err := load.Avg(); err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(l)
	}
}

func hostHandler(w http.ResponseWriter, r *http.Request) {
	if h, err := host.Info(); err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(h)
	}
}

type Process struct {
	Pid     int32                   `json:"pid"`
	Name    string                  `json:"name"`
	Cmdline string                  `json:"cmdline"`
	Cpu     float64                 `json:"cpu"`
	Mem     *process.MemoryInfoStat `json:"mem"`
}

func processHandler(w http.ResponseWriter, r *http.Request) {
	procs, err := process.Processes()
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	// Processes is not fully populated, only PIDs.
	// Lets fill up with names and CLI command line
	procList := make([]*Process, 0, len(procs))
	for idx := range procs {
		name, _ := procs[idx].Name()
		cmdline, _ := procs[idx].Cmdline()
		cpuTime, _ := procs[idx].CPUPercent()
		mem, _ := procs[idx].MemoryInfo()
		procList = append(procList, &Process{
			Pid:     procs[idx].Pid,
			Name:    name,
			Cmdline: cmdline,
			Cpu:     cpuTime,
			Mem:     mem,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(procList)
}
