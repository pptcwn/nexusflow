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

    score := 0.0

    // === WebGL Spoofing Detection ===
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

    // === Advanced Bot Detection ===
    if fp.IsHeadless || fp.Webdriver || fp.AutomationFlags {
        return Decision{Mode: "s", Score: 0, Confidence: 0, AllowProgressive: false}
    }

    // Mouse Pattern
    if fp.MousePatternScore < 40 && fp.MouseMoves > 12 {
        score -= 22
    }

    // Jerk
    if fp.JerkScore < 45 {
        score -= 20
    }

    // Acceleration Distribution
    if fp.AccelerationDistributionScore < 50 {
        score -= 22
    }

    // Hardware / Plugin
    if fp.HardwareConcurrency == 0 || fp.HardwareConcurrency > 32 {
        score -= 10
    }
    if fp.PluginCount == 0 {
        score -= 8
    }

    // Normal Scoring
    if fp.CanvasHash != "" { score += 18 }
    if fp.WebGLVendor != "" { score += 12 }
    if fp.HardwareConcurrency >= 2 { score += 8 }

    if fp.BehaviorScore > 0.75 { score += 32 }
    else if fp.BehaviorScore > 0.6 { score += 22 }
    else if fp.BehaviorScore > 0.45 { score += 12 }

    if fp.WatchTime > 5 { score += 12 }
    if fp.HasInteraction { score += 10 }
    if fp.TimeOnPage > 8 { score += 8 }

    if isGoodReferer(fp.Referer) { score += 8 }

    // ====================== ปรับเกณฑ์การตัดสินใจ ======================
    confidence := score / 115
    if confidence > 1 {
        confidence = 1
    }

    mode := "s"
    allowProgressive := false

    // เกณฑ์หลัก
    if score >= 58 && confidence >= 0.52 {
        mode = "m"
        allowProgressive = true
    }

    // เงื่อนไขพิเศษ: ถ้า WebGL + Mouse Pattern + Jerk ดีมาก
    if fp.WebGLDetection != nil &&
        !fp.WebGLDetection.Inconsistency &&
        fp.MousePatternScore > 70 &&
        fp.JerkScore > 65 {
        if score >= 50 {
            mode = "m"
            allowProgressive = true
        }
    }

    return Decision{
        Mode:             mode,
        Variant:          generateVariant(fp.IP),
        Score:            score,
        Confidence:       confidence,
        AllowProgressive: allowProgressive,
    }
}

func isGoodReferer(referer string) bool {
    good := []string{"facebook.com", "instagram.com", "google.com", "t.co", "youtube.com"}
    for _, g := range good {
        if strings.Contains(referer, g) {
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