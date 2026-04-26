import { motion } from "framer-motion";
import {
  CheckCircle2,
  Circle,
  Loader2,
  XCircle,
  Compass,
  Search,
  Brain,
  Layers,
  ListTree,
  Code2,
  Palette,
  ShieldCheck,
  Bug,
  Sparkles,
  Film,
  type LucideIcon,
} from "lucide-react";
import type { AgentPipelineId, AgentMilestone } from "@/hooks/useGeneration";

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — AgentPulseTimeline
//  Вертикальный список из 10 агентов. Пульсирует в реальном
//  времени, когда backend присылает event.Agent == "Coder" / "Security" / др.
//  Security ✅ зелёная галочка при прохождении Verification Gate.
//
//  Контракт: activeAgent + milestones из useGeneration (SSE).
//  Связь с бэкендом: generate_handler_sse.go → domain.EventBus.PublishStatus.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

interface AgentMeta {
  id: AgentPipelineId;
  label: string;
  role: string;
  icon: LucideIcon;
  accent: string; // tailwind color token
}

/** 10 каноничных агентов Workspace v3.0. architect/planner сведены в "Architect",
 *  чтобы уложиться в ровно 10 строк по ТЗ. */
const AGENTS_10: readonly AgentMeta[] = [
  { id: "director",    label: "Director",     role: "Оркестратор",        icon: Compass,     accent: "text-violet-400"    },
  { id: "researcher",  label: "Researcher",   role: "Анализ",             icon: Search,      accent: "text-amber-400"     },
  { id: "brain",       label: "Brain",        role: "Стратегия",          icon: Brain,       accent: "text-fuchsia-400"   },
  { id: "architect",   label: "Architect",    role: "Проектирование",     icon: Layers,      accent: "text-indigo-400"    },
  { id: "planner",     label: "Planner",      role: "DAG-план",           icon: ListTree,    accent: "text-sky-400"       },
  { id: "coder",       label: "Coder",        role: "Генерация кода",     icon: Code2,       accent: "text-emerald-400"   },
  { id: "designer",    label: "Designer",     role: "UI-ассеты",          icon: Palette,     accent: "text-pink-400"      },
  { id: "security",    label: "Security",     role: "Аудит безопасности", icon: ShieldCheck, accent: "text-emerald-400"   },
  { id: "tester",      label: "Tester",       role: "Прогон тестов",      icon: Bug,         accent: "text-orange-400"    },
  { id: "ui_reviewer", label: "UI Reviewer",  role: "UX / a11y",          icon: Sparkles,    accent: "text-teal-400"      },
] as const;

export interface AgentPulseTimelineProps {
  /** Агент, который прямо сейчас бежит (running) на бэке. */
  activeAgent: AgentPipelineId | null;
  /** Полный набор milestone-событий из SSE. */
  milestones: AgentMilestone[];
  /** Security Agent пропустил Verification Gate (зелёная галочка). */
  securityApproved: boolean;
  /** Tester Agent прошёл. */
  testerApproved?: boolean;
  /** UI Reviewer Agent прошёл. */
  uiReviewerApproved?: boolean;
}

type AgentRowState = "idle" | "running" | "completed" | "error";

/** Определяет состояние строки для конкретного агента по milestones + activeAgent. */
function resolveRowState(
  meta: AgentMeta,
  milestones: AgentMilestone[],
  activeAgent: AgentPipelineId | null,
): { state: AgentRowState; progress: number; message?: string } {
  if (activeAgent === meta.id) {
    const m = milestones.find((x) => normalizeAgent(x.agent) === meta.id);
    return { state: "running", progress: m?.progress ?? 0, message: m?.message };
  }
  const m = milestones.find((x) => normalizeAgent(x.agent) === meta.id);
  if (!m) return { state: "idle", progress: 0 };
  if (m.status === "completed") return { state: "completed", progress: 100, message: m.message };
  if (m.status === "error") return { state: "error", progress: m.progress ?? 0, message: m.message };
  return { state: "running", progress: m.progress ?? 0, message: m.message };
}

function normalizeAgent(raw: string): string {
  return (raw || "").toLowerCase().replace(/\s+/g, "_");
}

/** Приоритет зелёной галочки рядом с агентом верификации. */
function verificationCheckFor(
  meta: AgentMeta,
  securityApproved: boolean,
  testerApproved: boolean,
  uiReviewerApproved: boolean,
): boolean {
  if (meta.id === "security") return securityApproved;
  if (meta.id === "tester") return testerApproved;
  if (meta.id === "ui_reviewer") return uiReviewerApproved;
  return false;
}

const AgentPulseTimeline = ({
  activeAgent,
  milestones,
  securityApproved,
  testerApproved = false,
  uiReviewerApproved = false,
}: AgentPulseTimelineProps) => {
  return (
    <div className="space-y-1.5">
      <div className="flex items-center justify-between px-1">
        <h3 className="text-[10px] font-semibold uppercase tracking-[0.14em] text-muted-foreground/60">
          Agent Pulse
        </h3>
        <span className="text-[9px] text-muted-foreground/40">
          {activeAgent ? `▶ ${activeAgent}` : "idle"}
        </span>
      </div>

      <ol className="relative space-y-1">
        {/* Vertical connector line */}
        <div
          aria-hidden
          className="absolute left-[15px] top-3 bottom-3 w-px bg-gradient-to-b from-transparent via-border/40 to-transparent"
        />

        {AGENTS_10.map((meta, idx) => {
          const { state, progress, message } = resolveRowState(meta, milestones, activeAgent);
          const isPulsing = state === "running";
          const verified = verificationCheckFor(
            meta,
            securityApproved,
            testerApproved,
            uiReviewerApproved,
          );

          return (
            <AgentRow
              key={meta.id}
              meta={meta}
              state={state}
              progress={progress}
              message={message}
              isPulsing={isPulsing}
              verified={verified}
              index={idx}
            />
          );
        })}
      </ol>
    </div>
  );
};

// ─────────────────────────────────────────────────────────────────
//  Row
// ─────────────────────────────────────────────────────────────────

interface AgentRowProps {
  meta: AgentMeta;
  state: AgentRowState;
  progress: number;
  message?: string;
  isPulsing: boolean;
  verified: boolean;
  index: number;
}

const AgentRow = ({ meta, state, progress, message, isPulsing, verified, index }: AgentRowProps) => {
  const Icon = meta.icon;
  const dim = state === "idle" && !verified ? "opacity-55" : "opacity-100";

  return (
    <motion.li
      initial={{ opacity: 0, x: -6 }}
      animate={{ opacity: 1, x: 0 }}
      transition={{ duration: 0.28, delay: index * 0.03, ease: [0.22, 1, 0.36, 1] }}
      className={`relative flex items-center gap-2 pl-0.5 pr-2 py-1.5 rounded-lg transition-colors ${dim} ${
        isPulsing ? "bg-primary/5" : "hover:bg-secondary/20"
      }`}
    >
      {/* ── Icon + pulse halo ── */}
      <div className="relative w-8 h-8 shrink-0 flex items-center justify-center">
        {isPulsing && (
          <motion.span
            className="absolute inset-0 rounded-full bg-primary/25"
            initial={{ scale: 1, opacity: 0.6 }}
            animate={{ scale: 1.8, opacity: 0 }}
            transition={{ duration: 1.5, repeat: Infinity, ease: "easeOut" }}
          />
        )}
        <div
          className={`relative w-7 h-7 rounded-full flex items-center justify-center border ${
            state === "running"
              ? "bg-primary/15 border-primary/40"
              : state === "completed"
                ? "bg-emerald-500/15 border-emerald-500/40"
                : state === "error"
                  ? "bg-destructive/15 border-destructive/40"
                  : "bg-secondary/40 border-border/30"
          }`}
        >
          <Icon size={13} className={state === "idle" ? "text-muted-foreground/60" : meta.accent} />
        </div>
      </div>

      {/* ── Label + message ── */}
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-1.5">
          <p className="text-[11.5px] font-semibold text-foreground leading-tight">
            {meta.label}
          </p>
          {verified && (
            <motion.span
              initial={{ scale: 0, opacity: 0 }}
              animate={{ scale: 1, opacity: 1 }}
              transition={{ duration: 0.3, ease: [0.34, 1.56, 0.64, 1] }}
              title={`${meta.label} verified`}
            >
              <CheckCircle2 size={11} className="text-emerald-400" />
            </motion.span>
          )}
          {isPulsing && (
            <span className="px-1.5 py-[1px] rounded-full bg-primary/20 text-primary text-[8px] font-medium tracking-wider uppercase">
              live
            </span>
          )}
        </div>
        <p className="text-[9.5px] text-muted-foreground/55 truncate leading-tight">
          {state === "running" && message
            ? message
            : state === "completed"
              ? "✓ готово"
              : state === "error"
                ? "ошибка"
                : meta.role}
        </p>
      </div>

      {/* ── Trailing status ── */}
      <StatusBadge state={state} progress={progress} />
    </motion.li>
  );
};

const StatusBadge = ({ state, progress }: { state: AgentRowState; progress: number }) => {
  if (state === "completed") {
    return <CheckCircle2 size={14} className="text-emerald-400 shrink-0" />;
  }
  if (state === "error") {
    return <XCircle size={14} className="text-destructive shrink-0" />;
  }
  if (state === "running") {
    return (
      <div className="flex items-center gap-1 shrink-0">
        <Loader2 size={11} className="text-primary animate-spin" />
        {progress > 0 && (
          <span className="text-[9px] text-primary font-mono tabular-nums">{progress}%</span>
        )}
      </div>
    );
  }
  return <Circle size={11} className="text-muted-foreground/30 shrink-0" />;
};

export default AgentPulseTimeline;
