import { useEffect, useState, useRef, useMemo } from "react";
import { motion, AnimatePresence } from "framer-motion";

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — GenerationMagic v2
//  Neural Canvas: blueprint grid, agent thoughts, streaming
//  file terminal, tier progress. All CSS GPU-accelerated.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

interface StreamedFileEntry {
  name: string;
  size: number;
  receivedAt: Date;
}

interface MilestoneEntry {
  agent: string;
  message: string;
  progress: number;
  status: string;
}

interface GenerationMagicProps {
  logs?: string[];
  progress?: number;
  streamedFiles?: StreamedFileEntry[];
  milestones?: MilestoneEntry[];
  currentFSMState?: string;
}

// ── Agent display metadata ─────────────────────────
const AGENT_META: Record<string, { icon: string; color: string }> = {
  director:    { icon: "🎯", color: "text-violet-400" },
  researcher:  { icon: "🔍", color: "text-blue-400" },
  brain:       { icon: "🧠", color: "text-indigo-400" },
  architect:   { icon: "📐", color: "text-cyan-400" },
  planner:     { icon: "📋", color: "text-sky-400" },
  coder:       { icon: "💻", color: "text-emerald-400" },
  designer:    { icon: "🎨", color: "text-pink-400" },
  validator:   { icon: "✅", color: "text-green-400" },
  security:    { icon: "🛡️", color: "text-amber-400" },
  tester:      { icon: "🧪", color: "text-orange-400" },
  ui_reviewer: { icon: "👁️", color: "text-rose-400" },
  videographer:{ icon: "🎬", color: "text-fuchsia-400" },
};

const DEFAULT_AGENT = { icon: "⚡", color: "text-slate-400" };

// ── Tier detection from logs ───────────────────────
function detectTier(logs: string[]): { current: number; total: number } {
  for (let i = logs.length - 1; i >= 0; i--) {
    const m = logs[i].match(/Tier\s+(\d+)\s*\/\s*(\d+)/i);
    if (m) return { current: Number(m[1]), total: Number(m[2]) };
    const m2 = logs[i].match(/\[T(\d+)\]/);
    if (m2) return { current: Number(m2[1]) + 1, total: 6 };
  }
  return { current: 0, total: 6 };
}

// ── Pipeline stage from FSM state ──────────────────
type PipelineStage = "init" | "research" | "planning" | "coding" | "design" | "verification" | "done";

function fsmToStage(state: string): PipelineStage {
  const s = state.toLowerCase();
  if (/creat|init|idle/i.test(s)) return "init";
  if (/research/i.test(s)) return "research";
  if (/plan|architect|brain/i.test(s)) return "planning";
  if (/cod|generat|build/i.test(s)) return "coding";
  if (/design|visual|asset/i.test(s)) return "design";
  if (/verif|secur|test|review|valid/i.test(s)) return "verification";
  if (/complet|done|finish/i.test(s)) return "done";
  return "coding";
}

const STAGE_LABELS: Record<PipelineStage, { label: string; color: string }> = {
  init:         { label: "Initializing pipeline…", color: "text-slate-400" },
  research:     { label: "Researching domain…", color: "text-blue-400" },
  planning:     { label: "Designing architecture…", color: "text-indigo-400" },
  coding:       { label: "Generating code…", color: "text-emerald-400" },
  design:       { label: "Creating visual assets…", color: "text-pink-400" },
  verification: { label: "Verification gate…", color: "text-amber-400" },
  done:         { label: "Generation complete", color: "text-green-400" },
};

// File extension → color
function extColor(name: string): string {
  if (/\.tsx?$/.test(name)) return "text-blue-400";
  if (/\.jsx?$/.test(name)) return "text-yellow-400";
  if (/\.css$/.test(name)) return "text-pink-400";
  if (/\.json$/.test(name)) return "text-amber-400";
  if (/\.html?$/.test(name)) return "text-orange-400";
  if (/\.svg$/.test(name)) return "text-emerald-400";
  return "text-slate-400";
}

function formatBytes(b: number): string {
  if (b < 1024) return `${b}B`;
  return `${(b / 1024).toFixed(1)}KB`;
}

// ── Phantom wireframe blocks ───────────────────────
const PHANTOM_BLOCKS = [
  { x: "6%", y: "8%", w: "88%", h: "7%", label: "Header", delay: 0 },
  { x: "6%", y: "19%", w: "55%", h: "28%", label: "Hero", delay: 0.3 },
  { x: "65%", y: "19%", w: "29%", h: "13%", label: "CTA", delay: 0.5 },
  { x: "65%", y: "35%", w: "29%", h: "12%", label: "Stats", delay: 0.7 },
  { x: "6%", y: "52%", w: "28%", h: "18%", label: "Feature 1", delay: 0.9 },
  { x: "37%", y: "52%", w: "28%", h: "18%", label: "Feature 2", delay: 1.1 },
  { x: "68%", y: "52%", w: "26%", h: "18%", label: "Feature 3", delay: 1.3 },
  { x: "6%", y: "75%", w: "58%", h: "10%", label: "Grid", delay: 1.5 },
  { x: "68%", y: "75%", w: "26%", h: "10%", label: "Sidebar", delay: 1.7 },
  { x: "6%", y: "89%", w: "88%", h: "6%", label: "Footer", delay: 1.9 },
];

export default function GenerationMagic({
  logs = [],
  progress = 0,
  streamedFiles = [],
  milestones = [],
  currentFSMState = "idle",
}: GenerationMagicProps) {
  const fileLogRef = useRef<HTMLDivElement>(null);
  const [activeBlockIdx, setActiveBlockIdx] = useState(0);

  // Derived state
  const stage = useMemo(() => fsmToStage(currentFSMState), [currentFSMState]);
  const stageLabel = STAGE_LABELS[stage];
  const tier = useMemo(() => detectTier(logs), [logs]);
  const glowIntensity = useMemo(() => Math.max(0.1, progress / 100), [progress]);

  // Active milestones (last 4 running/completed)
  const activeMilestones = useMemo(
    () => milestones.filter((m) => m.status === "running" || m.status === "completed").slice(-4),
    [milestones],
  );

  // Last 12 streamed files for terminal
  const recentFiles = useMemo(() => streamedFiles.slice(-12), [streamedFiles]);

  // Auto-scroll file log
  useEffect(() => {
    fileLogRef.current?.scrollIntoView({ behavior: "smooth", block: "end" });
  }, [recentFiles]);

  // Cycle phantom block highlight
  useEffect(() => {
    const interval = setInterval(() => {
      setActiveBlockIdx((prev) => (prev + 1) % PHANTOM_BLOCKS.length);
    }, 2000);
    return () => clearInterval(interval);
  }, []);

  return (
    <div className="relative w-full h-full overflow-hidden bg-[#07070b]">
      {/* ══ Layer 0: Blueprint Grid (CSS only, GPU) ══════════════ */}
      <div
        className="absolute inset-0 will-change-transform"
        style={{
          backgroundImage:
            "linear-gradient(hsl(243 76% 58% / 0.06) 1px, transparent 1px), linear-gradient(90deg, hsl(243 76% 58% / 0.06) 1px, transparent 1px)",
          backgroundSize: "48px 48px",
        }}
      />
      <div
        className="absolute inset-0 opacity-30 will-change-transform"
        style={{
          backgroundImage:
            "linear-gradient(hsl(243 76% 58% / 0.03) 1px, transparent 1px), linear-gradient(90deg, hsl(243 76% 58% / 0.03) 1px, transparent 1px)",
          backgroundSize: "12px 12px",
        }}
      />

      {/* ══ Layer 1: Radial glow (intensity scales with progress) ══ */}
      <div
        className="absolute inset-0 pointer-events-none will-change-[opacity]"
        style={{
          background: `radial-gradient(ellipse 60% 50% at 50% 45%, hsla(243, 76%, 58%, ${glowIntensity * 0.12}) 0%, transparent 70%)`,
        }}
      />

      {/* ══ Layer 2: Phantom wireframe blocks ════════════════════ */}
      <div className="absolute inset-0 pointer-events-none" style={{ zIndex: 2 }}>
        {PHANTOM_BLOCKS.map((block, i) => (
          <motion.div
            key={i}
            className="absolute rounded-sm"
            style={{
              left: block.x,
              top: block.y,
              width: block.w,
              height: block.h,
              border: `1px solid hsla(243, 76%, 58%, ${activeBlockIdx === i ? 0.25 : 0.06})`,
              willChange: "opacity, border-color",
            }}
            animate={{
              opacity: activeBlockIdx === i ? [0.15, 0.3, 0.15] : 0.06,
            }}
            transition={{
              duration: activeBlockIdx === i ? 1.8 : 3,
              repeat: Infinity,
              ease: "easeInOut",
            }}
          >
            <div className={`w-full h-full rounded-sm transition-colors duration-700 ${activeBlockIdx === i ? "bg-indigo-500/[0.04]" : "bg-transparent"}`} />
            <span className="absolute top-0.5 left-1.5 text-[7px] font-mono text-indigo-400/25 select-none">
              {block.label}
            </span>
            {activeBlockIdx === i && (
              <motion.div
                className="absolute left-0 w-full h-px bg-gradient-to-r from-transparent via-indigo-500/30 to-transparent"
                animate={{ top: ["0%", "100%"] }}
                transition={{ duration: 1.2, repeat: Infinity, ease: "linear" }}
              />
            )}
          </motion.div>
        ))}
      </div>

      {/* ══ Layer 3: Scanning beams ══════════════════════════════ */}
      <motion.div
        className="absolute left-0 right-0 h-px pointer-events-none will-change-transform"
        style={{
          zIndex: 3,
          background: "linear-gradient(90deg, transparent, hsla(243,76%,58%,0.2) 30%, hsla(243,76%,58%,0.35) 50%, hsla(243,76%,58%,0.2) 70%, transparent)",
        }}
        animate={{ top: ["5%", "95%", "5%"] }}
        transition={{ duration: 7, repeat: Infinity, ease: "easeInOut" }}
      />
      <motion.div
        className="absolute top-0 bottom-0 w-px pointer-events-none will-change-transform"
        style={{
          zIndex: 3,
          background: "linear-gradient(180deg, transparent, hsla(220,80%,60%,0.12) 30%, hsla(220,80%,60%,0.18) 50%, hsla(220,80%,60%,0.12) 70%, transparent)",
        }}
        animate={{ left: ["10%", "90%", "10%"] }}
        transition={{ duration: 11, repeat: Infinity, ease: "easeInOut" }}
      />

      {/* ══ Layer 10: Center Hub ═════════════════════════════════ */}
      <div
        className="absolute left-1/2 top-[38%] -translate-x-1/2 -translate-y-1/2 flex flex-col items-center gap-3"
        style={{ zIndex: 10 }}
      >
        {/* Pulsing rings */}
        <div className="relative">
          <motion.div
            className="absolute -inset-8 sm:-inset-12 rounded-full border border-indigo-500/10"
            animate={{ scale: [1, 1.15, 1], opacity: [0.15, 0.03, 0.15] }}
            transition={{ duration: 3.5, repeat: Infinity }}
          />
          <motion.div
            className="absolute -inset-16 sm:-inset-24 rounded-full border border-indigo-500/5"
            animate={{ scale: [1, 1.08, 1], opacity: [0.08, 0.01, 0.08] }}
            transition={{ duration: 5, repeat: Infinity, delay: 0.5 }}
          />
          {/* Core logo */}
          <motion.div
            className="w-14 h-14 sm:w-16 sm:h-16 rounded-2xl bg-indigo-500/10 border border-indigo-500/25 flex items-center justify-center backdrop-blur-md relative overflow-hidden"
            animate={{
              boxShadow: [
                "0 0 20px hsla(243,76%,58%,0.1), 0 0 60px hsla(243,76%,58%,0.03)",
                "0 0 40px hsla(243,76%,58%,0.2), 0 0 80px hsla(243,76%,58%,0.06)",
                "0 0 20px hsla(243,76%,58%,0.1), 0 0 60px hsla(243,76%,58%,0.03)",
              ],
            }}
            transition={{ duration: 2.5, repeat: Infinity }}
          >
            <motion.div
              className="absolute inset-0 bg-gradient-to-br from-indigo-500/15 via-transparent to-violet-500/10"
              animate={{ opacity: [0.3, 0.8, 0.3] }}
              transition={{ duration: 3, repeat: Infinity }}
            />
            <span className="relative text-indigo-400 font-bold text-lg tracking-tight select-none">IC</span>
          </motion.div>
        </div>

        <motion.p
          className="text-indigo-400/50 text-[10px] font-semibold tracking-[0.3em] uppercase select-none"
          animate={{ opacity: [0.35, 0.7, 0.35] }}
          transition={{ duration: 2.5, repeat: Infinity }}
        >
          Istok Core
        </motion.p>

        {/* Stage label */}
        <motion.div
          key={stage}
          initial={{ opacity: 0, y: 6 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.35 }}
          className={`flex items-center gap-2 px-4 py-1.5 rounded-full border border-white/5 bg-white/[0.02] backdrop-blur-sm ${stageLabel.color}`}
        >
          <span className="relative flex h-2 w-2">
            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-current opacity-50" />
            <span className="relative inline-flex rounded-full h-2 w-2 bg-current opacity-80" />
          </span>
          <span className="text-[11px] font-medium select-none">{stageLabel.label}</span>
        </motion.div>

        {/* Tier progress */}
        {tier.current > 0 && (
          <div className="flex flex-col items-center gap-1.5 mt-1">
            <div className="flex items-center gap-1">
              {Array.from({ length: tier.total }, (_, i) => (
                <motion.div
                  key={i}
                  className={`h-1.5 rounded-full transition-all duration-500 ${
                    i < tier.current
                      ? "bg-indigo-500 w-6"
                      : i === tier.current
                        ? "bg-indigo-500/40 w-4"
                        : "bg-slate-700 w-3"
                  }`}
                  initial={i === tier.current - 1 ? { scale: 0.5 } : undefined}
                  animate={i === tier.current - 1 ? { scale: 1 } : undefined}
                  transition={{ duration: 0.3 }}
                />
              ))}
            </div>
            <span className="text-[9px] font-mono text-indigo-400/50 select-none tabular-nums">
              Tier {tier.current}/{tier.total}
            </span>
          </div>
        )}

        {/* Main progress bar */}
        {progress > 0 && (
          <div className="w-40 sm:w-52 flex flex-col items-center gap-1 mt-1">
            <div className="w-full h-[3px] rounded-full bg-white/5 overflow-hidden">
              <motion.div
                className="h-full rounded-full bg-gradient-to-r from-indigo-600 via-violet-500 to-indigo-600"
                style={{ backgroundSize: "200% 100%" }}
                initial={{ width: "0%" }}
                animate={{ width: `${progress}%`, backgroundPosition: ["0% 0%", "100% 0%"] }}
                transition={{
                  width: { duration: 0.6, ease: "easeOut" },
                  backgroundPosition: { duration: 2, repeat: Infinity, ease: "linear" },
                }}
              />
            </div>
            <span className="text-[10px] text-indigo-400/40 font-mono tabular-nums select-none">
              {progress}% · {streamedFiles.length} files
            </span>
          </div>
        )}
      </div>

      {/* ══ Layer 12: Agent Thoughts (left column) ═══════════════ */}
      <div
        className="absolute left-3 sm:left-5 top-16 bottom-16 w-56 sm:w-64 overflow-hidden pointer-events-none hidden md:flex flex-col gap-1.5"
        style={{ zIndex: 12 }}
      >
        <div className="text-[9px] font-mono uppercase tracking-widest text-indigo-400/30 mb-1">
          Agent Activity
        </div>
        <AnimatePresence mode="popLayout">
          {activeMilestones.map((m, i) => {
            const agentKey = m.agent.toLowerCase().replace(/\s+/g, "_");
            const meta = AGENT_META[agentKey] || DEFAULT_AGENT;
            const isRunning = m.status === "running";
            return (
              <motion.div
                key={`${m.agent}-${i}`}
                initial={{ opacity: 0, x: -12 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: -8 }}
                transition={{ duration: 0.3 }}
                className={`flex items-start gap-2 px-2.5 py-2 rounded-lg border backdrop-blur-sm ${
                  isRunning
                    ? "border-indigo-500/15 bg-indigo-500/[0.04]"
                    : "border-white/5 bg-white/[0.02]"
                }`}
              >
                <span className="text-sm shrink-0 mt-0.5">{meta.icon}</span>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-1.5">
                    <span className={`text-[10px] font-bold capitalize ${meta.color}`}>{m.agent}</span>
                    {isRunning && (
                      <span className="flex h-1.5 w-1.5">
                        <span className="animate-ping absolute h-1.5 w-1.5 rounded-full bg-indigo-400 opacity-50" />
                        <span className="relative rounded-full h-1.5 w-1.5 bg-indigo-400" />
                      </span>
                    )}
                    {m.status === "completed" && (
                      <span className="text-[8px] text-emerald-400/70">✓</span>
                    )}
                  </div>
                  <p className="text-[9px] text-slate-400/70 truncate leading-tight mt-0.5">
                    {m.message.replace(/^[^\w\u0400-\u04FF]*/, "").slice(0, 60)}
                  </p>
                </div>
              </motion.div>
            );
          })}
        </AnimatePresence>
      </div>

      {/* ══ Layer 13: Streaming File Terminal (right column) ═════ */}
      <div
        className="absolute right-3 sm:right-5 top-14 bottom-14 w-56 sm:w-72 lg:w-80 overflow-hidden pointer-events-none hidden sm:block"
        style={{ zIndex: 13 }}
      >
        {/* Terminal header */}
        <div className="flex items-center gap-2 px-3 py-1.5 rounded-t-lg border border-white/5 border-b-0 bg-white/[0.02] backdrop-blur-sm">
          <div className="flex gap-1">
            <div className="w-1.5 h-1.5 rounded-full bg-red-500/40" />
            <div className="w-1.5 h-1.5 rounded-full bg-yellow-500/40" />
            <div className="w-1.5 h-1.5 rounded-full bg-green-500/40" />
          </div>
          <span className="text-[8px] font-mono text-slate-500 uppercase tracking-wider">
            file stream
          </span>
          <span className="ml-auto text-[8px] font-mono text-indigo-400/40 tabular-nums">
            {streamedFiles.length} received
          </span>
        </div>
        {/* Terminal body */}
        <div className="border border-white/5 border-t-0 rounded-b-lg bg-[#0a0a12]/80 backdrop-blur-md p-2 max-h-[calc(100%-28px)] overflow-hidden">
          <div className="flex flex-col gap-px overflow-y-auto max-h-full scrollbar-none">
            <AnimatePresence mode="popLayout">
              {recentFiles.map((file, i) => (
                <motion.div
                  key={`${file.name}-${i}`}
                  initial={{ opacity: 0, x: 10, height: 0 }}
                  animate={{ opacity: 1, x: 0, height: "auto" }}
                  exit={{ opacity: 0 }}
                  transition={{ duration: 0.25 }}
                  className="flex items-center gap-1.5 py-0.5 font-mono"
                >
                  <span className="text-emerald-500/60 text-[9px]">✓</span>
                  <span className={`text-[9px] truncate flex-1 ${extColor(file.name)}`}>
                    {file.name}
                  </span>
                  <span className="text-[8px] text-slate-500/50 tabular-nums shrink-0">
                    {formatBytes(file.size)}
                  </span>
                </motion.div>
              ))}
            </AnimatePresence>
            <div ref={fileLogRef} />
            {streamedFiles.length === 0 && (
              <div className="flex items-center gap-2 py-3 justify-center">
                <motion.div
                  className="w-1 h-1 rounded-full bg-indigo-400/40"
                  animate={{ opacity: [0.2, 0.8, 0.2] }}
                  transition={{ duration: 1.2, repeat: Infinity }}
                />
                <span className="text-[9px] text-slate-500/40 font-mono">waiting for files…</span>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* ══ Layer 15: Bottom status bar ══════════════════════════ */}
      <div
        className="absolute bottom-0 left-0 right-0 pointer-events-none"
        style={{ zIndex: 15 }}
      >
        <div className="absolute inset-0 bg-gradient-to-t from-[#07070b] via-[#07070b]/60 to-transparent" />
        <div className="relative flex items-center justify-between px-4 sm:px-6 py-3">
          {/* Latest log message */}
          <div className="flex items-center gap-2 min-w-0 flex-1">
            <span className="relative flex h-2 w-2 shrink-0">
              <span className="animate-ping absolute h-full w-full rounded-full bg-indigo-400 opacity-40" />
              <span className="relative rounded-full h-2 w-2 bg-indigo-400/80" />
            </span>
            <AnimatePresence mode="wait">
              {logs.length > 0 && (
                <motion.span
                  key={logs[logs.length - 1]?.slice(0, 30)}
                  initial={{ opacity: 0, y: 4 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, y: -4 }}
                  className="text-[10px] sm:text-[11px] font-mono text-slate-300/60 truncate"
                >
                  {logs[logs.length - 1]?.replace(/^[^\w\u0400-\u04FF]*/, "").slice(0, 80)}
                </motion.span>
              )}
            </AnimatePresence>
          </div>
          {/* Right: file count + FSM state */}
          <div className="flex items-center gap-3 shrink-0 ml-4">
            <span className="text-[9px] font-mono text-slate-500/50 tabular-nums">
              {streamedFiles.length} files
            </span>
            <span className="text-[9px] font-mono text-indigo-400/40 uppercase">
              {currentFSMState}
            </span>
          </div>
        </div>
      </div>
    </div>
  );
}
