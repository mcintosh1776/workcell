package main

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mcintosh1776/workcell/internal/workcell"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(64)
	}

	switch os.Args[1] {
	case "run":
		if err := run(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "workcell: %v\n", err)
			os.Exit(1)
		}
	case "serve":
		if err := serve(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "workcell: %v\n", err)
			os.Exit(1)
		}
	case "version":
		fmt.Println("workcell dev")
	case "help", "--help", "-h":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		usage()
		os.Exit(64)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `usage:
  workcell run --profile fake -- <command> [args...]
  workcell serve [--addr 127.0.0.1:8787]
  workcell version`)
}

func run(args []string) error {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	profile := fs.String("profile", "fake", "profile to run")
	if err := fs.Parse(args); err != nil {
		return err
	}

	command := fs.Args()
	if len(command) > 0 && command[0] == "--" {
		command = command[1:]
	}

	runner := workcell.NewRunner(workcell.DefaultProfiles())
	job, err := runner.Run(context.Background(), workcell.SubmitJobRequest{
		Profile:        *profile,
		Command:        command,
		TimeoutSeconds: 300,
	})
	if err != nil {
		return err
	}

	encoded, err := json.MarshalIndent(job, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(encoded))
	if job.ExitCode != 0 {
		return fmt.Errorf("job failed with exit code %d", job.ExitCode)
	}
	return nil
}

func serve(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	addr := fs.String("addr", "127.0.0.1:8787", "listen address")
	if err := fs.Parse(args); err != nil {
		return err
	}

	runner := workcell.NewRunner(workcell.DefaultProfiles())
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "status": "ok"})
	})
	mux.HandleFunc("POST /v1/jobs", func(w http.ResponseWriter, r *http.Request) {
		var request workcell.SubmitJobRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
			return
		}
		job, err := runner.Run(r.Context(), request)
		if err != nil {
			writeError(w, http.StatusBadRequest, workcell.ErrorCode(err), err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"ok": true, "data": job})
	})
	mux.HandleFunc("GET /v1/jobs", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "data": runner.List()})
	})
	mux.HandleFunc("GET /v1/jobs/{jobId}", func(w http.ResponseWriter, r *http.Request) {
		job, ok := runner.Get(r.PathValue("jobId"))
		if !ok {
			writeError(w, http.StatusNotFound, "job_not_found", "job not found")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "data": job})
	})
	mux.HandleFunc("GET /v1/jobs/{jobId}/logs", func(w http.ResponseWriter, r *http.Request) {
		logs, ok := runner.Logs(r.PathValue("jobId"))
		if !ok {
			writeError(w, http.StatusNotFound, "job_not_found", "job not found")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "data": logs})
	})
	mux.HandleFunc("POST /v1/validation-jobs", func(w http.ResponseWriter, r *http.Request) {
		if !authorizeValidationJob(w, r) {
			return
		}
		var request workcell.ValidationWorkerRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
			return
		}
		result, err := runner.RunValidation(r.Context(), request)
		if err != nil {
			writeError(w, http.StatusBadRequest, workcell.ErrorCode(err), err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"ok": true, "data": result})
	})
	mux.HandleFunc("GET /v1/validation-jobs/{validationJobId}", func(w http.ResponseWriter, r *http.Request) {
		if !authorizeValidationJob(w, r) {
			return
		}
		result, ok := runner.ValidationResult(r.PathValue("validationJobId"))
		if !ok {
			writeError(w, http.StatusNotFound, "validation_job_not_found", "validation job not found")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "data": result})
	})

	server := &http.Server{
		Addr:              *addr,
		Handler:           requestLogger(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}
	slog.Info("workcell listening", "addr", *addr)
	return server.ListenAndServe()
}

func authorizeValidationJob(w http.ResponseWriter, r *http.Request) bool {
	token := validationAPIToken()
	if token == "" {
		writeError(w, http.StatusServiceUnavailable, "validation_api_not_configured", "validation job API token is not configured")
		return false
	}
	header := strings.TrimSpace(r.Header.Get("authorization"))
	if !strings.HasPrefix(strings.ToLower(header), "bearer ") {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authorization bearer token is required")
		return false
	}
	if subtle.ConstantTimeCompare([]byte(strings.TrimSpace(header[len("Bearer "):])), []byte(token)) != 1 {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authorization bearer token is invalid")
		return false
	}
	return true
}

func validationAPIToken() string {
	tokenFile := strings.TrimSpace(os.Getenv("WORKCELL_VALIDATION_API_TOKEN_FILE"))
	if tokenFile == "" {
		return ""
	}
	info, err := os.Stat(tokenFile)
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0o077 != 0 {
		return ""
	}
	data, err := os.ReadFile(tokenFile)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, map[string]any{
		"ok": false,
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		next.ServeHTTP(w, r)
		slog.Info("request", "method", r.Method, "path", r.URL.Path, "duration", strings.TrimSuffix(time.Since(started).String(), "0s"))
	})
}
