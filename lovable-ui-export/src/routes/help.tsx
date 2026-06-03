import { createFileRoute } from "@tanstack/react-router";
import { LifeBuoy, MessageCircle, BookOpen, Mail } from "lucide-react";
import { BackButton } from "@/components/features/BackButton";

export const Route = createFileRoute("/help")({
  component: HelpPage,
  head: () => ({ meta: [{ title: "Центр помощи — Исток" }] }),
});

function HelpPage() {
  const cards = [
    { icon: BookOpen, title: "Гайды", desc: "Научитесь запускать проекты в Истоке" },
    { icon: MessageCircle, title: "Сообщество", desc: "Задавайте вопросы и делитесь идеями" },
    { icon: Mail, title: "Связаться с поддержкой", desc: "Обычно отвечаем в течение суток" },
  ];
  return (
    <div className="min-h-screen bg-background">
      <div className="mx-auto max-w-4xl px-6 py-6">
        <BackButton to="/" />
        <div className="mt-6 flex items-center gap-3">
          <div className="grid h-12 w-12 place-items-center rounded-xl bg-gradient-primary text-primary-foreground">
            <LifeBuoy className="h-5 w-5" />
          </div>
          <div>
            <h1 className="text-2xl font-semibold">Центр помощи</h1>
            <p className="text-sm text-muted-foreground">Чем мы можем помочь?</p>
          </div>
        </div>
        <div className="mt-8 grid gap-4 sm:grid-cols-3">
          {cards.map((c) => (
            <div key={c.title} className="rounded-xl border border-border/60 bg-card/40 p-5 transition hover:border-primary/40">
              <c.icon className="h-5 w-5 text-primary" />
              <p className="mt-3 text-sm font-medium">{c.title}</p>
              <p className="mt-1 text-xs text-muted-foreground">{c.desc}</p>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
