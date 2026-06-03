import { useState } from "react";
import { motion } from "framer-motion";
import {
  Search, Database, CreditCard, MapPin, Send, Mail,
  Figma, Github, Inbox, Plug, Plus, Check, X as XIcon,
  FileText,
} from "lucide-react";
import { Dialog, DialogContent, DialogTitle, DialogDescription } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

interface ConnectorsModalProps {
  open: boolean;
  onOpenChange: (v: boolean) => void;
}

type AppConnector = {
  name: string;
  description: string;
  icon: typeof Database;
  tint: string;
  enabled?: boolean;
};

const appConnectors: AppConnector[] = [
  { name: "Cloud (Supabase)", description: "Postgres, авторизация, хранилище и realtime — встроено.", icon: Database, tint: "from-emerald-500/30 to-emerald-500/0", enabled: true },
  { name: "Stripe", description: "Платежи, подписки и клиентский портал.", icon: CreditCard, tint: "from-violet-500/30 to-violet-500/0", enabled: true },
  { name: "Google Maps", description: "Карты, места, маршруты и геокодинг.", icon: MapPin, tint: "from-sky-500/30 to-sky-500/0" },
  { name: "Telegram", description: "Отправляйте сообщения от бота и принимайте вебхуки.", icon: Send, tint: "from-cyan-500/30 to-cyan-500/0" },
  { name: "Resend", description: "Транзакционные письма с React-шаблонами.", icon: Mail, tint: "from-rose-500/30 to-rose-500/0", enabled: true },
  { name: "OpenAI", description: "Модели GPT для контента, кода и рассуждений.", icon: Inbox, tint: "from-fuchsia-500/30 to-fuchsia-500/0" },
];

type McpConnector = {
  name: string;
  description: string;
  icon: typeof Figma;
  tint: string;
};

const mcpConnectors: McpConnector[] = [
  { name: "Figma", description: "Подтягивайте фреймы, компоненты и токены в промты.", icon: Figma, tint: "from-pink-500/30 to-pink-500/0" },
  { name: "Linear", description: "Используйте задачи и проекты как контекст сборки.", icon: FileText, tint: "from-indigo-500/30 to-indigo-500/0" },
  { name: "Notion", description: "Подключайте страницы, базы и спецификации в чат.", icon: FileText, tint: "from-amber-400/30 to-amber-400/0" },
  { name: "GitHub", description: "Ссылайтесь на репозитории, задачи и diff-ы PR.", icon: Github, tint: "from-zinc-400/30 to-zinc-400/0" },
];

const ghostLogos = [Database, CreditCard, MapPin, Send, Mail, Figma, Github, Inbox, FileText, Plug];

export function ConnectorsModal({ open, onOpenChange }: ConnectorsModalProps) {
  const [q, setQ] = useState("");
  const filter = <T extends { name: string; description: string }>(items: T[]) =>
    items.filter(
      (i) =>
        i.name.toLowerCase().includes(q.toLowerCase()) ||
        i.description.toLowerCase().includes(q.toLowerCase()),
    );

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        className="max-w-5xl gap-0 overflow-hidden border-border/60 bg-card/95 p-0 backdrop-blur-2xl [&>button]:hidden"
      >
        <DialogTitle className="sr-only">Коннекторы и интеграции</DialogTitle>
        <DialogDescription className="sr-only">
          Подключите внешние сервисы и источники данных к рабочему пространству.
        </DialogDescription>
        {/* Header */}
        <div className="relative overflow-hidden border-b border-border/60 px-8 pb-8 pt-10">
          <div className="pointer-events-none absolute inset-0 opacity-[0.07]">
            <div className="grid h-full w-full grid-cols-10 place-items-center">
              {ghostLogos.concat(ghostLogos).map((Logo, i) => (
                <Logo key={i} className="h-8 w-8 text-foreground" />
              ))}
            </div>
          </div>
          <div className="pointer-events-none absolute inset-0 bg-gradient-to-b from-transparent via-card/40 to-card/95" />
          <div className="pointer-events-none absolute -left-24 -top-24 h-72 w-72 rounded-full bg-primary/20 blur-3xl" />
          <div className="pointer-events-none absolute -right-24 -top-32 h-80 w-80 rounded-full bg-fuchsia-500/20 blur-3xl" />

          <div className="relative flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
            <div className="max-w-xl">
              <div className="mb-2 inline-flex items-center gap-2 rounded-full border border-border/60 bg-background/40 px-3 py-1 text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
                <Plug className="h-3 w-3" /> Коннекторы
              </div>
              <h2 className="text-2xl font-semibold tracking-tight text-foreground sm:text-3xl">
                Стройте на том, чем уже пользуетесь
              </h2>
              <p className="mt-2 text-sm text-muted-foreground">
                Подключите Исток к своему стеку. Данные, платежи, дизайн и код — в один клик.
              </p>
            </div>
            <div className="relative w-full sm:w-72">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                placeholder="Поиск коннекторов…"
                value={q}
                onChange={(e) => setQ(e.target.value)}
                className="h-10 border-border/60 bg-background/60 pl-9"
              />
            </div>
            <button
              onClick={() => onOpenChange(false)}
              aria-label="Закрыть"
              className="absolute right-0 top-0 grid h-8 w-8 place-items-center rounded-md text-muted-foreground hover:bg-muted/40 hover:text-foreground"
            >
              <XIcon className="h-4 w-4" />
            </button>
          </div>
        </div>

        {/* Body */}
        <div className="max-h-[60vh] overflow-y-auto px-8 py-7">
          <SectionHeader title="Коннекторы приложений" subtitle="Подключают возможности в ваших сгенерированных приложениях." />
          <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
            {filter(appConnectors).map((c) => (
              <ConnectorCard key={c.name} connector={c} kind="app" />
            ))}
          </div>

          <div className="mt-9">
            <SectionHeader
              title="Коннекторы чата"
              subtitle="Подтягивают внешний контекст в Исток через MCP."
              badge="MCP"
            />
            <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
              {filter(mcpConnectors).map((c) => (
                <ConnectorCard key={c.name} connector={c} kind="mcp" />
              ))}
            </div>
          </div>
        </div>

        {/* Footer */}
        <div className="relative overflow-hidden border-t border-border/60 bg-elevated/40 px-8 py-4">
          <div className="pointer-events-none absolute inset-0 bg-gradient-to-r from-primary/5 via-transparent to-fuchsia-500/5" />
          <div className="relative flex flex-col items-start justify-between gap-3 sm:flex-row sm:items-center">
            <div>
              <p className="text-sm font-medium text-foreground">Не нашли нужный коннектор?</p>
              <p className="text-xs text-muted-foreground">Расскажите, что интегрировать дальше — мы выпускаем обновления еженедельно.</p>
            </div>
            <Button size="sm" variant="outline" className="border-border/60 bg-background/40">
              <Plus className="mr-1.5 h-3.5 w-3.5" /> Запросить коннектор
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}

function SectionHeader({
  title, subtitle, badge,
}: { title: string; subtitle: string; badge?: string }) {
  return (
    <div className="mb-3 flex items-end justify-between">
      <div>
        <div className="flex items-center gap-2">
          <h3 className="text-sm font-semibold tracking-tight text-foreground">{title}</h3>
          {badge && (
            <span className="rounded border border-primary/30 bg-primary/10 px-1.5 py-0.5 text-[9px] font-semibold uppercase tracking-wider text-primary">
              {badge}
            </span>
          )}
        </div>
        <p className="text-xs text-muted-foreground">{subtitle}</p>
      </div>
    </div>
  );
}

function ConnectorCard({
  connector, kind,
}: {
  connector: AppConnector | McpConnector;
  kind: "app" | "mcp";
}) {
  const Icon = connector.icon;
  const enabled = "enabled" in connector ? connector.enabled : false;
  return (
    <motion.button
      whileHover={{ y: -1 }}
      className={cn(
        "group relative flex items-center gap-3 overflow-hidden rounded-xl border border-border/60 bg-card/60 p-3.5 text-left transition-all hover:border-primary/40",
      )}
    >
      <div className={cn("absolute inset-0 bg-gradient-to-br opacity-60", connector.tint)} />
      <div className="relative grid h-10 w-10 shrink-0 place-items-center rounded-lg border border-border/60 bg-background/60">
        <Icon className="h-4.5 w-4.5 text-foreground" />
      </div>
      <div className="relative min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <p className="truncate text-sm font-medium text-foreground">{connector.name}</p>
          {kind === "mcp" && (
            <span className="rounded border border-primary/30 bg-primary/10 px-1 py-0.5 text-[9px] font-semibold uppercase tracking-wider text-primary">
              MCP
            </span>
          )}
        </div>
        <p className="truncate text-xs text-muted-foreground">{connector.description}</p>
      </div>
      <div className="relative shrink-0">
        {enabled ? (
          <Badge className="gap-1 border-emerald-500/30 bg-emerald-500/10 text-emerald-300 hover:bg-emerald-500/10">
            <Check className="h-3 w-3" /> Подключено
          </Badge>
        ) : (
          <span className="text-xs text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100">
            Подключить →
          </span>
        )}
      </div>
    </motion.button>
  );
}
