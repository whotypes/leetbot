package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/whotypes/leetbot/internal/data"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type CompanyData struct {
	Company    string               `json:"company"`
	Timeframes map[string][]Problem `json:"timeframes"`
}

type Problem struct {
	ID         int     `json:"id"`
	URL        string  `json:"url"`
	Title      string  `json:"title"`
	Difficulty string  `json:"difficulty"`
	Acceptance float64 `json:"acceptance"`
	Frequency  float64 `json:"frequency"`
}

type CompaniesList struct {
	Companies []string `json:"companies"`
}

type TimeframesList struct {
	Timeframes []string `json:"timeframes"`
}

var problemsData *data.ProblemsByCompany

func main() {
	var err error
	problemsData, err = data.LoadAllProblems()
	if err != nil {
		log.Fatalf("Failed to load problems data: %v", err)
	}

	r := mux.NewRouter()

	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/companies", getCompanies).Methods("GET")
	api.HandleFunc("/companies/{company}/timeframes", getTimeframes).Methods("GET")
	api.HandleFunc("/companies/{company}/problems", getProblems).Methods("GET")
	api.HandleFunc("/companies/{company}/timeframes/{timeframe}/problems", getProblemsByTimeframe).Methods("GET")
	api.HandleFunc("/all-problems", getAllProblems).Methods("GET")

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/dist/")))

	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
	)(r)

	// cloud run listens on 8080, lets expose $PORT
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create HTTP server with timeouts
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      corsHandler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Channel to listen for interrupt signal
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		fmt.Printf("Server starting on port %s\n", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	fmt.Println("Server is running. Press CTRL-C to exit.")
	<-done
	fmt.Println("Shutting down server...")

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	fmt.Println("Server exited")
}

func getCompanies(w http.ResponseWriter, r *http.Request) {
	companies := problemsData.GetAvailableCompanies()

	response := APIResponse{
		Success: true,
		Data:    CompaniesList{Companies: companies},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func getTimeframes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	company := vars["company"]

	timeframes := problemsData.GetAvailableTimeframes(company)

	response := APIResponse{
		Success: true,
		Data:    TimeframesList{Timeframes: timeframes},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func getProblems(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	company := vars["company"]

	// Get problems with priority (most recent timeframe with data)
	problems, timeframe := problemsData.GetProblemsWithPriority(company)

	if problems == nil {
		response := APIResponse{
			Success: false,
			Error:   "No problems found for company: " + company,
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
		return
	}

	// Convert to our API format
	apiProblems := make([]Problem, len(problems))
	for i, p := range problems {
		apiProblems[i] = Problem{
			ID:         p.ID,
			URL:        p.URL,
			Title:      p.Title,
			Difficulty: p.Difficulty,
			Acceptance: p.Acceptance,
			Frequency:  p.Frequency,
		}
	}

	response := APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"company":   company,
			"timeframe": timeframe,
			"problems":  apiProblems,
			"count":     len(apiProblems),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func getProblemsByTimeframe(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	company := vars["company"]
	timeframe := vars["timeframe"]

	problems := problemsData.GetProblems(company, timeframe)

	if problems == nil {
		response := APIResponse{
			Success: false,
			Error:   fmt.Sprintf("No problems found for company: %s, timeframe: %s", company, timeframe),
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
		return
	}

	// Convert to our API format
	apiProblems := make([]Problem, len(problems))
	for i, p := range problems {
		apiProblems[i] = Problem{
			ID:         p.ID,
			URL:        p.URL,
			Title:      p.Title,
			Difficulty: p.Difficulty,
			Acceptance: p.Acceptance,
			Frequency:  p.Frequency,
		}
	}

	response := APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"company":   company,
			"timeframe": timeframe,
			"problems":  apiProblems,
			"count":     len(apiProblems),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func getAllProblems(w http.ResponseWriter, r *http.Request) {
	allProblems := problemsData.GetAllProblems()

	allProblemsMap := make(map[string]map[string][]Problem)
	for company, timeframes := range allProblems {
		allProblemsMap[company] = make(map[string][]Problem)
		for timeframe, problems := range timeframes {
			apiProblems := make([]Problem, len(problems))
			for i, p := range problems {
				apiProblems[i] = Problem{
					ID:         p.ID,
					URL:        p.URL,
					Title:      p.Title,
					Difficulty: p.Difficulty,
					Acceptance: p.Acceptance,
					Frequency:  p.Frequency,
				}
			}
			allProblemsMap[company][timeframe] = apiProblems
		}
	}

	response := APIResponse{
		Success: true,
		Data:    allProblemsMap,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
