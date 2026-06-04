import { motion, AnimatePresence } from "framer-motion";
import { Check, Loader2, Circle, AlertCircle } from "lucide-react";
import { cn } from "@/lib/utils";
import type { Agent } from "@/lib/builderTypes";

function AgentIcon({ status }: { status: Agent["status"] }) {
  if (status === "working") return <Loader2 className="h-3.5 w-3.5 animate-spin text-primary" />;
  if (status === "thinking") return <Circle className="h-3.5 w-3.5 fill-muted-foreground/40 text-muted-foreground/40" />;
  if (status === "done") return <Check className="h-3.5 w-3.5 text-success" />;
  if (status === "error") return <AlertCircle className="h-3.5 w-3.5 text-destructive" />;
  return <Circle className="h-3.5 w-3.5 text-muted-foreground/30" />;
}

interface AgentPulseProps {
  agents: Agent[];
}

/** ИСТОК "Активность агентов" strip, fed by live SSE milestones. */
export function AgentPulse({ agents }: AgentPulseProps) {
  if (agents.length === 0) return null;
  const working = agents.filter((a) => a.status === "working").length;
  const done = agents.filter((a) => a.status === "done").length;

  return (
    <div className="border-t border-border/60 bg-panel">
      <div className="flex items-center justify-between px-3 py-2">
        <div className="flex items-center gap-2">
          <span className="font-mono text-[10px] uppercase tracking-wider text-muted-foreground">
            Активность агентов
          </span>
          <span className="rounded-md bg-primary/15 px-1.5 py-0.5 font-mono text-[10px] text-primary">
            {working} активн.
          </span>
        </div>
        <span className="font-mono text-[10px] text-muted-foreground">
          {done}/{agents.length} готово
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
  );
}
