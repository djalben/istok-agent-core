import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { BookOpen, Save } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { toast } from "sonner";

export const Route = createFileRoute("/settings/knowledge")({
  component: KnowledgePage,
  head: () => ({ meta: [{ title: "База знаний — Исток" }] }),
});

function KnowledgePage() {
  const [text, setText] = useState("");
  return (
    <div className="mx-auto max-w-3xl px-6 py-8">
      <div className="mb-6 flex items-center gap-3">
        <div className="grid h-10 w-10 place-items-center rounded-lg bg-primary/10 text-primary">
          <BookOpen className="h-5 w-5" />
        </div>
        <div>
          <h1 className="text-2xl font-semibold">База знаний</h1>
          <p className="mt-0.5 text-sm text-muted-foreground">
            Общие инструкции, которые Исток учитывает в каждом проекте рабочего пространства.
          </p>
        </div>
      </div>

      <div className="rounded-xl border border-border/60 bg-card/40 p-6">
        <label className="text-sm font-medium">Знания рабочего пространства</label>
        <p className="mt-1 text-xs text-muted-foreground">
          Например: бренд-гайд, тон-оф-войс, технические ограничения, любимые библиотеки.
        </p>
        <Textarea
          value={text}
          onChange={(e) => setText(e.target.value)}
          placeholder="Установите общие правила и настройки, которые Исток будет применять во всех ваших проектах..."
          className="mt-3 min-h-[260px] resize-y bg-background"
        />
        <div className="mt-4 flex items-center justify-between">
          <p className="text-[11px] text-muted-foreground">{text.length} / 8000 символов</p>
          <Button onClick={() => toast.success("База знаний сохранена")} className="bg-gradient-primary text-primary-foreground">
            <Save /> Сохранить
          </Button>
        </div>
      </div>
    </div>
  );
}
