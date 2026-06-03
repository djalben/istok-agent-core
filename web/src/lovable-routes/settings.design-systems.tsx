import { createFileRoute } from "@tanstack/react-router";
import { LayoutPanelLeft, Sparkles } from "lucide-react";
import { Button } from "@/components/ui/button";
import { TierBadge } from "@/components/features/SettingsPage";

export const Route = createFileRoute("/settings/design-systems")({
  component: DesignSystemsPage,
  head: () => ({ meta: [{ title: "Системы проектирования — Исток" }] }),
});

function DesignSystemsPage() {
  return (
    <div className="mx-auto max-w-4xl px-6 py-8">
      <div className="mb-8 flex items-center gap-3">
        <h1 className="text-2xl font-semibold">Системы проектирования</h1>
        <TierBadge tier="Enterprise" />
      </div>

      <div className="grid place-items-center rounded-2xl border border-dashed border-border/60 bg-card/20 px-6 py-20">
        <div className="grid h-14 w-14 place-items-center rounded-2xl bg-amber-500/10 text-amber-300">
          <LayoutPanelLeft className="h-6 w-6" />
        </div>
        <h2 className="mt-4 text-base font-medium">Стандартизируйте систему проектирования</h2>
        <p className="mt-1 max-w-md text-center text-sm text-muted-foreground">
          Загрузите токены, шрифты и компоненты, чтобы Исток применял ваш бренд во всех новых проектах.
        </p>
        <Button className="mt-5 bg-gradient-to-r from-amber-400 to-rose-500 text-white hover:opacity-90">
          <Sparkles /> Обновить до Enterprise
        </Button>
      </div>
    </div>
  );
}
