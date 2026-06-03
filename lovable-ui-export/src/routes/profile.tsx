import { createFileRoute, Link } from "@tanstack/react-router";
import { useMemo } from "react";
import { motion } from "framer-motion";
import { FolderOpen, Settings, Pencil, MapPin, Calendar, Link as LinkIcon } from "lucide-react";
import { Button } from "@/components/ui/button";
import { BackButton } from "@/components/features/BackButton";

export const Route = createFileRoute("/profile")({
  component: ProfilePage,
  head: () => ({ meta: [{ title: "Профиль — Исток" }] }),
});

function ProfilePage() {
  const cells = useMemo(() => {
    const arr: number[] = [];
    let seed = 7;
    for (let i = 0; i < 53 * 7; i++) {
      seed = (seed * 9301 + 49297) % 233280;
      const r = seed / 233280;
      arr.push(r > 0.94 ? 3 : r > 0.88 ? 2 : r > 0.82 ? 1 : 0);
    }
    return arr;
  }, []);

  const totalEdits = cells.reduce((s, v) => s + v, 0);
  const daysActive = cells.filter((v) => v > 0).length;

  const level = (v: number) =>
    v === 0
      ? "bg-muted/30"
      : v === 1
        ? "bg-primary/30"
        : v === 2
          ? "bg-primary/60"
          : "bg-primary";

  const months = ["Янв", "Фев", "Мар", "Апр", "Май", "Июн", "Июл", "Авг", "Сен", "Окт", "Ноя", "Дек"];

  return (
    <div className="min-h-screen bg-background">
      <div className="px-6 pt-4">
        <BackButton to="/" />
      </div>

      <div className="relative mx-6 mt-2 h-56 overflow-hidden rounded-2xl border border-border/60">
        <div className="absolute inset-0 bg-gradient-to-br from-fuchsia-500 via-violet-600 to-indigo-700" />
        <div className="absolute -left-20 -top-20 h-72 w-72 rounded-full bg-pink-400/40 blur-3xl" />
        <div className="absolute -bottom-24 right-10 h-72 w-72 rounded-full bg-cyan-400/30 blur-3xl" />
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_30%_70%,rgba(255,255,255,0.15),transparent_50%)]" />
      </div>

      <div className="mx-6 -mt-12 flex flex-wrap items-end justify-between gap-4 px-2">
        <div className="flex items-end gap-4">
          <motion.div
            initial={{ scale: 0.8, opacity: 0 }}
            animate={{ scale: 1, opacity: 1 }}
            className="grid h-24 w-24 place-items-center rounded-2xl border-4 border-background bg-gradient-primary text-2xl font-bold text-primary-foreground shadow-elegant"
          >
            АХ
          </motion.div>
          <div className="pb-1">
            <h1 className="text-2xl font-semibold">Александр</h1>
            <p className="text-sm text-muted-foreground">@alexandr</p>
            <div className="mt-1.5 flex flex-wrap gap-x-4 gap-y-1 text-[11px] text-muted-foreground">
              <span className="inline-flex items-center gap-1"><MapPin className="h-3 w-3" /> Москва</span>
              <span className="inline-flex items-center gap-1"><Calendar className="h-3 w-3" /> Зарегистрирован в марте 2025</span>
              <span className="inline-flex items-center gap-1"><LinkIcon className="h-3 w-3" /> istok.app/alex</span>
            </div>
          </div>
        </div>
        <div className="flex gap-2 pb-1">
          <Button variant="outline" size="sm"><Pencil /> Редактировать профиль</Button>
          <Button asChild variant="secondary" size="sm">
            <Link to="/settings/account"><Settings /> Настройки аккаунта</Link>
          </Button>
        </div>
      </div>

      <section className="mx-6 mt-10 grid gap-6 lg:grid-cols-[1fr_280px]">
        <div className="rounded-xl border border-border/60 bg-card/40 p-5">
          <div className="mb-4 flex items-baseline justify-between">
            <h2 className="text-sm font-semibold">{totalEdits} правок за последний год</h2>
            <p className="text-[11px] text-muted-foreground">Активность вкладов</p>
          </div>

          <div className="flex gap-3">
            <div className="flex flex-col justify-between py-1 text-[9px] text-muted-foreground">
              <span>Пн</span><span>Ср</span><span>Пт</span>
            </div>
            <div className="flex-1 overflow-x-auto">
              <div className="mb-1 grid grid-cols-12 text-[9px] text-muted-foreground">
                {months.map((m) => <span key={m}>{m}</span>)}
              </div>
              <div className="grid grid-flow-col grid-rows-7 gap-[3px]">
                {cells.map((v, i) => (
                  <div key={i} className={`h-2.5 w-2.5 rounded-[2px] ${level(v)}`} />
                ))}
              </div>
            </div>
          </div>

          <div className="mt-4 flex items-center justify-end gap-1.5 text-[10px] text-muted-foreground">
            Меньше
            <span className="h-2.5 w-2.5 rounded-[2px] bg-muted/30" />
            <span className="h-2.5 w-2.5 rounded-[2px] bg-primary/30" />
            <span className="h-2.5 w-2.5 rounded-[2px] bg-primary/60" />
            <span className="h-2.5 w-2.5 rounded-[2px] bg-primary" />
            Больше
          </div>
        </div>

        <div className="space-y-3">
          <StatCard label="Среднее в день" value={(totalEdits / 365).toFixed(2)} />
          <StatCard label="Текущая серия" value="4 дня" />
          <StatCard label="Активных дней" value={`${daysActive}`} />
        </div>
      </section>

      <section className="mx-6 mb-10 mt-8">
        <h2 className="mb-3 text-sm font-semibold">Опубликованные проекты</h2>
        <div className="grid place-items-center rounded-xl border border-dashed border-border/60 bg-card/20 py-16">
          <div className="grid h-12 w-12 place-items-center rounded-full bg-muted/40 text-muted-foreground">
            <FolderOpen className="h-5 w-5" />
          </div>
          <p className="mt-3 text-sm font-medium">Пока нет опубликованных проектов</p>
          <p className="mt-1 text-xs text-muted-foreground">Опубликуйте проект, чтобы показать его в профиле.</p>
        </div>
      </section>
    </div>
  );
}

function StatCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-xl border border-border/60 bg-card/40 p-4">
      <p className="text-[11px] uppercase tracking-wider text-muted-foreground">{label}</p>
      <p className="mt-1 text-2xl font-semibold">{value}</p>
    </div>
  );
}
