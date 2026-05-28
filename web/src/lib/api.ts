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

// ── Config ──────────────────────────────────────────────

// API base URL.
//   • Dev: VITE_API_BASE_URL пустая → используем "/api/v1" (Vite dev-proxy форвардит на backend).
//   • Prod (Vercel): VITE_API_BASE_URL = "https://your-backend.railway.app/api/v1".
//
// Без этого все вызовы (вкл. /auth/signup, /auth/login) на production
// бьют по самому Vercel → возвращается SPA index.html → JSON.parse() падает
// с "Unexpected token '<'" / "Unexpected token 'T'".
const API_BASE = (import.meta.env.VITE_API_BASE_URL as string | undefined) || "/api/v1";

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
  filePath: string;
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
  url?: string;
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
  ): () => void {
    console.log("DEBUG 1: Внутри функции generateProjectStream", { baseURL: this.baseURL, mode: request.mode, specLen: request.specification?.length });

    let abortController: AbortController | null = null;

    try {
      // Проверка токена
      const token = localStorage.getItem("auth_token");
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
        // Accumulate files sent individually via 'file' events (chunked delivery)
        const pendingFiles: Record<string, string> = {};
        let resultMeta: SSEResultMetaEvent | null = null;

        try {
          while (true) {
            const { done, value } = await reader.read();
            
            if (done) {
              console.log("🏁 SSE stream ended after", chunkCount, "chunks, resultDelivered=", resultDelivered, "pendingFiles=", Object.keys(pendingFiles).length);
              if (!resultDelivered && Object.keys(pendingFiles).length > 0) {
                console.log("🔧 Delivering accumulated files from stream end");
                resultDelivered = true;
                onResult({ files: pendingFiles, ...(resultMeta ?? {}) });
              }
              // Try fetching from server if session_id available
              if (!resultDelivered && request.session_id) {
                try {
                  console.log("📦 Stream ended — trying server file fetch for session", request.session_id);
                  const filesRes = await fetch(`${this.baseURL}/generate/files?session_id=${encodeURIComponent(request.session_id)}`);
                  if (filesRes.ok) {
                    const filesData = await filesRes.json() as { files?: Record<string, string> };
                    if (filesData.files && Object.keys(filesData.files).length > 0) {
                      console.log(`✅ Recovered ${Object.keys(filesData.files).length} files from server after stream end`);
                      resultDelivered = true;
                      onResult({ files: filesData.files, ...(resultMeta ?? {}) });
                    }
                  }
                } catch (e) { console.warn("File recovery fetch failed:", e); }
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
                    const htmlMatch = rawData.match(/<!DOCTYPE[\s\S]*<\/html>/i)
                      || rawData.match(/<html[\s\S]*<\/html>/i);
                    if (htmlMatch) {
                      console.log("✅ Extracted HTML from broken JSON:", htmlMatch[0].length, "chars");
                      pendingFiles["index.html"] = htmlMatch[0];
                    }
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
                  case "file_meta": {
                    // Server stores file content — we only get metadata via SSE
                    const fm = payload as { name?: string; size?: number };
                    if (typeof fm.name === "string") {
                      console.log(`📄 SSE file_meta: '${fm.name}' (${fm.size} bytes stored server-side)`);
                      if (onFile) onFile({ name: fm.name, size: fm.size ?? 0 });
                    }
                    break;
                  }
                  case "file": {
                    // Legacy fallback: small files might still arrive inline
                    const f = payload as Partial<SSEFileEvent>;
                    if (typeof f.name === "string" && typeof f.content === "string") {
                      console.log(`📄 SSE file (legacy): '${f.name}' (${f.content.length} chars)`);
                      pendingFiles[f.name] = f.content;
                      if (onFile) onFile({ name: f.name, size: f.content.length });
                    }
                    break;
                  }
                  case "file_start":
                  case "file_chunk":
                  case "file_end":
                  case "files_batch":
                    // Legacy chunked delivery — no longer used but ignored gracefully
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
                  case "replan": {
                    // Backend closed stream for re-planning — frontend should restart with enriched spec
                    const rp = payload as { feedback?: string; session_id?: string };
                    console.log("🔄 SSE replan event: feedback=", rp.feedback);
                    window.dispatchEvent(new CustomEvent("istok:replan", {
                      detail: { feedback: rp.feedback ?? "", session_id: rp.session_id ?? "" },
                    }));
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

                    // If we have legacy pending files (from inline delivery), use them
                    if (Object.keys(pendingFiles).length > 0) {
                      console.log("🎯 Delivering", Object.keys(pendingFiles).length, "inline files");
                      resultDelivered = true;
                      onResult({ files: pendingFiles, ...(resultMeta ?? {}) });
                      return;
                    }

                    // Fetch files from server via GET endpoint (new architecture)
                    const sid = donePayload.session_id;
                    if (sid && donePayload.file_count && donePayload.file_count > 0) {
                      console.log(`📦 Fetching ${donePayload.file_count} files from server for session ${sid}...`);
                      try {
                        const filesRes = await fetch(`${this.baseURL}/generate/files?session_id=${encodeURIComponent(sid)}`);
                        if (filesRes.ok) {
                          const filesData = await filesRes.json() as { files?: Record<string, string>; file_count?: number };
                          if (filesData.files && Object.keys(filesData.files).length > 0) {
                            console.log(`✅ Fetched ${Object.keys(filesData.files).length} files from server`);
                            resultDelivered = true;
                            onResult({ files: filesData.files, ...(resultMeta ?? {}) });
                            return;
                          }
                        } else {
                          console.error("❌ Failed to fetch files:", filesRes.status, await filesRes.text());
                        }
                      } catch (fetchErr) {
                        console.error("❌ File fetch error:", fetchErr);
                      }
                    }

                    if (!resultDelivered) {
                      console.error("⚠️ done received but no files available!");
                      onError(new Error("Stream completed but no result was received"));
                    }
                    return;
                  }
                }
              }
            }
          }
        } catch (readerErr) {
          console.error("🚨 КРИТИЧЕСКАЯ ОШИБКА SSE (reader loop):", readerErr);
          
          // Try fetching files from server (they may have been stored before disconnect)
          if (!resultDelivered && request.session_id) {
            try {
              console.log("📦 Disconnect recovery — fetching files for session", request.session_id);
              const filesRes = await fetch(`${this.baseURL}/generate/files?session_id=${encodeURIComponent(request.session_id)}`);
              if (filesRes.ok) {
                const filesData = await filesRes.json() as { files?: Record<string, string> };
                if (filesData.files && Object.keys(filesData.files).length > 0) {
                  console.log(`✅ Recovered ${Object.keys(filesData.files).length} files from server after disconnect`);
                  resultDelivered = true;
                  onResult({ files: filesData.files, ...(resultMeta ?? {}) });
                  if (onDisconnect) {
                    onDisconnect({ filesReceived: Object.keys(filesData.files).length, error: String(readerErr) });
                  }
                }
              }
            } catch (e) { console.warn("Disconnect recovery fetch failed:", e); }
          }

          if (!resultDelivered && Object.keys(pendingFiles).length > 0) {
            console.log("🔧 Partial delivery on disconnect:", Object.keys(pendingFiles).length, "files");
            resultDelivered = true;
            onResult({ files: pendingFiles, ...(resultMeta ?? {}) });
            if (onDisconnect) {
              onDisconnect({ filesReceived: Object.keys(pendingFiles).length, error: String(readerErr) });
            }
          } else if (!resultDelivered) {
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
      throw new Error(err.error || `HTTP ${res.status}`);
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
      throw new Error(err.error || `HTTP ${res.status}`);
    }
  }

  /**
   * Media Studio: generate a single image preview from a prompt.
   */
  async generateMediaPreview(prompt: string, width?: number, height?: number): Promise<string> {
    const res = await fetch(`${this.baseURL}/generate/media/preview`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ prompt, width, height }),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: "Preview generation failed" }));
      throw new Error(err.error || `HTTP ${res.status}`);
    }
    const data = await res.json();
    return data.url || "";
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
      throw new Error(err.error || `HTTP ${res.status}`);
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
