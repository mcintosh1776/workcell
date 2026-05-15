package workcell

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestRunnerFakeProfileSucceeds(t *testing.T) {
	runner := NewRunner(DefaultProfiles())

	job, err := runner.Run(context.Background(), SubmitJobRequest{
		Profile: "fake",
		Command: []string{"echo", "hello"},
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if job.State != JobSucceeded {
		t.Fatalf("State = %s, want %s", job.State, JobSucceeded)
	}
	if job.Backend != "fake" {
		t.Fatalf("Backend = %s, want fake", job.Backend)
	}
	if job.Cleanup.State != "complete" {
		t.Fatalf("Cleanup.State = %s, want complete", job.Cleanup.State)
	}
	if job.Error != "" {
		t.Fatalf("Error = %q, want empty for successful fake job", job.Error)
	}
	if job.Stdout != "echo hello" {
		t.Fatalf("Stdout = %q, want %q", job.Stdout, "echo hello")
	}
}

func TestRunnerRejectsInvalidProfile(t *testing.T) {
	runner := NewRunner(DefaultProfiles())

	_, err := runner.Run(context.Background(), SubmitJobRequest{
		Profile: "missing",
		Command: []string{"echo", "hello"},
	})
	if !errors.Is(err, ErrInvalidProfile) {
		t.Fatalf("error = %v, want ErrInvalidProfile", err)
	}
}

func TestRunnerRejectsEmptyCommand(t *testing.T) {
	runner := NewRunner(DefaultProfiles())

	_, err := runner.Run(context.Background(), SubmitJobRequest{
		Profile: "fake",
	})
	if !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("error = %v, want ErrInvalidCommand", err)
	}
}

func TestRunnerFakeProfileCanFail(t *testing.T) {
	runner := NewRunner(DefaultProfiles())

	job, err := runner.Run(context.Background(), SubmitJobRequest{
		Profile: "fake",
		Command: []string{"false"},
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if job.State != JobFailed {
		t.Fatalf("State = %s, want %s", job.State, JobFailed)
	}
	if job.ExitCode != 1 {
		t.Fatalf("ExitCode = %d, want 1", job.ExitCode)
	}
	if job.Stdout != "false" {
		t.Fatalf("Stdout = %q, want %q", job.Stdout, "false")
	}
	if job.Error != "" {
		t.Fatalf("Error = %q, want empty for command exit failure", job.Error)
	}
}

func TestRunnerListsJobsAndReturnsLogs(t *testing.T) {
	runner := NewRunner(DefaultProfiles())

	first, err := runner.Run(context.Background(), SubmitJobRequest{
		Profile: "fake",
		Command: []string{"echo", "first"},
	})
	if err != nil {
		t.Fatalf("first Run returned error: %v", err)
	}
	second, err := runner.Run(context.Background(), SubmitJobRequest{
		Profile: "fake",
		Command: []string{"echo", "second"},
	})
	if err != nil {
		t.Fatalf("second Run returned error: %v", err)
	}

	jobs := runner.List()
	if len(jobs) != 2 {
		t.Fatalf("List returned %d jobs, want 2", len(jobs))
	}
	if jobs[0].ID != second.ID {
		t.Fatalf("newest job = %s, want %s", jobs[0].ID, second.ID)
	}
	if jobs[1].ID != first.ID {
		t.Fatalf("oldest job = %s, want %s", jobs[1].ID, first.ID)
	}

	logs, ok := runner.Logs(first.ID)
	if !ok {
		t.Fatalf("Logs(%s) not found", first.ID)
	}
	if logs.Stdout != "echo first" {
		t.Fatalf("stdout = %q, want %q", logs.Stdout, "echo first")
	}
	if logs.Stderr != "" {
		t.Fatalf("stderr = %q, want empty", logs.Stderr)
	}
	if logs.Truncated {
		t.Fatal("logs unexpectedly marked truncated")
	}
}

func TestRunnerBackendFailurePreservesBackendExitAndError(t *testing.T) {
	runner := NewRunner(map[string]Profile{
		"podman-smoke": {
			ID:      "podman-smoke",
			Backend: "podman",
			BackendConfig: BackendConfig{
				Image:   "example.invalid/missing:latest",
				Timeout: 60,
			},
		},
	})
	runner.backends["podman"] = backendFunc{
		run: func(ctx context.Context, job Job, profile Profile) (int, string, string, error) {
			return 125, "", "image pull failed", &BackendError{Op: "create", Err: errors.New("image pull failed")}
		},
		cleanup: func(ctx context.Context, job Job, profile Profile) error {
			return nil
		},
	}

	job, err := runner.Run(context.Background(), SubmitJobRequest{
		Profile: "podman-smoke",
		Command: []string{"echo", "hello"},
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if job.State != JobFailed {
		t.Fatalf("State = %s, want %s", job.State, JobFailed)
	}
	if job.ExitCode != 125 {
		t.Fatalf("ExitCode = %d, want 125", job.ExitCode)
	}
	if job.Error == "" {
		t.Fatal("Error is empty, want backend failure detail")
	}
	if !strings.Contains(job.Error, "image pull failed") {
		t.Fatalf("Error = %q, want backend failure detail", job.Error)
	}
	if job.Stdout != "" {
		t.Fatalf("Stdout = %q, want empty stdout for backend failure", job.Stdout)
	}
	if job.Logs.StderrBytes == 0 {
		t.Fatal("StderrBytes = 0, want captured backend stderr")
	}
	if job.Stderr != "image pull failed" {
		t.Fatalf("Stderr = %q, want backend stderr detail", job.Stderr)
	}
}

func TestRunnerCleanupUsesFreshContextAfterRunContextExpires(t *testing.T) {
	runner := NewRunner(map[string]Profile{
		"podman-smoke": {
			ID:      "podman-smoke",
			Backend: "podman",
			BackendConfig: BackendConfig{
				Image:   "docker.io/library/alpine:3.20",
				Timeout: 60,
			},
		},
	})
	runner.backends["podman"] = backendFunc{
		run: func(ctx context.Context, job Job, profile Profile) (int, string, string, error) {
			if cancel, ok := ctx.Value(cancelContextKey{}).(context.CancelFunc); ok {
				cancel()
			}
			return 0, "ok", "", nil
		},
		cleanup: func(ctx context.Context, job Job, profile Profile) error {
			if err := ctx.Err(); err != nil {
				return err
			}
			return nil
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	ctx = context.WithValue(ctx, cancelContextKey{}, context.CancelFunc(cancel))

	job, err := runner.Run(ctx, SubmitJobRequest{
		Profile: "podman-smoke",
		Command: []string{"echo", "hello"},
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if job.State != JobSucceeded {
		t.Fatalf("State = %s, want %s", job.State, JobSucceeded)
	}
	if job.Cleanup.State != "complete" {
		t.Fatalf("Cleanup.State = %s, want complete", job.Cleanup.State)
	}
}

func TestRunnerValidationJobUsesSourceBundle(t *testing.T) {
	requireValidationCommands(t)
	repoDir, bundlePath, bundleSHA, headSHA := createValidationBundle(t)
	runner := NewRunner(DefaultProfiles())

	result, err := runner.RunValidation(context.Background(), ValidationWorkerRequest{
		ValidationJobID:   "validation-bundle-smoke",
		Repository:        "example/workcell",
		RepoURL:           repoDir,
		BaseRef:           "HEAD",
		HeadRef:           "HEAD",
		HeadSHA:           headSHA,
		ValidationProfile: "unit",
		Commands:          []string{"test -f README.md"},
		TimeoutSeconds:    30,
		NetworkPolicy:     "disabled",
		Mutation:          "none",
		SourceTransport:   "git-bundle",
		SourceBundlePath:  bundlePath,
		SourceBundleSHA:   bundleSHA,
	})
	if err != nil {
		t.Fatalf("RunValidation returned error: %v", err)
	}
	if result.Status != ValidationSucceeded {
		t.Fatalf("Status = %s, want %s; summary=%s", result.Status, ValidationSucceeded, result.Summary)
	}
	if result.WorkerBackend != "workcell" {
		t.Fatalf("WorkerBackend = %s, want workcell", result.WorkerBackend)
	}
	if result.HeadSHA != headSHA {
		t.Fatalf("HeadSHA = %s, want %s", result.HeadSHA, headSHA)
	}
	if len(result.DirtyTrackedFiles) != 0 {
		t.Fatalf("DirtyTrackedFiles = %v, want none", result.DirtyTrackedFiles)
	}

	stored, ok := runner.ValidationResult("validation-bundle-smoke")
	if !ok {
		t.Fatal("ValidationResult not found")
	}
	if stored.ValidationJobID != result.ValidationJobID {
		t.Fatalf("stored id = %s, want %s", stored.ValidationJobID, result.ValidationJobID)
	}
}

func TestRunnerValidationJobFailsOnDirtyTrackedFiles(t *testing.T) {
	requireValidationCommands(t)
	repoDir, bundlePath, bundleSHA, headSHA := createValidationBundle(t)
	runner := NewRunner(DefaultProfiles())

	result, err := runner.RunValidation(context.Background(), ValidationWorkerRequest{
		ValidationJobID:   "validation-dirty-smoke",
		Repository:        "example/workcell",
		RepoURL:           repoDir,
		BaseRef:           "HEAD",
		HeadRef:           "HEAD",
		HeadSHA:           headSHA,
		ValidationProfile: "unit",
		Commands:          []string{"printf dirty >> README.md"},
		TimeoutSeconds:    30,
		NetworkPolicy:     "disabled",
		Mutation:          "none",
		SourceTransport:   "git-bundle",
		SourceBundlePath:  bundlePath,
		SourceBundleSHA:   bundleSHA,
	})
	if err != nil {
		t.Fatalf("RunValidation returned error: %v", err)
	}
	if result.Status != ValidationFailed {
		t.Fatalf("Status = %s, want %s; summary=%s", result.Status, ValidationFailed, result.Summary)
	}
	if !reflect.DeepEqual(result.DirtyTrackedFiles, []string{"README.md"}) {
		t.Fatalf("DirtyTrackedFiles = %v, want [README.md]", result.DirtyTrackedFiles)
	}
}

func TestRunnerValidationJobBlocksOnBadBundleDigest(t *testing.T) {
	requireValidationCommands(t)
	repoDir, bundlePath, _, headSHA := createValidationBundle(t)
	runner := NewRunner(DefaultProfiles())

	result, err := runner.RunValidation(context.Background(), ValidationWorkerRequest{
		ValidationJobID:   "validation-bad-digest",
		Repository:        "example/workcell",
		RepoURL:           repoDir,
		BaseRef:           "HEAD",
		HeadRef:           "HEAD",
		HeadSHA:           headSHA,
		ValidationProfile: "unit",
		Commands:          []string{"test -f README.md"},
		TimeoutSeconds:    30,
		NetworkPolicy:     "disabled",
		Mutation:          "none",
		SourceTransport:   "git-bundle",
		SourceBundlePath:  bundlePath,
		SourceBundleSHA:   "0000000000000000000000000000000000000000000000000000000000000000",
	})
	if err != nil {
		t.Fatalf("RunValidation returned error: %v", err)
	}
	if result.Status != ValidationBlocked {
		t.Fatalf("Status = %s, want %s", result.Status, ValidationBlocked)
	}
	if result.Summary != "sourceBundleSha256 mismatch" {
		t.Fatalf("Summary = %q, want digest mismatch", result.Summary)
	}
}

type backendFunc struct {
	run     func(context.Context, Job, Profile) (int, string, string, error)
	cleanup func(context.Context, Job, Profile) error
}

func (backend backendFunc) Run(ctx context.Context, job Job, profile Profile) (int, string, string, error) {
	return backend.run(ctx, job, profile)
}

func (backend backendFunc) Cleanup(ctx context.Context, job Job, profile Profile) error {
	return backend.cleanup(ctx, job, profile)
}

type cancelContextKey struct{}

func requireValidationCommands(t *testing.T) {
	t.Helper()
	for _, command := range []string{"git", "bash"} {
		if _, err := exec.LookPath(command); err != nil {
			t.Skipf("%s not installed", command)
		}
	}
}

func createValidationBundle(t *testing.T) (repoDir string, bundlePath string, bundleSHA string, headSHA string) {
	t.Helper()
	repoDir = filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(repoDir, 0o700); err != nil {
		t.Fatal(err)
	}
	runTestCommand(t, repoDir, "git", "init", "--quiet")
	runTestCommand(t, repoDir, "git", "config", "user.email", "workcell@example.invalid")
	runTestCommand(t, repoDir, "git", "config", "user.name", "Workcell Test")
	if err := os.WriteFile(filepath.Join(repoDir, "README.md"), []byte("workcell\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runTestCommand(t, repoDir, "git", "add", "README.md")
	runTestCommand(t, repoDir, "git", "commit", "--quiet", "-m", "initial")
	headSHA = strings.TrimSpace(runTestCommand(t, repoDir, "git", "rev-parse", "HEAD"))
	bundlePath = filepath.Join(t.TempDir(), "source.bundle")
	runTestCommand(t, repoDir, "git", "bundle", "create", bundlePath, "HEAD")
	bundleBytes, err := os.ReadFile(bundlePath)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(bundleBytes)
	return repoDir, bundlePath, hex.EncodeToString(sum[:]), headSHA
}

func runTestCommand(t *testing.T, cwd string, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = cwd
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v failed: %v\n%s", name, args, err, output)
	}
	return string(output)
}
