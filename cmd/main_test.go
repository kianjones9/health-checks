package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestExecuteProbeSuccess(t *testing.T) {
	// Create a test server that always returns a 200 status code
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	probe := &Probe{
		URL:    server.URL,
		Method: "GET",
	}

	err := executeProbe(probe)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if probe.Availability.Successes != 1 {
		t.Errorf("Expected 1 success, got %d", probe.Availability.Successes)
	}

	if probe.Availability.Failures != 0 {
		t.Errorf("Expected 0 failures, got %d", probe.Availability.Failures)
	}
}

func TestExecuteProbeFailure(t *testing.T) {
	// Create a test server that always returns a 500 status code
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	probe := &Probe{
		URL:    server.URL,
		Method: "GET",
	}

	err := executeProbe(probe)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if probe.Availability.Successes != 0 {
		t.Errorf("Expected 0 successes, got %d", probe.Availability.Successes)
	}

	if probe.Availability.Failures != 1 {
		t.Errorf("Expected 1 failure, got %d", probe.Availability.Failures)
	}
}

func TestExecuteProbeTimeout(t *testing.T) {
	// Create a test server that delays the response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	probe := &Probe{
		URL:    server.URL,
		Method: "GET",
	}

	err := executeProbe(probe)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if probe.Availability.Successes != 0 {
		t.Errorf("Expected 0 successes, got %d", probe.Availability.Successes)
	}

	if probe.Availability.Failures != 1 {
		t.Errorf("Expected 1 failure, got %d", probe.Availability.Failures)
	}
}

func TestExecuteProbeWithHeadersAndBody(t *testing.T) {
	// Create a test server that checks headers and body
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Test-Header") != "test-value" {
			t.Errorf("Expected header X-Test-Header to be test-value, got %v", r.Header.Get("X-Test-Header"))
		}
		body := new(strings.Builder)
		_, err := io.Copy(body, r.Body)
		if err != nil {
			t.Fatalf("Error reading request body: %v", err)
		}
		if body.String() != "test-body" {
			t.Errorf("Expected body to be test-body, got %v", body.String())
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	probe := &Probe{
		URL:    server.URL,
		Method: "POST",
		Headers: map[string]string{
			"X-Test-Header": "test-value",
		},
		Body: "test-body",
	}

	err := executeProbe(probe)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if probe.Availability.Successes != 1 {
		t.Errorf("Expected 1 success, got %d", probe.Availability.Successes)
	}

	if probe.Availability.Failures != 0 {
		t.Errorf("Expected 0 failures, got %d", probe.Availability.Failures)
	}
}
