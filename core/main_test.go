package main

import (
	"encoding/json"
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

func TestDecideReturnsReviewForBlockedCountry(t *testing.T) {
	router := setupRouter(config.Default())
	req := httptest.NewRequest(http.MethodPost, "/decide", strings.NewReader(`{"ip":"203.0.113.10","country":"RU","referer":"https://google.com/search?q=example","canvasHash":"canvas-hash","webglVendor":"Intel Inc.","hardwareConcurrency":8,"pluginCount":3,"behaviorScore":0.8,"watchTime":9,"hasInteraction":true,"timeOnPage":15,"mouseMoves":30,"mousePatternScore":85,"jerkScore":82,"accelerationDistributionScore":78,"webglDetection":{"multipleCallsConsistent":true}}`))
	req.Header.Set("content-type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response JSON error: %v", err)
	}
	if body["mode"] != "review" {
		t.Fatalf("mode = %v, want review", body["mode"])
	}
	if body["allowProgressive"] != false {
		t.Fatalf("allowProgressive = %v, want false", body["allowProgressive"])
	}
}

func TestDecideRejectsRequestBodyOver64KB(t *testing.T) {
	router := setupRouter(config.Default())
	body := `{"ip":"203.0.113.10","padding":"` + strings.Repeat("x", 70*1024) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/decide", strings.NewReader(body))
	req.Header.Set("content-type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusRequestEntityTooLarge)
	}
}
