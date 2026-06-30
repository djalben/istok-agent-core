import { useEffect, useRef, useState, useMemo } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { Wand2, ArrowUp, Bot, User, Zap, ChevronDown, ChevronRight, MessageSquare, Hammer, Sparkles, Paperclip, Clapperboard, Brain, Copy, Check, FileCode, ListChecks } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { Textarea } from "@/components/ui/textarea";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { toast } from "sonner";
import type { ChatMessage } from "@/hooks/useGeneration";
import type { GenerationMode } from "@/lib/api";

interface BuilderChatPanelProps {
  messages: ChatMessage[];
  thinking: boolean;
  input: string;
  onInputChange: (v: string) => void;
  onSend: () => void;
  agentMode: GenerationMode;
  onModeChange: (m: GenerationMode) => void;
  videoEnabled?: boolean;
  onVideoEnabledChange?: (v: boolean) => void;
  projectName?: string;
  editMode?: boolean;
  onEditModeChange?: (v: boolean) => void;
  /** Raw engineering telemetry lines from the backend (LLM metrics, AST guards, pipeline stats). */
  telemetryLog?: string[];
  /** Current coder task progress from the backend, or null when idle. */
  taskProgress?: { completed: number; total: number } | null;
}

function terminalLineColor(line: string): string {
  const u = line.toUpperCase();
  if (u.includes("[LLM ERROR]") || u.includes("[LLM+REASON ERROR]")) return "text-red-400";
  if (u.includes("[LLM+REASON]") || u.includes("[LLM]")) return "text-emerald-400";
  if (u.includes("[AST GUARD]")) return "text-amber-400";
  if (u.includes("[CODER DONE]")) return "text-emerald-400/80";
  if (u.includes("[CODER]")) return "text-sky-400";
  if (u.includes("[PLANNING]")) return "text-blue-400";
  if (u.includes("[VALIDATION]")) return "text-amber-300";
  if (u.includes("[EXECUTION]")) return "text-zinc-500";
  return "text-zinc-400";
}

function fmtTime(d: Date): string {
  try {
    return d.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
  } catch {
    return "";
  }
}

const TAG_COLORS: Record<string, string> = {
  PLANNING: "text-blue-400 border-blue-500/40 bg-blue-500/10",
  EXECUTION: "text-emerald-400 border-emerald-500/40 bg-emerald-500/10",
  VALIDATION: "text-amber-400 border-amber-500/40 bg-amber-500/10",
};

function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false);
  const copy = () => {
    navigator.clipboard.writeText(text).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    });
  };
  return (
    <button
      onClick={copy}
      className="flex items-center gap-1 rounded px-1.5 py-0.5 text-[10px] text-muted-foreground transition-colors hover:text-foreground border border-border/50 hover:border-border"
    >
      {copied ? <Check className="h-3 w-3 text-success" /> : <Copy className="h-3 w-3" />}
      {copied ? "Скопировано" : "Копировать"}
    </button>
  );
}

// ── Message grouping (Cascade-style reasoning blocks) ────────
type MsgGroup =
  | { type: "thinking"; thoughts: ChatMessage[]; durationSec?: number; active: boolean }
  | { type: "msg"; message: ChatMessage };

function buildGroups(messages: ChatMessage[], thinking: boolean): MsgGroup[] {
  const groups: MsgGroup[] = [];
  let pending: ChatMessage[] = [];

  for (const m of messages) {
    if (m.kind === "thought") {
      pending.push(m);
    } else if (m.kind === "thought_duration") {
      if (pending.length > 0) {
        groups.push({ type: "thinking", thoughts: pending, durationSec: m.durationSec, active: false });
        pending = [];
      }
      // thought_duration absorbed into block above; not rendered standalone
    } else {
      if (pending.length > 0) {
        groups.push({ type: "thinking", thoughts: pending, active: false });
        pending = [];
      }
      groups.push({ type: "msg", message: m });
    }
  }

  if (pending.length > 0) {
    groups.push({ type: "thinking", thoughts: pending, active: thinking });
  }

  return groups;
}

// ── ThinkingBlock ─────────────────────────────────────────────
const TAG_ICONS: Record<string, string> = { PLANNING: "◆", EXECUTION: "▶", VALIDATION: "✓" };

function ThinkingBlock({
  thoughts,
  durationSec,
  active,
}: {
  thoughts: ChatMessage[];
  durationSec?: number;
  active: boolean;
}) {
  const [expanded, setExpanded] = useState(active);
  const bodyRef = useRef<HTMLDivElement | null>(null);

  // Expand when activated; auto-collapse when the block is "done"
  useEffect(() => {
    if (active) setExpanded(true);
  }, [active]);

  useEffect(() => {
    if (!active && durationSec !== undefined) setExpanded(false);
  }, [active, durationSec]);

  // Auto-scroll body while active
  useEffect(() => {
    if (active && bodyRef.current) {
      bodyRef.current.scrollTop = bodyRef.current.scrollHeight;
    }
  }, [thoughts.length, active]);

  return (
    <div className={`rounded-lg border overflow-hidden transition-colors ${
      active
        ? "border-zinc-700/60 bg-zinc-900/30"
        : "border-zinc-800/30 bg-zinc-950/20"
    }`}>
      {/* header */}
      <button
        type="button"
        onClick={() => setExpanded((v) => !v)}
        className="flex w-full items-center gap-2 px-3 py-2 text-left"
      >
        <Brain className={`h-3.5 w-3.5 shrink-0 transition-colors ${active ? "text-muted-foreground/70 animate-pulse" : "text-muted-foreground/30"}`} />
        <span className={`flex-1 font-mono text-xs ${active ? "text-muted-foreground/70" : "text-muted-foreground/40"}`}>
          {active
            ? "Thinking\u2026"
            : durationSec !== undefined
              ? `Thought for ${durationSec}s`
              : "Reasoning"}
        </span>
        {active && (
          <span className="flex items-center gap-0.5">
            {[0, 150, 300].map((d) => (
              <span
                key={d}
                className="h-1 w-1 rounded-full bg-muted-foreground/40 animate-pulse"
                style={{ animationDelay: `${d}ms` }}
              />
            ))}
          </span>
        )}
        {!active && (
          <span className="font-mono text-[9px] text-muted-foreground/25">{thoughts.length} steps</span>
        )}
        <ChevronDown
          className={`h-3 w-3 shrink-0 text-muted-foreground/30 transition-transform ${expanded ? "" : "-rotate-90"}`}
        />
      </button>

      {/* body */}
      <AnimatePresence initial={false}>
        {expanded && (
          <motion.div
            key="body"
            initial={{ height: 0 }}
            animate={{ height: "auto" }}
            exit={{ height: 0 }}
            transition={{ duration: 0.15, ease: "easeInOut" }}
            className="overflow-hidden"
          >
            <div
              ref={bodyRef}
              className="max-h-[220px] overflow-y-auto border-t border-zinc-800/30 px-3 py-2 space-y-1.5 scrollbar-thin scrollbar-thumb-zinc-800"
            >
              {thoughts.map((t) => (
                <div key={t.id} className="flex gap-1.5 min-w-0">
                  {t.thoughtTag && (
                    <span className="mt-px shrink-0 font-mono text-[9px] text-muted-foreground/25 select-none w-3">
                      {TAG_ICONS[t.thoughtTag] ?? "·"}
                    </span>
                  )}
                  <p className="font-mono text-[11px] leading-relaxed text-muted-foreground/50 break-words [overflow-wrap:anywhere]">
                    {t.content}
                  </p>
                </div>
              ))}
              {active && (
                <div className="flex items-center gap-1 pl-4">
                  <span className="h-1 w-4 rounded-full bg-muted-foreground/20 animate-pulse" />
                </div>
              )}
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}

// ── ThoughtDurationRow ───────────────────────────────────────
function ThoughtDurationRow({ durationSec, timestamp }: { durationSec: number; timestamp: Date }) {
  return (
    <div className="flex items-center gap-1.5 pl-1">
      <Brain className="h-3 w-3 shrink-0 text-muted-foreground/40" />
      <span className="font-mono text-[11px] text-muted-foreground/50">
        Thought for {durationSec}s
      </span>
      <span className="font-mono text-[10px] text-muted-foreground/30">{fmtTime(timestamp)}</span>
    </div>
  );
}

// ── ActionLogRow ─────────────────────────────────────────────
function ActionLogRow({
  summary,
  actionType,
  details,
  timestamp,
}: {
  summary: string;
  actionType: string;
  details: string[];
  timestamp: Date;
}) {
  const [open, setOpen] = useState(false);
  return (
    <div className="rounded-md border border-border/30 bg-elevated/20 px-2 py-1.5">
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className="flex w-full items-center gap-1.5 text-left"
      >
        <ListChecks className="h-3.5 w-3.5 shrink-0 text-sky-400/70" />
        <span className="min-w-0 flex-1 font-mono text-[11px] text-zinc-300 truncate">{summary}</span>
        <span className="font-mono text-[9px] text-muted-foreground/40 shrink-0 uppercase">{actionType}</span>
        <span className="font-mono text-[10px] text-muted-foreground/30 shrink-0 ml-1">{fmtTime(timestamp)}</span>
        {details.length > 0 && (
          open
            ? <ChevronDown className="h-3 w-3 shrink-0 text-muted-foreground/40" />
            : <ChevronRight className="h-3 w-3 shrink-0 text-muted-foreground/40" />
        )}
      </button>
      {open && details.length > 0 && (
        <ul className="mt-1.5 space-y-0.5 pl-5">
          {details.map((d, i) => (
            <li key={i} className="font-mono text-[10px] text-muted-foreground/60 truncate">• {d}</li>
          ))}
        </ul>
      )}
    </div>
  );
}

// ── CodeDiffRow ──────────────────────────────────────────────
function CodeDiffRow({
  filePath,
  diffHunk,
  additions,
  deletions,
  timestamp,
}: {
  filePath: string;
  diffHunk: string;
  additions: number;
  deletions: number;
  timestamp: Date;
}) {
  const [open, setOpen] = useState(false);
  const lines = diffHunk.split("\n");
  return (
    <div className="rounded-md border border-border/40 bg-[#0d1117] overflow-hidden">
      {/* header */}
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className="flex w-full items-center gap-2 px-2 py-1.5 border-b border-border/30 bg-zinc-900/60 hover:bg-zinc-900/80 transition-colors"
      >
        <FileCode className="h-3.5 w-3.5 shrink-0 text-sky-400/70" />
        <span className="min-w-0 flex-1 font-mono text-[11px] text-zinc-200 truncate text-left">{filePath}</span>
        <span className="font-mono text-[10px] text-emerald-400 shrink-0">+{additions}</span>
        {deletions > 0 && <span className="font-mono text-[10px] text-red-400 shrink-0">-{deletions}</span>}
        <span className="font-mono text-[10px] text-muted-foreground/30 shrink-0 ml-1">{fmtTime(timestamp)}</span>
        {open
          ? <ChevronDown className="h-3 w-3 shrink-0 text-muted-foreground/40" />
          : <ChevronRight className="h-3 w-3 shrink-0 text-muted-foreground/40" />
        }
      </button>
      {/* diff body */}
      {open && (
        <div className="max-h-[240px] overflow-y-auto px-0 py-1 scrollbar-thin scrollbar-thumb-zinc-800">
          {lines.map((line, i) => {
            const isAdd = line.startsWith("+") && !line.startsWith("+++");
            const isDel = line.startsWith("-") && !line.startsWith("---");
            const isHdr = line.startsWith("@@");
            return (
              <div
                key={i}
                className={`flex min-h-[18px] px-2 ${
                  isAdd ? "bg-emerald-950/50" : isDel ? "bg-red-950/40" : isHdr ? "bg-zinc-800/40" : ""
                }`}
              >
                <span className={`w-3 shrink-0 select-none font-mono text-[10px] leading-[18px] ${
                  isAdd ? "text-emerald-400" : isDel ? "text-red-400" : "text-zinc-600"
                }`}>
                  {isAdd ? "+" : isDel ? "-" : isHdr ? "" : " "}
                </span>
                <span className={`font-mono text-[10px] leading-[18px] break-all ${
                  isAdd ? "text-emerald-300" : isDel ? "text-red-300" : isHdr ? "text-amber-400/70" : "text-zinc-400"
                }`}>
                  {isAdd || isDel ? line.slice(1) : line}
                </span>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}

export function BuilderChatPanel({
  messages,
  thinking,
  input,
  onInputChange,
  onSend,
  agentMode,
  onModeChange,
  videoEnabled = false,
  onVideoEnabledChange,
  projectName,
  editMode = false,
  onEditModeChange,
  telemetryLog = [],
  taskProgress = null,
}: BuilderChatPanelProps) {
  const endRef = useRef<HTMLDivElement | null>(null);
  const terminalEndRef = useRef<HTMLDivElement | null>(null);
  const [terminalOpen, setTerminalOpen] = useState(true);

  useEffect(() => {
    endRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages.length, thinking]);

  useEffect(() => {
    if (terminalOpen) {
      terminalEndRef.current?.scrollIntoView({ behavior: "instant" });
    }
  }, [telemetryLog.length, terminalOpen]);

  const showTerminal = telemetryLog.length > 0 || thinking;

  const terminalLines = useMemo(() => telemetryLog.slice(-300), [telemetryLog]);
  const groups = useMemo(() => buildGroups(messages, thinking), [messages, thinking]);

  const send = () => {
    if (!input.trim() || thinking) return;
    onSend();
  };

  const enhance = () => {
    if (!input.trim()) return;
    onInputChange(
      `${input}\n\nСделай интерфейс премиальным и адаптированным под тёмную тему. Добавь аккуратную анимацию, осмысленные пустые состояния и удобную навигацию с клавиатуры.`,
    );
  };

  const isBuild = agentMode !== "code";

  return (
    <div className="flex h-full flex-col bg-panel">
      <div className="flex h-10 items-center justify-between border-b border-border/60 px-3">
        <div className="flex items-center gap-2">
          <div className="h-2 w-2 rounded-full bg-success animate-pulse" />
          <span className="text-xs font-medium text-muted-foreground">Диалог</span>
        </div>
        <span className="font-mono text-[10px] text-muted-foreground">
          {thinking ? "агенты работают…" : "готов"}
        </span>
      </div>

      {/* ── Terminal Panel ── */}
      <AnimatePresence initial={false}>
        {showTerminal && (
          <motion.div
            key="terminal"
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: terminalOpen ? 176 : 28, opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={{ duration: 0.18, ease: "easeInOut" }}
            className="shrink-0 overflow-hidden border-b border-zinc-800 bg-[#0a0a0f]"
            style={{ minHeight: 0 }}
          >
            {/* terminal header */}
            <div className="flex h-7 items-center justify-between border-b border-zinc-800/80 px-2">
              <div className="flex items-center gap-1.5">
                <span className="h-1.5 w-1.5 rounded-full bg-emerald-500 animate-pulse" />
                <span className="font-mono text-[9px] font-semibold uppercase tracking-widest text-zinc-500">
                  terminal
                </span>
                {thinking && (
                  <span className="font-mono text-[9px] text-zinc-600 animate-pulse">— running</span>
                )}
                {taskProgress && taskProgress.total > 0 && (
                  <span className="font-mono text-[9px] text-zinc-600">
                    · tasks {taskProgress.completed}/{taskProgress.total}
                  </span>
                )}
              </div>
              <div className="flex items-center gap-1">
                <span className="font-mono text-[9px] text-zinc-700">{telemetryLog.length} lines</span>
                <button
                  type="button"
                  onClick={() => setTerminalOpen((v) => !v)}
                  className="ml-1 grid h-4 w-4 place-items-center rounded text-zinc-600 hover:text-zinc-400"
                  aria-label={terminalOpen ? "Свернуть" : "Развернуть"}
                >
                  <ChevronDown className={`h-3 w-3 transition-transform ${terminalOpen ? "" : "-rotate-90"}`} />
                </button>
              </div>
            </div>
            {/* log lines */}
            {terminalOpen && (
              <div className="h-[148px] overflow-y-auto px-2 py-1 scrollbar-thin scrollbar-thumb-zinc-800">
                {terminalLines.map((line, i) => {
                  const tsMatch = line.match(/^\[(\d{2}:\d{2}:\d{2}\.\d+)\]\s*/);
                  const ts = tsMatch ? tsMatch[1] : null;
                  const body = ts ? line.slice(tsMatch![0].length) : line;
                  return (
                    <div key={i} className="flex gap-1 leading-4">
                      {ts && (
                        <span className="shrink-0 font-mono text-[10px] text-zinc-700 select-none">{ts}</span>
                      )}
                      <span className={`font-mono text-[10px] break-all ${terminalLineColor(body)}`}>{body}</span>
                    </div>
                  );
                })}
                {thinking && terminalLines.length === 0 && (
                  <span className="font-mono text-[10px] text-zinc-600 animate-pulse">waiting for telemetry…</span>
                )}
                <div ref={terminalEndRef} />
              </div>
            )}
          </motion.div>
        )}
      </AnimatePresence>

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
            {groups.map((g, gi) =>
              g.type === "thinking" ? (
                <motion.div
                  key={`thinking-${gi}`}
                  initial={{ opacity: 0, y: 6 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0 }}
                >
                  <ThinkingBlock
                    thoughts={g.thoughts}
                    durationSec={g.durationSec}
                    active={g.active}
                  />
                </motion.div>
              ) : (
                <motion.div
                  key={g.message.id}
                  initial={{ opacity: 0, y: 8 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0 }}
                >
                  {g.message.kind === "thought_duration" ? (
                    <ThoughtDurationRow durationSec={g.message.durationSec ?? 0} timestamp={g.message.timestamp} />
                  ) : g.message.kind === "action_log" ? (
                    <ActionLogRow
                      summary={g.message.content}
                      actionType={g.message.actionType ?? ""}
                      details={g.message.actionDetails ?? []}
                      timestamp={g.message.timestamp}
                    />
                  ) : g.message.kind === "code_diff" ? (
                    <CodeDiffRow
                      filePath={g.message.filePath ?? g.message.content}
                      diffHunk={g.message.diffHunk ?? ""}
                      additions={g.message.additions ?? 0}
                      deletions={g.message.deletions ?? 0}
                      timestamp={g.message.timestamp}
                    />
                  ) : g.message.kind === "postmortem" ? (
                    <div className="rounded-lg border border-border/60 bg-elevated/40 p-3">
                      <div className="mb-2 flex items-center justify-between">
                        <span className="text-xs font-semibold text-foreground">📋 Отчёт о генерации</span>
                        <CopyButton text={g.message.content} />
                      </div>
                      <pre className="whitespace-pre-wrap break-words font-mono text-[11px] leading-relaxed text-foreground/80 [overflow-wrap:anywhere]">{g.message.content}</pre>
                    </div>
                  ) : (
                    <div className="flex gap-2.5">
                      <div className={`mt-0.5 grid h-7 w-7 shrink-0 place-items-center rounded-md ${
                        g.message.role === "user" ? "bg-elevated" : "bg-gradient-primary shadow-glow"
                      }`}>
                        {g.message.role === "user"
                          ? <User className="h-3.5 w-3.5" />
                          : <Bot className="h-3.5 w-3.5 text-primary-foreground" />}
                      </div>
                      <div className="min-w-0 flex-1">
                        <div className="mb-0.5 flex items-baseline gap-2">
                          <span className="text-xs font-medium">{g.message.role === "user" ? "Вы" : "Исток"}</span>
                          <span className="font-mono text-[10px] text-muted-foreground">{fmtTime(g.message.timestamp)}</span>
                        </div>
                        <p className="whitespace-pre-wrap break-words [overflow-wrap:anywhere] text-sm leading-relaxed text-foreground/90">{g.message.content}</p>
                      </div>
                    </div>
                  )}
                </motion.div>
              )
            )}
          </AnimatePresence>
          <div ref={endRef} />
        </div>
      </ScrollArea>

      <div className="border-t border-border/60 p-3">
        {isBuild && (
          <div className="mb-2 flex items-center justify-between gap-3 rounded-xl border border-zinc-800 bg-zinc-900/40 px-3 py-2.5 transition-all duration-300 hover:border-zinc-700">
            <div className="flex min-w-0 items-start gap-2.5">
              <Clapperboard className="mt-0.5 h-4 w-4 shrink-0 text-emerald-400" />
              <div className="min-w-0">
                <label
                  htmlFor="video-toggle"
                  className="block cursor-pointer text-xs font-medium tracking-tight text-zinc-100"
                >
                  Генерация промо-ролика (Veo-3)
                </label>
                <p className="mt-0.5 text-[10px] leading-snug text-zinc-400">
                  Отключите для быстрого прототипирования. Включите для полного цикла (увеличивает время).
                </p>
              </div>
            </div>
            <Switch
              id="video-toggle"
              checked={videoEnabled}
              onCheckedChange={onVideoEnabledChange}
              aria-label="Генерация промо-ролика"
              className="shrink-0 transition-colors duration-300 data-[state=checked]:bg-emerald-500 data-[state=unchecked]:bg-zinc-800"
            />
          </div>
        )}
        <div className="rounded-xl border border-border bg-elevated p-2 focus-within:border-primary/60 focus-within:shadow-glow">
          <Textarea
            value={input}
            onChange={(e) => onInputChange(e.target.value)}
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
              <button
                type="button"
                onClick={() => toast("Вложения скоро будут доступны")}
                className="grid h-7 w-7 place-items-center rounded-md border border-border/70 bg-elevated/60 text-muted-foreground transition-all hover:border-primary/40 hover:text-primary"
                aria-label="Прикрепить"
              >
                <Paperclip className="h-3.5 w-3.5" />
              </button>
              <button
                type="button"
                onClick={() => onEditModeChange?.(!editMode)}
                className={`group flex h-7 items-center gap-1.5 rounded-md border px-2 text-xs transition-all ${
                  editMode
                    ? "border-primary/60 bg-primary/10 text-primary shadow-glow"
                    : "border-border/70 bg-elevated/60 text-muted-foreground hover:border-primary/40 hover:text-foreground"
                }`}
              >
                <Zap className={`h-3.5 w-3.5 ${editMode ? "fill-primary" : ""}`} />
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
              <div className="flex h-7 items-center overflow-hidden rounded-md border border-border/70 bg-elevated/60">
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <button
                      type="button"
                      className="flex h-full items-center gap-1 px-1.5 text-xs text-muted-foreground hover:text-foreground"
                    >
                      {isBuild ? (
                        <Hammer className="h-3.5 w-3.5 text-primary" />
                      ) : (
                        <MessageSquare className="h-3.5 w-3.5 text-primary" />
                      )}
                      <span className="font-medium">{isBuild ? "Сборка" : "Чат"}</span>
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
                        onModeChange("agent");
                      }}
                      className="flex cursor-pointer items-start gap-2 rounded-md px-2 py-1.5 text-xs"
                    >
                      <Hammer className="mt-0.5 h-3.5 w-3.5 text-primary" />
                      <div>
                        <div className="font-medium">Сборка</div>
                        <div className="text-[10px] text-muted-foreground">Спланировать, собрать и запустить</div>
                      </div>
                    </DropdownMenuItem>
                    <DropdownMenuItem
                      onSelect={(e) => {
                        e.preventDefault();
                        onModeChange("code");
                      }}
                      className="flex cursor-pointer items-start gap-2 rounded-md px-2 py-1.5 text-xs"
                    >
                      <MessageSquare className="mt-0.5 h-3.5 w-3.5 text-primary" />
                      <div>
                        <div className="font-medium">Чат</div>
                        <div className="text-[10px] text-muted-foreground">Обсудить без изменений</div>
                      </div>
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
                <Button
                  size="sm"
                  onClick={send}
                  disabled={!input.trim() || thinking}
                  className="h-7 gap-1 rounded-none rounded-r-md border-l border-border/70 bg-gradient-primary px-2.5 text-xs text-primary-foreground hover:opacity-90"
                >
                  <ArrowUp className="h-3.5 w-3.5" />
                </Button>
              </div>
            </div>
          </div>
        </div>
        <p className="mt-2 px-1 font-mono text-[10px] text-muted-foreground">
          Shift+Enter — новая строка · Enter — отправить
        </p>
      </div>
    </div>
  );
}
