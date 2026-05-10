import { useEffect, useState, useRef, useMemo } from "react";
import { motion, AnimatePresence } from "framer-motion";

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — GenerationMagic
//  Premium blueprint animation during project generation.
//  Grid, neural connections, phantom blocks, code stream.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

interface GenerationMagicProps {
  /** SSE status messages streamed from agents */
  logs?: string[];
  /** Current generation progress 0–100 */
  progress?: number;
}

// Phantom wireframe blocks — simulating page structure being "designed"
const PHANTOM_BLOCKS = [
  { x: "6%", y: "8%", w: "88%", h: "7%", label: "Header / Navigation", delay: 0 },
  { x: "6%", y: "19%", w: "55%", h: "28%", label: "Hero Section", delay: 0.3 },
  { x: "65%", y: "19%", w: "29%", h: "13%", label: "CTA Card", delay: 0.5 },
  { x: "65%", y: "35%", w: "29%", h: "12%", label: "Stats Widget", delay: 0.7 },
  { x: "6%", y: "52%", w: "28%", h: "18%", label: "Feature 1", delay: 0.9 },
  { x: "37%", y: "52%", w: "28%", h: "18%", label: "Feature 2", delay: 1.1 },
  { x: "68%", y: "52%", w: "26%", h: "18%", label: "Feature 3", delay: 1.3 },
  { x: "6%", y: "75%", w: "58%", h: "10%", label: "Content Grid", delay: 1.5 },
  { x: "68%", y: "75%", w: "26%", h: "10%", label: "Sidebar", delay: 1.7 },
  { x: "6%", y: "89%", w: "88%", h: "6%", label: "Footer", delay: 1.9 },
];

// Neural connection lines from center logo
const NEURAL_LINES = [
  { x1: "50%", y1: "50%", x2: "6%", y2: "12%" },
  { x1: "50%", y1: "50%", x2: "94%", y2: "8%" },
  { x1: "50%", y1: "50%", x2: "8%", y2: "88%" },
  { x1: "50%", y1: "50%", x2: "92%", y2: "92%" },
  { x1: "50%", y1: "50%", x2: "3%", y2: "50%" },
  { x1: "50%", y1: "50%", x2: "97%", y2: "50%" },
  { x1: "50%", y1: "50%", x2: "25%", y2: "4%" },
  { x1: "50%", y1: "50%", x2: "75%", y2: "96%" },
  { x1: "50%", y1: "50%", x2: "15%", y2: "65%" },
  { x1: "50%", y1: "50%", x2: "85%", y2: "35%" },
];

const CODE_SNIPPETS = [
  'import { createBrowserRouter } from "react-router-dom";',
  'import { QueryClient } from "@tanstack/react-query";',
  "const App = () => {",
  '  return <div className="min-h-screen">',
  "    <Toaster richColors />",
  "    <RouterProvider router={router} />",
  "  </div>;",
  "};",
  "export default App;",
  'import { Button } from "@/components/ui/button";',
  '  const [data, setData] = useState<Project[]>([]);',
  "  const { data, isLoading } = useQuery({",
  '    queryKey: ["projects"],',
  "    queryFn: fetchProjects,",
  "  });",
  '  <Card className="glass-panel p-6 space-y-4">',
  '  <Badge variant="secondary">Premium</Badge>',
  "  tailwind.config = { extend: { colors: { brand } } };",
  '  <motion.div animate={{ opacity: 1 }}>',
  "  const schema = z.object({ email: z.string().email() });",
];

// Pipeline stage detection from SSE logs
type PipelineStage = "planning" | "designing" | "execution" | "verification" | "idle";

const STAGE_META: Record<PipelineStage, { label: string; icon: string; color: string }> = {
  idle: { label: "Инициализация...", icon: "⚡", color: "text-muted-foreground/60" },
  planning: { label: "Планирование архитектуры", icon: "🧠", color: "text-blue-400" },
  designing: { label: "Генерация визуальных ассетов", icon: "🎨", color: "text-pink-400" },
  execution: { label: "Генерация кода", icon: "💻", color: "text-emerald-400" },
  verification: { label: "Верификация качества", icon: "🛡️", color: "text-amber-400" },
};

// Extract agent name from log message (e.g. "🧠 Директор — Ядро Истока" → "Director")
function extractAgent(log: string): string {
  const agentPatterns: [RegExp, string][] = [
    [/researcher|исследователь/i, "Researcher"],
    [/architect|архитект/i, "Architect"],
    [/planner|планировщик|DAG-план/i, "Planner"],
    [/director|директор/i, "Director"],
    [/brain|мозг|стратег/i, "Brain"],
    [/coder|кодер|код сгенерир/i, "Coder"],
    [/designer|дизайнер|визуальн/i, "Designer"],
    [/videographer|видеограф|промо/i, "Video"],
    [/validator|валидатор|verification/i, "Validator"],
    [/security|безопасност/i, "Security"],
    [/tester|тест/i, "Tester"],
    [/ui.?review/i, "UI Review"],
  ];
  for (const [pattern, name] of agentPatterns) {
    if (pattern.test(log)) return name;
  }
  return "System";
}

// Extract the meaningful message part (strip emoji prefix)
function extractMessage(log: string): string {
  return log.replace(/^[^\w\u0400-\u04FF]*/, "").slice(0, 80);
}

function detectStage(logs: string[]): PipelineStage {
  const last = (logs[logs.length - 1] || "").toLowerCase();
  const all = logs.join(" ").toLowerCase();
  if (/верификац|security|tester|ui.reviewer|quality/i.test(last)) return "verification";
  if (/кодер|coder|код|видеограф|многофайлов|группа.*файл/i.test(last)) return "execution";
  if (/designer|дизайн|визуальн|фотореалист|ассет/i.test(last)) return "designing";
  if (/план|planner|director|мозг|brain|архитект|strateg/i.test(last)) return "planning";
  if (/верификац|security|tester/i.test(all)) return "verification";
  if (/кодер|coder|многофайлов/i.test(all)) return "execution";
  if (/designer|дизайн/i.test(all)) return "designing";
  if (all.length > 0) return "planning";
  return "idle";
}

export default function GenerationMagic({ logs = [], progress = 0 }: GenerationMagicProps) {
  const [visibleLogs, setVisibleLogs] = useState<string[]>([]);
  const [codeLines, setCodeLines] = useState<string[]>([]);
  const [activeBlockIdx, setActiveBlockIdx] = useState(0);
  const logEndRef = useRef<HTMLDivElement>(null);

  // Stage detection
  const stage = useMemo(() => detectStage(logs), [logs]);
  const stageMeta = STAGE_META[stage];

  // Stream logs
  useEffect(() => {
    setVisibleLogs(logs.slice(-6));
  }, [logs]);

  // Auto-scroll logs
  useEffect(() => {
    logEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [visibleLogs]);

  // Cycling code snippets
  useEffect(() => {
    let idx = 0;
    const interval = setInterval(() => {
      setCodeLines((prev) => {
        const next = [...prev, CODE_SNIPPETS[idx % CODE_SNIPPETS.length]];
        idx++;
        return next.slice(-14);
      });
    }, 700);
    return () => clearInterval(interval);
  }, []);

  // Cycle active phantom block highlight
  useEffect(() => {
    const interval = setInterval(() => {
      setActiveBlockIdx((prev) => (prev + 1) % PHANTOM_BLOCKS.length);
    }, 2000);
    return () => clearInterval(interval);
  }, []);

  // Progress-based glow intensity
  const glowIntensity = useMemo(() => Math.max(0.1, progress / 100), [progress]);

  return (
    <div className="relative w-full h-full overflow-hidden bg-[#07070b]">
      {/* Animated grid background — brand-colored */}
      <div
        className="absolute inset-0"
        style={{
          backgroundImage:
            "linear-gradient(hsl(243 76% 58% / 0.05) 1px, transparent 1px), linear-gradient(90deg, hsl(243 76% 58% / 0.05) 1px, transparent 1px)",
          backgroundSize: "48px 48px",
        }}
      />
      {/* Secondary fine grid */}
      <div
        className="absolute inset-0 opacity-40"
        style={{
          backgroundImage:
            "linear-gradient(hsl(243 76% 58% / 0.02) 1px, transparent 1px), linear-gradient(90deg, hsl(243 76% 58% / 0.02) 1px, transparent 1px)",
          backgroundSize: "12px 12px",
        }}
      />

      {/* Radial glow from center */}
      <div
        className="absolute inset-0 pointer-events-none"
        style={{
          background: `radial-gradient(ellipse 60% 50% at 50% 50%, hsla(243, 76%, 58%, ${glowIntensity * 0.08}) 0%, transparent 70%)`,
        }}
      />

      {/* Neural connection lines (SVG) */}
      <svg className="absolute inset-0 w-full h-full pointer-events-none" style={{ zIndex: 1 }}>
        <defs>
          <linearGradient id="neural-grad" x1="0%" y1="0%" x2="100%" y2="0%">
            <stop offset="0%" stopColor="hsl(243 76% 58%)" stopOpacity="0.4" />
            <stop offset="50%" stopColor="hsl(220 80% 60%)" stopOpacity="0.2" />
            <stop offset="100%" stopColor="hsl(243 76% 58%)" stopOpacity="0" />
          </linearGradient>
        </defs>
        {NEURAL_LINES.map((line, i) => (
          <motion.line
            key={i}
            x1={line.x1}
            y1={line.y1}
            x2={line.x2}
            y2={line.y2}
            stroke="url(#neural-grad)"
            strokeWidth="0.6"
            strokeDasharray="4 6"
            initial={{ opacity: 0, pathLength: 0 }}
            animate={{
              opacity: [0, 0.3, 0.1, 0.3],
              pathLength: 1,
              strokeDashoffset: [0, -80],
            }}
            transition={{
              opacity: { duration: 4, repeat: Infinity, delay: i * 0.25 },
              pathLength: { duration: 1.5, delay: i * 0.15 },
              strokeDashoffset: { duration: 6, repeat: Infinity, ease: "linear" },
            }}
          />
        ))}
        {/* Pulse dots at neural endpoints */}
        {NEURAL_LINES.map((line, i) => (
          <motion.circle
            key={`dot-${i}`}
            cx={line.x2}
            cy={line.y2}
            r="2.5"
            fill="hsl(243 76% 58%)"
            initial={{ opacity: 0, scale: 0 }}
            animate={{ opacity: [0, 0.5, 0], scale: [0, 1.8, 0] }}
            transition={{ duration: 3, repeat: Infinity, delay: i * 0.35 + 0.8 }}
          />
        ))}
      </svg>

      {/* Phantom blueprint blocks with labels */}
      <div className="absolute inset-0 pointer-events-none" style={{ zIndex: 2 }}>
        {PHANTOM_BLOCKS.map((block, i) => (
          <motion.div
            key={i}
            className="absolute rounded-md"
            style={{
              left: block.x,
              top: block.y,
              width: block.w,
              height: block.h,
              border: "1px solid hsla(243, 76%, 58%, 0.1)",
            }}
            initial={{ opacity: 0, scale: 0.92 }}
            animate={{
              opacity: activeBlockIdx === i ? [0.12, 0.25, 0.12] : [0, 0.08, 0.04, 0.08],
              scale: activeBlockIdx === i ? [0.98, 1.01, 0.98] : [0.96, 1, 0.98, 1],
              borderColor: activeBlockIdx === i
                ? "hsla(243, 76%, 58%, 0.3)"
                : "hsla(243, 76%, 58%, 0.08)",
            }}
            transition={{
              duration: activeBlockIdx === i ? 2 : 3.5,
              repeat: Infinity,
              delay: block.delay,
              ease: "easeInOut",
            }}
          >
            <div className="w-full h-full rounded-md bg-primary/[0.02]" />
            {/* Block label */}
            <motion.span
              className="absolute top-1 left-2 text-[8px] sm:text-[9px] font-mono text-primary/30 select-none"
              animate={{ opacity: activeBlockIdx === i ? [0.3, 0.7, 0.3] : 0.2 }}
              transition={{ duration: 2, repeat: Infinity }}
            >
              {block.label}
            </motion.span>
            {/* Active scan line */}
            {activeBlockIdx === i && (
              <motion.div
                className="absolute left-0 w-full h-[1px] bg-gradient-to-r from-transparent via-primary/40 to-transparent"
                animate={{ top: ["0%", "100%"] }}
                transition={{ duration: 1.5, repeat: Infinity, ease: "linear" }}
              />
            )}
          </motion.div>
        ))}
      </div>

      {/* Center Logo + Neural Hub */}
      <div
        className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 flex flex-col items-center gap-3 sm:gap-4"
        style={{ zIndex: 10 }}
      >
        {/* Outer rings */}
        <div className="relative">
          <motion.div
            className="absolute -inset-6 sm:-inset-10 rounded-full"
            style={{ border: "1px solid hsla(243, 76%, 58%, 0.08)" }}
            animate={{ scale: [1, 1.12, 1], opacity: [0.2, 0.05, 0.2] }}
            transition={{ duration: 3.5, repeat: Infinity }}
          />
          <motion.div
            className="absolute -inset-12 sm:-inset-20 rounded-full"
            style={{ border: "1px solid hsla(243, 76%, 58%, 0.04)" }}
            animate={{ scale: [1, 1.08, 1], opacity: [0.1, 0.02, 0.1] }}
            transition={{ duration: 5, repeat: Infinity, delay: 0.5 }}
          />
          <motion.div
            className="absolute -inset-20 sm:-inset-32 rounded-full"
            style={{ border: "1px solid hsla(243, 76%, 58%, 0.02)" }}
            animate={{ scale: [1, 1.05, 1], opacity: [0.05, 0.01, 0.05] }}
            transition={{ duration: 7, repeat: Infinity, delay: 1 }}
          />

          {/* Core logo */}
          <motion.div
            className="w-14 h-14 sm:w-16 sm:h-16 rounded-2xl bg-primary/10 border border-primary/20 flex items-center justify-center backdrop-blur-md relative overflow-hidden"
            animate={{
              boxShadow: [
                "0 0 20px hsla(243,76%,58%,0.1), 0 0 60px hsla(243,76%,58%,0.03)",
                "0 0 40px hsla(243,76%,58%,0.2), 0 0 80px hsla(243,76%,58%,0.06)",
                "0 0 20px hsla(243,76%,58%,0.1), 0 0 60px hsla(243,76%,58%,0.03)",
              ],
            }}
            transition={{ duration: 2.5, repeat: Infinity }}
          >
            {/* Inner shimmer */}
            <motion.div
              className="absolute inset-0 bg-gradient-to-br from-primary/10 via-transparent to-primary/5"
              animate={{ opacity: [0.3, 0.8, 0.3] }}
              transition={{ duration: 3, repeat: Infinity }}
            />
            <span className="relative text-primary font-bold text-base sm:text-lg tracking-tight select-none">
              IC
            </span>
          </motion.div>
        </div>

        <motion.p
          className="text-primary/50 text-[10px] sm:text-xs font-medium tracking-[0.25em] uppercase select-none"
          animate={{ opacity: [0.3, 0.7, 0.3] }}
          transition={{ duration: 2.5, repeat: Infinity }}
        >
          Istok Core
        </motion.p>

        {/* Pipeline Stage Indicator */}
        <motion.div
          key={stage}
          initial={{ opacity: 0, y: 4 }}
          animate={{ opacity: 1, y: 0 }}
          className={`flex items-center gap-1.5 px-3 py-1 rounded-full border border-primary/10 bg-primary/[0.03] ${stageMeta.color}`}
        >
          <span className="text-sm">{stageMeta.icon}</span>
          <span className="text-[10px] sm:text-[11px] font-medium select-none">{stageMeta.label}</span>
        </motion.div>

        {/* Progress bar */}
        {progress > 0 && (
          <div className="w-36 sm:w-48 h-[3px] rounded-full bg-primary/10 overflow-hidden">
            <motion.div
              className="h-full rounded-full bg-gradient-to-r from-primary/60 via-primary/80 to-primary/60"
              initial={{ width: "0%" }}
              animate={{ width: `${progress}%` }}
              transition={{ duration: 0.6, ease: "easeOut" }}
            />
          </div>
        )}

        {/* Percentage text */}
        {progress > 0 && (
          <motion.span
            className="text-[10px] text-primary/40 font-mono tabular-nums select-none"
            animate={{ opacity: [0.4, 0.8, 0.4] }}
            transition={{ duration: 1.5, repeat: Infinity }}
          >
            {progress}%
          </motion.span>
        )}
      </div>

      {/* Streaming code overlay (right side) — hidden on very small screens */}
      <div
        className="absolute right-2 sm:right-4 top-12 bottom-12 w-48 sm:w-60 lg:w-72 overflow-hidden pointer-events-none hidden sm:block"
        style={{ zIndex: 5 }}
      >
        <div className="absolute inset-0 bg-gradient-to-l from-transparent via-transparent to-[#07070b]" />
        <div className="absolute inset-0 bg-gradient-to-b from-[#07070b] via-transparent to-[#07070b]" />
        <div className="flex flex-col gap-0.5 pt-2">
          <AnimatePresence mode="popLayout">
            {codeLines.map((line, i) => (
              <motion.div
                key={`${i}-${line}`}
                initial={{ opacity: 0, x: 16 }}
                animate={{ opacity: 0.12, x: 0 }}
                exit={{ opacity: 0, x: -8 }}
                transition={{ duration: 0.35 }}
                className="text-[9px] sm:text-[10px] font-mono text-primary/50 whitespace-nowrap"
              >
                <span className="text-primary/20 mr-2">{String(i + 1).padStart(2, "0")}</span>
                {line}
              </motion.div>
            ))}
          </AnimatePresence>
        </div>
      </div>

      {/* Agent status logs — prominent full-width overlay at bottom */}
      <div
        className="absolute bottom-0 left-0 right-0 max-h-44 overflow-hidden pointer-events-none"
        style={{ zIndex: 15 }}
      >
        <div className="absolute inset-0 bg-gradient-to-t from-[#07070b] via-[#07070b]/80 to-transparent" />
        <div className="relative flex flex-col gap-1 p-3 sm:p-4 pt-8">
          <AnimatePresence mode="popLayout">
            {visibleLogs.map((log, i) => (
              <motion.div
                key={`log-${i}-${log.slice(0, 30)}`}
                initial={{ opacity: 0, y: 8, x: -4 }}
                animate={{ opacity: 1, y: 0, x: 0 }}
                exit={{ opacity: 0, y: -6 }}
                transition={{ duration: 0.3 }}
                className="text-[10px] sm:text-[11px] font-mono text-foreground/70 truncate flex items-center gap-2"
              >
                <span className="w-1.5 h-1.5 rounded-full bg-primary shrink-0 animate-pulse" />
                <span className="text-primary/80 font-semibold">{extractAgent(log)}</span>
                <span className="text-muted-foreground/80">{extractMessage(log)}</span>
              </motion.div>
            ))}
          </AnimatePresence>
          <div ref={logEndRef} />
        </div>
      </div>

      {/* Horizontal scanning beam */}
      <motion.div
        className="absolute left-0 right-0 h-[1px] pointer-events-none"
        style={{
          zIndex: 3,
          background: "linear-gradient(90deg, transparent 0%, hsla(243, 76%, 58%, 0.15) 20%, hsla(243, 76%, 58%, 0.25) 50%, hsla(243, 76%, 58%, 0.15) 80%, transparent 100%)",
        }}
        animate={{ top: ["5%", "95%", "5%"] }}
        transition={{ duration: 8, repeat: Infinity, ease: "easeInOut" }}
      />

      {/* Vertical scanning beam (secondary) */}
      <motion.div
        className="absolute top-0 bottom-0 w-[1px] pointer-events-none"
        style={{
          zIndex: 3,
          background: "linear-gradient(180deg, transparent 0%, hsla(220, 80%, 60%, 0.1) 30%, hsla(220, 80%, 60%, 0.15) 50%, hsla(220, 80%, 60%, 0.1) 70%, transparent 100%)",
        }}
        animate={{ left: ["10%", "90%", "10%"] }}
        transition={{ duration: 12, repeat: Infinity, ease: "easeInOut" }}
      />
    </div>
  );
}
