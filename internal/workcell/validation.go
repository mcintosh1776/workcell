package workcell

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	validationWorkerBackend = "workcell"
	defaultValidationPath   = "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
)

type ValidationStatus string

const (
	ValidationSucceeded ValidationStatus = "succeeded"
	ValidationFailed    ValidationStatus = "failed"
	ValidationTimedOut  ValidationStatus = "timed_out"
	ValidationCancelled ValidationStatus = "cancelled"
	ValidationBlocked   ValidationStatus = "blocked"
)

// ValidationWorkerRequest is the generic validation-worker contract accepted by
// Workcell. It deliberately avoids platform-specific queue, bot, or tenant
// execution concepts; those remain caller metadata.
type ValidationWorkerRequest struct {
	ValidationJobID   string   `json:"validationJobId"`
	TenantID          string   `json:"tenantId,omitempty"`
	ProjectName       string   `json:"projectName,omitempty"`
	RepoTargetID      string   `json:"repoTargetId,omitempty"`
	Repository        string   `json:"repository"`
	RepoURL           string   `json:"repoUrl"`
	BaseRef           string   `json:"baseRef"`
	HeadRef           string   `json:"headRef"`
	HeadSHA           string   `json:"headSha"`
	AllowedWriteScope string   `json:"allowedWriteScope,omitempty"`
	ValidationProfile string   `json:"validationProfile"`
	Commands          []string `json:"commands"`
	TimeoutSeconds    int      `json:"timeoutSeconds"`
	NetworkPolicy     string   `json:"networkPolicy"`
	Mutation          string   `json:"mutation"`
	RequestedByBotID  string   `json:"requestedByBotId,omitempty"`
	SourceTaskID      string   `json:"sourceTaskId,omitempty"`
	SourcePRURL       string   `json:"sourcePrUrl,omitempty"`
	WorkingDirectory  string   `json:"workingDirectory,omitempty"`
	SourceTransport   string   `json:"sourceTransport,omitempty"`
	SourceBundlePath  string   `json:"sourceBundlePath,omitempty"`
	SourceBundleSHA   string   `json:"sourceBundleSha256,omitempty"`
}

type ValidationWorkerResult struct {
	ValidationJobID   string           `json:"validationJobId"`
	Status            ValidationStatus `json:"status"`
	Repository        string           `json:"repository,omitempty"`
	HeadRef           string           `json:"headRef,omitempty"`
	HeadSHA           string           `json:"headSha"`
	ValidationProfile string           `json:"validationProfile"`
	DirtyTrackedFiles []string         `json:"dirtyTrackedFiles"`
	ExitCode          *int             `json:"exitCode"`
	StartedAt         time.Time        `json:"startedAt"`
	CompletedAt       time.Time        `json:"completedAt"`
	DurationMs        int64            `json:"durationMs"`
	StdoutArtifact    string           `json:"stdoutArtifactPath"`
	StderrArtifact    string           `json:"stderrArtifactPath"`
	Summary           string           `json:"summary"`
	WorkerID          string           `json:"workerId"`
	WorkerBackend     string           `json:"workerBackend"`
	Mutation          string           `json:"mutation"`
}

func (runner *Runner) RunValidation(ctx context.Context, request ValidationWorkerRequest) (ValidationWorkerResult, error) {
	if err := validateValidationRequest(request); err != nil {
		return ValidationWorkerResult{}, err
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(request.TimeoutSeconds)*time.Second)
	defer cancel()

	started := time.Now().UTC()
	jobRoot, err := os.MkdirTemp("", "workcell-validation-"+safePathSegment(request.ValidationJobID)+"-")
	if err != nil {
		return ValidationWorkerResult{}, fmt.Errorf("%w: create validation job root: %v", ErrInvalidValidation, err)
	}
	workspaceDir := filepath.Join(jobRoot, "workspace")
	stdoutPath := filepath.Join(jobRoot, "stdout.txt")
	stderrPath := filepath.Join(jobRoot, "stderr.txt")
	homeDir := filepath.Join(jobRoot, "home")
	goCache := filepath.Join(jobRoot, "go-build")
	goModCache := filepath.Join(jobRoot, "go-mod")
	for _, path := range []string{workspaceDir, homeDir, goCache, goModCache} {
		if err := os.MkdirAll(path, 0o700); err != nil {
			return ValidationWorkerResult{}, fmt.Errorf("%w: create validation directory: %v", ErrInvalidValidation, err)
		}
	}

	var stdout, stderr bytes.Buffer
	status := ValidationSucceeded
	exitCode := 0
	summary := ""

	if err := runner.prepareValidationWorkspace(timeoutCtx, request, workspaceDir, &stdout, &stderr); err != nil {
		status = ValidationBlocked
		exitCode = commandExitCode(err)
		if exitCode == 0 {
			exitCode = 1
		}
		summary = err.Error()
	}

	runDir := workspaceDir
	workingDirectory := strings.TrimSpace(request.WorkingDirectory)
	if workingDirectory != "" && workingDirectory != "." {
		runDir = filepath.Join(workspaceDir, filepath.FromSlash(workingDirectory))
	}
	if status == ValidationSucceeded {
		if info, err := os.Stat(runDir); err != nil || !info.IsDir() {
			status = ValidationBlocked
			exitCode = 1
			summary = "workingDirectory does not exist"
		}
	}

	env := []string{
		"CI=1",
		"HOME=" + homeDir,
		"PATH=" + validationPath(),
		"GOCACHE=" + goCache,
		"GOMODCACHE=" + goModCache,
	}
	if status == ValidationSucceeded {
		for _, commandText := range request.Commands {
			stdout.WriteString("$ " + commandText + "\n")
			code, err := runCaptured(timeoutCtx, runDir, env, "bash", "-lc", commandText)
			stdout.Write(code.stdout)
			stderr.Write(code.stderr)
			if errors.Is(timeoutCtx.Err(), context.DeadlineExceeded) {
				status = ValidationTimedOut
				exitCode = 124
				summary = "validation timed out while running command: " + commandText
				break
			}
			if err != nil {
				status = ValidationFailed
				exitCode = code.exitCode
				if exitCode == 0 {
					exitCode = 1
				}
				summary = "validation command failed: " + commandText
				break
			}
		}
	}

	dirtyFiles := []string{}
	if _, err := os.Stat(filepath.Join(workspaceDir, ".git")); err == nil {
		code, _ := runCaptured(context.Background(), workspaceDir, env, "git", "status", "--porcelain", "--untracked-files=no")
		stdout.Write(code.stdout)
		stderr.Write(code.stderr)
		dirtyFiles = parseDirtyTrackedFiles(string(code.stdout))
	}
	if status == ValidationSucceeded && len(dirtyFiles) > 0 {
		status = ValidationFailed
		exitCode = 1
		summary = "validation left tracked files dirty"
	}

	if status == ValidationSucceeded {
		summary = fmt.Sprintf("Workcell validation succeeded for %s@%s profile=%s", request.Repository, request.HeadSHA, request.ValidationProfile)
	} else if summary == "" {
		summary = fmt.Sprintf("Workcell validation %s", status)
	}

	if err := os.WriteFile(stdoutPath, stdout.Bytes(), 0o600); err != nil {
		return ValidationWorkerResult{}, fmt.Errorf("write stdout artifact: %w", err)
	}
	if err := os.WriteFile(stderrPath, stderr.Bytes(), 0o600); err != nil {
		return ValidationWorkerResult{}, fmt.Errorf("write stderr artifact: %w", err)
	}

	completed := time.Now().UTC()
	result := ValidationWorkerResult{
		ValidationJobID:   request.ValidationJobID,
		Status:            status,
		Repository:        request.Repository,
		HeadRef:           request.HeadRef,
		HeadSHA:           request.HeadSHA,
		ValidationProfile: request.ValidationProfile,
		DirtyTrackedFiles: dirtyFiles,
		ExitCode:          &exitCode,
		StartedAt:         started,
		CompletedAt:       completed,
		DurationMs:        completed.Sub(started).Milliseconds(),
		StdoutArtifact:    stdoutPath,
		StderrArtifact:    stderrPath,
		Summary:           summary,
		WorkerID:          "workcell-daemon",
		WorkerBackend:     validationWorkerBackend,
		Mutation:          "none",
	}
	runner.mu.Lock()
	runner.validationResults[result.ValidationJobID] = result
	runner.mu.Unlock()
	return result, nil
}

func (runner *Runner) ValidationResult(id string) (ValidationWorkerResult, bool) {
	runner.mu.RLock()
	defer runner.mu.RUnlock()
	result, ok := runner.validationResults[id]
	return cloneValidationResult(result), ok
}

func (runner *Runner) prepareValidationWorkspace(ctx context.Context, request ValidationWorkerRequest, workspaceDir string, stdout, stderr io.Writer) error {
	source := strings.TrimSpace(request.RepoURL)
	if request.SourceBundlePath != "" {
		if err := verifyBundleDigest(request.SourceBundlePath, request.SourceBundleSHA); err != nil {
			return err
		}
		source = request.SourceBundlePath
	}
	if source == "" {
		return fmt.Errorf("repoUrl is required")
	}

	code, err := runCaptured(ctx, "", nil, "git", "clone", "--quiet", source, workspaceDir)
	stdout.Write(code.stdout)
	stderr.Write(code.stderr)
	if err != nil {
		return fmt.Errorf("failed to clone repository")
	}
	if request.SourceBundlePath == "" {
		code, err = runCaptured(ctx, workspaceDir, nil, "git", "fetch", "--quiet", "origin", request.BaseRef, request.HeadRef)
		stdout.Write(code.stdout)
		stderr.Write(code.stderr)
		if err != nil {
			return fmt.Errorf("failed to fetch requested refs")
		}
	}
	code, err = runCaptured(ctx, workspaceDir, nil, "git", "checkout", "--quiet", request.HeadSHA)
	stdout.Write(code.stdout)
	stderr.Write(code.stderr)
	if err != nil {
		return fmt.Errorf("failed to checkout requested head SHA")
	}
	return nil
}

func validateValidationRequest(request ValidationWorkerRequest) error {
	missing := []string{}
	required := map[string]string{
		"validationJobId":   request.ValidationJobID,
		"repository":        request.Repository,
		"repoUrl":           request.RepoURL,
		"baseRef":           request.BaseRef,
		"headRef":           request.HeadRef,
		"headSha":           request.HeadSHA,
		"validationProfile": request.ValidationProfile,
	}
	for field, value := range required {
		if strings.TrimSpace(value) == "" {
			missing = append(missing, field)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("%w: missing required fields: %s", ErrInvalidValidation, strings.Join(missing, ", "))
	}
	if !strings.Contains(request.Repository, "/") {
		return fmt.Errorf("%w: repository must be owner/name", ErrInvalidValidation)
	}
	if request.Mutation != "" && request.Mutation != "none" {
		return fmt.Errorf("%w: mutation must be none", ErrInvalidValidation)
	}
	if len(request.Commands) == 0 {
		return fmt.Errorf("%w: commands are required", ErrInvalidValidation)
	}
	for _, command := range request.Commands {
		if strings.TrimSpace(command) == "" {
			return fmt.Errorf("%w: commands must not be empty", ErrInvalidValidation)
		}
	}
	if request.TimeoutSeconds < 1 || request.TimeoutSeconds > 3600 {
		return fmt.Errorf("%w: timeoutSeconds must be from 1 to 3600", ErrInvalidValidation)
	}
	if request.NetworkPolicy != "" {
		switch request.NetworkPolicy {
		case "disabled", "restricted", "enabled":
		default:
			return fmt.Errorf("%w: invalid networkPolicy", ErrInvalidValidation)
		}
	}
	if request.SourceBundlePath != "" && request.SourceTransport != "" && request.SourceTransport != "git-bundle" {
		return fmt.Errorf("%w: sourceTransport must be git-bundle when sourceBundlePath is set", ErrInvalidValidation)
	}
	workingDirectory := strings.TrimSpace(request.WorkingDirectory)
	if filepath.IsAbs(workingDirectory) || containsParentPath(workingDirectory) {
		return fmt.Errorf("%w: workingDirectory is unsafe", ErrInvalidValidation)
	}
	return nil
}

func verifyBundleDigest(path string, want string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("sourceBundlePath does not exist")
	}
	defer file.Close()
	if strings.TrimSpace(want) == "" {
		return nil
	}
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return fmt.Errorf("failed to read sourceBundlePath")
	}
	got := hex.EncodeToString(hash.Sum(nil))
	if got != strings.ToLower(strings.TrimSpace(want)) {
		return fmt.Errorf("sourceBundleSha256 mismatch")
	}
	return nil
}

type commandResult struct {
	exitCode int
	stdout   []byte
	stderr   []byte
}

func runCaptured(ctx context.Context, cwd string, env []string, name string, args ...string) (commandResult, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	if cwd != "" {
		cmd.Dir = cwd
	}
	if env != nil {
		cmd.Env = env
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return commandResult{
		exitCode: commandExitCode(err),
		stdout:   stdout.Bytes(),
		stderr:   stderr.Bytes(),
	}, err
}

func commandExitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return 1
}

func parseDirtyTrackedFiles(output string) []string {
	files := []string{}
	for _, line := range strings.Split(output, "\n") {
		if len(line) < 4 {
			continue
		}
		file := strings.TrimSpace(line[3:])
		if file != "" {
			files = append(files, file)
		}
	}
	return files
}

func cloneValidationResult(result ValidationWorkerResult) ValidationWorkerResult {
	result.DirtyTrackedFiles = append([]string(nil), result.DirtyTrackedFiles...)
	return result
}

func safePathSegment(value string) string {
	var builder strings.Builder
	for _, char := range value {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '-' || char == '_' || char == '.' {
			builder.WriteRune(char)
		} else {
			builder.WriteRune('-')
		}
	}
	segment := strings.Trim(builder.String(), "-.")
	if segment == "" {
		return "job"
	}
	if len(segment) > 80 {
		return segment[:80]
	}
	return segment
}

func containsParentPath(path string) bool {
	if path == "" {
		return false
	}
	for _, part := range strings.Split(filepath.Clean(path), string(filepath.Separator)) {
		if part == ".." {
			return true
		}
	}
	return false
}

func validationPath() string {
	if path := strings.TrimSpace(os.Getenv("PATH")); path != "" {
		return path
	}
	return defaultValidationPath
}
