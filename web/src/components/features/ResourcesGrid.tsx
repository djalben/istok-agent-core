import { motion } from "framer-motion";
import { Sparkles, ArrowRight } from "lucide-react";
import { cn } from "@/lib/utils";

interface Template {
  title: string;
  description: string;
  category: string;
  gradient: string;
  pattern: "blob" | "grid" | "wave" | "mesh";
}

const templates: Template[] = [
  { title: "Платформа мероприятий", description: "Билеты, регистрации и расписание для офлайн-событий.",                  category: "Маркетплейс",  gradient: "from-fuchsia-500 via-violet-500 to-indigo-600",     pattern: "mesh" },
  { title: "Портфолио архитектора", description: "Журнальная вёрстка с кейсами, планами и фотографиями.",                  category: "Портфолио",    gradient: "from-stone-300 via-stone-500 to-stone-800",        pattern: "grid" },
  { title: "SaaS-дашборд",          description: "Метрики, биллинг и управление командой для продукта.",                   category: "Внутреннее",   gradient: "from-sky-400 via-blue-600 to-indigo-700",          pattern: "wave" },
  { title: "AI-чат стартер",        description: "Стриминговый чат с инструментами, памятью и сменой моделей.",            category: "ИИ",           gradient: "from-emerald-400 via-teal-500 to-cyan-600",        pattern: "blob" },
  { title: "Сайт ресторана",        description: "Меню, бронь столиков и история шефа с эффектным hero.",                  category: "HoReCa",       gradient: "from-amber-400 via-orange-500 to-rose-600",        pattern: "mesh" },
  { title: "Рассылка",              description: "Редактор, архив и платные подписки — удобно читать.",                    category: "Издательство", gradient: "from-rose-400 via-pink-500 to-fuchsia-600",        pattern: "wave" },
  { title: "Мобильный банк",        description: "Счета, переводы и бюджет с глянцевым интерфейсом.",                      category: "Финтех",       gradient: "from-zinc-200 via-slate-400 to-zinc-700",          pattern: "grid" },
  { title: "Недвижимость",          description: "Объявления, карта и профили агентов, готовые к брендированию.",          category: "Маркетплейс",  gradient: "from-lime-400 via-emerald-500 to-teal-700",        pattern: "blob" },
  { title: "Платформа курсов",      description: "Уроки, прогресс и тесты с оплатой через Stripe.",                        category: "Образование",  gradient: "from-violet-400 via-purple-600 to-indigo-700",     pattern: "mesh" },
  { title: "Крипто-кошелёк",        description: "Мультичейн-балансы, отправка/приём и цены в реальном времени.",          category: "Финтех",       gradient: "from-yellow-300 via-amber-500 to-orange-600",      pattern: "wave" },
  { title: "Календарь бронирования", description: "Запись клиентов для студий, тренеров и сервисов.",                       category: "Продуктивность", gradient: "from-cyan-300 via-sky-500 to-blue-700",          pattern: "grid" },
  { title: "Лендинг",                description: "Hero, сетка фич, тарифы и FAQ — заточен под конверсию.",                category: "Маркетинг",    gradient: "from-pink-400 via-rose-500 to-red-600",            pattern: "mesh" },
  { title: "Документация DevTools",  description: "MDX-документация с поиском и тёмными код-блоками.",                     category: "Документация", gradient: "from-slate-300 via-zinc-500 to-slate-800",         pattern: "grid" },
  { title: "Форум сообщества",       description: "Темы, голоса, модерация и репутация пользователей.",                    category: "Социальное",   gradient: "from-orange-400 via-red-500 to-rose-700",          pattern: "blob" },
  { title: "Музыкальный плеер",      description: "Библиотека, плейлисты и эффектный экран «сейчас играет».",              category: "Медиа",        gradient: "from-purple-400 via-fuchsia-600 to-pink-700",      pattern: "wave" },
];

const categories = ["Все", "Маркетплейс", "Портфолио", "Внутреннее", "ИИ", "Финтех", "Маркетинг"];

export function ResourcesGrid() {
  return (
    <div>
      {/* Header */}
      <div className="relative overflow-hidden border-b border-border/60 px-8 py-12">
        <div className="pointer-events-none absolute -left-24 top-1/2 h-72 w-72 -translate-y-1/2 rounded-full bg-primary/15 blur-3xl" />
        <div className="pointer-events-none absolute -right-24 -top-12 h-80 w-80 rounded-full bg-fuchsia-500/15 blur-3xl" />
        <div className="relative mx-auto max-w-7xl">
          <div className="inline-flex items-center gap-2 rounded-full border border-border/60 bg-background/40 px-3 py-1 text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
            <Sparkles className="h-3 w-3 text-primary" /> Шаблоны
          </div>
          <h1 className="mt-3 text-3xl font-semibold tracking-tight sm:text-4xl">
            Начните с проверенной формы
          </h1>
          <p className="mt-2 max-w-xl text-sm text-muted-foreground">
            Подобранные точки старта от команды Истока. Откройте любой шаблон —
            клонируйте промт и адаптируйте под ваш бренд за секунды.
          </p>
          <div className="mt-6 flex flex-wrap gap-2">
            {categories.map((c, i) => (
              <button
                key={c}
                className={cn(
                  "rounded-full border px-3 py-1 text-xs transition-colors",
                  i === 0
                    ? "border-primary/40 bg-primary/10 text-foreground"
                    : "border-border/60 bg-background/40 text-muted-foreground hover:text-foreground",
                )}
              >
                {c}
              </button>
            ))}
          </div>
        </div>
      </div>

      {/* Grid */}
      <div className="mx-auto max-w-7xl px-8 py-8">
        <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3">
          {templates.map((t, i) => (
            <TemplateCard key={t.title} template={t} index={i} />
          ))}
        </div>
      </div>
    </div>
  );
}

function TemplateCard({ template, index }: { template: Template; index: number }) {
  return (
    <motion.button
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: Math.min(index * 0.03, 0.3), duration: 0.25 }}
      whileHover={{ y: -3 }}
      className="group relative overflow-hidden rounded-2xl border border-border/60 bg-card/60 text-left transition-all hover:border-primary/40 hover:shadow-glow"
    >
      <div className="relative h-44 overflow-hidden">
        <div className={cn("absolute inset-0 bg-gradient-to-br", template.gradient)} />
        <Pattern kind={template.pattern} />
        <div className="absolute left-3 top-3 flex gap-1.5">
          <span className="h-2 w-2 rounded-full bg-white/40" />
          <span className="h-2 w-2 rounded-full bg-white/40" />
          <span className="h-2 w-2 rounded-full bg-white/40" />
        </div>
        <div className="absolute inset-x-3 bottom-3 flex h-16 items-end gap-2">
          <div className="h-6 w-16 rounded bg-white/30 backdrop-blur" />
          <div className="h-6 flex-1 rounded bg-white/15 backdrop-blur" />
        </div>
        <div className="absolute inset-0 bg-gradient-to-t from-background/30 to-transparent" />
      </div>
      <div className="space-y-1 p-4">
        <div className="flex items-center justify-between gap-2">
          <h3 className="truncate text-sm font-semibold text-foreground">{template.title}</h3>
          <span className="shrink-0 rounded border border-border/60 bg-background/60 px-1.5 py-0.5 text-[9px] uppercase tracking-wider text-muted-foreground">
            {template.category}
          </span>
        </div>
        <p className="line-clamp-2 text-xs text-muted-foreground">{template.description}</p>
        <div className="flex items-center justify-between pt-2">
          <span className="text-[10px] uppercase tracking-wider text-muted-foreground">Шаблон</span>
          <span className="inline-flex items-center gap-1 text-xs text-primary opacity-0 transition-opacity group-hover:opacity-100">
            Использовать <ArrowRight className="h-3 w-3" />
          </span>
        </div>
      </div>
    </motion.button>
  );
}

function Pattern({ kind }: { kind: Template["pattern"] }) {
  if (kind === "grid") {
    return (
      <div
        className="absolute inset-0 opacity-25"
        style={{
          backgroundImage:
            "linear-gradient(rgba(255,255,255,.18) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,.18) 1px, transparent 1px)",
          backgroundSize: "22px 22px",
        }}
      />
    );
  }
  if (kind === "wave") {
    return (
      <svg className="absolute inset-x-0 bottom-0 h-24 w-full opacity-50" viewBox="0 0 400 100" preserveAspectRatio="none">
        <path d="M0,60 C100,20 200,100 400,40 L400,100 L0,100 Z" fill="rgba(255,255,255,0.15)" />
        <path d="M0,70 C120,40 220,90 400,60 L400,100 L0,100 Z" fill="rgba(255,255,255,0.1)" />
      </svg>
    );
  }
  if (kind === "blob") {
    return (
      <>
        <div className="absolute -left-10 top-2 h-32 w-32 rounded-full bg-white/30 blur-2xl" />
        <div className="absolute -right-6 bottom-2 h-24 w-24 rounded-full bg-white/20 blur-2xl" />
      </>
    );
  }
  return (
    <>
      <div className="absolute -left-12 -top-12 h-40 w-40 rounded-full bg-white/30 blur-3xl" />
      <div className="absolute right-0 top-1/2 h-32 w-32 -translate-y-1/2 rounded-full bg-black/20 blur-3xl" />
      <div className="absolute bottom-0 left-1/2 h-28 w-28 -translate-x-1/2 rounded-full bg-white/20 blur-2xl" />
    </>
  );
}
