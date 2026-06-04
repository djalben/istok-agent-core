import { useState } from "react";
import { createFileRoute } from "@tanstack/react-router";
import { toast } from "sonner";
import {
  Pencil,
  GitFork,
  ArrowRightLeft,
  UserRoundCog,
  EyeOff,
  ShieldCheck,
  CloudOff,
  Trash2,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Badge } from "@/components/ui/badge";

export const Route = createFileRoute("/settings/project")({
  head: () => ({
    meta: [{ title: "Настройки проекта — Исток" }],
  }),
  component: ProjectSettingsPage,
});

const stats: Array<{ label: string; value: string }> = [
  { label: "Название проекта", value: "Лендинг TaxiGo" },
  { label: "Создатель", value: "Анна Соколова" },
  { label: "Технологический стек", value: "Next.js · Tailwind · Cloud" },
  { label: "Количество правок", value: "184" },
  { label: "URL поддомен", value: "taxigo.istok.app" },
  { label: "Создано в", value: "12 марта 2026" },
  { label: "Количество сообщений", value: "326" },
  { label: "Использованы кредиты", value: "12.4 / 100" },
];

function ProjectSettingsPage() {
  const [tags, setTags] = useState("маркетинг, лендинг, такси");
  const [publicRemix, setPublicRemix] = useState(true);
  const [hideBadge, setHideBadge] = useState(false);
  const [disableAnalytics, setDisableAnalytics] = useState(false);
  const [autoSecurity, setAutoSecurity] = useState(true);

  return (
    <div className="mx-auto max-w-4xl px-8 py-10">
      <header className="mb-8">
        <p className="text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
          Проект
        </p>
        <h1 className="mt-1 text-3xl font-semibold tracking-tight">Обзор</h1>
        <p className="mt-1.5 text-sm text-muted-foreground">
          Метаданные, видимость и опасные действия для проекта.
        </p>
      </header>

      {/* Stats grid */}
      <section className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4">
        {stats.map((s) => (
          <div
            key={s.label}
            className="rounded-lg border border-border/60 bg-card/40 px-3 py-2.5"
          >
            <p className="text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
              {s.label}
            </p>
            <p className="mt-1 truncate text-sm font-medium text-foreground">
              {s.value}
            </p>
          </div>
        ))}
      </section>

      {/* Tags */}
      <section className="mt-8 space-y-1.5">
        <Label htmlFor="tags">Метки</Label>
        <Input
          id="tags"
          value={tags}
          onChange={(e) => setTags(e.target.value)}
          placeholder="Через запятую: маркетинг, лендинг"
          className="h-9 text-sm"
        />
        <p className="text-xs text-muted-foreground">
          Метки помогают группировать проекты в дашборде.
        </p>
      </section>

      {/* Settings rows */}
      <section className="mt-8 divide-y divide-border/60 rounded-lg border border-border/60 bg-card/30">
        <SwitchRow
          title="Публичный ремикс"
          description="Разрешить любому пользователю создать ремикс этого проекта."
          checked={publicRemix}
          onCheckedChange={setPublicRemix}
        />
        <ActionRow
          title="Категория проекта"
          description="Текущая: Маркетинговый сайт"
          actionLabel="Изменить"
          onClick={() => toast("Открываю выбор категории")}
        />
        <SwitchRow
          title='Скрыть значок "Сделано в ИСТОК"'
          description="Убрать брендированный бейдж со страницы опубликованного приложения."
          checked={hideBadge}
          onCheckedChange={setHideBadge}
        />
        <ActionRow
          icon={<Pencil className="h-3.5 w-3.5" />}
          title="Переименовать проект"
          description="Изменить публичное название проекта."
          actionLabel="Переименовать"
          onClick={() => toast("Откройте меню карточки для переименования")}
        />
        <ActionRow
          icon={<GitFork className="h-3.5 w-3.5" />}
          title="Проект ремикса"
          description="Создать копию с собственной историей."
          actionLabel="Ремикс"
          onClick={() => toast("Создание ремикса")}
        />
        <ActionRow
          icon={<ArrowRightLeft className="h-3.5 w-3.5" />}
          title="Переместить рабочее пространство"
          description="Перенести проект в другое пространство команды."
          actionLabel="Переместить"
          onClick={() => toast("Открываю перемещение")}
        />
        <ActionRow
          icon={<UserRoundCog className="h-3.5 w-3.5" />}
          title="Передача права собственности"
          description="Назначить нового владельца проекта."
          actionLabel="Передать"
          onClick={() => toast("Передача прав")}
        />
        <SwitchRow
          icon={<EyeOff className="h-3.5 w-3.5" />}
          title="Отключить аналитику"
          description="Не собирать события посетителей опубликованного приложения."
          checked={disableAnalytics}
          onCheckedChange={setDisableAnalytics}
        />
        <SwitchRow
          icon={<ShieldCheck className="h-3.5 w-3.5" />}
          title="Автоматическое устранение проблем безопасности"
          description="Агенты будут чинить уязвимости без подтверждения."
          checked={autoSecurity}
          onCheckedChange={setAutoSecurity}
        />
        <ActionRow
          icon={<CloudOff className="h-3.5 w-3.5" />}
          title="Снять проект с публикации"
          description="Опубликованное приложение станет недоступным."
          actionLabel="Снять с публикации"
          onClick={() => toast.warning("Проект снят с публикации")}
        />
      </section>

      {/* Danger zone */}
      <section className="mt-8 rounded-lg border border-destructive/40 bg-destructive/5 p-5">
        <div className="mb-3 flex items-center gap-2">
          <Badge variant="destructive" className="uppercase tracking-wider">
            Опасная зона
          </Badge>
        </div>
        <div className="flex items-start justify-between gap-4">
          <div>
            <h3 className="text-sm font-semibold text-foreground">Удалить проект</h3>
            <p className="mt-1 text-xs text-muted-foreground">
              Это действие необратимо. Все файлы, история чата и публикации будут
              удалены без возможности восстановления.
            </p>
          </div>
          <Button
            variant="destructive"
            className="shrink-0 gap-2"
            onClick={() => toast.error("Проект удалён")}
          >
            <Trash2 className="h-4 w-4" /> Удалить проект
          </Button>
        </div>
      </section>
    </div>
  );
}

function SwitchRow({
  icon,
  title,
  description,
  checked,
  onCheckedChange,
}: {
  icon?: React.ReactNode;
  title: string;
  description: string;
  checked: boolean;
  onCheckedChange: (v: boolean) => void;
}) {
  return (
    <div className="flex items-start justify-between gap-4 px-4 py-3.5">
      <div className="flex min-w-0 items-start gap-2.5">
        {icon && <span className="mt-0.5 text-muted-foreground">{icon}</span>}
        <div className="min-w-0">
          <p className="text-sm font-medium text-foreground">{title}</p>
          <p className="mt-0.5 text-xs text-muted-foreground">{description}</p>
        </div>
      </div>
      <Switch checked={checked} onCheckedChange={onCheckedChange} />
    </div>
  );
}

function ActionRow({
  icon,
  title,
  description,
  actionLabel,
  onClick,
}: {
  icon?: React.ReactNode;
  title: string;
  description: string;
  actionLabel: string;
  onClick: () => void;
}) {
  return (
    <div className="flex items-start justify-between gap-4 px-4 py-3.5">
      <div className="flex min-w-0 items-start gap-2.5">
        {icon && <span className="mt-0.5 text-muted-foreground">{icon}</span>}
        <div className="min-w-0">
          <p className="text-sm font-medium text-foreground">{title}</p>
          <p className="mt-0.5 text-xs text-muted-foreground">{description}</p>
        </div>
      </div>
      <Button variant="outline" size="sm" onClick={onClick} className="shrink-0">
        {actionLabel}
      </Button>
    </div>
  );
}
