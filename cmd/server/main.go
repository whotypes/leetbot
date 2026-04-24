// Command server is the leetbot.org HTTP API and static file server.
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/whotypes/leetbot/internal/data"
)

type envelope struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

type api struct {
	pbc *data.ProblemsByCompany
}

func main() {
	pbc, err := data.LoadAllProblems()
	if err != nil {
		log.Fatalf("load embedded data: %v", err)
	}

	s := &api{pbc: pbc}
	r := mux.NewRouter()
	r.PathPrefix("/api/").Handler(http.StripPrefix("/api", s.apiRouter()))
	r.PathPrefix("/").Handler(s.spaFileServer("web/dist"))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           loggingMiddleware(r),
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Printf("listening on :%s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server: %v", err)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t0 := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(t0))
	})
}

func (s *api) apiRouter() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/companies", s.handleCompanies).Methods(http.MethodGet)
	r.HandleFunc("/all-problems", s.handleAllProblems).Methods(http.MethodGet)
	r.Methods(http.MethodGet).Path("/companies/{company}/timeframes").HandlerFunc(s.handleTimeframes)
	r.Methods(http.MethodGet).Path("/companies/{company}/timeframes/{timeframe}/problems").HandlerFunc(s.handleProblems)
	r.Methods(http.MethodGet).Path("/companies/{company}/problems").HandlerFunc(s.handleProblemsPriority)
	return r
}

func (s *api) handleCompanies(w http.ResponseWriter, _ *http.Request) {
	updated := data.DataLastUpdatedRFC3339()
	writeJSON(w, http.StatusOK, envelope{
		Success: true,
		Data: map[string]any{
			"dataLastUpdated": updated,
			"companies":       s.pbc.CompanyCatalog(),
		},
	})
}

func (s *api) handleAllProblems(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, envelope{Success: true, Data: s.pbc.GetAllProblems()})
}

func (s *api) handleTimeframes(w http.ResponseWriter, r *http.Request) {
	company, err := pathVar(mux.Vars(r), "company")
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if !data.InWebCatalog(s.pbc, company) {
		writeErr(w, http.StatusNotFound, "unknown company")
		return
	}
	tf := data.TimeframesForWeb(s.pbc, company)
	writeJSON(w, http.StatusOK, envelope{
		Success: true,
		Data:    map[string]any{"timeframes": tf},
	})
}

func (s *api) handleProblems(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	company, err := pathVar(vars, "company")
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	timeframe, err := pathVar(vars, "timeframe")
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if !data.InWebCatalog(s.pbc, company) {
		writeErr(w, http.StatusNotFound, "unknown company")
		return
	}
	normTF := data.NormalizeTimeframe(timeframe)
	probs := s.pbc.GetProblems(company, normTF)
	if probs == nil {
		probs = []data.Problem{}
	}
	local := s.pbc.CompanyHasLocalData(company)
	emptyTF := local && len(probs) == 0
	writeJSON(w, http.StatusOK, envelope{
		Success: true,
		Data: map[string]any{
			"company":             strings.ToLower(strings.TrimSpace(company)),
			"timeframe":           normTF,
			"problems":            probs,
			"count":               len(probs),
			"companyHasLocalData": local,
			"emptyTimeframe":      emptyTF,
		},
	})
}

func (s *api) handleProblemsPriority(w http.ResponseWriter, r *http.Request) {
	company, err := pathVar(mux.Vars(r), "company")
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if !data.InWebCatalog(s.pbc, company) {
		writeErr(w, http.StatusNotFound, "unknown company")
		return
	}
	probs, tf := s.pbc.GetProblemsWithPriority(company)
	if probs == nil {
		probs = []data.Problem{}
	}
	local := s.pbc.CompanyHasLocalData(company)
	writeJSON(w, http.StatusOK, envelope{
		Success: true,
		Data: map[string]any{
			"company":             strings.ToLower(strings.TrimSpace(company)),
			"timeframe":           tf,
			"problems":            probs,
			"count":               len(probs),
			"companyHasLocalData": local,
		},
	})
}

func pathVar(vars map[string]string, key string) (string, error) {
	v := strings.TrimSpace(vars[key])
	if v == "" {
		return "", errInvalid("missing " + key)
	}
	return v, nil
}

type pathErr string

func (e pathErr) Error() string { return string(e) }
func errInvalid(s string) error { return pathErr(s) }

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, envelope{Success: false, Error: msg})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func (s *api) spaFileServer(root string) http.Handler {
	fs := http.FileServer(http.Dir(root))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		// real build assets: let the file server 404 for missing files
		if strings.HasPrefix(r.URL.Path, "/assets/") {
			fs.ServeHTTP(w, r)
			return
		}
		p := r.URL.Path
		if p == "" || p == "/" {
			serveIndexHTML(w, root)
			return
		}
		rel := strings.TrimPrefix(p, "/")
		local := filepath.Join(root, rel)
		if st, err := os.Stat(local); err == nil && !st.IsDir() {
			fs.ServeHTTP(w, r)
			return
		}
		serveIndexHTML(w, root)
	})
}

func serveIndexHTML(w http.ResponseWriter, root string) {
	p := filepath.Join(root, "index.html")
	b, err := os.ReadFile(p)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(b)
}
