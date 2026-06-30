/**
 * ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 *  ИСТОК АГЕНТ — API Client Layer
 *  Единый модуль связи фронтенда с Go-бэкендом.
 *
 *  Режим работы:
 *    Подключен к реальному Go backend на localhost:8080
 *    Поддержка SSE для стриминга Reasoning шагов
 * ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 */

import { parseAgentText } from "./sse-parsers";
import {
  ProjectListResponseSchema,
  ProjectDetailSchema,
  ProjectSummarySchema,
  UserProfileSchema,
  FolderListResponseSchema,
  WorkspaceListResponseSchema,
  safeParseContract,
  type ProjectSummary,
  type ProjectDetail,
  type UserProfile,
  type CreateProjectRequest,
  type UpdateProjectRequest,
  type RemixProjectRequest,
  type Folder,
  type Workspace,
} from "./contracts";

/** Error subclass that preserves the HTTP status code from failed API responses. */
export class ApiError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

// ── Config ──────────────────────────────────────────────

// API base URL.
//   • Dev: VITE_API_BASE_URL пустая → используем "/api/v1" (Vite dev-proxy форвардит на backend).
//   • Prod (Vercel): VITE_API_BASE_URL = "https://your-backend.railway.app/api/v1".
//
// Без этого все вызовы (вкл. /auth/signup, /auth/login) на production
// бьют по самому Vercel → возвращается SPA index.html → JSON.parse() падает
// с "Unexpected token '<'" / "Unexpected token 'T'".
const API_BASE = (import.meta.env.VITE_API_BASE_URL as string | undefined) || "/api/v1";

/** Public API base (e.g. "https://…/api/v1" in prod, "/api/v1" in dev). */
export const API_BASE_URL = API_BASE;

console.log("🔌 API URL:", API_BASE, "| mode:", import.meta.env.MODE);

// ── Types ───────────────────────────────────────────────

export type GenerationMode = "agent" | "code" | "synthesis";

export interface GenerateRequest {
  specification?: string;
  url?: string;
  messages?: Array<{ role: string; content: string }>;
  mode?: GenerationMode; // "agent" = Инновационное проектирование | "code" = Быстрая генерация | "synthesis" = Адаптивный синтез
  session_id?: string;      // unique session for checkpoint/resume
  resume?: boolean;         // true = resume from last checkpoint
  existing_files?: string[]; // files already received by client
  generate_video?: boolean; // true = генерировать промо-ролик (Veo-3, последовательно); false = быстрый прототип
}

export interface GenerateResponse {
  projectId?: string;
  status?: string;
  files?: Record<string, string>;
  code?: string;
  message?: string;
  /** Server-side metadata from `result_meta` SSE event. */
  duration?: string;
  assets?: string;
  video?: string;
  file_count?: number;
}

// ── SSE event payload types ─────────────────────────────

export interface SSEStatusEvent {
  agent: string;
  status: string;
  state?: string;
  message?: unknown;
  progress?: number;
  timestamp?: string;
}

export interface SSEFileEvent {
  name: string;
  content: string;
}

export interface SSEResultMetaEvent {
  file_count?: number;
  assets?: string;
  video?: string;
  duration?: string;
}

export interface SSEErrorEvent {
  message?: unknown;
}

export interface SSEThoughtEvent {
  agent: string;
  tag: string; // "PLANNING" | "EXECUTION" | "VALIDATION"
  message: string;
  timestamp?: string;
}

export interface SSEPostMortemEvent {
  report: string;
  timestamp?: string;
}

/** FSM state transition emitted by backend `events.PublishFSMTransition`. */
export interface SSEFSMEvent {
  from?: string;
  to?: string;
  state?: string;
  reason?: string;
  message?: unknown;
  agent?: string;
  timestamp?: string;
}

/** Result delivered to onResult callback (legacy single-blob result event). */
export type SSEResultEvent = GenerateResponse;

/** JSON patch from Editor Agent (Chat-to-Modify). */
export interface FilePatch {
  file_path: string;
  search: string;
  replace: string;
}

/** Media asset for design review (Media Studio). */
export interface MediaAsset {
  id: string;
  type: "image" | "video";
  placement: string; // "hero" | "og" | "logo" | "promo_video"
  label: string;
  prompt: string;
  preview_url?: string;    // stock photo URL (default) or AI-generated preview
  stock_keywords?: string; // comma-separated keywords for Unsplash
  url?: string;            // final generated URL (post-approval)
}

export interface AgentStats {
  model: string;
  modelVersion: string;
  responseTimeMs: number;
  crawlerNodesFound: number;
  generatedFilesCount: number;
  tokensUsed: number;
  costRub: number;
  status: string;
  uptime: string;
}

export interface ReasoningStep {
  step: number;
  type: string;
  description: string;
  status: string;
}

export interface ProjectFiles {
  [filename: string]: string;
}

export interface SignupRequest {
  email: string;
  password: string;
  display_name?: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface User {
  id: string;
  email: string;
  display_name: string;
  created_at: string;
}

// ── Helpers ─────────────────────────────────────────────

/** Local alias preserving previous call-site name. */
const extractMessage = (raw: unknown): string => parseAgentText(raw, /* stripThoughts */ false);

// ── API Client ──────────────────────────────────────────

class IstokAPI {
  private baseURL: string;

  constructor(baseURL: string) {
    this.baseURL = baseURL.replace(/\/+$/, ""); // trim trailing slashes to prevent //generate/stream 404
  }

  /**
   * Генерация проекта с поддержкой SSE стриминга
   */
  async generateProject(
    request: GenerateRequest,
    onReasoningStep?: (step: ReasoningStep) => void,
    onProgress?: (message: string) => void
  ): Promise<GenerateResponse> {
    try {
      const response = await fetch(`${this.baseURL}/generate`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(request),
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || "Ошибка генерации проекта");
      }

      const data = await response.json();
      return data;
    } catch (error) {
      console.error("Generate project error:", error);
      throw error;
    }
  }

  /**
   * Генерация проекта с SSE стримингом (S-Tier Orchestrator)
   */
  generateProjectStream(
    request: GenerateRequest,
    onStatus: (status: {
      agent: string;
      status: string;
      state?: string;
      message: string;
      progress: number;
      timestamp?: string;
    }) => void,
    onResult: (result: GenerateResponse) => void,
    onError: (error: Error) => void,
    onFSM?: (transition: {
      from?: string;
      to?: string;
      state?: string;
      reason?: string;
      agent?: string;
      message?: string;
    }) => void,
    onFile?: (file: { name: string; size: number }) => void,
    onDisconnect?: (info: { filesReceived: number; error: string }) => void,
    onThought?: (thought: SSEThoughtEvent) => void,
    onPostMortem?: (pm: SSEPostMortemEvent) => void,
  ): () => void {
    console.log("DEBUG 1: Внутри функции generateProjectStream", { baseURL: this.baseURL, mode: request.mode, specLen: request.specification?.length });

    let abortController: AbortController | null = null;

    try {
      // Проверка токена
      const token = localStorage.getItem("istok_token");
      if (!token) {
        console.warn("ТОКЕН НЕ НАЙДЕН — продолжаем без авторизации (public endpoint)");
      } else {
        console.log("DEBUG 1.1: Токен найден, длина:", token.length);
      }

      const streamURL = `${this.baseURL}/generate/stream`;
      console.log("DEBUG 1.2: streamURL =", streamURL);
      console.log("🔗 SSE connecting:", streamURL, "| body:", JSON.stringify(request).substring(0, 200));

      abortController = new AbortController();

      fetch(streamURL, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...(token ? { "Authorization": `Bearer ${token}` } : {}),
        },
        body: JSON.stringify(request),
        signal: abortController.signal,
      }).then(async (response) => {
        console.log("DEBUG 1.3: fetch завершился, status =", response.status, "ok =", response.ok);
        if (!response.ok) {
          const body = await response.text().catch(() => "");
          console.error(`🚨 SSE HTTP ${response.status} from ${streamURL}:`, body);
          throw new Error(`HTTP ${response.status}: ${body || response.statusText}`);
        }
        console.log("✅ SSE connected, status:", response.status, "content-type:", response.headers.get("content-type"));

        const reader = response.body?.getReader();
        const decoder = new TextDecoder();

        if (!reader) {
          throw new Error("No response body — browser may not support ReadableStream");
        }

        let buffer = "";
        let chunkCount = 0;
        let resultDelivered = false;
        let resultMeta: SSEResultMetaEvent | null = null;

        try {
          while (true) {
            const { done, value } = await reader.read();
            
            if (done) {
              console.log("🏁 SSE stream ended after", chunkCount, "chunks, resultDelivered=", resultDelivered);
              // Stream ended cleanly — fetch files if not already delivered
              if (!resultDelivered && request.session_id) {
                const files = await this.pollForFiles(request.session_id, (msg) => {
                  onStatus({ agent: "system", status: "polling", message: msg, progress: 90 });
                });
                if (files && Object.keys(files).length > 0) {
                  resultDelivered = true;
                  onResult({ files, ...(resultMeta ?? {}) });
                }
              }
              if (!resultDelivered) {
                onError(new Error("SSE stream ended without delivering result"));
              }
              break;
            }

            const chunk = decoder.decode(value, { stream: true });
            chunkCount++;
            if (chunkCount <= 15) console.log(`📦 SSE chunk #${chunkCount} (${chunk.length} bytes):`, chunk.substring(0, 200));

            buffer += chunk;
            const lines = buffer.split("\n\n");
            buffer = lines.pop() || "";

            for (const line of lines) {
              if (!line.trim()) continue;
              if (line.startsWith(":")) continue;

              const eventMatch = line.match(/^event: (.+)$/m);
              const dataMatch = line.match(/^data: (.+)$/m);

              if (eventMatch && dataMatch) {
                const event = eventMatch[1].trim();
                const rawData = dataMatch[1];
                let data: unknown;
                try { data = JSON.parse(rawData); } catch (e) {
                  console.warn(`⚠️ SSE JSON parse error for event '${event}':`, e, "raw_len:", rawData.length, "first200:", rawData.substring(0, 200));
                  if (event === "file" || event === "result") {
                    // Broken JSON — skip (files fetched via GET endpoint)
                  }
                  continue;
                }

                const payload = (data ?? {}) as Record<string, unknown>;
                switch (event) {
                  case "status": {
                    const s = payload as Partial<SSEStatusEvent>;
                    onStatus({
                      agent: String(s.agent ?? ""),
                      status: String(s.status ?? ""),
                      state: typeof s.state === "string" ? s.state : undefined,
                      message: extractMessage(s.message),
                      progress: Number(s.progress ?? 0),
                      timestamp: typeof s.timestamp === "string" ? s.timestamp : undefined,
                    });
                    break;
                  }
                  case "file_meta":
                  case "file":
                  case "file_start":
                  case "file_chunk":
                  case "file_end":
                  case "files_batch":
                    // Files are stored server-side — client fetches via GET on done/disconnect
                    break;
                  case "result_meta": {
                    const m = payload as SSEResultMetaEvent;
                    console.log("📋 SSE result_meta received:", m.file_count, "files, duration:", m.duration);
                    resultMeta = m;
                    break;
                  }
                  case "result": {
                    const r = payload as SSEResultEvent;
                    console.log("🎯 SSE result event received, files:", Object.keys(r.files ?? {}));
                    resultDelivered = true;
                    onResult(r);
                    break;
                  }
                  case "user_action": {
                    // Human-in-the-loop: architecture approval request
                    const ua = payload as { agent?: string; message?: string; draft_plan?: string; session_id?: string; progress?: number };
                    onStatus({
                      agent: String(ua.agent ?? "architect"),
                      status: "user_action",
                      message: extractMessage(ua.message),
                      progress: Number(ua.progress ?? 35),
                    });
                    // Emit a custom event for the approval UI
                    window.dispatchEvent(new CustomEvent("istok:user_action", {
                      detail: { draft_plan: ua.draft_plan ?? "", session_id: ua.session_id ?? "" },
                    }));
                    break;
                  }
                  case "media_approval": {
                    // Human-in-the-loop: media asset approval (design review / Media Studio)
                    const ma = payload as { agent?: string; message?: string; media_assets?: MediaAsset[]; session_id?: string; progress?: number };
                    onStatus({
                      agent: String(ma.agent ?? "designer"),
                      status: "media_approval",
                      message: extractMessage(ma.message),
                      progress: Number(ma.progress ?? 70),
                    });
                    window.dispatchEvent(new CustomEvent("istok:media_approval", {
                      detail: { media_assets: ma.media_assets ?? [], session_id: ma.session_id ?? "" },
                    }));
                    break;
                  }
                  case "insufficient_funds": {
                    // Pause & Resume: backend paused due to credit exhaustion
                    const ifp = payload as { session_id?: string; message?: string };
                    onStatus({
                      agent: "system",
                      status: "insufficient_funds",
                      message: extractMessage(ifp.message) || "Недостаточно средств для продолжения генерации",
                      progress: 0,
                    });
                    window.dispatchEvent(new CustomEvent("istok:insufficient_funds", {
                      detail: { session_id: ifp.session_id ?? "" },
                    }));
                    break;
                  }
                  case "replan": {
                    // Backend closed stream for re-planning — frontend should restart with enriched spec
                    const rp = payload as { feedback?: string; session_id?: string };
                    console.log("🔄 SSE replan event: feedback=", rp.feedback);
                    window.dispatchEvent(new CustomEvent("istok:replan", {
                      detail: { feedback: rp.feedback ?? "", session_id: rp.session_id ?? "" },
                    }));
                    break;
                  }
                  case "thought": {
                    if (onThought) {
                      const t = payload as Partial<SSEThoughtEvent>;
                      onThought({
                        agent: String(t.agent ?? ""),
                        tag: String(t.tag ?? ""),
                        message: extractMessage(t.message),
                        timestamp: typeof t.timestamp === "string" ? t.timestamp : undefined,
                      });
                    }
                    break;
                  }
                  case "postmortem": {
                    if (onPostMortem) {
                      const pm = payload as Partial<SSEPostMortemEvent>;
                      onPostMortem({
                        report: extractMessage(pm.report),
                        timestamp: typeof pm.timestamp === "string" ? pm.timestamp : undefined,
                      });
                    }
                    break;
                  }
                  case "fsm": {
                    const fsm = payload as SSEFSMEvent;
                    if (onFSM) {
                      onFSM({
                        from: typeof fsm.from === "string" ? fsm.from : undefined,
                        to: typeof fsm.to === "string" ? fsm.to : undefined,
                        state: typeof fsm.state === "string" ? fsm.state : undefined,
                        reason: typeof fsm.reason === "string" ? fsm.reason : undefined,
                        agent: typeof fsm.agent === "string" ? fsm.agent : undefined,
                        message: extractMessage(fsm.message),
                      });
                    }
                    break;
                  }
                  case "error": {
                    const e = payload as SSEErrorEvent;
                    const errMsg = extractMessage(e.message) || "Unknown error";
                    // Intercept replan_requested — not a fatal error, trigger feedback loop
                    if (errMsg.includes("replan_requested")) {
                      console.log("🔄 SSE error contains replan_requested — dispatching istok:replan");
                      window.dispatchEvent(new CustomEvent("istok:replan", {
                        detail: { feedback: "", session_id: "" },
                      }));
                      break;
                    }
                    onError(new Error(errMsg));
                    break;
                  }
                  case "done": {
                    const donePayload = payload as { session_id?: string; file_count?: number };
                    console.log("✅ SSE done event received, session_id=", donePayload.session_id, "file_count=", donePayload.file_count);
                    if (resultDelivered) return;

                    const sid = donePayload.session_id || request.session_id;
                    if (sid) {
                      const files = await this.pollForFiles(sid, (msg) => {
                        onStatus({ agent: "system", status: "polling", message: msg, progress: 95 });
                      });
                      if (files && Object.keys(files).length > 0) {
                        resultDelivered = true;
                        onResult({ files, ...(resultMeta ?? {}) });
                        return;
                      }
                    }
                    if (!resultDelivered) {
                      onError(new Error("Stream completed but no files could be fetched"));
                    }
                    return;
                  }
                }
              }
            }
          }
        } catch (readerErr) {
          console.error("🚨 SSE disconnect:", readerErr);

          // Poll server for files — generation continues server-side after disconnect
          if (!resultDelivered && request.session_id) {
            console.log("📦 SSE disconnected — polling server for files (generation still running)...");
            onStatus({ agent: "system", status: "recovering", message: "⏳ Получение файлов с сервера...", progress: 95 });
            const files = await this.pollForFiles(request.session_id, (msg) => {
              onStatus({ agent: "system", status: "polling", message: msg, progress: 90 });
            });
            if (files && Object.keys(files).length > 0) {
              console.log(`✅ Recovered ${Object.keys(files).length} files from server`);
              resultDelivered = true;
              onResult({ files, ...(resultMeta ?? {}) });
              if (onDisconnect) {
                onDisconnect({ filesReceived: Object.keys(files).length, error: String(readerErr) });
              }
            }
          }

          if (!resultDelivered) {
            if (onDisconnect) {
              onDisconnect({ filesReceived: 0, error: String(readerErr) });
            }
            onError(readerErr instanceof Error ? readerErr : new Error(String(readerErr)));
          }
        }
      }).catch((error) => {
        console.error("🚨 SSE fetch/connect error:", error?.message || error, "| URL:", `${this.baseURL}/generate/stream`);
        onError(error instanceof Error ? error : new Error(String(error)));
      });
    } catch (outerErr) {
      console.error("КРИТИЧЕСКИЙ СБОЙ ВНУТРИ API (generateProjectStream):", outerErr);
      onError(outerErr instanceof Error ? outerErr : new Error(String(outerErr)));
    }

    return () => {
      console.log("Stream cancelled via abort");
      abortController?.abort();
    };
  }

  /**
   * Poll server for files until generation is complete (or timeout).
   * Relays last_status from the backend so frontend can show real-time orchestrator progress.
   * Returns files map or null if failed.
   */
  private async pollForFiles(
    sessionId: string,
    onStatusUpdate?: (msg: string) => void,
    maxWaitMs = 1_200_000,
    intervalMs = 3_000,
  ): Promise<Record<string, string> | null> {
    const start = Date.now();
    let lastCount = 0;
    let lastStatusMsg = "";
    let lastPendingKind = "";

    while (Date.now() - start < maxWaitMs) {
      try {
        const res = await fetch(`${this.baseURL}/generate/files?session_id=${encodeURIComponent(sessionId)}`);
        if (!res.ok) {
          console.warn(`📦 Poll: HTTP ${res.status} — retrying...`);
          await new Promise(r => setTimeout(r, intervalMs));
          continue;
        }
        const data = await res.json() as {
          files?: Record<string, string>;
          file_count?: number;
          complete?: boolean;
          last_status?: string;
          pending_action?: { kind: string; draft_plan?: string; media_assets?: unknown; session_id: string };
        };
        const count = data.file_count ?? Object.keys(data.files ?? {}).length;
        const elapsed = Math.round((Date.now() - start) / 1000);

        // Relay backend status to UI
        if (data.last_status && data.last_status !== lastStatusMsg) {
          lastStatusMsg = data.last_status;
          console.log(`📦 Poll: backend status → ${lastStatusMsg}`);
          if (onStatusUpdate) {
            onStatusUpdate(lastStatusMsg);
          }
        }

        // Trigger HITL modals during polling fallback (dispatch once per gate)
        if (data.pending_action && data.pending_action.kind !== lastPendingKind) {
          lastPendingKind = data.pending_action.kind;
          const pa = data.pending_action;
          if (pa.kind === "user_action") {
            console.log("📦 Poll: detected pending user_action — dispatching istok:user_action");
            window.dispatchEvent(new CustomEvent("istok:user_action", {
              detail: { draft_plan: pa.draft_plan ?? "", session_id: pa.session_id },
            }));
          } else if (pa.kind === "media_approval") {
            console.log("📦 Poll: detected pending media_approval — dispatching istok:media_approval");
            window.dispatchEvent(new CustomEvent("istok:media_approval", {
              detail: { media_assets: pa.media_assets ?? [], session_id: pa.session_id },
            }));
          } else if (pa.kind === "insufficient_funds") {
            console.log("📦 Poll: detected pending insufficient_funds — dispatching istok:insufficient_funds");
            window.dispatchEvent(new CustomEvent("istok:insufficient_funds", {
              detail: { session_id: pa.session_id },
            }));
          }
        } else if (!data.pending_action && lastPendingKind) {
          lastPendingKind = ""; // gate cleared — allow re-dispatch if a new one appears
        }

        console.log(`📦 Poll: ${count} files, complete=${data.complete} (${elapsed}s elapsed)`);

        if (data.complete && data.files && count > 0) {
          return data.files;
        }

        // complete=true but 0 files = error
        if (data.complete && count === 0) {
          console.error(`📦 Poll: complete=true but 0 files — generation failed. Last status: ${lastStatusMsg}`);
          if (onStatusUpdate) {
            onStatusUpdate(lastStatusMsg || "❌ Генерация завершилась без файлов");
          }
          return null;
        }

        // Still generating — log progress
        if (count > lastCount) {
          lastCount = count;
          console.log(`📦 Poll: ${count} files so far, waiting for completion...`);
        }
      } catch (e) {
        console.warn("📦 Poll error:", e);
      }
      // Adaptive interval: 3s for first 60s, then 5s
      const wait = (Date.now() - start > 60_000) ? 5_000 : intervalMs;
      await new Promise(r => setTimeout(r, wait));
    }

    // Timeout — return whatever we have
    console.warn(`📦 Poll timeout after ${maxWaitMs}ms — fetching final state`);
    try {
      const res = await fetch(`${this.baseURL}/generate/files?session_id=${encodeURIComponent(sessionId)}`);
      if (res.ok) {
        const data = await res.json() as { files?: Record<string, string> };
        if (data.files && Object.keys(data.files).length > 0) return data.files;
      }
    } catch { /* ignore */ }
    return null;
  }

  /**
   * Получение статистики агента
   */
  async getStats(): Promise<AgentStats> {
    try {
      const response = await fetch(`${this.baseURL}/stats`);
      if (!response.ok) {
        throw new Error("Failed to fetch stats");
      }
      return await response.json();
    } catch (error) {
      console.error("Get stats error:", error);
      throw error;
    }
  }

  /**
   * Health check
   */
  async healthCheck(): Promise<{ status: string; uptime: string }> {
    try {
      const response = await fetch(`${this.baseURL}/health`);
      if (!response.ok) {
        throw new Error("Health check failed");
      }
      return await response.json();
    } catch (error) {
      console.error("Health check error:", error);
      throw error;
    }
  }

  /**
   * Преобразование сообщений чата в формат для API
   */
  formatMessages(messages: Array<{ role: string; content: string }>) {
    return messages.map((msg) => ({
      role: msg.role === "user" ? "user" : "assistant",
      content: msg.content,
    }));
  }

  /**
   * Генерация кода из истории чата
   */
  async generateFromChat(
    messages: Array<{ role: string; content: string }>,
    mode: GenerationMode = "code"
  ): Promise<GenerateResponse> {
    const formattedMessages = this.formatMessages(messages);
    
    const lastUserMessage = formattedMessages
      .filter((m) => m.role === "user")
      .pop();

    if (!lastUserMessage) {
      throw new Error("No user message found");
    }

    return this.generateProject({
      specification: lastUserMessage.content,
      messages: formattedMessages,
      mode,
    });
  }

  /**
   * Регистрация нового пользователя
   */
  async signup(request: SignupRequest): Promise<AuthResponse> {
    try {
      const response = await fetch(`${this.baseURL}/auth/signup`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(request),
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || "Ошибка регистрации");
      }

      const data = await response.json();
      
      // Сохраняем токен в localStorage
      if (data.token) {
        localStorage.setItem("istok_token", data.token);
        localStorage.setItem("istok_user", JSON.stringify(data.user));
      }
      
      return data;
    } catch (error) {
      console.error("Signup error:", error);
      throw error;
    }
  }

  /**
   * Вход пользователя
   */
  async login(request: LoginRequest): Promise<AuthResponse> {
    try {
      const response = await fetch(`${this.baseURL}/auth/login`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(request),
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || "Ошибка входа");
      }

      const data = await response.json();
      
      // Сохраняем токен в localStorage
      if (data.token) {
        localStorage.setItem("istok_token", data.token);
        localStorage.setItem("istok_user", JSON.stringify(data.user));
      }
      
      return data;
    } catch (error) {
      console.error("Login error:", error);
      throw error;
    }
  }

  /**
   * Получение текущего пользователя
   */
  async getMe(): Promise<User> {
    try {
      const token = localStorage.getItem("istok_token");
      if (!token) {
        throw new Error("Токен не найден");
      }

      const response = await fetch(`${this.baseURL}/auth/me`, {
        method: "GET",
        headers: {
          "Authorization": `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error("Не авторизован");
      }

      return await response.json();
    } catch (error) {
      console.error("Get me error:", error);
      throw error;
    }
  }

  /**
   * Выход пользователя
   */
  logout(): void {
    localStorage.removeItem("istok_token");
    localStorage.removeItem("istok_user");
  }

  /**
   * Проверка авторизации
   */
  isAuthenticated(): boolean {
    return !!localStorage.getItem("istok_token");
  }

  /**
   * Railway deploy — отправляет project_name + files в POST /api/v1/deploy/railway.
   * Бэкенд вызывает Railway GraphQL API и возвращает status + deploy_url + logs_url.
   */
  async deployToRailway(payload: {
    project_name?: string;
    files: Array<{ path: string; content: string }>;
    env_vars?: Record<string, string>;
  }): Promise<{
    status: "queued" | "deploying" | "success" | "failed" | "unavailable";
    service_id?: string;
    deploy_url?: string;
    logs_url?: string;
    message?: string;
    error?: string;
  }> {
    const res = await fetch(`${this.baseURL}/deploy/railway`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    const data = await res.json().catch(() => ({}));
    return data;
  }

  /**
   * Human-in-the-Loop: submit architecture approval decision.
   */
  async approveArchitecture(sessionId: string, approved: boolean, feedback?: string): Promise<void> {
    const res = await fetch(`${this.baseURL}/generate/approve`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ session_id: sessionId, approved, feedback: feedback || undefined }),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: "Approval failed" }));
      throw new ApiError(err.error || `HTTP ${res.status}`, res.status);
    }
  }

  /**
   * Human-in-the-Loop: submit media asset approval decision (Media Studio).
   */
  async approveMedia(sessionId: string, approved: boolean, assets: MediaAsset[]): Promise<void> {
    const res = await fetch(`${this.baseURL}/generate/approve_media`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ session_id: sessionId, approved, assets }),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: "Media approval failed" }));
      throw new ApiError(err.error || `HTTP ${res.status}`, res.status);
    }
  }

  /**
   * Pause & Resume: signal the backend to resume generation after insufficient funds pause.
   */
  async resumeFunds(sessionId: string): Promise<void> {
    const res = await fetch(`${this.baseURL}/generate/resume_funds`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ session_id: sessionId }),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: "Resume failed" }));
      throw new ApiError(err.error || `HTTP ${res.status}`, res.status);
    }
  }

  /**
   * Media Studio: generate a single image preview from a prompt.
   */
  async generateMediaPreview(assetId: string, prompt: string, width?: number, height?: number): Promise<{ url: string; source: "ai" | "stock" }> {
    const res = await fetch(`${this.baseURL}/generate/media/preview`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ asset_id: assetId, prompt, width, height }),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: "Preview generation failed" }));
      throw new Error(err.error || `HTTP ${res.status}`);
    }
    const data = await res.json();
    return { url: data.url || "", source: data.source || "stock" };
  }

  /**
   * Pre-flight: enhance a short idea into a detailed specification.
   * Optionally includes a competitor reference URL for visual/structural analysis.
   */
  async enhancePrompt(prompt: string, referenceURL?: string): Promise<string> {
    const res = await fetch(`${this.baseURL}/prompt/enhance`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ prompt, reference_url: referenceURL || undefined }),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: "Enhancement failed" }));
      throw new ApiError(err.message || err.error || `HTTP ${res.status}`, res.status);
    }
    const data = await res.json();
    return data.enhanced || "";
  }

  /**
   * Interactive Editor Agent: send chat message with current files, receive JSON patches.
   */
  async sendEditorMessage(
    sessionId: string,
    message: string,
    files: Record<string, string>,
  ): Promise<FilePatch[]> {
    const res = await fetch(`${this.baseURL}/editor/chat`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ session_id: sessionId, message, files }),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: "Editor request failed" }));
      throw new Error(err.error || `HTTP ${res.status}`);
    }
    const data = await res.json();
    return data.patches || [];
  }

  /**
   * Surgical Component Edit: send single file + prompt, receive updated code.
   * Used by Inspector (point-and-click visual editor).
   */
  async editComponent(
    filePath: string,
    currentCode: string,
    prompt: string,
  ): Promise<{ file_path: string; new_code: string }> {
    const res = await fetch(`${this.baseURL}/generate/edit`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        file_path: filePath,
        current_code: currentCode,
        prompt,
      }),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: "Edit request failed" }));
      throw new Error(err.error || `HTTP ${res.status}`);
    }
    return res.json();
  }

  // ── Layer 1: Projects & Profile (real DB-backed reads) ──

  /** Authorization header from the stored JWT, if present. */
  private authHeaders(): Record<string, string> {
    const token = localStorage.getItem("istok_token");
    return token ? { Authorization: `Bearer ${token}` } : {};
  }

  /**
   * GET /api/v1/projects — list the authenticated user's projects.
   */
  async getProjects(): Promise<ProjectSummary[]> {
    const res = await fetch(`${this.baseURL}/projects`, {
      headers: { ...this.authHeaders() },
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new ApiError(err.error || `HTTP ${res.status}`, res.status);
    }
    const data = await res.json().catch(() => ({}));
    const parsed = safeParseContract(ProjectListResponseSchema, data, "GET /projects");
    return parsed.ok ? parsed.data.projects : [];
  }

  /**
   * GET /api/v1/projects/:id — full project (metadata + files) for the builder.
   */
  async getProject(id: string): Promise<ProjectDetail | null> {
    const res = await fetch(`${this.baseURL}/projects/${encodeURIComponent(id)}`, {
      headers: { ...this.authHeaders() },
    });
    if (res.status === 404) return null;
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new ApiError(err.error || `HTTP ${res.status}`, res.status);
    }
    const data = await res.json().catch(() => ({}));
    const parsed = safeParseContract(ProjectDetailSchema, data, "GET /projects/:id");
    return parsed.ok ? parsed.data : null;
  }

  /**
   * POST /api/v1/projects — persist a newly generated project (metadata + files).
   */
  async createProject(payload: CreateProjectRequest): Promise<ProjectDetail> {
    const res = await fetch(`${this.baseURL}/projects`, {
      method: "POST",
      headers: { "Content-Type": "application/json", ...this.authHeaders() },
      body: JSON.stringify(payload),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new ApiError(err.error || `HTTP ${res.status}`, res.status);
    }
    const data = await res.json().catch(() => ({}));
    const parsed = safeParseContract(ProjectDetailSchema, data, "POST /projects");
    if (!parsed.ok) throw new ApiError("Некорректный ответ при создании проекта", 502);
    return parsed.data;
  }

  /**
   * PATCH /api/v1/projects/:id — partial update (rename / move / transfer / files).
   */
  async updateProject(id: string, patch: UpdateProjectRequest): Promise<ProjectSummary> {
    const res = await fetch(`${this.baseURL}/projects/${encodeURIComponent(id)}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json", ...this.authHeaders() },
      body: JSON.stringify(patch),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new ApiError(err.error || `HTTP ${res.status}`, res.status);
    }
    const data = await res.json().catch(() => ({}));
    const parsed = safeParseContract(ProjectSummarySchema, data, "PATCH /projects/:id");
    if (!parsed.ok) throw new ApiError("Некорректный ответ при обновлении проекта", 502);
    return parsed.data;
  }

  /**
   * POST /api/v1/projects/:id/remix — clone a project into a new one.
   */
  async remixProject(id: string, payload: RemixProjectRequest): Promise<ProjectSummary> {
    const res = await fetch(`${this.baseURL}/projects/${encodeURIComponent(id)}/remix`, {
      method: "POST",
      headers: { "Content-Type": "application/json", ...this.authHeaders() },
      body: JSON.stringify(payload),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new ApiError(err.error || `HTTP ${res.status}`, res.status);
    }
    const data = await res.json().catch(() => ({}));
    const parsed = safeParseContract(ProjectSummarySchema, data, "POST /projects/:id/remix");
    if (!parsed.ok) throw new ApiError("Некорректный ответ при ремиксе проекта", 502);
    return parsed.data;
  }

  /**
   * DELETE /api/v1/projects/:id — remove a project owned by the user.
   */
  async deleteProject(id: string): Promise<void> {
    const res = await fetch(`${this.baseURL}/projects/${encodeURIComponent(id)}`, {
      method: "DELETE",
      headers: { ...this.authHeaders() },
    });
    if (!res.ok && res.status !== 404) {
      const err = await res.json().catch(() => ({}));
      throw new ApiError(err.error || `HTTP ${res.status}`, res.status);
    }
  }

  /**
   * GET /api/v1/user/profile — profile + aggregated stats for the profile page.
   */
  async getUserProfile(): Promise<UserProfile> {
    const res = await fetch(`${this.baseURL}/user/profile`, {
      headers: { ...this.authHeaders() },
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new ApiError(err.error || `HTTP ${res.status}`, res.status);
    }
    const data = await res.json().catch(() => ({}));
    const parsed = safeParseContract(UserProfileSchema, data, "GET /user/profile");
    if (!parsed.ok) throw new ApiError("Некорректный ответ профиля", 502);
    return parsed.data;
  }

  /**
   * GET /api/v1/folders — folders for the "move to folder" selector.
   */
  async getFolders(): Promise<Folder[]> {
    const res = await fetch(`${this.baseURL}/folders`, { headers: { ...this.authHeaders() } });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new ApiError(err.error || `HTTP ${res.status}`, res.status);
    }
    const data = await res.json().catch(() => ({}));
    const parsed = safeParseContract(FolderListResponseSchema, data, "GET /folders");
    return parsed.ok ? parsed.data.folders : [];
  }

  /**
   * GET /api/v1/workspaces — workspaces for the "transfer" selector.
   */
  async getWorkspaces(): Promise<Workspace[]> {
    const res = await fetch(`${this.baseURL}/workspaces`, { headers: { ...this.authHeaders() } });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new ApiError(err.error || `HTTP ${res.status}`, res.status);
    }
    const data = await res.json().catch(() => ({}));
    const parsed = safeParseContract(WorkspaceListResponseSchema, data, "GET /workspaces");
    return parsed.ok ? parsed.data.workspaces : [];
  }

  /**
   * Получение сохраненного пользователя
   */
  getCurrentUser(): User | null {
    const userStr = localStorage.getItem("istok_user");
    if (!userStr) return null;
    try {
      return JSON.parse(userStr);
    } catch {
      return null;
    }
  }
}

// ── Export Singleton ────────────────────────────────────

export const api = new IstokAPI(API_BASE);
export default api;
