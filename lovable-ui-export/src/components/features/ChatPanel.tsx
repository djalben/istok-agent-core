import { useEffect, useRef, useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { Wand2, ArrowUp, Bot, User, Zap, Mic, ChevronDown, MessageSquare, Hammer } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { ScrollArea } from "@/components/ui/scroll-area";
import { AttachmentMenu } from "./AttachmentMenu";
import { type ChatMessage } from "@/lib/mockData";
import { Sparkles } from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { toast } from "sonner";

interface ChatPanelProps {
  onSubmit: () => void;
  initialMessages?: ChatMessage[];
  onMessagesChange?: (messages: ChatMessage[]) => void;
  projectName?: string;
}

export function ChatPanel({ onSubmit, initialMessages = [], onMessagesChange, projectName }: ChatPanelProps) {
  const [messages, setMessages] = useState<ChatMessage[]>(initialMessages);
  const [input, setInput] = useState("");
  const [visualEditing, setVisualEditing] = useState(false);
  const [mode, setMode] = useState<"build" | "chat">("build");
  const [recording, setRecording] = useState(false);
  const endRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    endRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages.length]);

  const updateMessages = (next: ChatMessage[]) => {
    setMessages(next);
    onMessagesChange?.(next);
  };

  const send = () => {
    const text = input.trim();
    if (!text) return;
    const now = new Date().toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
    updateMessages([
      ...messages,
      { id: crypto.randomUUID(), role: "user", content: text, timestamp: now },
      {
        id: crypto.randomUUID(),
        role: "assistant",
        content: "Принял. Поднимаю команду агентов и готовлю план перед тем, как написать код.",
        timestamp: now,
      },
    ]);
    setInput("");
    onSubmit();
  };

  const enhance = () => {
    if (!input.trim()) return;
    setInput((p) =>
      `${p}\n\nСделай интерфейс премиальным и адаптированным под тёмную тему. Добавь аккуратную анимацию, осмысленные пустые состояния и удобную навигацию с клавиатуры.`,
    );
  };

  return (
    <div className="flex h-full flex-col bg-panel">
      <div className="flex h-10 items-center justify-between border-b border-border/60 px-3">
        <div className="flex items-center gap-2">
          <div className="h-2 w-2 rounded-full bg-success animate-pulse" />
          <span className="text-xs font-medium text-muted-foreground">Диалог</span>
        </div>
        <span className="font-mono text-[10px] text-muted-foreground">сессия · 42 мин</span>
      </div>

      <ScrollArea className="flex-1 px-3">
        <div className="space-y-4 py-4">
          {messages.length === 0 ? (
            <div className="flex flex-col items-center justify-center px-4 py-12 text-center">
              <div className="grid h-12 w-12 place-items-center rounded-xl bg-gradient-primary shadow-glow">
                <Sparkles className="h-5 w-5 text-primary-foreground" />
              </div>
              <h3 className="mt-4 text-sm font-semibold text-foreground">
                {projectName ? `Продолжить ${projectName}` : "Начать новый проект"}
              </h3>
              <p className="mt-1 max-w-[260px] text-xs text-muted-foreground">
                Опишите приложение, которое хотите создать. Команда агентов Истока спланирует, соберёт и запустит его.
              </p>
            </div>
          ) : null}
          <AnimatePresence initial={false}>
            {messages.map((m) => (
              <motion.div
                key={m.id}
                initial={{ opacity: 0, y: 8 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0 }}
              >
                {m.role === "system" ? (
                  <div className="rounded-md border border-dashed border-border bg-muted/30 px-3 py-2 text-center font-mono text-[11px] text-muted-foreground">
                    {m.content}
                  </div>
                ) : (
                  <div className="flex gap-2.5">
                    <div className={`mt-0.5 grid h-7 w-7 shrink-0 place-items-center rounded-md ${
                      m.role === "user" ? "bg-elevated" : "bg-gradient-primary shadow-glow"
                    }`}>
                      {m.role === "user" ? <User className="h-3.5 w-3.5" /> : <Bot className="h-3.5 w-3.5 text-primary-foreground" />}
                    </div>
                    <div className="min-w-0 flex-1">
                      <div className="mb-0.5 flex items-baseline gap-2">
                        <span className="text-xs font-medium">{m.role === "user" ? "Вы" : "Исток"}</span>
                        <span className="font-mono text-[10px] text-muted-foreground">{m.timestamp}</span>
                      </div>
                      <p className="text-sm leading-relaxed text-foreground/90">{m.content}</p>
                    </div>
                  </div>
                )}
              </motion.div>
            ))}
          </AnimatePresence>
          <div ref={endRef} />
        </div>
      </ScrollArea>


      <div className="border-t border-border/60 p-3">
        <div className="rounded-xl border border-border bg-elevated p-2 focus-within:border-primary/60 focus-within:shadow-glow">
          <Textarea
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter" && !e.shiftKey) {
                e.preventDefault();
                send();
              }
            }}
            placeholder="Опишите, что собрать, улучшить или исправить…"
            className="min-h-[72px] resize-none border-0 bg-transparent p-2 text-sm shadow-none focus-visible:ring-0"
          />
          <div className="flex items-center justify-between gap-2 px-1 pt-1">
            <div className="flex items-center gap-1">
              <AttachmentMenu />
              <button
                type="button"
                onClick={() => {
                  setVisualEditing((v) => !v);
                  toast(visualEditing ? "Визуальное редактирование выключено" : "Визуальное редактирование включено");
                }}
                className={`group flex h-7 items-center gap-1.5 rounded-md border px-2 text-xs transition-all ${
                  visualEditing
                    ? "border-primary/60 bg-primary/10 text-primary shadow-glow"
                    : "border-border/70 bg-elevated/60 text-muted-foreground hover:border-primary/40 hover:text-foreground"
                }`}
              >
                <Zap className={`h-3.5 w-3.5 ${visualEditing ? "fill-primary" : ""}`} />
                Визуальное редактирование
              </button>
              <Button
                variant="ghost"
                size="sm"
                onClick={enhance}
                className="h-7 gap-1.5 px-2 text-xs text-muted-foreground hover:text-primary"
              >
                <Wand2 className="h-3.5 w-3.5" /> Улучшить
              </Button>
            </div>
            <div className="flex items-center gap-1">
              <button
                type="button"
                onClick={() => {
                  setRecording((r) => !r);
                  toast(recording ? "Запись голоса остановлена" : "Слушаю…");
                }}
                className={`grid h-7 w-7 place-items-center rounded-md border transition-all ${
                  recording
                    ? "border-destructive/70 bg-destructive/15 text-destructive"
                    : "border-border/70 bg-elevated/60 text-muted-foreground hover:border-primary/40 hover:text-primary"
                }`}
                aria-label="Голосовой ввод"
              >
                <Mic className={`h-3.5 w-3.5 ${recording ? "animate-pulse" : ""}`} />
              </button>
              <div className="flex h-7 items-center overflow-hidden rounded-md border border-border/70 bg-elevated/60">
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <button
                      type="button"
                      className="flex h-full items-center gap-1 px-1.5 text-xs text-muted-foreground hover:text-foreground"
                    >
                      {mode === "build" ? (
                        <Hammer className="h-3.5 w-3.5 text-primary" />
                      ) : (
                        <MessageSquare className="h-3.5 w-3.5 text-primary" />
                      )}
                      <span className="font-medium">{mode === "build" ? "Сборка" : "Чат"}</span>
                      <ChevronDown className="h-3 w-3" />
                    </button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent
                    align="end"
                    sideOffset={6}
                    className="w-44 border-border/70 bg-popover/95 p-1.5 backdrop-blur-xl"
                  >
                    <DropdownMenuItem
                      onSelect={(e) => {
                        e.preventDefault();
                        setMode("build");
                      }}
                      className="flex cursor-pointer items-start gap-2 rounded-md px-2 py-1.5 text-xs"
                    >
                      <Hammer className="mt-0.5 h-3.5 w-3.5 text-primary" />
                      <div>
                        <div className="font-medium">Сборка</div>
                        <div className="text-[10px] text-muted-foreground">
                          Спланировать, собрать и запустить
                        </div>
                      </div>
                    </DropdownMenuItem>
                    <DropdownMenuItem
                      onSelect={(e) => {
                        e.preventDefault();
                        setMode("chat");
                      }}
                      className="flex cursor-pointer items-start gap-2 rounded-md px-2 py-1.5 text-xs"
                    >
                      <MessageSquare className="mt-0.5 h-3.5 w-3.5 text-primary" />
                      <div>
                        <div className="font-medium">Чат</div>
                        <div className="text-[10px] text-muted-foreground">
                          Обсудить без изменений
                        </div>
                      </div>
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
                <Button
                  size="sm"
                  onClick={send}
                  disabled={!input.trim()}
                  className="h-7 gap-1 rounded-none rounded-r-md border-l border-border/70 bg-gradient-primary px-2.5 text-xs text-primary-foreground hover:opacity-90"
                >
                  <ArrowUp className="h-3.5 w-3.5" />
                </Button>
              </div>
            </div>
          </div>
        </div>
        <p className="mt-2 px-1 font-mono text-[10px] text-muted-foreground">
          Shift+Enter — новая строка · Esc — отменить генерацию
        </p>
      </div>
    </div>
  );
}
