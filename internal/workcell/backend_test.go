package workcell

import (
	"context"
	"testing"
)

func TestPodmanBackendRunCommand(t *testing.T) {
	backend := NewPodmanBackend("docker.io/library/alpine:3.20")
	
	ctx := context.Background()
	job := Job{
		ID:      "test-job-123",
		Command: []string{"echo", "hello"},
	}
	
	exitCode, stdout, stderr, err := backend.Run(ctx, job)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("ExitCode = %d, want 0", exitCode)
	}
	if stdout != "hello
" {  The user wants me to fix a compile error in the workcell repo. The error is:
```
internal/workcell/backend_test.go:24:15: string literal not terminated
```

I've prepared the task branch. Now I need to look at the file to see the issue.

From the show file result, I can see on line 24-25:
```go
	if stdout != "hello
" {
```

This is a broken string literal - the string spans two lines without proper escaping or concatenation. It should be:
```go
	if stdout != "hello
" {
```

Let me fix this by replacing the broken multi-line string with the proper escaped version.

write file internal/workcell/backend_test.go find: 	if stdout != "hello
" {
 replace: 	if stdout != "hello
" {

Wait, the command format might not handle the newline in the find pattern well. Let me try a different approach - I'll replace just the opening part and the closing part separately, or find a pattern that works.

Actually, looking at the verified command result, the file content shows:
```
	if stdout != "hello
" {
```

So the string is literally broken across lines. Let me try to match this exactly.

write file internal/workcell/backend_test.go find: stdout != "hello
 replace: stdout != "hello
" {

Hmm, that might still have issues. Let me try replacing the whole if statement line or use a different strategy.

Actually, let me try with the exact content including the newline:

write file internal/workcell/backend_test.go find: 	if stdout != "hello
" {
 replace: 	if stdout != "hello
" {

Or maybe I should try:

write file internal/workcell/backend_test.go find: "hello
" {
 replace: "hello
" {

Let me try this more targeted approach.  write file internal/workcell/backend_test.go find: "hello
" { replace: "hello
" {
" {
		t.Fatalf("Stdout = %q, want "hello\n"", stdout)
	}
	if stderr != "" {
		t.Fatalf("Stderr = %q, want empty", stderr)
	}
}

func TestPodmanBackendRunFalseCommand(t *testing.T) {
	backend := NewPodmanBackend("docker.io/library/alpine:3.20")
	
	ctx := context.Background()
	job := Job{
		ID:      "test-job-456",
		Command: []string{"false"},
	}
	
	exitCode, stdout, stderr, err := backend.Run(ctx, job)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if exitCode != 1 {
		t.Fatalf("ExitCode = %d, want 1", exitCode)
	}
	// Cleanup should succeed even for failed commands
	err = backend.Cleanup(ctx, job)
	if err != nil {
		t.Fatalf("Cleanup returned error: %v", err)
	}
}

func TestPodmanBackendCleanup(t *testing.T) {
	backend := NewPodmanBackend("docker.io/library/alpine:3.20")
	
	ctx := context.Background()
	job := Job{
		ID:      "test-job-789",
		Command: []string{"echo", "test"},
	}
	
	// Cleanup should succeed even if container was already removed by --rm
	err := backend.Cleanup(ctx, job)
	if err != nil {
		t.Fatalf("Cleanup returned error: %v", err)
	}
}