package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"nexusflow/core/config"
)

func TestBehaviorReturnsNoContentForValidJSON(t *testing.T) {
	router := setupRouter(config.Default())
	req := httptest.NewRequest(http.MethodPost, "/behavior", strings.NewReader(`{"mouseMoves":3}`))
	req.Header.Set("content-type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestBehaviorReturnsBadRequestForMalformedJSON(t *testing.T) {
	router := setupRouter(config.Default())
	req := httptest.NewRequest(http.MethodPost, "/behavior", strings.NewReader(`{"mouseMoves":`))
	req.Header.Set("content-type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
