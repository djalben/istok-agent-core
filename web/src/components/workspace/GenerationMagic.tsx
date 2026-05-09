import { useEffect, useState, useRef } from "react";
import { motion, AnimatePresence } from "framer-motion";

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — GenerationMagic
//  Blueprint-анимация во время генерации проекта.
//  Сетка, нейронные связи, бегущий код, лого ИСТОК.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

interface GenerationMagicProps {
  /** SSE status messages streamed from agents */
  logs?: string[];
  /** Current generation progress 0–100 */
  progress?: number;
}

// Phantom block positions for the "blueprint wireframe"
const PHANTOM_BLOCKS = [
  { x: "8%", y: "12%", w: "35%", h: "8%", delay: 0 },
  { x: "50%", y: "12%", w: "42%", h: "8%", delay: 0.2 },
  { x: "8%", y: "26%", w: "84%", h: "22%", delay: 0.4 },
  { x: "8%", y: "54%", w: "40%", h: "14%", delay: 0.6 },
  { x: "54%", y: "54%", w: "38%", h: "14%", delay: 0.8 },
  { x: "8%", y: "74%", w: "26%", h: "10%", delay: 1.0 },
  { x: "38%", y: "74%", w: "26%", h: "10%", delay: 1.2 },
  { x: "68%", y: "74%", w: "24%", h: "10%", delay: 1.4 },
  { x: "8%", y: "88%", w: "84%", h: "6%", delay: 1.6 },
];

// Neural connection lines from center logo to edges
const NEURAL_LINES = [
  { x1: "50%", y1: "50%", x2: "5%", y2: "10%" },
  { x1: "50%", y1: "50%", x2: "95%", y2: "15%" },
  { x1: "50%", y1: "50%", x2: "10%", y2: "90%" },
  { x1: "50%", y1: "50%", x2: "90%", y2: "85%" },
  { x1: "50%", y1: "50%", x2: "5%", y2: "50%" },
  { x1: "50%", y1: "50%", x2: "95%", y2: "50%" },
  { x1: "50%", y1: "50%", x2: "30%", y2: "5%" },
  { x1: "50%", y1: "50%", x2: "70%", y2: "95%" },
];

const CODE_SNIPPETS = [
  'import { useState } from "react";',
  "const App = () => {",
  '  return <div className="app">',
  "    <Header />",
  "    <MainContent />",
  "    <Footer />",
  "  </div>;",
  "};",
  "export default App;",
  '<button onClick={handleSubmit}>',
  "  const [data, setData] = useState([]);",
  "  useEffect(() => { fetchData(); }, []);",
  '  <nav className="flex items-center">',
  "  const router = createBrowserRouter([",
  '    { path: "/", element: <Home /> },',
  "  ]);",
  "  tailwind.config = { theme: { extend: {} } };",
  '  <Card className="glass p-6">',
];

export default function GenerationMagic({ logs = [], progress = 0 }: GenerationMagicProps) {
  const [visibleLogs, setVisibleLogs] = useState<string[]>([]);
  const [codeLines, setCodeLines] = useState<string[]>([]);
  const logEndRef = useRef<HTMLDivElement>(null);

  // Stream logs with a slight delay for effect
  useEffect(() => {
    setVisibleLogs(logs.slice(-8));
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
        return next.slice(-12);
      });
    }, 800);
    return () => clearInterval(interval);
  }, []);

  return (
    <div className="relative w-full h-full overflow-hidden bg-[#08080c]">
      {/* Grid background */}
      <div
        className="absolute inset-0 opacity-[0.06]"
        style={{
          backgroundImage:
            "linear-gradient(hsl(263 70% 58% / 0.4) 1px, transparent 1px), linear-gradient(90deg, hsl(263 70% 58% / 0.4) 1px, transparent 1px)",
          backgroundSize: "40px 40px",
        }}
      />

      {/* Neural connection lines (SVG) */}
      <svg className="absolute inset-0 w-full h-full" style={{ zIndex: 1 }}>
        {NEURAL_LINES.map((line, i) => (
          <motion.line
            key={i}
            x1={line.x1}
            y1={line.y1}
            x2={line.x2}
            y2={line.y2}
            stroke="hsl(263 70% 58%)"
            strokeWidth="0.5"
            strokeDasharray="6 8"
            initial={{ opacity: 0, pathLength: 0 }}
            animate={{
              opacity: [0, 0.25, 0.1, 0.25],
              pathLength: 1,
              strokeDashoffset: [0, -100],
            }}
            transition={{
              opacity: { duration: 3, repeat: Infinity, delay: i * 0.3 },
              pathLength: { duration: 2, delay: i * 0.2 },
              strokeDashoffset: {
                duration: 8,
                repeat: Infinity,
                ease: "linear",
              },
            }}
          />
        ))}
        {/* Glowing pulse dots at neural endpoints */}
        {NEURAL_LINES.map((line, i) => (
          <motion.circle
            key={`dot-${i}`}
            cx={line.x2}
            cy={line.y2}
            r="3"
            fill="hsl(263 70% 58%)"
            initial={{ opacity: 0, scale: 0 }}
            animate={{
              opacity: [0, 0.6, 0],
              scale: [0, 1.5, 0],
            }}
            transition={{
              duration: 2.5,
              repeat: Infinity,
              delay: i * 0.4 + 1,
            }}
          />
        ))}
      </svg>

      {/* Phantom blueprint blocks */}
      <div className="absolute inset-0" style={{ zIndex: 2 }}>
        {PHANTOM_BLOCKS.map((block, i) => (
          <motion.div
            key={i}
            className="absolute rounded-lg border border-primary/10"
            style={{
              left: block.x,
              top: block.y,
              width: block.w,
              height: block.h,
            }}
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{
              opacity: [0, 0.15, 0.08, 0.15],
              scale: [0.95, 1, 0.98, 1],
            }}
            transition={{
              duration: 3,
              repeat: Infinity,
              delay: block.delay,
              ease: "easeInOut",
            }}
          >
            <div className="w-full h-full rounded-lg bg-primary/[0.03]" />
            {/* Scan line effect */}
            <motion.div
              className="absolute top-0 left-0 w-full h-[2px] bg-gradient-to-r from-transparent via-primary/30 to-transparent"
              animate={{ top: ["0%", "100%"] }}
              transition={{
                duration: 2,
                repeat: Infinity,
                delay: block.delay + 0.5,
                ease: "linear",
              }}
            />
          </motion.div>
        ))}
      </div>

      {/* Center Logo + Neural Hub */}
      <div
        className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 flex flex-col items-center gap-4"
        style={{ zIndex: 10 }}
      >
        {/* Glow rings */}
        <div className="relative">
          <motion.div
            className="absolute -inset-8 rounded-full border border-primary/10"
            animate={{ scale: [1, 1.15, 1], opacity: [0.15, 0.05, 0.15] }}
            transition={{ duration: 3, repeat: Infinity }}
          />
          <motion.div
            className="absolute -inset-16 rounded-full border border-primary/5"
            animate={{ scale: [1, 1.1, 1], opacity: [0.1, 0.02, 0.1] }}
            transition={{ duration: 4, repeat: Infinity, delay: 0.5 }}
          />
          <motion.div
            className="w-16 h-16 rounded-2xl bg-primary/10 border border-primary/20 flex items-center justify-center backdrop-blur-sm"
            animate={{
              boxShadow: [
                "0 0 20px hsla(263,70%,58%,0.1)",
                "0 0 40px hsla(263,70%,58%,0.2)",
                "0 0 20px hsla(263,70%,58%,0.1)",
              ],
            }}
            transition={{ duration: 2, repeat: Infinity }}
          >
            <span className="text-primary font-bold text-lg tracking-tight select-none">
              IC
            </span>
          </motion.div>
        </div>

        <motion.p
          className="text-primary/60 text-xs font-medium tracking-[0.2em] uppercase select-none"
          animate={{ opacity: [0.4, 0.8, 0.4] }}
          transition={{ duration: 2, repeat: Infinity }}
        >
          Istok Core
        </motion.p>

        {/* Progress bar */}
        {progress > 0 && (
          <div className="w-48 h-1 rounded-full bg-primary/10 overflow-hidden">
            <motion.div
              className="h-full rounded-full bg-primary/40"
              initial={{ width: "0%" }}
              animate={{ width: `${progress}%` }}
              transition={{ duration: 0.5 }}
            />
          </div>
        )}
      </div>

      {/* Streaming code overlay (right side, faded) */}
      <div
        className="absolute right-4 top-16 bottom-16 w-64 overflow-hidden pointer-events-none"
        style={{ zIndex: 5 }}
      >
        <div className="absolute inset-0 bg-gradient-to-l from-transparent to-[#08080c]" />
        <div className="absolute inset-0 bg-gradient-to-b from-[#08080c] via-transparent to-[#08080c]" />
        <div className="flex flex-col gap-1 pt-2">
          <AnimatePresence mode="popLayout">
            {codeLines.map((line, i) => (
              <motion.div
                key={`${i}-${line}`}
                initial={{ opacity: 0, x: 20 }}
                animate={{ opacity: 0.15, x: 0 }}
                exit={{ opacity: 0, x: -10 }}
                transition={{ duration: 0.4 }}
                className="text-[10px] font-mono text-primary/60 whitespace-nowrap"
              >
                {line}
              </motion.div>
            ))}
          </AnimatePresence>
        </div>
      </div>

      {/* Agent status logs (bottom overlay) */}
      {visibleLogs.length > 0 && (
        <div
          className="absolute bottom-4 left-4 right-4 max-h-28 overflow-hidden pointer-events-none"
          style={{ zIndex: 15 }}
        >
          <div className="absolute inset-0 bg-gradient-to-b from-transparent to-[#08080c]/80" />
          <div className="flex flex-col gap-0.5">
            <AnimatePresence mode="popLayout">
              {visibleLogs.map((log, i) => (
                <motion.div
                  key={`log-${i}-${log.slice(0, 20)}`}
                  initial={{ opacity: 0, y: 8 }}
                  animate={{ opacity: 0.5, y: 0 }}
                  exit={{ opacity: 0 }}
                  transition={{ duration: 0.3 }}
                  className="text-[10px] font-mono text-muted-foreground/50 truncate"
                >
                  {log}
                </motion.div>
              ))}
            </AnimatePresence>
            <div ref={logEndRef} />
          </div>
        </div>
      )}

      {/* Horizontal scanning beam */}
      <motion.div
        className="absolute left-0 right-0 h-[1px] bg-gradient-to-r from-transparent via-primary/20 to-transparent"
        style={{ zIndex: 3 }}
        animate={{ top: ["10%", "90%", "10%"] }}
        transition={{ duration: 6, repeat: Infinity, ease: "easeInOut" }}
      />
    </div>
  );
}
