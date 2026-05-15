package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthorizeValidationJobFailsClosedWithoutConfiguredToken(t *testing.T) {
	t.Setenv("WORKCELL_VALIDATION_API_TOKEN", "")
	request := httptest.NewRequest(http.MethodPost, "/v1/validation-jobs", nil)
	response := httptest.NewRecorder()

	if authorizeValidationJob(response, request) {
		t.Fatal("authorizeValidationJob returned true without configured token")
	}
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
}

func TestAuthorizeValidationJobRequiresBearerToken(t *testing.T) {
	t.Setenv("WORKCELL_VALIDATION_API_TOKEN", "secret-token")
	request := httptest.NewRequest(http.MethodPost, "/v1/validation-jobs", nil)
	response := httptest.NewRecorder()

	if authorizeValidationJob(response, request) {
		t.Fatal("authorizeValidationJob returned true without bearer token")
	}
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func TestAuthorizeValidationJobAcceptsConfiguredBearerToken(t *testing.T) {
	t.Setenv("WORKCELL_VALIDATION_API_TOKEN", "secret-token")
	request := httptest.NewRequest(http.MethodPost, "/v1/validation-jobs", nil)
	request.Header.Set("authorization", "Bearer secret-token")
	response := httptest.NewRecorder()

	if !authorizeValidationJob(response, request) {
		t.Fatal("authorizeValidationJob returned false for configured bearer token")
	}
}
