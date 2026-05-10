package main

import (
	"context"
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
			fmt.Fprintf(os.Stderr, "workcell: %v
", err)
			os.Exit(1)
		}
	case "serve":
		if err := serve(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "workcell: %v
", err)
			os.Exit(1)
		}
	case "init":
		if err := initCmd(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "workcell: %v
", err)
			os.Exit(1)
		}
	case "version":
		fmt.Println("workcell dev")
	case "help", "--help", "-h":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s

", os.Args[1])
		usage()
		os.Exit(64)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `usage:
  workcell init [--runtime fake|podman|incus]
  workcell run --profile fake -- <command> [args...]
  workcell serve [--addr 127.0.0.1:8787]
  workcell version`)
}

func initCmd(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	runtime := fs.String("runtime", "fake", "runtime to configure (fake, podman, incus)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	configPath := "workcell.yaml"
	
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("config file already exists: %s", configPath)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("cannot check config file: %w", err)
	}

	var content string
	switch *runtime {
	case "fake":
		content = `profiles:
  fake:
    backend: fake
    description: In-process backend for testing
`
	case "podman":
		content = `profiles:
  podman-smoke:
    backend: podman
    image: docker.io/library/alpine:3.20
    timeoutSeconds: 300
`
	case "incus":
		content = `profiles:
  incus-smoke:
    backend: incus
    image: images:ubuntu/24.04
    timeoutSeconds: 900
`
	default:
		return fmt.Errorf("unsupported runtime: %s", *runtime)
	}

	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	fmt.Printf("Initialized workcell.yaml with %s runtime
", *runtime)
	return nil
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
		writeJSON(w, http.StatusAccepted, job)
	})
	mux.HandleFunc("GET /v1/jobs/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		job, err := runner.GetJob(r.Context(), id)
		if err != nil {
			writeError(w, http.StatusNotFound, workcell.ErrorCode(err), err.Error())
			return
		}
		writeJSON(w, http.StatusOK, job)
	})

	slog.Info("workcell server listening", "addr", *addr)
	return http.ListenAndServe(*addr, mux)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}

func writeError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, map[string]any{
		"error":   code,
		"message": message,
	})
}