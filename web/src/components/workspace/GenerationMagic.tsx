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

// ── Terminator Data Feed constants (CSS-animated, zero JS overhead) ──
const HEX_CHARS = "0123456789ABCDEF";
const MATRIX_COLS = Array.from({ length: 8 }, (_, i) => ({
  chars: Array.from({ length: 30 }, () =>
    HEX_CHARS[Math.floor(Math.random() * 16)] +
    HEX_CHARS[Math.floor(Math.random() * 16)]
  ).join("\n"),
  left: `${i * 12.5}%`,
  duration: `${5 + (i % 3) * 2.5}s`,
  delay: `${i * 0.3}s`,
}));

const CODE_LINES = [
  "func (o *Orchestrator) Generate(ctx) {",
  "  fsm := domain.NewTaskFSM()",
  "  plan := o.planner.CreateDAG(spec)",
  "  for _, tier := range plan.Tiers {",
  "    code := o.coder.Chunk(tier)",
  "    o.events.PublishFile(code)",
  "  }",
  "  gate := NewVerificationGate()",
  "  gate.Verify(ctx, allFiles)",
  "}",
  "",
  "const App: FC = () => {",
  "  const { data } = useGeneration()",
  "  return (",
  "    <Layout sidebar={<AgentPanel />}>",
  "      <Hero content={data.hero} />",
  "      <Features items={data.feat} />",
  "      <Footer />",
  "    </Layout>",
  "  )",
  "}",
];

const WIRE_RECTS = [
  { x: 3, y: 3, w: 94, h: 10 },
  { x: 3, y: 17, w: 55, h: 30 },
  { x: 62, y: 17, w: 35, h: 14 },
  { x: 62, y: 35, w: 35, h: 12 },
  { x: 3, y: 51, w: 30, h: 22 },
  { x: 35, y: 51, w: 30, h: 22 },
  { x: 67, y: 51, w: 30, h: 22 },
  { x: 3, y: 78, w: 94, h: 19 },
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

  // Last 12 streamed files for terminal
  const recentFiles = useMemo(() => streamedFiles.slice(-12), [streamedFiles]);

  // Agent status map for Terminator feeds
  const agentFeeds = useMemo(() => {
    const map: Record<string, { active: boolean; progress: number; msg: string }> = {};
    for (const m of milestones) {
      const key = m.agent.toLowerCase().replace(/\s+/g, "_");
      map[key] = { active: m.status === "running", progress: m.progress, msg: m.message };
    }
    return map;
  }, [milestones]);
  const showPlannerFeed = stage === "init" || stage === "planning" || !!agentFeeds.planner || !!agentFeeds.director;
  const showCoderFeed = stage === "coding" || !!agentFeeds.coder;
  const showDesignerFeed = !!agentFeeds.designer || stage === "design";
  const showVideoFeed = !!agentFeeds.videographer;

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

      {/* ══ Layer 12: Terminator Data Feed (left column) ═════════ */}
      <div
        className="absolute left-3 sm:left-5 top-14 bottom-14 w-56 sm:w-64 overflow-hidden pointer-events-none hidden md:flex flex-col"
        style={{ zIndex: 12 }}
      >
        {/* CSS-only keyframes — GPU-accelerated, zero JS overhead */}
        <style>{`
          @keyframes tMatrixFall{0%{transform:translateY(-50%)}100%{transform:translateY(0%)}}
          @keyframes tCodeScroll{0%{transform:translateY(0)}100%{transform:translateY(-50%)}}
          @keyframes tScanBeam{0%{transform:translateY(-100%)}100%{transform:translateY(300%)}}
          @keyframes tWireDash{0%{stroke-dashoffset:200}100%{stroke-dashoffset:0}}
          @keyframes tFlicker{0%,100%{opacity:.7}33%{opacity:.4}66%{opacity:.9}}
        `}</style>

        <div className="relative flex-1 rounded-lg border border-cyan-500/10 bg-[#050a0e]/80 backdrop-blur-md overflow-hidden flex flex-col">
          {/* Micro scanning grid */}
          <div
            className="absolute inset-0 pointer-events-none opacity-[0.025]"
            style={{
              backgroundImage:
                "linear-gradient(rgba(0,255,200,1) 1px,transparent 1px),linear-gradient(90deg,rgba(0,255,200,1) 1px,transparent 1px)",
              backgroundSize: "6px 6px",
            }}
          />
          {/* Vertical scan beam */}
          <div
            className="absolute left-0 right-0 h-10 pointer-events-none will-change-transform"
            style={{
              background:
                "linear-gradient(180deg,transparent,rgba(0,255,200,0.04) 40%,rgba(0,255,200,0.07) 50%,rgba(0,255,200,0.04) 60%,transparent)",
              animation: "tScanBeam 3.5s linear infinite",
            }}
          />

          {/* Header */}
          <div className="relative px-3 py-2 border-b border-cyan-500/10 flex items-center gap-2 shrink-0">
            <div className="w-1.5 h-1.5 rounded-full bg-cyan-400 animate-pulse" />
            <span className="text-[8px] font-mono uppercase tracking-[0.2em] text-cyan-400/70">
              Neural Data Feed
            </span>
          </div>

          {/* Feed sections */}
          <div className="relative flex-1 overflow-y-auto scrollbar-none p-1.5 flex flex-col gap-1.5">
            <AnimatePresence mode="popLayout">
              {/* ── PLANNER: Hex Matrix Rain ── */}
              {showPlannerFeed && (
                <motion.div
                  key="feed-planner"
                  initial={{ opacity: 0, y: -8 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, height: 0 }}
                  transition={{ duration: 0.35 }}
                  className="rounded border border-cyan-500/8 bg-cyan-950/20 overflow-hidden shrink-0"
                >
                  <div className="flex items-center gap-1.5 px-2 py-1 border-b border-cyan-500/5">
                    <span className="text-[7px] font-mono font-bold text-cyan-400/80 tracking-wider">PLANNER</span>
                    {(agentFeeds.planner?.active || agentFeeds.director?.active) && (
                      <span className="w-1 h-1 rounded-full bg-cyan-400 animate-pulse" />
                    )}
                    <span className="ml-auto text-[7px] font-mono text-cyan-400/30 tabular-nums">
                      {agentFeeds.planner?.progress ?? agentFeeds.director?.progress ?? 0}%
                    </span>
                  </div>
                  <div className="relative h-14 overflow-hidden">
                    {MATRIX_COLS.map((col, i) => (
                      <div
                        key={i}
                        className="absolute top-0 font-mono text-[7px] leading-[1.1] text-cyan-400/40 whitespace-pre select-none"
                        style={{
                          left: col.left,
                          animation: `tMatrixFall ${col.duration} linear ${col.delay} infinite`,
                          willChange: "transform",
                        }}
                      >
                        {col.chars}
                      </div>
                    ))}
                    <div className="absolute bottom-0 left-0 right-0 h-4 bg-gradient-to-t from-[#050a0e] to-transparent" />
                  </div>
                </motion.div>
              )}

              {/* ── CODER: Scrolling Code Stream ── */}
              {showCoderFeed && (
                <motion.div
                  key="feed-coder"
                  initial={{ opacity: 0, y: -8 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, height: 0 }}
                  transition={{ duration: 0.35 }}
                  className="rounded border border-emerald-500/8 bg-emerald-950/20 overflow-hidden shrink-0"
                >
                  <div className="flex items-center gap-1.5 px-2 py-1 border-b border-emerald-500/5">
                    <span className="text-[7px] font-mono font-bold text-emerald-400/80 tracking-wider">CODER</span>
                    {agentFeeds.coder?.active && (
                      <span className="w-1 h-1 rounded-full bg-emerald-400 animate-pulse" />
                    )}
                    <span className="ml-auto text-[7px] font-mono text-emerald-400/30 tabular-nums">
                      {streamedFiles.length} files
                    </span>
                  </div>
                  <div className="relative h-[72px] overflow-hidden">
                    <div
                      className="font-mono text-[7px] leading-[1.3] whitespace-pre select-none px-2 pt-1"
                      style={{
                        animation: "tCodeScroll 15s linear infinite",
                        willChange: "transform",
                      }}
                    >
                      {CODE_LINES.map((line, i) => (
                        <div key={i} className={line.startsWith("  ") ? "text-emerald-400/35" : "text-emerald-300/50"}>
                          {line || "\u00A0"}
                        </div>
                      ))}
                      {CODE_LINES.map((line, i) => (
                        <div key={`dup-${i}`} className={line.startsWith("  ") ? "text-emerald-400/35" : "text-emerald-300/50"}>
                          {line || "\u00A0"}
                        </div>
                      ))}
                    </div>
                    {/* CRT scanline overlay */}
                    <div
                      className="absolute inset-0 pointer-events-none"
                      style={{
                        animation: "tFlicker 3s step-end infinite",
                        background: "linear-gradient(transparent 50%, rgba(0,255,120,0.012) 50%)",
                        backgroundSize: "100% 4px",
                      }}
                    />
                    <div className="absolute bottom-0 left-0 right-0 h-4 bg-gradient-to-t from-[#050a0e] to-transparent" />
                  </div>
                </motion.div>
              )}

              {/* ── DESIGNER: Wireframe + Circular Progress ── */}
              {showDesignerFeed && (
                <motion.div
                  key="feed-designer"
                  initial={{ opacity: 0, y: -8 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, height: 0 }}
                  transition={{ duration: 0.35 }}
                  className="rounded border border-pink-500/8 bg-pink-950/20 overflow-hidden shrink-0"
                >
                  <div className="flex items-center gap-1.5 px-2 py-1 border-b border-pink-500/5">
                    <span className="text-[7px] font-mono font-bold text-pink-400/80 tracking-wider">DESIGNER</span>
                    {agentFeeds.designer?.active && (
                      <span className="w-1 h-1 rounded-full bg-pink-400 animate-pulse" />
                    )}
                    <span className="ml-auto text-[7px] font-mono text-pink-400/30 tabular-nums">
                      {agentFeeds.designer?.progress ?? 0}%
                    </span>
                  </div>
                  <div className="relative h-16 p-1.5 flex items-center gap-2">
                    <svg viewBox="0 0 100 100" className="w-20 h-14 shrink-0">
                      {WIRE_RECTS.map((r, i) => (
                        <rect
                          key={i}
                          x={r.x} y={r.y} width={r.w} height={r.h}
                          fill="none" stroke="rgba(236,72,153,0.2)" strokeWidth="0.5" rx="1"
                          style={{ strokeDasharray: 200, animation: `tWireDash ${2 + i * 0.3}s ease-out ${i * 0.15}s forwards` }}
                        />
                      ))}
                    </svg>
                    <div className="relative w-10 h-10 shrink-0">
                      <svg viewBox="0 0 36 36" className="w-full h-full -rotate-90">
                        <circle cx="18" cy="18" r="15" fill="none" stroke="rgba(236,72,153,0.08)" strokeWidth="2" />
                        <circle
                          cx="18" cy="18" r="15" fill="none"
                          stroke="rgba(236,72,153,0.5)" strokeWidth="2"
                          strokeDasharray={`${(agentFeeds.designer?.progress ?? 0) * 0.94} 94`}
                          strokeLinecap="round" className="transition-all duration-700"
                        />
                      </svg>
                      <span className="absolute inset-0 flex items-center justify-center text-[7px] font-mono text-pink-400/60 tabular-nums">
                        {agentFeeds.designer?.progress ?? 0}%
                      </span>
                    </div>
                  </div>
                </motion.div>
              )}

              {/* ── VIDEOGRAPHER: Frame Render + Progress ── */}
              {showVideoFeed && (
                <motion.div
                  key="feed-video"
                  initial={{ opacity: 0, y: -8 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, height: 0 }}
                  transition={{ duration: 0.35 }}
                  className="rounded border border-fuchsia-500/8 bg-fuchsia-950/20 overflow-hidden shrink-0"
                >
                  <div className="flex items-center gap-1.5 px-2 py-1 border-b border-fuchsia-500/5">
                    <span className="text-[7px] font-mono font-bold text-fuchsia-400/80 tracking-wider">VIDEOGRAPHER</span>
                    {agentFeeds.videographer?.active && (
                      <span className="w-1 h-1 rounded-full bg-fuchsia-400 animate-pulse" />
                    )}
                  </div>
                  <div className="relative h-12 p-1.5 flex items-center gap-2">
                    <div className="flex-1 flex gap-0.5 overflow-hidden">
                      {Array.from({ length: 6 }, (_, i) => (
                        <motion.div
                          key={i}
                          className="w-6 h-8 rounded-sm border border-fuchsia-500/15 bg-fuchsia-500/[0.03] flex items-center justify-center"
                          animate={{ opacity: [0.3, 0.8, 0.3] }}
                          transition={{ duration: 1.5, delay: i * 0.2, repeat: Infinity }}
                        >
                          <span className="text-[6px] font-mono text-fuchsia-400/40">F{i + 1}</span>
                        </motion.div>
                      ))}
                    </div>
                    <div className="relative w-9 h-9 shrink-0">
                      <svg viewBox="0 0 36 36" className="w-full h-full -rotate-90">
                        <circle cx="18" cy="18" r="15" fill="none" stroke="rgba(192,38,211,0.08)" strokeWidth="2" />
                        <circle
                          cx="18" cy="18" r="15" fill="none"
                          stroke="rgba(192,38,211,0.5)" strokeWidth="2"
                          strokeDasharray={`${(agentFeeds.videographer?.progress ?? 0) * 0.94} 94`}
                          strokeLinecap="round" className="transition-all duration-700"
                        />
                      </svg>
                      <span className="absolute inset-0 flex items-center justify-center text-[7px] font-mono text-fuchsia-400/60 tabular-nums">
                        {agentFeeds.videographer?.progress ?? 0}%
                      </span>
                    </div>
                  </div>
                </motion.div>
              )}
            </AnimatePresence>
          </div>
        </div>
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
