import { createFileRoute } from "@tanstack/react-router";
import { DashboardShell } from "@/components/features/DashboardShell";

export const Route = createFileRoute("/")({
  head: () => ({
    meta: [
      { title: "Исток — AI-сборщик сайтов и приложений" },
      { name: "description", content: "Исток управляет командой ИИ-агентов, которые исследуют, проектируют и собирают full-stack приложения из одного запроса." },
      { property: "og:title", content: "Исток — AI-сборщик сайтов и приложений" },
      { property: "og:description", content: "Создавайте full-stack приложения из одного запроса с помощью команды ИИ-агентов." },
    ],
  }),
  component: Index,
});

function Index() {
  return <DashboardShell />;
}
