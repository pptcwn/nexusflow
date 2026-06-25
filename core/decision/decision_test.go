package decision

import (
	"testing"

	"nexusflow/core/config"
	"nexusflow/core/fingerprint"
)

func TestMakeDecisionReturnsReviewForAutomationSignals(t *testing.T) {
	cfg := config.Default()
	fp := fingerprint.Fingerprint{IsHeadless: true, Webdriver: true}

	got := MakeDecision(fp, cfg)

	if got.Mode != "review" {
		t.Fatalf("Mode = %q, want review", got.Mode)
	}
	if got.AllowProgressive {
		t.Fatalf("AllowProgressive = true, want false")
	}
}

func TestMakeDecisionAllowsTrustedHumanSignals(t *testing.T) {
	cfg := config.Default()
	fp := fingerprint.Fingerprint{
		IP:                            "203.0.113.10",
		Referer:                       "https://google.com/search?q=example",
		CanvasHash:                    "canvas-hash",
		WebGLVendor:                   "Intel Inc.",
		HardwareConcurrency:           8,
		PluginCount:                   3,
		BehaviorScore:                 0.8,
		WatchTime:                     9,
		HasInteraction:                true,
		TimeOnPage:                    15,
		MouseMoves:                    30,
		MousePatternScore:             85,
		JerkScore:                     82,
		AccelerationDistributionScore: 78,
		WebGLDetection: &fingerprint.WebGLDetection{
			MultipleCallsConsistent: true,
		},
	}

	got := MakeDecision(fp, cfg)

	if got.Mode != "standard" {
		t.Fatalf("Mode = %q, want standard", got.Mode)
	}
	if !got.AllowProgressive {
		t.Fatalf("AllowProgressive = false, want true")
	}
	if got.Confidence <= 0 {
		t.Fatalf("Confidence = %v, want > 0", got.Confidence)
	}
}
