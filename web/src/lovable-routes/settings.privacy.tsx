import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { Switch } from "@/components/ui/switch";

export const Route = createFileRoute("/settings/privacy")({
  component: PrivacyPage,
  head: () => ({ meta: [{ title: "Конфиденциальность и безопасность — Исток" }] }),
});

type Toggle = { id: string; title: string; desc: string; default?: boolean };

const groups: { title: string; items: Toggle[] }[] = [
  {
    title: "Доступ и членство",
    items: [
      { id: "sso", title: "Требовать SSO", desc: "Все участники входят только через корпоративный провайдер.", default: false },
      { id: "guests", title: "Разрешить гостей", desc: "Внешние пользователи могут просматривать проекты по ссылке.", default: true },
      { id: "domain-lock", title: "Ограничить регистрацию доменом", desc: "Только @istok.app могут присоединиться к рабочему пространству.", default: true },
    ],
  },
  {
    title: "ИИ",
    items: [
      { id: "ai-train", title: "Использовать мои данные для улучшения моделей", desc: "Анонимизированные подсказки помогают сделать агента лучше.", default: false },
      { id: "ai-context", title: "Делиться кодом проекта с ИИ", desc: "Агент видит файлы для более точных ответов.", default: true },
      { id: "ai-retention", title: "Хранить историю чатов 30 дней", desc: "Иначе сообщения удаляются после закрытия сессии.", default: true },
    ],
  },
  {
    title: "Автоматизация безопасности",
    items: [
      { id: "secret-scan", title: "Сканирование секретов в коммитах", desc: "Блокировать публикацию ключей и токенов.", default: true },
      { id: "deps-scan", title: "Сканирование зависимостей", desc: "Уведомлять о CVE в установленных пакетах.", default: true },
      { id: "auto-patch", title: "Автоматические патчи безопасности", desc: "Обновлять зависимости с критическими уязвимостями.", default: false },
    ],
  },
  {
    title: "Совместное использование",
    items: [
      { id: "public-share", title: "Разрешить публичные ссылки", desc: "Участники могут делиться проектами без авторизации.", default: true },
      { id: "embed", title: "Разрешить встраивание (iframe)", desc: "Опубликованные проекты можно вставлять на сторонние сайты.", default: true },
      { id: "watermark", title: "Удалить плашку «Сделано в Истоке»", desc: "Доступно на тарифе Pro и выше.", default: false },
    ],
  },
  {
    title: "Разъёмы MCP",
    items: [
      { id: "mcp-allow", title: "Разрешить установку MCP-серверов", desc: "Участники могут добавлять собственные источники данных.", default: true },
      { id: "mcp-review", title: "Ручное одобрение новых серверов", desc: "Администратор подтверждает каждый MCP перед использованием.", default: false },
    ],
  },
  {
    title: "Защита данных",
    items: [
      { id: "encrypt-rest", title: "Шифрование данных в покое", desc: "AES-256 для всех файлов проектов.", default: true },
      { id: "audit-export", title: "Экспорт журналов в SIEM", desc: "Передавать события в подключённую систему мониторинга.", default: false },
      { id: "data-region", title: "Хранить данные в России", desc: "Все артефакты размещаются в локальном дата-центре.", default: true },
    ],
  },
];

function Row({ t }: { t: Toggle }) {
  const [on, setOn] = useState(!!t.default);
  return (
    <div className="flex items-center justify-between gap-4 py-3.5 first:pt-0 last:pb-0">
      <div className="min-w-0">
        <p className="text-sm font-medium">{t.title}</p>
        <p className="mt-0.5 text-xs text-muted-foreground">{t.desc}</p>
      </div>
      <Switch checked={on} onCheckedChange={setOn} />
    </div>
  );
}

function PrivacyPage() {
  return (
    <div className="mx-auto max-w-3xl px-6 py-8">
      <div className="mb-8">
        <h1 className="text-2xl font-semibold">Конфиденциальность и безопасность</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Тонко настройте, какие данные собираются и кто имеет доступ к рабочему пространству.
        </p>
      </div>

      <div className="space-y-5">
        {groups.map((g) => (
          <section key={g.title} className="rounded-xl border border-border/60 bg-card/40 p-6">
            <h2 className="mb-2 text-sm font-semibold">{g.title}</h2>
            <div className="divide-y divide-border/60">
              {g.items.map((t) => <Row key={t.id} t={t} />)}
            </div>
          </section>
        ))}
      </div>
    </div>
  );
}
