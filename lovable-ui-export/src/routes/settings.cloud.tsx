import { createFileRoute } from "@tanstack/react-router";
import { Cloud, Sparkles, Plus, TrendingUp } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Progress } from "@/components/ui/progress";

export const Route = createFileRoute("/settings/cloud")({
  component: CloudBalancePage,
  head: () => ({ meta: [{ title: "Баланс Облака и ИИ — Исток" }] }),
});

function BalanceCard({
  icon: Icon, title, used, total, suffix, accent,
}: { icon: React.ComponentType<{ className?: string }>; title: string; used: number; total: number; suffix: string; accent: string }) {
  const pct = Math.min(100, (used / total) * 100);
  return (
    <div className="rounded-xl border border-border/60 bg-card/40 p-6">
      <div className="mb-4 flex items-center gap-3">
        <span className={`grid h-9 w-9 place-items-center rounded-lg bg-gradient-to-br text-white ${accent}`}>
          <Icon className="h-4 w-4" />
        </span>
        <h2 className="text-sm font-semibold">{title}</h2>
      </div>
      <div className="flex items-baseline gap-2">
        <span className="text-3xl font-semibold tracking-tight">${used.toFixed(2)}</span>
        <span className="text-xs text-muted-foreground">из ${total.toFixed(2)} {suffix}</span>
      </div>
      <Progress value={pct} className="mt-3 h-2" />
      <p className="mt-2 text-xs text-muted-foreground">Сбрасывается в начале следующего месяца.</p>
    </div>
  );
}

function CloudBalancePage() {
  return (
    <div className="mx-auto max-w-4xl px-6 py-8">
      <div className="mb-8 flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold">Баланс Облака и ИИ</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Использование по тарифу с оплатой по факту. Каждое рабочее пространство получает бесплатные кредиты ежемесячно.
          </p>
        </div>
        <Button className="bg-gradient-primary text-primary-foreground"><Plus /> Пополнить счёт</Button>
      </div>

      <div className="grid gap-4 sm:grid-cols-2">
        <BalanceCard icon={Cloud} title="Облако" used={3.42} total={25} suffix="бесплатно в месяц" accent="from-cyan-500 to-blue-600" />
        <BalanceCard icon={Sparkles} title="ИИ Шлюз" used={0.18} total={1} suffix="бесплатно в месяц" accent="from-fuchsia-500 to-violet-600" />
      </div>

      <div className="mt-5 rounded-xl border border-border/60 bg-card/40 p-6">
        <div className="mb-4 flex items-center gap-2">
          <TrendingUp className="h-4 w-4 text-primary" />
          <h2 className="text-sm font-semibold">Активность за 30 дней</h2>
        </div>
        <div className="flex h-32 items-end gap-1">
          {Array.from({ length: 30 }).map((_, i) => {
            const h = 18 + ((i * 13) % 70);
            return <div key={i} className="flex-1 rounded-t bg-gradient-to-t from-primary/30 to-primary/80" style={{ height: `${h}%` }} />;
          })}
        </div>
        <div className="mt-2 flex justify-between text-[10px] text-muted-foreground">
          <span>30 дней назад</span><span>Сегодня</span>
        </div>
      </div>
    </div>
  );
}
