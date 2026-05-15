package serve

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/srmdn/orbital/internal/scan"
)

//go:embed templates/*
var templatesFS embed.FS

func Run(version string) error {
	home := os.Getenv("HOME")

	mu := &sync.RWMutex{}
	ready := &atomic.Bool{}
	var dataPtr *pageData

	tmpl := template.Must(template.New("index.html").Funcs(template.FuncMap{
		"formatSize":   scan.FormatSize,
		"homeRelative": homeRelative,
		"barPct":       barPct,
	}).ParseFS(templatesFS, "templates/index.html"))

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if !ready.Load() {
			tmpl.Execute(w, &pageData{Home: home, Loading: true})
			return
		}
		mu.RLock()
		d := dataPtr
		mu.RUnlock()
		tmpl.Execute(w, d)
	})

	mux.HandleFunc("/api/scan", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if !ready.Load() {
			json.NewEncoder(w).Encode(map[string]interface{}{"loading": true})
			return
		}
		mu.RLock()
		defer mu.RUnlock()
		json.NewEncoder(w).Encode(dataPtr)
	})

	mux.HandleFunc("/api/refresh", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		ready.Store(false)
		data := scanAndBuild(home, version)
		mu.Lock()
		dataPtr = data
		mu.Unlock()
		ready.Store(true)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	})

	mux.HandleFunc("/api/clean", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		r.ParseForm()
		if r.FormValue("confirm") != "yes-delete" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "confirmation required — send confirm=yes-delete",
			})
			return
		}

		mu.RLock()
		var entries []scan.Entry
		for _, tg := range dataPtr.Tiers {
			entries = append(entries, tg.Entries...)
		}
		mu.RUnlock()

		var freed int64
		var failed int
		for _, e := range entries {
			if !e.Cleanable {
				continue
			}
			// Safety: only delete paths under home directory
			rel, err := filepath.Rel(home, e.Path)
			if err != nil || strings.HasPrefix(rel, "..") {
				failed++
				continue
			}
			if err := os.RemoveAll(e.Path); err != nil {
				failed++
			} else {
				freed += e.SizeMB
			}
		}

		ready.Store(false)
		data := scanAndBuild(home, version)
		mu.Lock()
		dataPtr = data
		mu.Unlock()
		ready.Store(true)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"freedMB": freed,
			"failed":  failed,
			"totals":  data.Totals,
		})
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	url := fmt.Sprintf("http://localhost:%d", port)

	fmt.Printf("\n  🚀 orbital dashboard → %s\n", url)
	fmt.Println("  scanning...")

	go func() {
		exec.Command("open", url).Run()
	}()

	go func() {
		data := scanAndBuild(home, version)
		mu.Lock()
		dataPtr = data
		mu.Unlock()
		ready.Store(true)
		fmt.Println("  ✅ scan complete.")
		fmt.Println("  Press Ctrl+C to stop.")
	}()

	return http.Serve(listener, mux)
}

func scanAndBuild(home, version string) *pageData {
	entries, t1, t2, t3, t4 := scan.Collect(home)
	gitFound, gitSize := scan.HasGitTrap(home)

	var maxSize int64
	for _, e := range entries {
		if e.SizeMB > maxSize {
			maxSize = e.SizeMB
		}
	}

	tierNums := []int{scan.TierSafe, scan.TierReinst, scan.TierApp, scan.TierManual}
	tierTotals := []int64{t1, t2, t3, t4}
	var groups []tierGroup
	for i, tn := range tierNums {
		var groupEntries []scan.Entry
		for _, e := range entries {
			if e.Tier == tn {
				groupEntries = append(groupEntries, e)
			}
		}
		groups = append(groups, tierGroup{
			Num:     tn,
			Label:   scan.TierLabel(tn),
			Total:   tierTotals[i],
			Entries: groupEntries,
		})
	}

	allEmpty := true
	for _, tg := range groups {
		if len(tg.Entries) > 0 {
			allEmpty = false
		}
	}

	return &pageData{
		Home:     home,
		Version:  version,
		Tiers:    groups,
		AllEmpty: allEmpty,
		MaxSize:  maxSize,
		Totals: totals{
			Tier1:     t1,
			Tier2:     t2,
			Tier3:     t3,
			Tier4:     t4,
			All:       t1 + t2 + t3 + t4,
			Cleanable: t1 + t2,
		},
		GitTrap: gitTrap{Found: gitFound, SizeMB: gitSize},
	}
}

type tierGroup struct {
	Num     int
	Label   string
	Total   int64
	Entries []scan.Entry
}

type pageData struct {
	Home     string
	Version  string
	Tiers    []tierGroup
	AllEmpty bool
	MaxSize  int64
	Totals   totals
	GitTrap  gitTrap
	Loading  bool
}

type totals struct {
	Tier1     int64 `json:"tier1"`
	Tier2     int64 `json:"tier2"`
	Tier3     int64 `json:"tier3"`
	Tier4     int64 `json:"tier4"`
	All       int64 `json:"all"`
	Cleanable int64 `json:"cleanable"`
}

type gitTrap struct {
	Found  bool  `json:"found"`
	SizeMB int64 `json:"sizeMB"`
}

func homeRelative(home, path string) string {
	rel, err := filepath.Rel(home, path)
	if err != nil || strings.HasPrefix(rel, "..") {
		return path
	}
	return rel
}

func barPct(size, max int64) int {
	if max == 0 {
		return 0
	}
	pct := int(size * 100 / max)
	if pct < 2 {
		return 2
	}
	return pct
}
