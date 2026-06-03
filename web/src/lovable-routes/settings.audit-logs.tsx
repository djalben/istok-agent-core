import { createFileRoute } from "@tanstack/react-router";
import { Clock, PhoneCall } from "lucide-react";
import { Button } from "@/components/ui/button";
import { TierBadge } from "@/components/features/SettingsPage";

export const Route = createFileRoute("/settings/audit-logs")({
  component: AuditLogsPage,
  head: () => ({ meta: [{ title: "Журналы аудита — Исток" }] }),
});

function AuditLogsPage() {
  return (
    <div className="mx-auto max-w-4xl px-6 py-8">
      <div className="mb-8 flex items-center gap-3">
        <h1 className="text-2xl font-semibold">Журналы аудита</h1>
        <TierBadge tier="Enterprise" />
      </div>

      <div className="grid place-items-center rounded-2xl border border-dashed border-border/60 bg-card/20 px-6 py-20">
        <div className="grid h-14 w-14 place-items-center rounded-2xl bg-amber-500/10 text-amber-300">
          <Clock className="h-6 w-6" />
        </div>
        <h2 className="mt-4 text-base font-medium">Отслеживайте каждое изменение в рабочем пространстве</h2>
        <p className="mt-1 max-w-md text-center text-sm text-muted-foreground">
          Полная история действий участников: входы, изменения ролей, публикации, удаления проектов и API-вызовы.
        </p>
        <Button className="mt-5 bg-gradient-to-r from-amber-400 to-rose-500 text-white hover:opacity-90">
          <PhoneCall /> Поговорите с отделом продаж
        </Button>
      </div>
    </div>
  );
}
