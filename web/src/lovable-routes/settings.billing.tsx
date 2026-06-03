import { createFileRoute } from "@tanstack/react-router";
import { Check, Sparkles, Zap, Rocket, Building2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";


export const Route = createFileRoute("/settings/billing")({
  head: () => ({
    meta: [
      { title: "Тарифы — Исток" },
      { name: "description", content: "Выберите тариф Истока, который подходит вашей команде." },
      { property: "og:title", content: "Тарифы — Исток" },
      { property: "og:description", content: "Сравните тарифы Free, Pro, Business и Enterprise." },
    ],
  }),
  component: BillingPage,
});

const tiers = [
  {
    id: "free", name: "Free", price: "$0", cadence: "навсегда", icon: Sparkles,
    blurb: "Для самостоятельного знакомства с Истоком.",
    features: ["5 генераций в день", "Публичные проекты", "Поддержка сообщества"],
    cta: "Текущий тариф", disabled: true, accent: "from-muted to-muted",
  },
  {
    id: "pro", name: "Pro", price: "$20", cadence: "в месяц", icon: Zap,
    blurb: "Для тех, кто запускает свои проекты.",
    features: ["100 генераций / мес.", "Приватные проекты", "Свои домены", "Приоритетная очередь"],
    cta: "Перейти на Pro", accent: "from-violet-500 to-fuchsia-500",
  },
  {
    id: "business", name: "Business", price: "$80", cadence: "в месяц", icon: Rocket,
    blurb: "Для стартапов и команд.",
    features: ["Безлимитные генерации", "Командные рабочие пространства", "SSO и роли", "Продвинутая аналитика", "Премиум-поддержка"],
    cta: "Перейти на Business", featured: true, accent: "from-fuchsia-500 to-amber-400",
  },
  {
    id: "enterprise", name: "Enterprise", price: "По запросу", cadence: "годовой контракт", icon: Building2,
    blurb: "Для организаций с особыми требованиями.",
    features: ["SLA и выделенная поддержка", "Self-hosted раннеры", "Ревью безопасности", "Индивидуальный договор"],
    cta: "Связаться с отделом продаж", accent: "from-slate-600 to-slate-800",
  },
];

function BillingPage() {
  return (
    <div className="mx-auto max-w-7xl px-6 py-8">
        <div className="mb-10 max-w-2xl">

          <div className="inline-flex items-center gap-2 rounded-full border border-border/60 bg-elevated/60 px-3 py-1 text-xs text-muted-foreground">
            <Sparkles className="h-3 w-3 text-primary" /> Тарифы и оплата
          </div>
          <h1 className="mt-4 text-4xl font-semibold tracking-tight">
            Выберите тариф, который <span className="text-gradient">растёт вместе с вами</span>
          </h1>
          <p className="mt-2 text-sm text-muted-foreground">
            Меняйте или отменяйте подписку в любой момент. Во все платные тарифы включены безлимитные соавторы.
          </p>
        </div>

        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-4">
          {tiers.map((t) => (
            <div
              key={t.id}
              className={cn(
                "relative overflow-hidden rounded-2xl border bg-card p-6 transition-all",
                t.featured
                  ? "border-primary/60 shadow-glow ring-1 ring-primary/30"
                  : "border-border/60 hover:border-border",
              )}
            >
              {t.featured && (
                <div className="absolute right-4 top-4 rounded-full bg-gradient-primary px-2 py-0.5 text-[10px] font-medium uppercase tracking-wider text-primary-foreground">
                  Популярный
                </div>
              )}
              <div
                aria-hidden
                className={cn(
                  "absolute -right-12 -top-12 h-32 w-32 rounded-full bg-gradient-to-br opacity-20 blur-2xl",
                  t.accent,
                )}
              />

              <div className="relative">
                <div className="mb-4 inline-flex items-center gap-2">
                  <span className={cn("grid h-8 w-8 place-items-center rounded-lg bg-gradient-to-br text-white", t.accent)}>
                    <t.icon className="h-4 w-4" />
                  </span>
                  <span className="text-sm font-medium">{t.name}</span>
                </div>

                <div className="flex items-baseline gap-1.5">
                  <span className="text-3xl font-semibold tracking-tight">{t.price}</span>
                  <span className="text-xs text-muted-foreground">{t.cadence}</span>
                </div>
                <p className="mt-2 text-sm text-muted-foreground">{t.blurb}</p>

                <Button
                  disabled={t.disabled}
                  className={cn(
                    "mt-5 w-full",
                    t.featured && "bg-gradient-primary text-primary-foreground shadow-glow hover:opacity-90",
                  )}
                  variant={t.featured ? "default" : "outline"}
                >
                  {t.cta}
                </Button>

                <ul className="mt-5 space-y-2">
                  {t.features.map((f) => (
                    <li key={f} className="flex items-start gap-2 text-sm text-foreground/90">
                      <Check className="mt-0.5 h-3.5 w-3.5 shrink-0 text-primary" />
                      <span>{f}</span>
                    </li>
                  ))}
                </ul>
              </div>
            </div>
          ))}
        </div>
    </div>
  );
}

