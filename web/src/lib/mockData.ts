export type AgentStatus = "idle" | "thinking" | "working" | "done" | "error";

export interface Agent {
  id: string;
  name: string;
  role: string;
  status: AgentStatus;
  task: string;
}

export interface ChatMessage {
  id: string;
  role: "user" | "assistant" | "system";
  content: string;
  timestamp: string;
}

export interface FileNode {
  name: string;
  path: string;
  type: "file" | "folder";
  language?: string;
  content?: string;
  children?: FileNode[];
}

export interface Project {
  id: string;
  name: string;
  description: string;
  updatedAt: string;
  gradient: string;
  framework: string;
}

export const mockProjects: Project[] = [
  {
    id: "1",
    name: "Лендинг TaxiGo",
    description: "Маркетинговый сайт сервиса такси со сценарием бронирования",
    updatedAt: "2 часа назад",
    framework: "Next.js",
    gradient: "bg-gradient-to-br from-violet-500 to-fuchsia-600",
  },
  {
    id: "2",
    name: "Аналитика Lumen",
    description: "Дашборд KPI в реальном времени с графиками и фильтрами",
    updatedAt: "Вчера",
    framework: "React",
    gradient: "bg-gradient-to-br from-cyan-400 to-blue-600",
  },
  {
    id: "3",
    name: "Helix CRM",
    description: "Воронка клиентов с канбаном и AI-инсайтами",
    updatedAt: "3 дня назад",
    framework: "Remix",
    gradient: "bg-gradient-to-br from-amber-400 to-orange-600",
  },
  {
    id: "4",
    name: "NoteOrbit",
    description: "Markdown-заметки с семантическим поиском",
    updatedAt: "На прошлой неделе",
    framework: "Astro",
    gradient: "bg-gradient-to-br from-emerald-400 to-teal-600",
  },
  {
    id: "5",
    name: "Pulse Newsletter",
    description: "Платформа публикаций с платными подписками",
    updatedAt: "На прошлой неделе",
    framework: "Next.js",
    gradient: "bg-gradient-to-br from-rose-400 to-red-600",
  },
  {
    id: "6",
    name: "Forge DevTools",
    description: "Маркетинговый сайт для open-source CLI",
    updatedAt: "2 недели назад",
    framework: "Vite",
    gradient: "bg-gradient-to-br from-indigo-400 to-violet-600",
  },
];

export const initialAgents: Agent[] = [
  { id: "director", name: "Директор", role: "Управляет сборкой", status: "done", task: "План утверждён" },
  { id: "researcher", name: "Исследователь", role: "Собирает контекст и источники", status: "done", task: "Проиндексировано 12 источников" },
  { id: "architect", name: "Архитектор", role: "Проектирует структуру файлов", status: "working", task: "Создаю каркас маршрутов…" },
  { id: "coder", name: "Кодер", role: "Пишет компоненты и логику", status: "thinking", task: "В очереди: компонент Hero" },
  { id: "designer", name: "Дизайнер", role: "Применяет дизайн-токены", status: "idle", task: "Жду готовый макет" },
];

export const initialChat: ChatMessage[] = [
  {
    id: "1",
    role: "user",
    content: "Собери мне лендинг для сервиса такси TaxiGo. Нужны hero-блок, фичи, тарифы и виджет бронирования.",
    timestamp: "10:42",
  },
  {
    id: "2",
    role: "assistant",
    content: "Готовлю план. Подниму проект на Next.js с hero-блоком, сеткой фич, прозрачными тарифами и интерактивным виджетом бронирования. Подготовлю Markdown-бриф для вашего подтверждения перед написанием кода.",
    timestamp: "10:42",
  },
  {
    id: "3",
    role: "system",
    content: "Директор передал задачу Архитектору → Кодеру. Ожидается подтверждение бизнес-плана.",
    timestamp: "10:43",
  },
];

export const mockFiles: FileNode[] = [
  {
    name: "src",
    path: "src",
    type: "folder",
    children: [
      {
        name: "app",
        path: "src/app",
        type: "folder",
        children: [
          {
            name: "page.tsx",
            path: "src/app/page.tsx",
            type: "file",
            language: "tsx",
            content: `import { Hero } from "@/components/Hero";
import { Features } from "@/components/Features";
import { Pricing } from "@/components/Pricing";
import { BookingWidget } from "@/components/BookingWidget";

export default function Home() {
  return (
    <main className="min-h-screen bg-background">
      <Hero />
      <BookingWidget />
      <Features />
      <Pricing />
    </main>
  );
}
`,
          },
          {
            name: "layout.tsx",
            path: "src/app/layout.tsx",
            type: "file",
            language: "tsx",
            content: `export const metadata = {
  title: "TaxiGo — Поездки по запросу",
  description: "Закажите поездку за секунды. Прозрачные цены и водители, которым можно доверять.",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="ru">
      <body>{children}</body>
    </html>
  );
}
`,
          },
        ],
      },
      {
        name: "components",
        path: "src/components",
        type: "folder",
        children: [
          {
            name: "Hero.tsx",
            path: "src/components/Hero.tsx",
            type: "file",
            language: "tsx",
            content: `export function Hero() {
  return (
    <section className="relative isolate overflow-hidden py-32">
      <div className="mx-auto max-w-5xl px-6 text-center">
        <h1 className="text-6xl font-bold tracking-tight">
          Поездки по запросу.<br />
          <span className="text-gradient">Прозрачная цена.</span>
        </h1>
        <p className="mt-6 text-xl text-muted-foreground">
          Нажмите кнопку и поехали. Никаких сюрпризов с тарифами.
        </p>
      </div>
    </section>
  );
}
`,
          },
          {
            name: "BookingWidget.tsx",
            path: "src/components/BookingWidget.tsx",
            type: "file",
            language: "tsx",
            content: `"use client";
import { useState } from "react";

export function BookingWidget() {
  const [pickup, setPickup] = useState("");
  const [dropoff, setDropoff] = useState("");
  return (
    <div className="mx-auto -mt-12 max-w-2xl rounded-2xl border bg-card p-6 shadow-xl">
      <h3 className="text-lg font-semibold">Куда едем?</h3>
      {/* форма бронирования */}
    </div>
  );
}
`,
          },
        ],
      },
      {
        name: "styles.css",
        path: "src/styles.css",
        type: "file",
        language: "css",
        content: `:root {
  --background: #0a0a0f;
  --foreground: #fafafa;
}
`,
      },
    ],
  },
  {
    name: "package.json",
    path: "package.json",
    type: "file",
    language: "json",
    content: `{
  "name": "taxigo-landing",
  "version": "0.1.0",
  "private": true,
  "scripts": {
    "dev": "next dev",
    "build": "next build",
    "start": "next start"
  }
}
`,
  },
  {
    name: "README.md",
    path: "README.md",
    type: "file",
    language: "md",
    content: `# Лендинг TaxiGo\n\nСгенерировано в Истоке.\n`,
  },
];

export const businessPlanMarkdown = `# TaxiGo — Бриф бизнес-плана

## Видение
Сервис такси с **прозрачными ценами без накруток**, который одинаково ценит и водителей, и пассажиров.

## Целевая аудитория
- Городские пассажиры (25–45 лет), уставшие от динамических тарифов
- Водители с частичной занятостью, которым важен предсказуемый доход

## Ключевые страницы
1. **Лендинг** — hero-блок, ценностное предложение, виджет бронирования
2. **Тарифы** — фиксированные ставки с выбором города
3. **Водителям** — калькулятор дохода и регистрация
4. **О сервисе** — команда и обещание безопасности

## Технический план
- **Фреймворк:** Next.js 15 (App Router)
- **Стили:** Tailwind v4 с кастомными дизайн-токенами
- **Анимация:** Framer Motion для hero и появления секций
- **Формы:** React Hook Form + Zod-валидация

## Метрики успеха
- Время до первого бронирования меньше 90 секунд
- LCP < 1.5с на 4G
- Конверсия в регистрацию водителей > 8%

> **Подтвердите, чтобы начать генерацию кода.** Агент-кодер готов к запуску.
`;
