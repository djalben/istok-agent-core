/**
 * ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 *  ИСТОК АГЕНТ — API Client Layer
 *  Единый модуль связи фронтенда с Go-бэкендом.
 *
 *  Режим работы:
 *    USE_MOCKS = true  → возвращает заглушки (dev)
 *    USE_MOCKS = false → реальные fetch к API_BASE
 *
 *  Windsurf / Lovable: импортируйте `api` (singleton)
 *  и вызывайте методы напрямую.
 * ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 */

// ── Config ──────────────────────────────────────────────

// Используем переменные окружения для гибкости деплоя
const API_BASE = import.meta.env.VITE_API_BASE_URL || "http://localhost:8080/api/v1";
// Временно включаем моки для локальной разработки, пока бэкенд не задеплоен
const USE_MOCKS = import.meta.env.VITE_USE_MOCKS === "true" || true; // ⚠️ Измените на false после деплоя бэкенда
const MOCK_LATENCY_MS = 350; // имитация сетевой задержки

// Логируем конфигурацию при инициализации
console.log("🔌 API Configuration:", {
  API_BASE,
  USE_MOCKS: USE_MOCKS ? "✅ MOCKS (dev mode)" : "🌐 REAL BACKEND",
  env: import.meta.env.MODE,
});

// ── Helpers ─────────────────────────────────────────────

function uid(): string {
  return crypto.randomUUID?.() ?? `${Date.now()}-${Math.random().toString(36).slice(2, 9)}`;
}

function delay(ms: number): Promise<void> {
  return new Promise((r) => setTimeout(r, ms));
}

// ── Error Types ─────────────────────────────────────────

export class ApiError extends Error {
  constructor(
    public readonly status: number,
    public readonly code: string,
    message: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

// ── Interfaces ──────────────────────────────────────────

/** Запрос на генерацию проекта */
export interface GenerateRequest {
  /** URL конкурента для анализа */
  url?: string;
  /** Текстовая спецификация проекта */
  specification?: string;
  /** Целевой фреймворк (по умолчанию next) */
  framework?: "next" | "nuxt" | "remix" | "astro";
  /** Язык интерфейса */
  locale?: string;
}

/** Ответ на запрос генерации */
export interface GenerateResponse {
  projectId: string;
  status: ProjectStatus;
  estimatedTimeMs: number;
}

/** Статусы жизненного цикла проекта */
export type ProjectStatus =
  | "queued"
  | "crawling"
  | "analyzing"
  | "generating"
  | "building"
  | "ready"
  | "error";

/** Статистика проекта — отдаётся бэкендом */
export interface ProjectStats {
  projectId: string;
  model: string;
  modelVersion: string;
  responseTimeMs: number;
  crawlerNodesFound: number;
  generatedFilesCount: number;
  tokensUsed: number;
  costRub: number;
  status: ProjectStatus;
  createdAt: string; // ISO 8601
  updatedAt: string;
}

/** Сообщение в чате агента */
export interface AgentMessage {
  id: string;
  projectId: string;
  role: "user" | "agent" | "system";
  content: string;
  timestamp: string; // ISO 8601
  status: "pending" | "streaming" | "complete" | "error";
  metadata?: Record<string, unknown>;
}

/** Отправка сообщения агенту */
export interface SendMessageRequest {
  content: string;
  /** Контекст для агента (выбранные файлы, URL и т.д.) */
  context?: Record<string, unknown>;
}

/** Баланс токенов пользователя */
export interface TokenBalance {
  currentRub: number;
  totalRub: number;
  tokensRemaining: number;
  plan: "free" | "pro" | "enterprise";
  resetsAt: string; // ISO 8601
}

/** Сводка агента (здоровье системы) */
export interface AgentHealthResponse {
  activeProjects: number;
  totalGenerated: number;
  uptimeSeconds: number;
  modelVersion: string;
  avgResponseTimeMs: number;
  queueDepth: number;
  gpuUtilization: number; // 0-1
}

/** Сгенерированный файл проекта */
export interface GeneratedFile {
  path: string;
  language: string;
  sizeBytes: number;
  preview: string; // первые ~200 символов
}

/** Полный проект */
export interface Project {
  id: string;
  name: string;
  stats: ProjectStats;
  messages: AgentMessage[];
  files: GeneratedFile[];
  previewUrl: string | null;
}

// ── Mock Data ───────────────────────────────────────────

const MOCK_PROJECT_ID = "proj_a1b2c3d4e5";

const MOCK_STATS: ProjectStats = {
  projectId: MOCK_PROJECT_ID,
  model: "Claude 4.6 Thinking",
  modelVersion: "4.6.0-rc.2",
  responseTimeMs: 142,
  crawlerNodesFound: 847,
  generatedFilesCount: 23,
  tokensUsed: 18_420,
  costRub: 2_340,
  status: "ready",
  createdAt: "2026-03-25T10:30:00Z",
  updatedAt: "2026-03-25T10:32:14Z",
};

const MOCK_BALANCE: TokenBalance = {
  currentRub: 65_000,
  totalRub: 100_000,
  tokensRemaining: 1_240_000,
  plan: "pro",
  resetsAt: "2026-04-01T00:00:00Z",
};

const MOCK_HEALTH: AgentHealthResponse = {
  activeProjects: 3,
  totalGenerated: 1_247,
  uptimeSeconds: 864_000,
  modelVersion: "4.6.0-rc.2",
  avgResponseTimeMs: 142,
  queueDepth: 0,
  gpuUtilization: 0.34,
};

const MOCK_MESSAGES: AgentMessage[] = [
  {
    id: uid(),
    projectId: MOCK_PROJECT_ID,
    role: "system",
    content: "Сессия инициализирована. Модель: Claude 4.6 Thinking. Готов к анализу.",
    timestamp: new Date(Date.now() - 60_000).toISOString(),
    status: "complete",
  },
  {
    id: uid(),
    projectId: MOCK_PROJECT_ID,
    role: "user",
    content: "Проанализируй сайт конкурента и создай улучшенную версию с современным дизайном",
    timestamp: new Date(Date.now() - 45_000).toISOString(),
    status: "complete",
  },
  {
    id: uid(),
    projectId: MOCK_PROJECT_ID,
    role: "agent",
    content: "Запускаю глубокий анализ... Обнаружено 847 узлов. Извлекаю структуру, палитру и UX-паттерны. Генерирую оптимизированную архитектуру проекта.",
    timestamp: new Date(Date.now() - 30_000).toISOString(),
    status: "complete",
  },
  {
    id: uid(),
    projectId: MOCK_PROJECT_ID,
    role: "agent",
    content: "Сгенерировано 23 файла. Проект готов к предпросмотру. Используется Next.js 15, Tailwind CSS v4, и оптимизированная структура компонентов.",
    timestamp: new Date(Date.now() - 15_000).toISOString(),
    status: "complete",
  },
];

const MOCK_FILES: GeneratedFile[] = [
  { path: "app/layout.tsx", language: "tsx", sizeBytes: 1_240, preview: "export default function RootLayout({ children }…" },
  { path: "app/page.tsx", language: "tsx", sizeBytes: 3_870, preview: "import { Hero } from '@/components/hero'…" },
  { path: "components/hero.tsx", language: "tsx", sizeBytes: 2_110, preview: "export function Hero() { return <section…" },
  { path: "tailwind.config.ts", language: "ts", sizeBytes: 890, preview: "import type { Config } from 'tailwindcss'…" },
  { path: "lib/utils.ts", language: "ts", sizeBytes: 340, preview: "import { clsx } from 'clsx'…" },
];

// ── HTTP Transport ──────────────────────────────────────

async function request<T>(
  method: "GET" | "POST" | "PUT" | "DELETE",
  path: string,
  body?: unknown,
): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    method,
    headers: {
      "Content-Type": "application/json",
      // "Authorization": `Bearer ${getToken()}`, // ← раскомментировать при авторизации
    },
    body: body ? JSON.stringify(body) : undefined,
  });

  if (!res.ok) {
    const err = await res.json().catch(() => ({ code: "UNKNOWN", message: res.statusText }));
    throw new ApiError(res.status, err.code ?? "UNKNOWN", err.message ?? "Ошибка запроса");
  }

  return res.json();
}

// ── API Client Class ────────────────────────────────────

class IstokApiClient {
  // ─── Генерация проекта ───────────────────────────────

  /** Запустить генерацию проекта */
  async generateProject(req: GenerateRequest): Promise<GenerateResponse> {
    if (USE_MOCKS) {
      await delay(MOCK_LATENCY_MS);
      return {
        projectId: MOCK_PROJECT_ID,
        status: "crawling",
        estimatedTimeMs: 14_000,
      };
    }
    return request<GenerateResponse>("POST", "/projects/generate", req);
  }

  /** Получить статус / статистику проекта */
  async getProjectStats(projectId: string): Promise<ProjectStats> {
    if (USE_MOCKS) {
      await delay(MOCK_LATENCY_MS);
      return { ...MOCK_STATS, projectId };
    }
    return request<ProjectStats>("GET", `/projects/${projectId}/stats`);
  }

  /** Получить полный проект (статистика + файлы + сообщения) */
  async getProject(projectId: string): Promise<Project> {
    if (USE_MOCKS) {
      await delay(MOCK_LATENCY_MS);
      return {
        id: projectId,
        name: "Конкурентный Анализ #1",
        stats: { ...MOCK_STATS, projectId },
        messages: MOCK_MESSAGES,
        files: MOCK_FILES,
        previewUrl: "http://localhost:3000",
      };
    }
    return request<Project>("GET", `/projects/${projectId}`);
  }

  // ─── Агент / Чат ────────────────────────────────────

  /** Получить историю сообщений проекта */
  async getMessages(projectId: string): Promise<AgentMessage[]> {
    if (USE_MOCKS) {
      await delay(MOCK_LATENCY_MS);
      return MOCK_MESSAGES.map((m) => ({ ...m, projectId }));
    }
    return request<AgentMessage[]>("GET", `/projects/${projectId}/messages`);
  }

  /** Отправить сообщение агенту */
  async sendMessage(projectId: string, req: SendMessageRequest): Promise<AgentMessage> {
    if (USE_MOCKS) {
      await delay(MOCK_LATENCY_MS * 3);
      return {
        id: uid(),
        projectId,
        role: "agent",
        content: "Принято. Запускаю анализ и генерацию. Ожидаемое время: ~15 секунд.",
        timestamp: new Date().toISOString(),
        status: "complete",
      };
    }
    return request<AgentMessage>("POST", `/projects/${projectId}/messages`, req);
  }

  // ─── Биллинг ────────────────────────────────────────

  /** Текущий баланс токенов */
  async getBalance(): Promise<TokenBalance> {
    if (USE_MOCKS) {
      await delay(MOCK_LATENCY_MS);
      return MOCK_BALANCE;
    }
    return request<TokenBalance>("GET", "/billing/balance");
  }

  // ─── Системные ──────────────────────────────────────

  /** Здоровье агента и общая статистика */
  async getHealth(): Promise<AgentHealthResponse> {
    if (USE_MOCKS) {
      await delay(MOCK_LATENCY_MS);
      return MOCK_HEALTH;
    }
    return request<AgentHealthResponse>("GET", "/agent/health");
  }

  /** Список сгенерированных файлов проекта */
  async getFiles(projectId: string): Promise<GeneratedFile[]> {
    if (USE_MOCKS) {
      await delay(MOCK_LATENCY_MS);
      return MOCK_FILES;
    }
    return request<GeneratedFile[]>("GET", `/projects/${projectId}/files`);
  }
}

// ── Singleton ───────────────────────────────────────────

export const api = new IstokApiClient();

// ── Re-exports для обратной совместимости ───────────────
// Компоненты, которые уже используют старые моки,
// могут импортировать их отсюда.

export { MOCK_STATS, MOCK_BALANCE, MOCK_MESSAGES, MOCK_FILES, MOCK_HEALTH };
