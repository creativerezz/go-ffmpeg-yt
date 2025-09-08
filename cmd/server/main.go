package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-ffmpeg-yt/internal/captions"
)

type captionsRequest struct {
	URL  string `json:"url"`
	Lang string `json:"lang"`
	// format can be: text or vtt
	Format string `json:"format"`
}

type captionsResponse struct {
	Source   string `json:"source"`
	Language string `json:"language"`
	Format   string `json:"format"`
	Content  string `json:"content"`
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("/captions", captionsHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           logRequests(mux),
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 15 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func captionsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req := captionsRequest{}
	if r.Method == http.MethodGet {
		req.URL = r.URL.Query().Get("url")
		req.Lang = r.URL.Query().Get("lang")
		req.Format = r.URL.Query().Get("format")
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("invalid json: %v", err), http.StatusBadRequest)
			return
		}
	}

	if req.URL == "" {
		http.Error(w, "missing url", http.StatusBadRequest)
		return
	}
	if req.Lang == "" {
		req.Lang = "en"
	}
	if req.Format == "" {
		req.Format = "text"
	}

	res, err := captions.FetchWithYtDlp(ctx, req.URL, req.Lang, req.Format)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to fetch captions: %v", err), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(captionsResponse{
		Source:   res.Source,
		Language: req.Lang,
		Format:   req.Format,
		Content:  res.Content,
	})
}

