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

// ── Config ──────────────────────────────────────────────

const API_BASE = import.meta.env.VITE_API_BASE_URL || 
  (import.meta.env.MODE === "development" 
    ? "http://localhost:8080/api/v1" 
    : (() => {
        console.error("🚨 КРИТИЧЕСКАЯ ОШИБКА: VITE_API_BASE_URL не установлен в production!");
        console.error("Добавьте переменную окружения VITE_API_BASE_URL в Vercel Dashboard");
        throw new Error("VITE_API_BASE_URL не установлен. Приложение не может работать без backend URL.");
      })()
  );

console.log("🔌 API URL:", import.meta.env.VITE_API_BASE_URL || "(fallback)", "→", API_BASE, "| mode:", import.meta.env.MODE);

// ── Types ───────────────────────────────────────────────

export type GenerationMode = "agent" | "code" | "synthesis";

export interface GenerateRequest {
  specification?: string;
  url?: string;
  messages?: Array<{ role: string; content: string }>;
  mode?: GenerationMode; // "agent" = Инновационное проектирование | "code" = Быстрая генерация | "synthesis" = Адаптивный синтез
}

export interface GenerateResponse {
  projectId: string;
  status: string;
  files?: Record<string, string>;
  code?: string;
  message?: string;
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

/**
 * Safely extract a string from any SSE message field.
 * Claude 3.7 Thinking can return objects like:
 *   { type: "thinking", thinking: "..." }
 *   { type: "text", text: "..." }
 *   { content: [...], reasoning_content: "..." }
 */
function extractMessage(raw: unknown): string {
  if (raw == null) return "";
  if (typeof raw === "string") return raw;
  if (typeof raw === "number" || typeof raw === "boolean") return String(raw);
  if (typeof raw === "object") {
    const obj = raw as Record<string, unknown>;
    const candidate =
      obj.text ??
      obj.content ??
      obj.reasoning_content ??
      obj.thinking ??
      obj.message ??
      obj.description ??
      obj.output;
    if (candidate != null && typeof candidate !== "object") return String(candidate);
    if (typeof candidate === "object") return extractMessage(candidate);
    return JSON.stringify(raw);
  }
  return String(raw);
}

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
      message: string;
      progress: number;
      timestamp?: string;
    }) => void,
    onResult: (result: GenerateResponse) => void,
    onError: (error: Error) => void
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

        try {
          while (true) {
            const { done, value } = await reader.read();
            
            if (done) {
              console.log("🏁 SSE stream ended after", chunkCount, "chunks");
              break;
            }

            const chunk = decoder.decode(value, { stream: true });
            chunkCount++;
            if (chunkCount <= 10) console.log(`📦 SSE chunk #${chunkCount} (${chunk.length} bytes):`, chunk.substring(0, 200));

            buffer += chunk;
            const lines = buffer.split("\n\n");
            buffer = lines.pop() || "";

            for (const line of lines) {
              if (!line.trim()) continue;
              if (line.startsWith(":")) continue;

              const eventMatch = line.match(/^event: (.+)$/m);
              const dataMatch = line.match(/^data: (.+)$/m);

              if (eventMatch && dataMatch) {
                const event = eventMatch[1];
                let data: any;
                try { data = JSON.parse(dataMatch[1]); } catch (e) {
                  console.warn("⚠️ SSE JSON parse error:", e, "raw:", dataMatch[1].substring(0, 100));
                  continue;
                }

                switch (event) {
                  case "status":
                    onStatus({
                      ...data,
                      message: extractMessage(data?.message),
                      agent: String(data?.agent ?? ""),
                      status: String(data?.status ?? ""),
                      progress: Number(data?.progress ?? 0),
                    });
                    break;
                  case "result":
                    onResult(data);
                    break;
                  case "error":
                    onError(new Error(extractMessage(data?.message) || "Unknown error"));
                    break;
                  case "done":
                    console.log("✅ SSE done event received");
                    return;
                }
              }
            }
          }
        } catch (readerErr) {
          console.error("🚨 КРИТИЧЕСКАЯ ОШИБКА SSE (reader loop):", readerErr);
          onError(readerErr instanceof Error ? readerErr : new Error(String(readerErr)));
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
   * Генерация с SSE стримингом (для будущей реализации)
   */
  generateProjectStreamOld(
    request: GenerateRequest,
    onReasoningStep: (step: ReasoningStep) => void,
    onProgress: (message: string) => void,
    onComplete: (response: GenerateResponse) => void,
    onError: (error: Error) => void
  ): () => void {
    const eventSource = new EventSource(
      `${this.baseURL}/generate/stream?${new URLSearchParams({
        specification: request.specification || "",
        url: request.url || "",
      })}`
    );

    eventSource.addEventListener("reasoning", (event) => {
      try {
        const step: ReasoningStep = JSON.parse(event.data);
        onReasoningStep(step);
      } catch (e) {
        console.error("Failed to parse reasoning step:", e);
      }
    });

    eventSource.addEventListener("progress", (event) => {
      onProgress(event.data);
    });

    eventSource.addEventListener("complete", (event) => {
      try {
        const response: GenerateResponse = JSON.parse(event.data);
        onComplete(response);
        eventSource.close();
      } catch (e) {
        console.error("Failed to parse complete event:", e);
      }
    });

    eventSource.addEventListener("error", (event) => {
      onError(new Error("Stream error"));
      eventSource.close();
    });

    // Возвращаем функцию для отмены
    return () => eventSource.close();
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
