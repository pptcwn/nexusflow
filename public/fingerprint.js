let behavior = {
    watchTime: 0,
    mouseMoves: 0,
    scrolls: 0,
    clicks: 0,
    timeOnPage: 0,
    hasInteraction: false,
    startTime: Date.now(),
    videoWatched: false,

    // === Advanced Bot Detection ===
    isHeadless: false,
    webdriver: false,
    automationFlags: false,
    pluginCount: 0,
    languageCount: 0,

    // === Mouse Pattern Analysis ===
    mousePatternScore: 100,
    straightMoves: 0,
    erraticMoves: 0,
    averageSpeed: 0,
    lastMouseX: 0,
    lastMouseY: 0,
    lastMouseTime: Date.now(),

    // === Jerk Analysis ===
    jerkScore: 100,
    lastSpeed: 0,
    jerkValues: [],

    // === Acceleration Distribution Analysis ===
    accelerationValues: [],
    accelerationVariance: 0,
    accelerationMean: 0,
    accelerationDistributionScore: 100,

    // === WebGL & Audio Detection ===
    webglDetection: null,
    audioContext: null,
};

// ====================== 1. Advanced Bot Detection ======================
function detectAdvancedBots() {
    if (navigator.webdriver === true) {
        behavior.webdriver = true;
    }

    const flags = [
        window.callPhantom,
        window._phantom,
        window.Buffer,
        window.spawn,
        window.domAutomation,
    ];
    if (flags.some(Boolean)) {
        behavior.automationFlags = true;
    }

    if (/HeadlessChrome/.test(navigator.userAgent)) {
        behavior.isHeadless = true;
    }

    behavior.pluginCount = navigator.plugins ? navigator.plugins.length : 0;
    behavior.languageCount = navigator.languages ? navigator.languages.length : 0;
}

// ====================== 2. Mouse Pattern + Jerk + Acceleration ======================
function analyzeMouseMovement(e) {
    const now = Date.now();
    const dx = e.clientX - behavior.lastMouseX;
    const dy = e.clientY - behavior.lastMouseY;
    const dt = now - behavior.lastMouseTime;

    if (dt === 0 || dt > 150) return;

    const distance = Math.sqrt(dx * dx + dy * dy);
    const currentSpeed = distance / dt;

    // --- Mouse Pattern ---
    const isStraight = (Math.abs(dx) > 50 && Math.abs(dy) < 8) ||
                       (Math.abs(dy) > 50 && Math.abs(dx) < 8);
    if (isStraight) behavior.straightMoves++;
    else behavior.erraticMoves++;

    behavior.averageSpeed = (behavior.averageSpeed * 0.7) + (currentSpeed * 0.3);

    // --- Jerk Analysis ---
    const acceleration = (currentSpeed - behavior.lastSpeed) / dt;
    const jerk = Math.abs(acceleration - (behavior.lastSpeed / dt));

    behavior.jerkValues.push(jerk);
    if (behavior.jerkValues.length > 20) behavior.jerkValues.shift();

    if (behavior.jerkValues.length > 5) {
        const avgJerk = behavior.jerkValues.reduce((a, b) => a + b, 0) / behavior.jerkValues.length;
        let jScore = 100;
        if (avgJerk < 0.0005) jScore -= 35;
        else if (avgJerk > 0.05) jScore -= 20;
        behavior.jerkScore = Math.max(0, jScore);
    }

    // --- Acceleration Distribution ---
    behavior.accelerationValues.push(acceleration);
    if (behavior.accelerationValues.length > 30) behavior.accelerationValues.shift();

    if (behavior.accelerationValues.length > 8) {
        const sum = behavior.accelerationValues.reduce((a, b) => a + b, 0);
        const mean = sum / behavior.accelerationValues.length;
        let variance = 0;
        for (let val of behavior.accelerationValues) {
            variance += Math.pow(val - mean, 2);
        }
        variance /= behavior.accelerationValues.length;

        behavior.accelerationMean = mean;
        behavior.accelerationVariance = variance;

        let distScore = 100;
        if (variance < 0.0001) distScore -= 40;
        else if (variance < 0.001) distScore -= 25;

        const unique = new Set(behavior.accelerationValues.map(v => v.toFixed(5)));
        if (unique.size < 6 && behavior.accelerationValues.length > 15) distScore -= 20;

        behavior.accelerationDistributionScore = Math.max(0, distScore);
    }

    // --- Mouse Pattern Score ---
    let mScore = 100;
    mScore -= behavior.straightMoves * 4;
    mScore += behavior.erraticMoves * 1.5;
    mScore -= Math.max(0, (behavior.averageSpeed - 2) * 8);
    behavior.mousePatternScore = Math.max(0, Math.min(100, mScore));

    // อัพเดทค่าล่าสุด
    behavior.lastSpeed = currentSpeed;
    behavior.lastMouseX = e.clientX;
    behavior.lastMouseY = e.clientY;
    behavior.lastMouseTime = now;
}

// ====================== 3. WebGL Spoofing Detection ======================
function detectWebGLSpoofing() {
    const canvas = document.createElement('canvas');
    const gl = canvas.getContext('webgl') || canvas.getContext('experimental-webgl');

    if (!gl) return { supported: false };

    const result = {
        supported: true,
        inconsistency: false,
        suspiciousVendor: false,
        multipleCallsConsistent: true,
        extensionCount: 0,
    };

    try {
        const calls = [];
        for (let i = 0; i < 3; i++) {
            calls.push({
                vendor: gl.getParameter(gl.VENDOR),
                renderer: gl.getParameter(gl.RENDERER),
            });
        }

        const first = calls[0];
        for (let i = 1; i < calls.length; i++) {
            if (calls[i].vendor !== first.vendor || calls[i].renderer !== first.renderer) {
                result.multipleCallsConsistent = false;
                result.inconsistency = true;
                break;
            }
        }

        const suspicious = ['Microsoft', 'Google Inc.', 'ANGLE'];
        if (suspicious.some(v => calls[0].vendor && calls[0].vendor.includes(v))) {
            result.suspiciousVendor = true;
            result.inconsistency = true;
        }

        result.extensionCount = (gl.getSupportedExtensions() || []).length;
        if (result.extensionCount < 5 || result.extensionCount > 60) {
            result.inconsistency = true;
        }
    } catch (e) {
        result.error = true;
    }

    return result;
}

// ====================== 4. AudioContext Detection ======================
function detectAudioContext() {
    try {
        const AudioContext = window.AudioContext || window.webkitAudioContext;
        if (!AudioContext) return { supported: false };

        const audioCtx = new AudioContext();
        const result = {
            supported: true,
            sampleRate: audioCtx.sampleRate,
            maxChannelCount: audioCtx.destination.maxChannelCount,
        };
        audioCtx.close().catch(() => {});
        return result;
    } catch (e) {
        return { supported: false, error: true };
    }
}

// ====================== 5. Video + Normal Behavior ======================
function initVideoAndBehavior() {
    const videos = document.querySelectorAll('video');
    videos.forEach(video => {
        video.addEventListener('timeupdate', () => {
            if (video.currentTime > behavior.watchTime) {
                behavior.watchTime = video.currentTime;
            }
            if (video.currentTime > 3) {
                behavior.videoWatched = true;
                behavior.hasInteraction = true;
            }
        });
        video.addEventListener('play', () => behavior.hasInteraction = true);
        video.addEventListener('click', () => {
            behavior.clicks++;
            behavior.hasInteraction = true;
        });
    });

    document.addEventListener('mousemove', analyzeMouseMovement);
    document.addEventListener('mousemove', () => behavior.mouseMoves++);
    document.addEventListener('scroll', () => behavior.scrolls++);
    document.addEventListener('click', () => {
        behavior.clicks++;
        behavior.hasInteraction = true;
    });
}

// ====================== 6. Send Data ======================
function sendBehaviorData() {
    behavior.timeOnPage = (Date.now() - behavior.startTime) / 1000;

    fetch('/behavior', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(buildBehaviorPayload()),
        keepalive: true
    }).catch(() => {});
}

function buildBehaviorPayload() {
    return {
        watchTime: roundMetric(behavior.watchTime),
        mouseMoves: behavior.mouseMoves,
        scrolls: behavior.scrolls,
        clicks: behavior.clicks,
        timeOnPage: roundMetric(behavior.timeOnPage),
        timeOnPageBucket: bucketSeconds(behavior.timeOnPage),
        hasInteraction: behavior.hasInteraction,
        videoWatched: behavior.videoWatched,
        isHeadless: behavior.isHeadless,
        webdriver: behavior.webdriver,
        automationFlags: behavior.automationFlags,
        pluginCount: behavior.pluginCount,
        languageCount: behavior.languageCount,
        mousePatternScore: roundMetric(behavior.mousePatternScore),
        averageSpeed: roundMetric(behavior.averageSpeed),
        jerkScore: roundMetric(behavior.jerkScore),
        accelerationMean: roundMetric(behavior.accelerationMean),
        accelerationVariance: roundMetric(behavior.accelerationVariance),
        accelerationDistributionScore: roundMetric(behavior.accelerationDistributionScore),
        webglDetection: behavior.webglDetection,
        audioContext: behavior.audioContext,
    };
}

function roundMetric(value) {
    if (typeof value !== 'number' || !Number.isFinite(value)) return 0;
    return Math.round(value * 1000) / 1000;
}

function bucketSeconds(seconds) {
    if (seconds < 5) return '0-5';
    if (seconds < 15) return '5-15';
    if (seconds < 30) return '15-30';
    if (seconds < 60) return '30-60';
    return '60+';
}

// ====================== 7. Initialization ======================
function initTracking() {
    detectAdvancedBots();
    behavior.webglDetection = detectWebGLSpoofing();
    behavior.audioContext = detectAudioContext();
    initVideoAndBehavior();

    setInterval(sendBehaviorData, 3000);
    window.addEventListener('beforeunload', sendBehaviorData);
}

window.addEventListener('load', initTracking);
