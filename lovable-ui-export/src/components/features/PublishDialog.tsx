import { useEffect, useRef, useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { Check, Copy, ExternalLink, Globe, Loader2, Rocket, Sparkles, Terminal } from "lucide-react";
import { toast } from "sonner";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

interface PublishDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  projectName?: string;
}

type Phase = "idle" | "building" | "deployed";

const BUILD_LOG = [
  "▲ Vercel CLI 34.2.1",
  "→ Подключение к проекту istok/web ...",
  "✓ Подключено к istok/web (создан .vercel)",
  "→ Проверка окружения сборки",
  "  Node v20.11.1 · bun 1.1.30 · регион iad1",
  "→ Установка зависимостей",
  "  ✓ разрешено 412 пакетов за 1.4с",
  "→ Запуск команды сборки: `bun run build`",
  "  vite v7.0.4 сборка для production...",
  "  ✓ 1284 модулей трансформировано.",
  "  dist/assets/index-9f2a.css   42.18 kB │ gzip: 8.91 kB",
  "  dist/assets/index-8c10.js   312.04 kB │ gzip: 96.22 kB",
  "  ✓ собрано за 4.82с",
  "→ Загрузка артефактов сборки (1.2 MB) ...",
  "  ✓ Загружено · 412 файлов",
  "→ Подготовка edge-функций ... готово",
  "→ Назначение доменов ...",
  "✓ Production-деплой готов",
];

export function PublishDialog({ open, onOpenChange, projectName = "istok-app" }: PublishDialogProps) {
  const [tab, setTab] = useState<"subdomain" | "custom">("subdomain");
  const [subdomain, setSubdomain] = useState(projectName.toLowerCase().replace(/\s+/g, "-"));
  const [customDomain, setCustomDomain] = useState("");
  const [phase, setPhase] = useState<Phase>("idle");
  const [logs, setLogs] = useState<string[]>([]);
  const [copied, setCopied] = useState(false);
  const logRef = useRef<HTMLDivElement>(null);

  const deployUrl =
    tab === "custom" && customDomain ? `https://${customDomain}` : `https://${subdomain}.istok.app`;

  useEffect(() => {
    if (phase !== "building") return;
    setLogs([]);
    let i = 0;
    const id = setInterval(() => {
      setLogs((prev) => [...prev, BUILD_LOG[i]]);
      i++;
      if (i >= BUILD_LOG.length) {
        clearInterval(id);
        setTimeout(() => {
          setPhase("deployed");
          toast.success("Деплой опубликован", { description: deployUrl });
        }, 500);
      }
    }, 230);
    return () => clearInterval(id);
  }, [phase, deployUrl]);

  useEffect(() => {
    logRef.current?.scrollTo({ top: logRef.current.scrollHeight, behavior: "smooth" });
  }, [logs]);

  useEffect(() => {
    if (!open) {
      setTimeout(() => {
        setPhase("idle");
        setLogs([]);
      }, 250);
    }
  }, [open]);

  const copyUrl = async () => {
    await navigator.clipboard.writeText(deployUrl);
    setCopied(true);
    toast.success("Ссылка скопирована");
    setTimeout(() => setCopied(false), 1800);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-xl border-border/80 bg-card p-0">
        <DialogHeader className="border-b border-border/60 p-6 pb-4">
          <DialogTitle className="flex items-center gap-2 text-base font-semibold">
            <Rocket className="h-4 w-4 text-primary" />
            Опубликовать в production
          </DialogTitle>
          <DialogDescription className="text-xs text-muted-foreground">
            Запустите проект на глобальной edge-сети за секунды.
          </DialogDescription>
        </DialogHeader>

        <AnimatePresence mode="wait">
          {phase === "idle" && (
            <motion.div
              key="idle"
              initial={{ opacity: 0, y: 6 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0 }}
              className="space-y-5 p-6"
            >
              <Tabs value={tab} onValueChange={(v) => setTab(v as "subdomain" | "custom")}>
                <TabsList className="grid w-full grid-cols-2">
                  <TabsTrigger value="subdomain">Поддомен</TabsTrigger>
                  <TabsTrigger value="custom">Свой домен</TabsTrigger>
                </TabsList>
                <TabsContent value="subdomain" className="mt-4 space-y-3">
                  <label className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
                    Ваш адрес на istok.app
                  </label>
                  <div className="flex items-center overflow-hidden rounded-md border border-border/60 bg-elevated/40 focus-within:border-primary/60">
                    <Input
                      value={subdomain}
                      onChange={(e) => setSubdomain(e.target.value.replace(/[^a-z0-9-]/gi, "").toLowerCase())}
                      className="h-10 border-0 bg-transparent font-mono text-sm focus-visible:ring-0"
                      placeholder="my-project"
                    />
                    <span className="border-l border-border/60 bg-background/40 px-3 py-2 font-mono text-sm text-muted-foreground">
                      .istok.app
                    </span>
                  </div>
                  <p className="text-xs text-muted-foreground">
                    Бесплатный SSL, глобальный CDN и preview-деплои без настройки — включены.
                  </p>
                </TabsContent>
                <TabsContent value="custom" className="mt-4 space-y-3">
                  <label className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
                    Ваш домен
                  </label>
                  <Input
                    value={customDomain}
                    onChange={(e) => setCustomDomain(e.target.value.trim())}
                    placeholder="www.example.com"
                    className="h-10 font-mono text-sm"
                  />
                  <div className="rounded-md border border-border/60 bg-elevated/40 p-3 font-mono text-[11px] text-muted-foreground">
                    <div className="mb-1 text-foreground">Добавьте DNS-записи для подтверждения владения:</div>
                    <div>A   @   76.76.21.21</div>
                    <div>CNAME   www   cname.istok.app</div>
                  </div>
                </TabsContent>
              </Tabs>

              <div className="rounded-lg border border-border/60 bg-elevated/40 p-3 text-xs text-muted-foreground">
                <div className="mb-1 flex items-center gap-1.5 text-foreground">
                  <Sparkles className="h-3 w-3 text-primary" /> Предпросмотр деплоя
                </div>
                <span className="font-mono">{deployUrl}</span>
              </div>

              <Button
                onClick={() => setPhase("building")}
                disabled={tab === "custom" && !customDomain}
                className="h-10 w-full bg-gradient-primary text-primary-foreground shadow-glow hover:opacity-90"
              >
                <Rocket className="h-4 w-4" /> Опубликовать на Vercel
              </Button>
            </motion.div>
          )}

          {phase === "building" && (
            <motion.div
              key="building"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="space-y-4 p-6"
            >
              <div className="flex items-center gap-3 rounded-lg border border-primary/30 bg-primary/5 p-3">
                <Loader2 className="h-4 w-4 animate-spin text-primary" />
                <div className="flex-1">
                  <p className="text-sm font-medium">Сборка и деплой…</p>
                  <p className="font-mono text-[11px] text-muted-foreground">{deployUrl}</p>
                </div>
              </div>
              <div
                ref={logRef}
                className="h-64 overflow-y-auto rounded-md border border-border/60 bg-background/80 p-3 font-mono text-[11px] leading-relaxed text-muted-foreground"
              >
                <div className="mb-2 flex items-center gap-1.5 text-foreground">
                  <Terminal className="h-3 w-3" /> build.log
                </div>
                {logs.map((line, i) => (
                  <motion.div
                    key={i}
                    initial={{ opacity: 0, x: -4 }}
                    animate={{ opacity: 1, x: 0 }}
                    className={
                      line.startsWith("✓")
                        ? "text-emerald-400"
                        : line.startsWith("→")
                          ? "text-primary"
                          : "text-muted-foreground"
                    }
                  >
                    {line}
                  </motion.div>
                ))}
                <div className="mt-1 inline-block h-3 w-1.5 animate-pulse bg-primary" />
              </div>
            </motion.div>
          )}

          {phase === "deployed" && (
            <motion.div
              key="deployed"
              initial={{ opacity: 0, scale: 0.96 }}
              animate={{ opacity: 1, scale: 1 }}
              exit={{ opacity: 0 }}
              className="space-y-5 p-6 text-center"
            >
              <div className="relative mx-auto grid h-16 w-16 place-items-center rounded-full bg-gradient-primary shadow-glow">
                <Check className="h-8 w-8 text-primary-foreground" />
                <motion.div
                  className="absolute inset-0 rounded-full border border-primary/40"
                  animate={{ scale: [1, 1.6], opacity: [0.6, 0] }}
                  transition={{ duration: 1.8, repeat: Infinity }}
                />
              </div>
              <div>
                <h3 className="text-lg font-semibold">Деплой опубликован</h3>
                <p className="mt-1 text-xs text-muted-foreground">
                  Собрано за 4.82с · Раздаётся из 18 edge-регионов
                </p>
              </div>
              <div className="flex items-center gap-2 rounded-md border border-border/60 bg-elevated/40 p-2 pl-3">
                <Globe className="h-3.5 w-3.5 text-primary" />
                <span className="flex-1 truncate text-left font-mono text-xs">{deployUrl}</span>
                <Button variant="ghost" size="sm" className="h-7 gap-1.5" onClick={copyUrl} aria-label="Скопировать ссылку">
                  {copied ? <Check className="h-3.5 w-3.5 text-primary" /> : <Copy className="h-3.5 w-3.5" />}
                </Button>
                <Button
                  size="sm"
                  className="h-7 gap-1.5 bg-gradient-primary text-primary-foreground"
                  onClick={() => window.open(deployUrl, "_blank")}
                >
                  <ExternalLink className="h-3.5 w-3.5" /> Открыть
                </Button>
              </div>
              <Button variant="outline" className="w-full" onClick={() => onOpenChange(false)}>
                Готово
              </Button>
            </motion.div>
          )}
        </AnimatePresence>
      </DialogContent>
    </Dialog>
  );
}
