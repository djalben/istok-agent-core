import { motion, AnimatePresence } from "framer-motion";
import { Check, Loader2, Circle, AlertCircle, ArrowLeft, ArrowRight, RotateCw, Lock, Sparkles } from "lucide-react";
import { cn } from "@/lib/utils";
import type { Agent } from "@/lib/mockData";

interface PreviewPanelProps {
  agents: Agent[];
  clean?: boolean;
}

function AgentIcon({ status }: { status: Agent["status"] }) {
  if (status === "working") return <Loader2 className="h-3.5 w-3.5 animate-spin text-primary" />;
  if (status === "thinking") return <Circle className="h-3.5 w-3.5 fill-muted-foreground/40 text-muted-foreground/40" />;
  if (status === "done") return <Check className="h-3.5 w-3.5 text-success" />;
  if (status === "error") return <AlertCircle className="h-3.5 w-3.5 text-destructive" />;
  return <Circle className="h-3.5 w-3.5 text-muted-foreground/30" />;
}

export function PreviewPanel({ agents, clean = false }: PreviewPanelProps) {
  return (
    <div className="flex h-full flex-col bg-panel">
      <div className="flex h-10 items-center justify-between border-b border-border/60 px-3">
        <div className="flex items-center gap-2">
          <div className="h-2 w-2 rounded-full bg-warning animate-pulse" />
          <span className="text-xs font-medium text-muted-foreground">Живой предпросмотр</span>
        </div>
        <span className="font-mono text-[10px] text-muted-foreground">localhost:5173</span>
      </div>

      <div className="flex items-center gap-2 border-b border-border/60 bg-background/40 px-3 py-2">
        <div className="flex gap-1">
          <ArrowLeft className="h-3.5 w-3.5 text-muted-foreground/50" />
          <ArrowRight className="h-3.5 w-3.5 text-muted-foreground/50" />
          <RotateCw className="h-3.5 w-3.5 text-muted-foreground/50" />
        </div>
        <div className="flex flex-1 items-center gap-1.5 rounded-md border border-border/60 bg-elevated px-2 py-1 font-mono text-[11px] text-muted-foreground">
          <Lock className="h-3 w-3 text-success" />
          taxigo.preview.istok.dev
        </div>
      </div>

      <div className="relative flex-1 overflow-hidden bg-background">
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_30%_10%,rgba(139,92,246,0.18),transparent_60%)]" />
        {clean ? (
          <div className="relative flex h-full flex-col items-center justify-center px-6 py-10 text-center">
            <motion.div
              initial={{ opacity: 0, scale: 0.9 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ duration: 0.5 }}
              className="relative"
            >
              <motion.div
                className="absolute inset-0 -z-10 rounded-full bg-primary/30 blur-3xl"
                animate={{ opacity: [0.4, 0.8, 0.4], scale: [0.9, 1.1, 0.9] }}
                transition={{ duration: 3, repeat: Infinity, ease: "easeInOut" }}
              />
              <div className="grid h-20 w-20 place-items-center rounded-2xl bg-gradient-primary shadow-glow">
                <Sparkles className="h-9 w-9 text-primary-foreground" />
              </div>
            </motion.div>
            <h2 className="mt-6 text-2xl font-semibold tracking-tight">
              Начните <span className="text-gradient">проект</span>
            </h2>
            <p className="mx-auto mt-2 max-w-xs text-sm text-muted-foreground">
              Опишите в чате, что вы хотите построить. Живой предпросмотр появится здесь, пока агенты Истока работают.
            </p>
            <div className="mt-6 flex items-center gap-1.5 rounded-full border border-border/80 bg-elevated px-3 py-1 font-mono text-[10px] uppercase tracking-wider text-muted-foreground">
              <span className="h-1.5 w-1.5 rounded-full bg-primary animate-pulse" /> Ждём промт
            </div>
          </div>
        ) : (
          <div className="relative flex h-full flex-col items-center justify-start overflow-auto px-6 py-10">
            <div className="text-center">
              <div className="inline-flex items-center gap-2 rounded-full border border-border/80 bg-elevated px-3 py-1 text-xs text-muted-foreground">
                <span className="h-1.5 w-1.5 rounded-full bg-primary" /> Уже в 12 городах
              </div>
              <h1 className="mt-6 text-4xl font-semibold tracking-tight md:text-5xl">
                Поездки по запросу.
                <br />
                <span className="text-gradient">Без накруток.</span>
              </h1>
              <p className="mx-auto mt-3 max-w-md text-sm text-muted-foreground">
                Нажмите кнопку, садитесь и поехали. Прозрачные фиксированные цены для пассажиров и предсказуемый доход для водителей.
              </p>
              <div className="mt-6 flex justify-center gap-2">
                <button className="rounded-md bg-gradient-primary px-4 py-2 text-xs font-medium text-primary-foreground shadow-glow">
                  Заказать поездку
                </button>
                <button className="rounded-md border border-border bg-elevated px-4 py-2 text-xs">
                  Стать водителем
                </button>
              </div>
            </div>

            <div className="mt-8 grid w-full max-w-md grid-cols-3 gap-2">
              {["Эконом", "Комфорт", "XL"].map((tier, i) => (
                <div key={tier} className="rounded-lg border border-border/80 bg-card p-3 text-left">
                  <div className="text-[10px] uppercase text-muted-foreground">{tier}</div>
                  <div className="mt-1 font-mono text-sm">{(8 + i * 4).toFixed(0)} ₽</div>
                  <div className="text-[10px] text-muted-foreground">≈ 3 мин</div>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>


      <div className="border-t border-border/60 bg-panel">
        <div className="flex items-center justify-between px-3 py-2">
          <div className="flex items-center gap-2">
            <span className="font-mono text-[10px] uppercase tracking-wider text-muted-foreground">
              Активность агентов
            </span>
            <span className="rounded-md bg-primary/15 px-1.5 py-0.5 font-mono text-[10px] text-primary">
              {agents.filter((a) => a.status === "working").length} активн.
            </span>
          </div>
          <span className="font-mono text-[10px] text-muted-foreground">
            {agents.filter((a) => a.status === "done").length}/{agents.length} готово
          </span>
        </div>
        <div className="max-h-44 overflow-y-auto scrollbar-thin">
          <AnimatePresence initial={false}>
            {agents.map((a) => (
              <motion.div
                key={a.id}
                layout
                initial={{ opacity: 0, x: -8 }}
                animate={{ opacity: 1, x: 0 }}
                className={cn(
                  "flex items-center gap-3 border-t border-border/40 px-3 py-2",
                  a.status === "working" && "bg-primary/5",
                )}
              >
                <div className="grid h-6 w-6 place-items-center rounded-md bg-elevated">
                  <AgentIcon status={a.status} />
                </div>
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2">
                    <span className="text-xs font-medium">{a.name}</span>
                    <span className="font-mono text-[10px] text-muted-foreground">{a.role}</span>
                  </div>
                  <div className="truncate text-[11px] text-muted-foreground">{a.task}</div>
                </div>
                {a.status === "working" && (
                  <div className="h-1 w-12 overflow-hidden rounded-full bg-muted">
                    <motion.div
                      className="h-full bg-gradient-primary"
                      animate={{ x: ["-100%", "100%"] }}
                      transition={{ duration: 1.2, repeat: Infinity, ease: "linear" }}
                      style={{ width: "60%" }}
                    />
                  </div>
                )}
              </motion.div>
            ))}
          </AnimatePresence>
        </div>
      </div>
    </div>
  );
}
