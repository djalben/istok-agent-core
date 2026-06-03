import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { Sparkles, Trash2, ShieldCheck, KeyRound } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { Progress } from "@/components/ui/progress";
import { BackButton } from "@/components/features/BackButton";

export const Route = createFileRoute("/settings/account")({
  component: AccountPage,
  head: () => ({ meta: [{ title: "Настройки аккаунта — Исток" }] }),
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

function AccountPage() {
  const [visibility, setVisibility] = useState("public");
  const [sound, setSound] = useState("chime");
  const [suggestions, setSuggestions] = useState(true);
  const [twoFA, setTwoFA] = useState(false);

  const soundLabel: Record<string, string> = {
    off: "Выкл.",
    chime: "Колокольчик",
    ding: "Дзынь",
  };

  return (
    <div className="mx-auto max-w-3xl px-6 py-8">
      <div className="mb-8">
        <h1 className="text-2xl font-semibold">Профиль аккаунта</h1>
        <p className="mt-1 text-sm text-muted-foreground">Управляйте профилем, предпочтениями и безопасностью.</p>
      </div>

      <div className="space-y-5">

          <Section title="Уровень Vibe-кодинга" description="Получайте опыт за каждый выпуск.">
            <div className="flex items-center gap-3">
              <div className="grid h-10 w-10 place-items-center rounded-lg bg-gradient-primary text-primary-foreground">
                <Sparkles className="h-4 w-4" />
              </div>
              <div className="flex-1">
                <div className="flex items-baseline justify-between text-xs">
                  <span className="font-medium">Уровень 7 · Vibe-архитектор</span>
                  <span className="text-muted-foreground">1 420 / 2 000 XP</span>
                </div>
                <Progress value={71} className="mt-2 h-2" />
              </div>
            </div>
          </Section>

          <Section title="Профиль">
            <div className="grid gap-4 sm:grid-cols-2">
              <div>
                <Label htmlFor="username">Имя пользователя</Label>
                <Input id="username" defaultValue="alexandr" className="mt-1.5" />
              </div>
              <div>
                <Label htmlFor="email">Email</Label>
                <Input id="email" type="email" defaultValue="alex@istok.app" className="mt-1.5" />
              </div>
            </div>
          </Section>

          <Section title="Видимость профиля" description="Выберите, кто может видеть ваш профиль и проекты.">
            <RadioGroup value={visibility} onValueChange={setVisibility} className="space-y-2">
              {[
                { v: "public", t: "Публичный", d: "Любой может посмотреть профиль" },
                { v: "unlisted", t: "По ссылке", d: "Только у кого есть ссылка" },
                { v: "private", t: "Приватный", d: "Только вы" },
              ].map((o) => (
                <label key={o.v} className="flex cursor-pointer items-start gap-3 rounded-lg border border-border/60 p-3 hover:bg-muted/30">
                  <RadioGroupItem value={o.v} id={o.v} className="mt-0.5" />
                  <div>
                    <p className="text-sm font-medium">{o.t}</p>
                    <p className="text-xs text-muted-foreground">{o.d}</p>
                  </div>
                </label>
              ))}
            </RadioGroup>
          </Section>

          <Section title="Подсказки в чате" description="Показывать AI-подсказки в окне сообщения.">
            <div className="flex items-center justify-between">
              <p className="text-sm">Включить подсказки</p>
              <Switch checked={suggestions} onCheckedChange={setSuggestions} />
            </div>
          </Section>

          <Section title="Звук завершения генерации">
            <RadioGroup value={sound} onValueChange={setSound} className="grid gap-2 sm:grid-cols-3">
              {["off", "chime", "ding"].map((s) => (
                <label key={s} className="flex cursor-pointer items-center gap-2 rounded-lg border border-border/60 p-3 hover:bg-muted/30">
                  <RadioGroupItem value={s} id={`s-${s}`} />
                  <span className="text-sm">{soundLabel[s]}</span>
                </label>
              ))}
            </RadioGroup>
          </Section>

          <Section title="Привязанные аккаунты">
            <div className="flex items-center justify-between rounded-lg border border-border/60 p-3">
              <div className="flex items-center gap-3">
                <div className="grid h-9 w-9 place-items-center rounded-md bg-white">
                  <svg viewBox="0 0 24 24" className="h-4 w-4"><path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 0 1-2.2 3.32v2.76h3.56c2.08-1.92 3.28-4.74 3.28-8.09Z"/><path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.56-2.76c-.99.66-2.25 1.06-3.72 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84A11 11 0 0 0 12 23Z"/><path fill="#FBBC05" d="M5.84 14.11A6.6 6.6 0 0 1 5.48 12c0-.73.13-1.45.36-2.11V7.05H2.18A11 11 0 0 0 1 12c0 1.77.43 3.45 1.18 4.95l3.66-2.84Z"/><path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.06 14.97 1 12 1A11 11 0 0 0 2.18 7.05l3.66 2.84C6.71 7.31 9.14 5.38 12 5.38Z"/></svg>
                </div>
                <div>
                  <p className="text-sm font-medium">Google</p>
                  <p className="text-xs text-muted-foreground">alex@gmail.com · Подключено</p>
                </div>
              </div>
              <Button variant="outline" size="sm">Отключить</Button>
            </div>
          </Section>

          <Section title="Двухфакторная аутентификация" description="Добавьте дополнительный уровень защиты аккаунта.">
            <div className="flex items-center justify-between rounded-lg border border-border/60 p-3">
              <div className="flex items-center gap-3">
                <div className="grid h-9 w-9 place-items-center rounded-md bg-emerald-500/10 text-emerald-400">
                  <ShieldCheck className="h-4 w-4" />
                </div>
                <div>
                  <p className="text-sm font-medium">Приложение-аутентификатор</p>
                  <p className="text-xs text-muted-foreground">{twoFA ? "Включено" : "Не настроено"}</p>
                </div>
              </div>
              <Switch checked={twoFA} onCheckedChange={setTwoFA} />
            </div>
            <Button variant="outline" size="sm" className="mt-3"><KeyRound /> Управлять кодами восстановления</Button>
          </Section>

          <section className="rounded-xl border border-destructive/30 bg-destructive/5 p-6">
            <div className="mb-4">
              <h2 className="text-sm font-semibold text-destructive">Опасная зона</h2>
              <p className="mt-0.5 text-xs text-muted-foreground">После удаления аккаунта восстановить его нельзя.</p>
            </div>
            <div className="flex items-center justify-between rounded-lg border border-destructive/30 bg-background/40 p-3">
              <div>
                <p className="text-sm font-medium">Удалить аккаунт</p>
                <p className="text-xs text-muted-foreground">Полностью удалить аккаунт и все данные.</p>
              </div>
            <Button variant="destructive" size="sm"><Trash2 /> Удалить аккаунт</Button>
          </div>
        </section>
      </div>
    </div>
  );
}

