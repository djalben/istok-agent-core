/**
 * ИСТОК — Test Connection to Railway Backend
 * Run: node test-connection.js
 */

const BACKEND = "https://web-production-18f7f.up.railway.app/api/v1";

async function test() {
  console.log("🔗 Testing connection to:", BACKEND);
  console.log("─".repeat(50));

  // 1. Health check
  try {
    const res = await fetch(`${BACKEND}/health`);
    const data = await res.json();
    console.log(`✅ /health → HTTP ${res.status}`, data);
  } catch (e) {
    console.error("❌ /health FAILED:", e.message);
  }

  // 2. CORS preflight simulation
  try {
    const origin = "https://istok-agent-core-dtoqkzr8x-djalbens-projects.vercel.app";
    const res = await fetch(`${BACKEND}/health`, {
      method: "OPTIONS",
      headers: {
        "Origin": origin,
        "Access-Control-Request-Method": "POST",
        "Access-Control-Request-Headers": "Content-Type",
      },
    });
    const acao = res.headers.get("access-control-allow-origin");
    const acam = res.headers.get("access-control-allow-methods");
    console.log(`✅ CORS preflight → HTTP ${res.status}`);
    console.log(`   Access-Control-Allow-Origin: ${acao}`);
    console.log(`   Access-Control-Allow-Methods: ${acam}`);
    if (acao !== origin && acao !== "*") {
      console.error(`❌ CORS BLOCKED! Origin ${origin} not in allowed list. Got: ${acao}`);
    }
  } catch (e) {
    console.error("❌ CORS preflight FAILED:", e.message);
  }

  // 3. POST to /generate/stream (just check it accepts the request)
  try {
    const res = await fetch(`${BACKEND}/generate/stream`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ specification: "test", mode: "code" }),
    });
    console.log(`✅ /generate/stream → HTTP ${res.status} Content-Type: ${res.headers.get("content-type")}`);
  } catch (e) {
    console.error("❌ /generate/stream FAILED:", e.message);
  }

  console.log("─".repeat(50));
  console.log("Done.");
}

test();
