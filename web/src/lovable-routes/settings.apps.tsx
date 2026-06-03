import { createFileRoute } from "@tanstack/react-router";
import { Apple, Monitor, ExternalLink } from "lucide-react";
import { Button } from "@/components/ui/button";
import { BackButton } from "@/components/features/BackButton";

export const Route = createFileRoute("/settings/apps")({
  component: AppsSettings,
  head: () => ({ meta: [{ title: "Приложения — Исток" }] }),
});

function Section({ title, description, children }: { title: string; description?: string; children: React.ReactNode }) {
  return (
    <section className="rounded-xl border border-border/60 bg-card/40 p-6">
      <div className="mb-4">
        <h2 className="text-sm font-semibold">{title}</h2>
        {description && <p className="mt-0.5 text-xs text-muted-foreground">{description}</p>}
      </div>
      {children}
    </section>
  );
}

function AppsSettings() {
  return (
    <div className="mx-auto max-w-3xl px-4 py-8 sm:px-6">
      <div className="mb-8">
        <h1 className="text-2xl font-semibold">Устройства и приложения</h1>
        <p className="mt-1 text-sm text-muted-foreground">Подключите свои аккаунты и установите десктоп-приложение.</p>
      </div>

      <div className="space-y-5">

          <Section title="Чат-боты и мессенджеры">
            <div className="flex flex-col gap-3 rounded-lg border border-border/60 p-4 sm:flex-row sm:items-center sm:justify-between">
              <div className="flex items-center gap-3">
                <div className="grid h-10 w-10 place-items-center rounded-lg bg-[#229ED9]/15 text-[#229ED9]">
                  <svg viewBox="0 0 24 24" className="h-5 w-5" fill="currentColor"><path d="M9.04 15.36 8.9 19.4c.42 0 .6-.18.82-.4l1.97-1.88 4.08 2.99c.75.41 1.28.2 1.48-.69l2.68-12.57v-.01c.24-1.11-.4-1.55-1.13-1.28L3.18 10.2c-1.08.42-1.06 1.03-.18 1.3l4.05 1.26 9.4-5.93c.44-.29.85-.13.52.16"/></svg>
                </div>
                <div>
                  <p className="text-sm font-medium">Telegram</p>
                  <p className="text-xs text-muted-foreground">Получайте уведомления о завершённых сборках.</p>
                </div>
              </div>
              <Button variant="outline" size="sm">Подключить</Button>
            </div>
          </Section>

          <Section title="Десктоп-приложение" description="Исток для macOS и Windows — быстрее, нативные горячие клавиши.">
            <div className="grid gap-3 sm:grid-cols-2">
              <div className="flex items-center justify-between rounded-lg border border-border/60 p-4">
                <div className="flex items-center gap-3">
                  <div className="grid h-10 w-10 place-items-center rounded-lg bg-elevated">
                    <Apple className="h-5 w-5" />
                  </div>
                  <div>
                    <p className="text-sm font-medium">Исток для macOS</p>
                    <p className="text-xs text-muted-foreground">Universal · 84 МБ</p>
                  </div>
                </div>
                <Button size="sm" variant="outline">
                  Скачать <ExternalLink />
                </Button>
              </div>
              <div className="flex items-center justify-between rounded-lg border border-border/60 p-4">
                <div className="flex items-center gap-3">
                  <div className="grid h-10 w-10 place-items-center rounded-lg bg-elevated">
                    <Monitor className="h-5 w-5" />
                  </div>
                  <div>
                    <p className="text-sm font-medium">Исток для Windows</p>
                    <p className="text-xs text-muted-foreground">x64 · 92 МБ</p>
                  </div>
                </div>
                <Button size="sm" variant="outline">
                  Скачать <ExternalLink />
                </Button>
              </div>
            </div>
          </Section>
      </div>
    </div>
  );
}

