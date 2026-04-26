import { motion, AnimatePresence } from "framer-motion";
import { ShieldCheck, Cpu, Sparkles } from "lucide-react";
import AgentPulseTimeline from "./AgentPulseTimeline";
import type { AgentPipelineId, AgentMilestone } from "@/hooks/useGeneration";

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — MilestonesPanel
//  Правая панель Workspace: 10-агентный таймлайн, FSM-state индикатор
//  и Verified-бейдж после прохождения Security Verification Gate.
//
//  Чисто презентационный компонент. Все данные — из useGeneration (SSE).
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

export interface MilestonesPanelProps {
  activeAgent: AgentPipelineId | null;
  milestones: AgentMilestone[];
  currentFSMState: string;
  securityApproved: boolean;
  testerApproved: boolean;
  uiReviewerApproved: boolean;
}

const MilestonesPanel = ({
  activeAgent,
  milestones,
  currentFSMState,
  securityApproved,
  testerApproved,
  uiReviewerApproved,
}: MilestonesPanelProps) => {
  const allVerified = securityApproved && testerApproved && uiReviewerApproved;
  const isRunning = activeAgent !== null;

  return (
    <div className="flex flex-col h-full gap-3">
      {/* ── Header: pipeline + FSM state ───────────────────── */}
      <header className="px-1 space-y-2">
        <div className="flex items-center gap-2">
          <Cpu size={13} className="text-primary" />
          <h2 className="text-[12px] font-bold tracking-tight text-foreground">
            Pipeline
          </h2>
          <span
            className={`ml-auto text-[9px] font-mono uppercase tracking-wider px-1.5 py-0.5 rounded-md border ${
              isRunning
                ? "bg-primary/10 border-primary/30 text-primary"
                : "bg-secondary/40 border-border/30 text-muted-foreground/60"
            }`}
            title="Current FSM state"
          >
            {currentFSMState || "idle"}
          </span>
        </div>

        {/* Verified banner — appears once Security Gate passes. */}
        <AnimatePresence>
          {securityApproved && (
            <motion.div
              key="verified-banner"
              initial={{ opacity: 0, y: -6, scale: 0.97 }}
              animate={{ opacity: 1, y: 0, scale: 1 }}
              exit={{ opacity: 0, y: -4 }}
              transition={{ duration: 0.32, ease: [0.34, 1.56, 0.64, 1] }}
              className="relative overflow-hidden rounded-lg border border-emerald-500/35 bg-gradient-to-r from-emerald-500/10 via-emerald-500/5 to-transparent px-2.5 py-1.5"
            >
              <div className="flex items-center gap-2">
                <div className="w-6 h-6 shrink-0 rounded-full bg-emerald-500/20 border border-emerald-500/40 flex items-center justify-center">
                  <ShieldCheck size={12} className="text-emerald-400" />
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-1.5">
                    <p className="text-[11px] font-bold text-emerald-300 leading-none">
                      Verified
                    </p>
                    {allVerified && (
                      <Sparkles size={9} className="text-emerald-300/80" />
                    )}
                  </div>
                  <p className="text-[9px] text-emerald-300/70 leading-tight mt-0.5">
                    {allVerified
                      ? "Security · Tests · UI — all clean"
                      : "Code passed Security Gate"}
                  </p>
                </div>
                <span className="text-[8.5px] uppercase tracking-[0.14em] text-emerald-300/60 font-semibold">
                  ✓ ok
                </span>
              </div>
              {/* subtle pulse glow when fully verified */}
              {allVerified && (
                <motion.span
                  aria-hidden
                  className="pointer-events-none absolute inset-0 rounded-lg ring-1 ring-emerald-400/30"
                  initial={{ opacity: 0.3 }}
                  animate={{ opacity: [0.3, 0.6, 0.3] }}
                  transition={{ duration: 2.4, repeat: Infinity, ease: "easeInOut" }}
                />
              )}
            </motion.div>
          )}
        </AnimatePresence>
      </header>

      {/* ── Body: 10-agent timeline ─────────────────────────── */}
      <div className="flex-1 min-h-0 overflow-y-auto pr-0.5">
        <AgentPulseTimeline
          activeAgent={activeAgent}
          milestones={milestones}
          securityApproved={securityApproved}
          testerApproved={testerApproved}
          uiReviewerApproved={uiReviewerApproved}
        />
      </div>

      {/* ── Footer: progress hint ───────────────────────────── */}
      <footer className="px-1 pt-1 border-t border-glass-border/20">
        <p className="text-[9px] text-muted-foreground/50 leading-snug">
          Live SSE · 10 agents · Verification Gate
        </p>
      </footer>
    </div>
  );
};

export default MilestonesPanel;
