import { createFileRoute } from "@tanstack/react-router";
import { BookOpen, Sparkles, Code2, Database } from "lucide-react";
import { BackButton } from "@/components/features/BackButton";

export const Route = createFileRoute("/docs")({
  component: DocsPage,
  head: () => ({ meta: [{ title: "Документация — Исток" }] }),
});

function DocsPage() {
  const sections = [
    { icon: Sparkles, title: "Быстрый старт", desc: "Запустите первый проект за пару минут." },
    { icon: Code2, title: "API сборщика", desc: "Программный доступ к среде Истока." },
    { icon: Database, title: "Cloud и база данных", desc: "Схемы, RLS, edge-функции." },
    { icon: BookOpen, title: "Гайды", desc: "Лучшие практики разработки с ИИ." },
  ];
  return (
    <div className="min-h-screen bg-background">
      <div className="mx-auto max-w-4xl px-6 py-6">
        <BackButton to="/" />
        <div className="mt-6">
          <h1 className="text-3xl font-semibold">Документация</h1>
          <p className="mt-1 text-sm text-muted-foreground">Всё, что нужно для работы с Истоком.</p>
        </div>
        <div className="mt-8 grid gap-4 sm:grid-cols-2">
          {sections.map((s) => (
            <div key={s.title} className="group rounded-xl border border-border/60 bg-card/40 p-5 transition hover:border-primary/40">
              <s.icon className="h-5 w-5 text-primary" />
              <p className="mt-3 text-sm font-medium">{s.title}</p>
              <p className="mt-1 text-xs text-muted-foreground">{s.desc}</p>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
