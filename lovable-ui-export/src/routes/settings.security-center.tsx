import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { ShieldCheck, Sparkles } from "lucide-react";
import { Button } from "@/components/ui/button";
import { TierBadge } from "@/components/features/SettingsPage";
import { cn } from "@/lib/utils";

export const Route = createFileRoute("/settings/security-center")({
  component: SecurityCenterPage,
  head: () => ({ meta: [{ title: "Центр безопасности — Исток" }] }),
});

const tabs = [
  "Анализ кода",
  "Безопасность цепочки поставок",
  "Секреты",
  "Уязвимости",
  "Политики",
];

function SecurityCenterPage() {
  const [active, setActive] = useState(tabs[0]);
  return (
    <div className="mx-auto max-w-5xl px-6 py-8">
      <div className="mb-6 flex items-center gap-3">
        <h1 className="text-2xl font-semibold">Центр безопасности</h1>
        <TierBadge tier="Business" />
      </div>

      <div className="mb-8 flex flex-wrap gap-1 border-b border-border/60">
        {tabs.map((t) => (
          <button
            key={t}
            onClick={() => setActive(t)}
            className={cn(
              "relative px-3 py-2 text-xs transition-colors",
              active === t
                ? "text-foreground after:absolute after:inset-x-3 after:bottom-0 after:h-0.5 after:bg-primary"
                : "text-muted-foreground hover:text-foreground",
            )}
          >
            {t}
          </button>
        ))}
      </div>

      <div className="grid place-items-center rounded-2xl border border-dashed border-border/60 bg-card/20 px-6 py-20">
        <div className="grid h-14 w-14 place-items-center rounded-2xl bg-violet-500/10 text-violet-300">
          <ShieldCheck className="h-6 w-6" />
        </div>
        <h2 className="mt-4 text-base font-medium">{active} — доступно на тарифе Business</h2>
        <p className="mt-1 max-w-md text-center text-sm text-muted-foreground">
          Получите автоматические проверки безопасности, ревью pull-request'ов и контроль политик.
        </p>
        <Button className="mt-5 bg-gradient-to-r from-violet-500 to-fuchsia-500 text-white hover:opacity-90">
          <Sparkles /> Переход на бизнес-версию
        </Button>
      </div>
    </div>
  );
}
