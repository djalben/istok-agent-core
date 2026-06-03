import { useState } from "react";
import { useNavigate } from "@tanstack/react-router";
import { motion } from "framer-motion";
import { Plus, Mic, ArrowUp, Hammer, ChevronDown, MessageSquare, Sparkles } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { cn } from "@/lib/utils";

export function DashboardHero() {
  const navigate = useNavigate();
  const [prompt, setPrompt] = useState("");
  const [mode, setMode] = useState<"build" | "chat">("build");

  const submit = () => {
    navigate({ to: "/builder" });
  };

  return (
    <section className="relative isolate overflow-hidden">
      {/* Mesh gradient */}
      <div className="pointer-events-none absolute inset-0 -z-10">
        <div className="absolute inset-0 bg-background" />
        <div className="absolute -top-32 left-1/2 h-[680px] w-[1100px] -translate-x-1/2 rounded-full bg-[radial-gradient(circle_at_30%_30%,rgba(236,72,153,0.45),transparent_55%),radial-gradient(circle_at_70%_40%,rgba(99,102,241,0.55),transparent_60%),radial-gradient(circle_at_50%_70%,rgba(14,165,233,0.45),transparent_60%)] blur-3xl" />
        <div className="absolute inset-x-0 top-0 h-[520px] bg-[radial-gradient(ellipse_at_top,rgba(168,85,247,0.18),transparent_60%)]" />
        <div
          className="absolute inset-0 opacity-[0.04]"
          style={{
            backgroundImage:
              "linear-gradient(rgba(255,255,255,0.4) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,0.4) 1px, transparent 1px)",
            backgroundSize: "44px 44px",
          }}
        />
        <div className="absolute inset-x-0 bottom-0 h-48 bg-gradient-to-t from-background to-transparent" />
      </div>

      <div className="mx-auto flex max-w-3xl flex-col items-center px-6 pb-20 pt-24 text-center">
        <motion.div
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          className="inline-flex items-center gap-2 rounded-full border border-border/60 bg-card/40 px-3 py-1 text-[11px] text-muted-foreground backdrop-blur"
        >
          <Sparkles className="h-3 w-3 text-primary" />
          Исток v0.4.2 · мульти-агентные сборки стали в 2 раза быстрее
        </motion.div>

        <motion.h1
          initial={{ opacity: 0, y: 16 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.12 }}
          className="mt-6 text-5xl font-semibold tracking-tight text-foreground sm:text-6xl"
        >
          Что будем создавать, <span className="text-gradient">Александр?</span>
        </motion.h1>
        <motion.p
          initial={{ opacity: 0, y: 16 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.15 }}
          className="mt-3 max-w-xl text-sm text-muted-foreground"
        >
          Опишите, что вы хотите создать — агенты Истока исследуют, спроектируют и реализуют это.
        </motion.p>

        <motion.div
          initial={{ opacity: 0, y: 20, scale: 0.98 }}
          animate={{ opacity: 1, y: 0, scale: 1 }}
          transition={{ delay: 0.2, type: "spring", stiffness: 200, damping: 24 }}
          className="relative mt-10 w-full"
        >
          <div className="absolute -inset-4 rounded-3xl bg-gradient-to-br from-primary/30 via-fuchsia-500/20 to-cyan-500/30 opacity-60 blur-2xl" />
          <div className="relative rounded-2xl border border-border/60 bg-card/80 p-3 shadow-2xl backdrop-blur-xl">
            <div className="flex items-start gap-2">
              <Button variant="ghost" size="icon" className="h-9 w-9 shrink-0 rounded-lg text-muted-foreground hover:text-foreground">
                <Plus className="h-4 w-4" />
              </Button>
              <textarea
                value={prompt}
                onChange={(e) => setPrompt(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === "Enter" && !e.shiftKey) {
                    e.preventDefault();
                    submit();
                  }
                }}
                placeholder="Попросите Исток построить..."
                rows={2}
                className="flex-1 resize-none bg-transparent px-1 py-2 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none"
              />
            </div>
            <div className="mt-2 flex items-center justify-end gap-2">
              <div className="flex items-center gap-1.5">

                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button
                      variant="outline"
                      size="sm"
                      className="h-8 gap-1.5 border-border/60 bg-elevated/40 text-xs"
                    >
                      {mode === "build" ? <Hammer className="h-3.5 w-3.5 text-primary" /> : <MessageSquare className="h-3.5 w-3.5 text-primary" />}
                      {mode === "build" ? "Сборка" : "Чат"}
                      <ChevronDown className="h-3 w-3 opacity-60" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end" className="w-44">
                    <DropdownMenuItem onClick={() => setMode("build")} className="gap-2">
                      <Hammer className="h-3.5 w-3.5" /> Режим сборки
                    </DropdownMenuItem>
                    <DropdownMenuItem onClick={() => setMode("chat")} className="gap-2">
                      <MessageSquare className="h-3.5 w-3.5" /> Режим чата
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
                <Button variant="ghost" size="icon" className="h-8 w-8 rounded-lg text-muted-foreground hover:text-foreground">
                  <Mic className="h-4 w-4" />
                </Button>
                <Button
                  size="icon"
                  onClick={submit}
                  className={cn(
                    "h-8 w-8 rounded-lg bg-gradient-primary text-primary-foreground shadow-glow transition hover:opacity-90",
                    !prompt && "opacity-70",
                  )}
                >
                  <ArrowUp className="h-4 w-4" />
                </Button>
              </div>
            </div>
          </div>

          <div className="mt-4 flex flex-wrap items-center justify-center gap-1.5 text-[11px] text-muted-foreground">
            {[
              "Лендинг для подписки на кофе",
              "CRM в реальном времени с канбаном",
              "Markdown-блог с авторизацией",
              "Внутренняя админ-панель",
            ].map((s) => (
              <button
                key={s}
                onClick={() => setPrompt(s)}
                className="rounded-full border border-border/60 bg-card/40 px-2.5 py-1 backdrop-blur transition hover:border-primary/40 hover:text-foreground"
              >
                {s}
              </button>
            ))}
          </div>
        </motion.div>
      </div>
    </section>
  );
}
