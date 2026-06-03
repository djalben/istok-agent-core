import { createFileRoute } from "@tanstack/react-router";
import { Sparkles, Plus, Brain, Wand2, Code2, FileSearch } from "lucide-react";
import { Button } from "@/components/ui/button";

export const Route = createFileRoute("/settings/skills")({
  component: SkillsPage,
  head: () => ({ meta: [{ title: "Навыки — Исток" }] }),
});

const skills = [
  { icon: Brain, name: "Анализ требований", desc: "Превращает грубое ТЗ в чёткий план.", enabled: true },
  { icon: Code2, name: "Рефакторинг кода", desc: "Безопасное переписывание модулей.", enabled: true },
  { icon: Wand2, name: "Полировка UI", desc: "Доводит интерфейс до Premium-уровня.", enabled: false },
  { icon: FileSearch, name: "Поиск по проекту", desc: "Глубокий семантический поиск по файлам.", enabled: true },
];

function SkillsPage() {
  return (
    <div className="mx-auto max-w-4xl px-6 py-8">
      <div className="relative mb-8 overflow-hidden rounded-2xl border border-border/60 p-8">
        <div className="absolute inset-0 bg-gradient-to-br from-fuchsia-600/30 via-violet-600/20 to-indigo-700/30" />
        <div className="absolute -right-20 -top-20 h-64 w-64 rounded-full bg-fuchsia-500/30 blur-3xl" />
        <div className="absolute -bottom-20 -left-10 h-64 w-64 rounded-full bg-indigo-500/30 blur-3xl" />
        <div className="relative">
          <div className="inline-flex items-center gap-2 rounded-full border border-white/15 bg-white/5 px-3 py-1 text-[11px] backdrop-blur">
            <Sparkles className="h-3 w-3" /> Навыки агента
          </div>
          <h1 className="mt-3 text-3xl font-semibold tracking-tight">Научите Исток, как вы работаете</h1>
          <p className="mt-1 max-w-xl text-sm text-muted-foreground">
            Подключайте навыки — небольшие модули, которые меняют поведение агента под ваши процессы и стек.
          </p>
        </div>
      </div>

      <div className="mb-4 flex items-center justify-between">
        <h2 className="text-sm font-semibold">Установленные навыки</h2>
        <Button variant="outline" size="sm"><Plus /> Добавить навык</Button>
      </div>

      <div className="grid gap-3 sm:grid-cols-2">
        {skills.map((s) => (
          <div key={s.name} className="flex items-start gap-3 rounded-xl border border-border/60 bg-card/40 p-4">
            <div className="grid h-9 w-9 place-items-center rounded-lg bg-primary/10 text-primary">
              <s.icon className="h-4 w-4" />
            </div>
            <div className="min-w-0 flex-1">
              <div className="flex items-center justify-between gap-2">
                <p className="truncate text-sm font-medium">{s.name}</p>
                <span className={`rounded-full px-2 py-0.5 text-[10px] ${s.enabled ? "bg-emerald-500/15 text-emerald-400" : "bg-muted/60 text-muted-foreground"}`}>
                  {s.enabled ? "Включён" : "Выключен"}
                </span>
              </div>
              <p className="mt-0.5 text-xs text-muted-foreground">{s.desc}</p>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
