import { createFileRoute } from "@tanstack/react-router";
import { LayoutGrid, Sparkles } from "lucide-react";
import { Button } from "@/components/ui/button";
import { TierBadge } from "@/components/features/SettingsPage";

export const Route = createFileRoute("/settings/templates")({
  component: TemplatesPage,
  head: () => ({ meta: [{ title: "Шаблоны — Исток" }] }),
});

function TemplatesPage() {
  return (
    <div className="mx-auto max-w-4xl px-6 py-8">
      <div className="mb-8 flex items-center gap-3">
        <h1 className="text-2xl font-semibold">Шаблоны</h1>
        <TierBadge tier="Business" />
      </div>

      <div className="grid place-items-center rounded-2xl border border-dashed border-border/60 bg-card/20 px-6 py-20">
        <div className="grid h-14 w-14 place-items-center rounded-2xl bg-violet-500/10 text-violet-300">
          <LayoutGrid className="h-6 w-6" />
        </div>
        <h2 className="mt-4 text-base font-medium">Используйте проекты повторно как шаблоны рабочей области</h2>
        <p className="mt-1 max-w-md text-center text-sm text-muted-foreground">
          Сохраняйте проекты как шаблоны и запускайте новые с готовыми настройками, кодом и подключениями.
        </p>
        <Button className="mt-5 bg-gradient-to-r from-violet-500 to-fuchsia-500 text-white hover:opacity-90">
          <Sparkles /> Обновить до Business
        </Button>
      </div>
    </div>
  );
}
