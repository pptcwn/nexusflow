package fingerprint

type Fingerprint struct {
	IP                            string                 `json:"ip"`
	Country                       string                 `json:"country"`
	UserAgent                     string                 `json:"userAgent"`
	Referer                       string                 `json:"referer"`
	Timestamp                     int64                  `json:"timestamp"`
	CanvasHash                    string                 `json:"canvasHash"`
	WebGLVendor                   string                 `json:"webglVendor"`
	HardwareConcurrency           int                    `json:"hardwareConcurrency"`
	BehaviorScore                 float64                `json:"behaviorScore"`
	WatchTime                     float64                `json:"watchTime"`
	HasInteraction                bool                   `json:"hasInteraction"`
	TimeOnPage                    float64                `json:"timeOnPage"`
	IsHeadless                    bool                   `json:"isHeadless"`
	Webdriver                     bool                   `json:"webdriver"`
	AutomationFlags               bool                   `json:"automationFlags"`
	PluginCount                   int                    `json:"pluginCount"`
	MouseMoves                    int                    `json:"mouseMoves"`
	MousePatternScore             float64                `json:"mousePatternScore"`
	JerkScore                     float64                `json:"jerkScore"`
	AccelerationDistributionScore float64                `json:"accelerationDistributionScore"`
	WebGLDetection                *WebGLDetection        `json:"webglDetection"`
	AudioContext                  map[string]interface{} `json:"audioContext"`
}

type WebGLDetection struct {
	Inconsistency           bool `json:"inconsistency"`
	SuspiciousVendor        bool `json:"suspiciousVendor"`
	MultipleCallsConsistent bool `json:"multipleCallsConsistent"`
}
