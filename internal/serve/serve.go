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

	"github.com/srmdn/orbital/internal/scan"
)

//go:embed templates/*
var templatesFS embed.FS

func Run() error {
	home := os.Getenv("HOME")

	fmt.Print("\n  scanning your mac...")
	data := scanAndBuild(home)
	fmt.Println(" done.")

	mu := &sync.RWMutex{}

	tmpl := template.Must(template.New("index.html").Funcs(template.FuncMap{
		"formatSize":   scan.FormatSize,
		"tierLabel":    scan.TierLabel,
		"tierTotal":    tierTotal,
		"homeRelative": homeRelative,
		"barPct":       barPct,
	}).ParseFS(templatesFS, "templates/index.html"))

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		mu.RLock()
		d := data
		mu.RUnlock()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.Execute(w, d)
	})

	mux.HandleFunc("/api/scan", func(w http.ResponseWriter, r *http.Request) {
		mu.RLock()
		defer mu.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	})

	mux.HandleFunc("/api/refresh", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		mu.Lock()
		data = scanAndBuild(home)
		mu.Unlock()

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
		entries := data.Entries
		mu.RUnlock()

		var freed int64
		var failed int
		for _, e := range entries {
			if !e.Cleanable {
				continue
			}
			if err := os.RemoveAll(e.Path); err != nil {
				failed++
			} else {
				freed += e.SizeMB
			}
		}

		mu.Lock()
		data = scanAndBuild(home)
		mu.Unlock()

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

	fmt.Printf("  🚀 orbital dashboard → %s\n", url)
	fmt.Println("  Press Ctrl+C to stop.")

	go func() {
		exec.Command("open", url).Run()
	}()

	return http.Serve(listener, mux)
}

func scanAndBuild(home string) *pageData {
	entries, t1, t2, t3, t4 := scan.Collect(home)
	gitFound, gitSize := scan.HasGitTrap(home)

	var maxSize int64
	for _, e := range entries {
		if e.SizeMB > maxSize {
			maxSize = e.SizeMB
		}
	}

	return &pageData{
		Home:    home,
		Entries: entries,
		MaxSize: maxSize,
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

type pageData struct {
	Home    string
	Entries []scan.Entry
	MaxSize int64
	Totals  totals
	GitTrap gitTrap
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

func tierTotal(t totals, tier int) int64 {
	switch tier {
	case 1:
		return t.Tier1
	case 2:
		return t.Tier2
	case 3:
		return t.Tier3
	case 4:
		return t.Tier4
	}
	return 0
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
