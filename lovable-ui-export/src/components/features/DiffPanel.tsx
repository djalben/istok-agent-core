import { GitCommit, Minus, Plus } from "lucide-react";
import { ScrollArea } from "@/components/ui/scroll-area";

const oldCode = `export function Hero() {
  return (
    <section className="py-16">
      <h1 className="text-3xl font-bold">
        Добро пожаловать
      </h1>
      <p className="text-muted-foreground">
        Старое описание продукта.
      </p>
    </section>
  );
}`;

const newCode = `export function Hero() {
  return (
    <section className="relative py-24">
      <h1 className="text-5xl font-semibold tracking-tight">
        Создавайте быстрее с Истоком
      </h1>
      <p className="mt-3 max-w-xl text-muted-foreground">
        Команда ИИ-агентов проектирует и собирает приложения за минуты.
      </p>
      <Button size="lg" className="mt-6">Начать</Button>
    </section>
  );
}`;

function Side({
  title,
  code,
  variant,
}: {
  title: string;
  code: string;
  variant: "removed" | "added";
}) {
  const lines = code.split("\n");
  const isAdd = variant === "added";
  return (
    <div className="flex min-w-0 flex-1 flex-col border-border/60">
      <div className="flex h-9 items-center justify-between border-b border-border/60 bg-panel/40 px-3 text-xs">
        <span className="font-medium text-muted-foreground">{title}</span>
        <span
          className={
            isAdd
              ? "rounded-md bg-emerald-500/10 px-1.5 py-0.5 font-mono text-[10px] text-emerald-400"
              : "rounded-md bg-rose-500/10 px-1.5 py-0.5 font-mono text-[10px] text-rose-400"
          }
        >
          {isAdd ? "+ добавлено" : "− удалено"}
        </span>
      </div>
      <ScrollArea className="flex-1">
        <pre className="min-h-full font-mono text-[12.5px] leading-6">
          {lines.map((line, i) => (
            <div
              key={i}
              className={
                "flex items-start gap-2 px-3 py-px " +
                (isAdd ? "bg-emerald-500/5" : "bg-rose-500/5")
              }
            >
              <span className="w-6 select-none text-right text-muted-foreground/50">
                {i + 1}
              </span>
              {isAdd ? (
                <Plus className="mt-1 h-3 w-3 shrink-0 text-emerald-400/80" />
              ) : (
                <Minus className="mt-1 h-3 w-3 shrink-0 text-rose-400/80" />
              )}
              <code
                className={
                  "whitespace-pre " +
                  (isAdd ? "text-emerald-100/90" : "text-rose-100/80 line-through decoration-rose-400/40")
                }
              >
                {line || " "}
              </code>
            </div>
          ))}
        </pre>
      </ScrollArea>
    </div>
  );
}

export function DiffPanel() {
  return (
    <div className="flex h-full flex-col bg-background">
      <div className="flex h-10 items-center justify-between border-b border-border/60 bg-panel px-3">
        <div className="flex items-center gap-2">
          <GitCommit className="h-3.5 w-3.5 text-primary" />
          <span className="text-xs font-medium text-muted-foreground">
            Сравнение · src/components/Hero.tsx
          </span>
        </div>
        <span className="font-mono text-[10px] text-muted-foreground">
          +6 −4 строки
        </span>
      </div>
      <div className="flex min-h-0 flex-1 divide-x divide-border/60">
        <Side title="До · main" code={oldCode} variant="removed" />
        <Side title="После · current" code={newCode} variant="added" />
      </div>
    </div>
  );
}
