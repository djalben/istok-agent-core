import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { Upload, Trash2, LogOut } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { BackButton } from "@/components/features/BackButton";
import { toast } from "sonner";

export const Route = createFileRoute("/settings/workspace")({
  component: WorkspaceSettings,
  head: () => ({ meta: [{ title: "Настройки рабочего пространства — Исток" }] }),
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

function WorkspaceSettings() {
  const [name, setName] = useState("Рабочее пространство Александра");
  const [handle, setHandle] = useState("alexandr");
  const [limit, setLimit] = useState("100");

  return (
    <div className="mx-auto max-w-3xl px-4 py-8 sm:px-6">
      <div className="mb-8">
        <h1 className="text-2xl font-semibold">Настройки рабочего пространства</h1>
        <p className="mt-1 text-sm text-muted-foreground">Управляйте брендом, лимитами и доступом команды.</p>
      </div>

      <div className="space-y-5">

          <Section title="Логотип" description="Загрузите квадратное изображение PNG или JPG, не менее 256×256.">
            <div className="flex items-center gap-4">
              <div className="grid h-16 w-16 place-items-center rounded-xl bg-gradient-primary text-lg font-semibold text-primary-foreground">
                A
              </div>
              <Button variant="outline" size="sm" onClick={() => toast("Выбор файла…")}>
                <Upload /> Загрузить логотип
              </Button>
            </div>
          </Section>

          <Section title="Основное">
            <div className="grid gap-4 sm:grid-cols-2">
              <div>
                <Label htmlFor="ws-name">Название рабочего пространства</Label>
                <Input id="ws-name" value={name} onChange={(e) => setName(e.target.value)} className="mt-1.5" />
              </div>
              <div>
                <Label htmlFor="ws-handle">Идентификатор (handle)</Label>
                <div className="mt-1.5 flex items-center rounded-md border border-input bg-background focus-within:ring-2 focus-within:ring-ring">
                  <span className="px-3 text-sm text-muted-foreground">istok.app/</span>
                  <Input
                    id="ws-handle"
                    value={handle}
                    onChange={(e) => setHandle(e.target.value)}
                    className="border-0 bg-transparent focus-visible:ring-0"
                  />
                </div>
              </div>
            </div>
          </Section>

          <Section title="Лимит кредитов по умолчанию" description="Сколько кредитов доступно каждому участнику в месяц.">
            <div className="max-w-[200px]">
              <Input
                type="number"
                value={limit}
                onChange={(e) => setLimit(e.target.value)}
                min={0}
              />
            </div>
          </Section>

          <section className="rounded-xl border border-destructive/30 bg-destructive/5 p-6">
            <div className="mb-4">
              <h2 className="text-sm font-semibold text-destructive">Опасная зона</h2>
              <p className="mt-0.5 text-xs text-muted-foreground">Эти действия необратимы.</p>
            </div>
            <div className="space-y-3">
              <div className="flex flex-col gap-3 rounded-lg border border-destructive/30 bg-background/40 p-3 sm:flex-row sm:items-center sm:justify-between">
                <div>
                  <p className="text-sm font-medium">Покинуть рабочее пространство</p>
                  <p className="text-xs text-muted-foreground">Вы потеряете доступ ко всем проектам внутри.</p>
                </div>
                <Button variant="outline" size="sm" className="border-destructive/40 text-destructive hover:bg-destructive/10">
                  <LogOut /> Покинуть
                </Button>
              </div>
              <div className="flex flex-col gap-3 rounded-lg border border-destructive/30 bg-background/40 p-3 sm:flex-row sm:items-center sm:justify-between">
                <div>
                  <p className="text-sm font-medium">Удалить рабочее пространство</p>
                  <p className="text-xs text-muted-foreground">Полностью удалить рабочее пространство и его проекты.</p>
                </div>
                <Button variant="destructive" size="sm"><Trash2 /> Удалить</Button>
              </div>
            </div>
          </section>
      </div>
    </div>
  );
}

