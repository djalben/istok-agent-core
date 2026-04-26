import { motion, AnimatePresence } from "framer-motion";
import { Clock, RotateCcw, Trash2, X, History } from "lucide-react";
import type { CloudProject } from "@/lib/projectSync";

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — VersionHistoryDialog
//  Модалка с списком сохранённых версий проекта.
//  Backed by CloudProject (Supabase) — каждая запись = снимок генерации.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

export interface VersionHistoryDialogProps {
  open: boolean;
  projects: CloudProject[];
  onClose: () => void;
  onRestore: (project: CloudProject) => void;
  onDelete: (id: string) => void;
}

const VersionHistoryDialog = ({
  open,
  projects,
  onClose,
  onRestore,
  onDelete,
}: VersionHistoryDialogProps) => {
  return (
    <AnimatePresence>
      {open && (
        <>
          <motion.div
            key="overlay"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={onClose}
            className="fixed inset-0 z-50 bg-background/60 backdrop-blur-sm"
          />
          <motion.div
            key="dialog"
            role="dialog"
            aria-label="Version history"
            initial={{ opacity: 0, scale: 0.96, y: 8 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.98, y: 4 }}
            transition={{ duration: 0.25, ease: [0.22, 1, 0.36, 1] }}
            className="fixed top-1/2 left-1/2 z-50 w-[min(560px,92vw)] max-h-[78vh] -translate-x-1/2 -translate-y-1/2 glass-panel rounded-2xl shadow-xl overflow-hidden flex flex-col"
          >
            <header className="flex items-center justify-between px-4 py-3 border-b border-glass-border/30">
              <div className="flex items-center gap-2">
                <History size={14} className="text-primary" />
                <h2 className="text-[13px] font-semibold">Version History</h2>
                <span className="text-[10px] text-muted-foreground/60">
                  {projects.length} {projects.length === 1 ? "version" : "versions"}
                </span>
              </div>
              <button
                onClick={onClose}
                className="w-6 h-6 rounded-md flex items-center justify-center text-muted-foreground hover:text-foreground hover:bg-secondary/40 transition-colors"
              >
                <X size={14} />
              </button>
            </header>

            <div className="flex-1 overflow-y-auto px-2 py-2">
              {projects.length === 0 ? (
                <div className="py-10 text-center text-[12px] text-muted-foreground/50">
                  <History size={22} className="mx-auto mb-2 opacity-40" />
                  No versions yet. Generate a project — snapshots will appear here.
                </div>
              ) : (
                <ul className="space-y-1">
                  {projects.map((p) => (
                    <VersionRow
                      key={p.id}
                      project={p}
                      onRestore={() => {
                        onRestore(p);
                        onClose();
                      }}
                      onDelete={() => onDelete(p.id)}
                    />
                  ))}
                </ul>
              )}
            </div>

            <footer className="px-4 py-2 border-t border-glass-border/30 text-[10px] text-muted-foreground/50">
              Versions are stored in cloud per user. Restoring loads the snapshot into the workspace.
            </footer>
          </motion.div>
        </>
      )}
    </AnimatePresence>
  );
};

interface VersionRowProps {
  project: CloudProject;
  onRestore: () => void;
  onDelete: () => void;
}

const VersionRow = ({ project, onRestore, onDelete }: VersionRowProps) => {
  const createdAt = new Date(project.created_at);
  const dateLabel = formatRelative(createdAt);

  return (
    <li className="group flex items-center gap-2 px-2.5 py-2 rounded-lg hover:bg-secondary/30 transition-colors">
      <div className="w-7 h-7 rounded-lg bg-primary/10 flex items-center justify-center shrink-0">
        <Clock size={12} className="text-primary" />
      </div>
      <div className="flex-1 min-w-0">
        <p className="text-[12px] font-medium text-foreground truncate">
          {project.prompt || "Untitled"}
        </p>
        <p className="text-[10px] text-muted-foreground/60">{dateLabel}</p>
      </div>
      <button
        onClick={onRestore}
        className="flex items-center gap-1 h-6 px-2 rounded-md text-[10px] text-primary bg-primary/10 hover:bg-primary/20 transition-colors"
        title="Restore"
      >
        <RotateCcw size={10} />
        <span className="hidden sm:inline">Restore</span>
      </button>
      <button
        onClick={onDelete}
        className="opacity-0 group-hover:opacity-100 w-6 h-6 rounded-md flex items-center justify-center text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-all"
        title="Delete version"
      >
        <Trash2 size={11} />
      </button>
    </li>
  );
};

function formatRelative(d: Date): string {
  const diffMs = Date.now() - d.getTime();
  const min = Math.floor(diffMs / 60_000);
  if (min < 1) return "just now";
  if (min < 60) return `${min}m ago`;
  const h = Math.floor(min / 60);
  if (h < 24) return `${h}h ago`;
  const days = Math.floor(h / 24);
  if (days < 7) return `${days}d ago`;
  return d.toLocaleDateString();
}

export default VersionHistoryDialog;
