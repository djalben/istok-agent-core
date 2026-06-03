import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { Sparkles, Palette, Code2, Lightbulb, Layers, Wrench, BookOpen } from "lucide-react";
import { BackButton } from "@/components/features/BackButton";
import {
  Accordion, AccordionContent, AccordionItem, AccordionTrigger,
} from "@/components/ui/accordion";
import { cn } from "@/lib/utils";

export const Route = createFileRoute("/prompts")({
  head: () => ({
    meta: [
      { title: "Библиотека промтов · Исток" },
      { name: "description", content: "Лучшие практики и советы по составлению промтов для агентов Истока." },
    ],
  }),
  component: PromptsPage,
});

type Category = {
  id: string;
  label: string;
  icon: typeof Sparkles;
  tips: { q: string; a: string }[];
};

const categories: Category[] = [
  {
    id: "strategy",
    label: "Стратегия",
    icon: Lightbulb,
    tips: [
      {
        q: "Начните проект правильно",
        a: "Опишите конечную цель, аудиторию и ключевую ценность одним абзацем. Чёткая постановка помогает агентам сразу выбрать архитектуру и стек.",
      },
      {
        q: "Используйте базу знаний",
        a: "Прикрепите документы, ссылки и заметки — модель будет опираться на них и не выдумывать факты.",
      },
      {
        q: "Разбивайте большие задачи",
        a: "Итеративные шаги дают предсказуемый результат: «сначала каркас, потом авторизация, затем биллинг».",
      },
    ],
  },
  {
    id: "design",
    label: "Дизайн",
    icon: Palette,
    tips: [
      {
        q: "Задавайте тональность визуала",
        a: "Укажите референсы, настроение и тип типографики. Например: «минимализм, тёмная тема, акцент фуксия».",
      },
      {
        q: "Просите варианты, а не правки",
        a: "Когда не уверены — попросите 2–3 направления. Так быстрее найдёте нужный визуальный язык.",
      },
    ],
  },
  {
    id: "functionality",
    label: "Функциональность",
    icon: Code2,
    tips: [
      {
        q: "Описывайте поведение, а не реализацию",
        a: "«Пользователь видит карточку и может пометить её как избранную» лучше, чем «добавь useState и handler».",
      },
      {
        q: "Указывайте граничные случаи",
        a: "Что показать при ошибке, пустом списке, отсутствии прав — это экономит несколько итераций.",
      },
    ],
  },
  {
    id: "architecture",
    label: "Архитектура",
    icon: Layers,
    tips: [
      {
        q: "Фиксируйте контракты данных",
        a: "Поля сущностей и связи помогают агентам сразу собрать корректную схему и API.",
      },
    ],
  },
  {
    id: "debugging",
    label: "Отладка",
    icon: Wrench,
    tips: [
      {
        q: "Прикладывайте ошибки полностью",
        a: "Стек-трейс, шаги воспроизведения и ожидаемое поведение — лучший способ быстро получить фикс.",
      },
    ],
  },
];

function PromptsPage() {
  const [active, setActive] = useState(categories[0].id);
  const current = categories.find((c) => c.id === active)!;

  return (
    <div className="min-h-screen bg-background">
      <div className="border-b border-border/60 bg-card/30 px-6 py-3">
        <BackButton />
      </div>
      <div className="mx-auto grid max-w-6xl grid-cols-1 gap-8 px-6 py-12 lg:grid-cols-[220px_1fr]">
        <aside>
          <div className="mb-4 flex items-center gap-2">
            <BookOpen className="h-4 w-4 text-primary" />
            <h2 className="text-sm font-semibold tracking-tight">Категории</h2>
          </div>
          <nav className="space-y-1">
            {categories.map((c) => (
              <button
                key={c.id}
                onClick={() => setActive(c.id)}
                className={cn(
                  "flex w-full items-center gap-2.5 rounded-md px-2.5 py-1.5 text-sm transition-colors",
                  active === c.id
                    ? "bg-primary/10 text-foreground"
                    : "text-muted-foreground hover:bg-muted/40 hover:text-foreground",
                )}
              >
                <c.icon className={cn("h-4 w-4", active === c.id && "text-primary")} />
                {c.label}
              </button>
            ))}
          </nav>
        </aside>

        <main>
          <div className="mb-6">
            <div className="inline-flex items-center gap-2 rounded-full border border-border/60 bg-elevated/40 px-3 py-1 text-[11px] text-muted-foreground">
              <Sparkles className="h-3 w-3 text-primary" /> Библиотека промтов
            </div>
            <h1 className="mt-3 text-3xl font-semibold tracking-tight">{current.label}</h1>
            <p className="mt-1.5 text-sm text-muted-foreground">
              Практические советы, которые делают взаимодействие с агентами Истока предсказуемым и быстрым.
            </p>
          </div>

          <Accordion type="single" collapsible className="rounded-xl border border-border/60 bg-card/40">
            {current.tips.map((t, i) => (
              <AccordionItem key={i} value={`item-${i}`} className="border-border/60 px-4">
                <AccordionTrigger className="text-left text-sm font-medium hover:no-underline">
                  {t.q}
                </AccordionTrigger>
                <AccordionContent className="text-sm leading-relaxed text-muted-foreground">
                  {t.a}
                </AccordionContent>
              </AccordionItem>
            ))}
          </Accordion>
        </main>
      </div>
    </div>
  );
}
