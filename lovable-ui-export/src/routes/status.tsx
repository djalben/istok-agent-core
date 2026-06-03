import { createFileRoute } from "@tanstack/react-router";
import { Activity } from "lucide-react";
import { BackButton } from "@/components/features/BackButton";

export const Route = createFileRoute("/status")({
  component: StatusPage,
  head: () => ({ meta: [{ title: "Статус — Исток" }] }),
});

const services = [
  { name: "API", status: "Работает" },
  { name: "Среда сборки", status: "Работает" },
  { name: "AI Gateway", status: "Работает" },
  { name: "База данных", status: "Работает" },
  { name: "Авторизация", status: "Работает" },
];

function StatusPage() {
  return (
    <div className="min-h-screen bg-background">
      <div className="mx-auto max-w-3xl px-6 py-6">
        <BackButton to="/" />
        <div className="mt-6 flex items-center gap-3">
          <div className="grid h-12 w-12 place-items-center rounded-xl bg-emerald-500/15 text-emerald-400">
            <Activity className="h-5 w-5" />
          </div>
          <div>
            <h1 className="text-2xl font-semibold">Все системы работают</h1>
            <p className="text-sm text-muted-foreground">Обновлено только что</p>
          </div>
        </div>
        <div className="mt-8 divide-y divide-border/60 rounded-xl border border-border/60 bg-card/40">
          {services.map((s) => (
            <div key={s.name} className="flex items-center justify-between p-4">
              <p className="text-sm font-medium">{s.name}</p>
              <span className="inline-flex items-center gap-1.5 text-xs text-emerald-400">
                <span className="h-1.5 w-1.5 rounded-full bg-emerald-400" /> {s.status}
              </span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
