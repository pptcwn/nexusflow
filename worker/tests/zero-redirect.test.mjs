import assert from "node:assert/strict";
import worker from "../.tmp/worker-test.mjs";

const env = {
  BACKEND_URL: "https://backend.example",
  STANDARD_ORIGIN: "https://standard.example",
  REVIEW_ORIGIN: "https://review.example",
};
const redirectStatuses = new Set([301, 302, 303, 307, 308]);

await test("high-confidence standard decision fetches standard origin with no redirect", async () => {
  const calls = stubFetch({
    decision: { mode: "standard", confidence: 0.8, allowProgressive: true },
  });

  const response = await worker.fetch(new Request("https://edge.example/products?a=1"), env);

  assertNoRedirect(response);
  assert.equal(calls.backend.length, 1);
  assert.equal(calls.origin.length, 1);
  assert.equal(calls.origin[0], "https://standard.example/products?a=1");
});

await test("low-confidence standard decision stays on review origin", async () => {
  const calls = stubFetch({
    decision: { mode: "standard", confidence: 0.2, allowProgressive: true },
  });

  const response = await worker.fetch(new Request("https://edge.example/products"), env);

  assertNoRedirect(response);
  assert.equal(calls.backend.length, 1);
  assert.equal(calls.origin.length, 1);
  assert.equal(calls.origin[0], "https://review.example/products");
});

await test("backend failure falls back to review origin with no redirect", async () => {
  const calls = stubFetch({ backendThrows: true });

  const response = await worker.fetch(new Request("https://edge.example/fallback"), env);

  assertNoRedirect(response);
  assert.equal(calls.backend.length, 1);
  assert.equal(calls.origin.length, 1);
  assert.equal(calls.origin[0], "https://review.example/fallback");
});

await test("suspicious Cloudflare signal routes to review origin without backend call", async () => {
  const calls = stubFetch({
    decision: { mode: "standard", confidence: 0.99, allowProgressive: true },
  });
  const request = new Request("https://edge.example/check", {
    headers: { "cf-bot-score": "10" },
  });

  const response = await worker.fetch(request, env);

  assertNoRedirect(response);
  assert.equal(calls.backend.length, 0);
  assert.equal(calls.origin.length, 1);
  assert.equal(calls.origin[0], "https://review.example/check");
});

function stubFetch({ decision, backendThrows = false }) {
  const calls = { backend: [], origin: [] };

  globalThis.fetch = async (input) => {
    const url = input instanceof Request ? input.url : String(input);

    if (url.startsWith(env.BACKEND_URL)) {
      calls.backend.push(url);
      if (backendThrows) {
        throw new Error("backend unavailable");
      }
      return Response.json(decision);
    }

    calls.origin.push(url);
    return new Response(`origin:${new URL(url).origin}`, { status: 200 });
  };

  return calls;
}

function assertNoRedirect(response) {
  assert.equal(redirectStatuses.has(response.status), false, `got redirect status ${response.status}`);
}

async function test(name, fn) {
  try {
    await fn();
    console.log(`ok - ${name}`);
  } catch (error) {
    console.error(`not ok - ${name}`);
    throw error;
  }
}
