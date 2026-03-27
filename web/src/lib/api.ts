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

const API_BASE = import.meta.env.VITE_API_BASE_URL || "http://localhost:8080/api/v1";

console.log("🔌 API Configuration:", {
  API_BASE,
  mode: import.meta.env.MODE,
});

// ── Types ───────────────────────────────────────────────

export interface GenerateRequest {
  specification?: string;
  url?: string;
  messages?: Array<{ role: string; content: string }>;
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

// ── API Client ──────────────────────────────────────────

class IstokAPI {
  private baseURL: string;

  constructor(baseURL: string) {
    this.baseURL = baseURL;
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
   * Генерация с SSE стримингом (для будущей реализации)
   */
  generateProjectStream(
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
    messages: Array<{ role: string; content: string }>
  ): Promise<GenerateResponse> {
    const formattedMessages = this.formatMessages(messages);
    
    // Берем последнее сообщение пользователя как specification
    const lastUserMessage = formattedMessages
      .filter((m) => m.role === "user")
      .pop();

    if (!lastUserMessage) {
      throw new Error("No user message found");
    }

    return this.generateProject({
      specification: lastUserMessage.content,
      messages: formattedMessages,
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
