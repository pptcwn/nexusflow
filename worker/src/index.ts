export interface Env {
  BACKEND_URL: string;
  STANDARD_ORIGIN: string;
  REVIEW_ORIGIN: string;
}

type Decision = {
  mode: "review" | "standard";
  confidence: number;
  allowProgressive: boolean;
};

const REVIEW_DECISION: Decision = {
  mode: "review",
  confidence: 0,
  allowProgressive: false,
};

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url);

    const botScore = request.headers.get("cf-bot-score");
    const threatScore = request.headers.get("cf-threat-score");
    if (isSuspiciousCloudflareSignal(botScore, threatScore)) {
      return fetchOrigin(request, env.REVIEW_ORIGIN, url);
    }

    const fingerprint = {
      ip: request.headers.get("cf-connecting-ip") || "",
      country: request.headers.get("cf-ipcountry") || "XX",
      userAgent: request.headers.get("user-agent") || "",
      referer: request.headers.get("referer") || "",
      timestamp: Date.now(),
    };

    const decision = await getDecision(fingerprint, env.BACKEND_URL);
    const origin = decision.mode === "standard" ? env.STANDARD_ORIGIN : env.REVIEW_ORIGIN;

    return fetchOrigin(request, origin, url);
  },
};

function isSuspiciousCloudflareSignal(botScore: string | null, threatScore: string | null): boolean {
  const parsedBotScore = botScore ? Number.parseInt(botScore, 10) : null;
  const parsedThreatScore = threatScore ? Number.parseInt(threatScore, 10) : null;

  return (
    (parsedBotScore !== null && parsedBotScore < 20) ||
    (parsedThreatScore !== null && parsedThreatScore > 45)
  );
}

async function getDecision(fp: Record<string, unknown>, backendUrl: string): Promise<Decision> {
  try {
    const res = await fetch(`${backendUrl}/decide`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(fp),
    });
    if (!res.ok) {
      return REVIEW_DECISION;
    }

    const decision = (await res.json()) as Partial<Decision>;
    if (decision.mode !== "standard" && decision.mode !== "review") {
      return REVIEW_DECISION;
    }

    return {
      mode: decision.mode,
      confidence: typeof decision.confidence === "number" ? decision.confidence : 0,
      allowProgressive: decision.allowProgressive === true,
    };
  } catch {
    return REVIEW_DECISION;
  }
}

function fetchOrigin(request: Request, origin: string, requestUrl: URL): Promise<Response> {
  const target = new URL(requestUrl.pathname + requestUrl.search, origin);
  return fetch(new Request(target, request));
}
