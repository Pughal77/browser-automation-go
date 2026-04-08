package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

type VisitRequest struct {
	URL     string `json:"url"`
	Timeout int    `json:"timeout,omitempty"` // timeout in seconds
}

type VisitResponse struct {
	Markdown string           `json:"markdown"`
	Metadata MetadataResponse `json:"metadata"`
}

type MetadataResponse struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func main() {
	fmt.Println("Checking Chromium availability (may download if missing)...")
	l := launcher.New().Headless(true).Set("headless", "new")
	launchURL, err := l.Launch()
	if err == nil {
		fmt.Println("Chromium is ready for local testing!")
		// Connect and close immediately to clean up the process
		browser := rod.New().ControlURL(launchURL).MustConnect()
		browser.MustClose()
	} else {
		fmt.Printf("Warning: failed to pre-warm Chromium: %v\n", err)
	}

	http.HandleFunc("/visit", visitHandler)

	fmt.Println("Server starting on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func visitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req VisitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		sendError(w, "URL is required", http.StatusBadRequest)
		return
	}

	timeoutSec := req.Timeout
	if timeoutSec <= 0 || timeoutSec > 300 {
		timeoutSec = 60 // default 30 seconds
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(timeoutSec)*time.Second)
	defer cancel()

	// 1. Scrape the page
	scrapeRes, err := Scrape(ctx, req.URL, time.Duration(timeoutSec)*time.Second)
	if err != nil {
		sendError(w, fmt.Sprintf("Scraping failed: %v", err), http.StatusInternalServerError)
		return
	}

	// 2. Convert to Markdown
	convRes, err := Convert(scrapeRes.HTML, req.URL)
	if err != nil {
		sendError(w, fmt.Sprintf("Conversion failed: %v", err), http.StatusInternalServerError)
		return
	}

	// 3. Response
	// Prefer scraper metadata if conversion fails to provide good metadata
	title := convRes.Title
	if title == "" {
		title = scrapeRes.Title
	}
	description := convRes.Description
	if description == "" {
		description = scrapeRes.Description
	}

	res := VisitResponse{
		Markdown: convRes.Markdown,
		Metadata: MetadataResponse{
			Title:       title,
			URL:         req.URL,
			Description: description,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func sendError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{Error: msg})
}
