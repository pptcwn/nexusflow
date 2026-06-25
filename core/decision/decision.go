package decision

import (
	"strings"

	"nexusflow/core/config"
	"nexusflow/core/fingerprint"
)

type Decision struct {
	Mode             string  `json:"mode"`
	Variant          int     `json:"variant"`
	Score            float64 `json:"score"`
	Confidence       float64 `json:"confidence"`
	AllowProgressive bool    `json:"allowProgressive"`
}

func MakeDecision(fp fingerprint.Fingerprint, cfg *config.Config) Decision {
	if cfg.KillSwitchEnabled {
		return Decision{Mode: cfg.GlobalMode, Confidence: 0, AllowProgressive: false}
	}

	if fp.IsHeadless || fp.Webdriver || fp.AutomationFlags {
		return Decision{Mode: "review", Score: 0, Confidence: 0, AllowProgressive: false}
	}

	score := scoreFingerprint(fp)
	confidence := score / 115
	if confidence > 1 {
		confidence = 1
	}
	if confidence < 0 {
		confidence = 0
	}

	mode := "review"
	allowProgressive := false
	if score >= 58 && confidence >= 0.52 {
		mode = "standard"
		allowProgressive = true
	}

	return Decision{
		Mode:             mode,
		Variant:          generateVariant(fp.IP),
		Score:            score,
		Confidence:       confidence,
		AllowProgressive: allowProgressive,
	}
}

func scoreFingerprint(fp fingerprint.Fingerprint) float64 {
	score := 0.0

	if fp.WebGLDetection != nil {
		if fp.WebGLDetection.Inconsistency {
			score -= 20
		}
		if fp.WebGLDetection.SuspiciousVendor {
			score -= 15
		}
		if !fp.WebGLDetection.MultipleCallsConsistent {
			score -= 18
		}
	}

	if fp.MousePatternScore < 40 && fp.MouseMoves > 12 {
		score -= 22
	}
	if fp.JerkScore < 45 {
		score -= 20
	}
	if fp.AccelerationDistributionScore < 50 {
		score -= 22
	}
	if fp.HardwareConcurrency == 0 || fp.HardwareConcurrency > 32 {
		score -= 10
	}
	if fp.PluginCount == 0 {
		score -= 8
	}
	if fp.CanvasHash != "" {
		score += 18
	}
	if fp.WebGLVendor != "" {
		score += 12
	}
	if fp.HardwareConcurrency >= 2 {
		score += 8
	}
	if fp.BehaviorScore > 0.75 {
		score += 32
	} else if fp.BehaviorScore > 0.6 {
		score += 22
	} else if fp.BehaviorScore > 0.45 {
		score += 12
	}
	if fp.WatchTime > 5 {
		score += 12
	}
	if fp.HasInteraction {
		score += 10
	}
	if fp.TimeOnPage > 8 {
		score += 8
	}
	if isGoodReferer(fp.Referer) {
		score += 8
	}

	return score
}

func isGoodReferer(referer string) bool {
	good := []string{"facebook.com", "instagram.com", "google.com", "t.co", "youtube.com"}
	for _, value := range good {
		if strings.Contains(strings.ToLower(referer), value) {
			return true
		}
	}
	return false
}

func generateVariant(ip string) int {
	hash := 0
	for _, char := range ip {
		hash = (hash*31 + int(char)) % 10
	}
	return hash
}
