import { useState, useEffect } from "react";
import { CheckCircle2, XCircle, ChevronDown, ChevronRight } from "lucide-react";
import { toast } from "sonner";
import { api } from "@/lib/api";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog";

interface ApprovalPayload {
  draft_plan: string;
  session_id: string;
}

const ArchitectureApprovalModal = () => {
  const [open, setOpen] = useState(false);
  const [payload, setPayload] = useState<ApprovalPayload | null>(null);
  const [expanded, setExpanded] = useState(true);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    const handler = (e: Event) => {
      const detail = (e as CustomEvent<ApprovalPayload>).detail;
      if (detail?.session_id) {
        setPayload(detail);
        setOpen(true);
      }
    };
    window.addEventListener("istok:user_action", handler);
    return () => window.removeEventListener("istok:user_action", handler);
  }, []);

  const handleApprove = async () => {
    if (!payload) return;
    setSubmitting(true);
    try {
      await api.approveArchitecture(payload.session_id, true);
      toast.success("✅ Архитектура утверждена!");
      setOpen(false);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Ошибка утверждения");
    } finally {
      setSubmitting(false);
    }
  };

  const handleReject = async () => {
    if (!payload) return;
    setSubmitting(true);
    try {
      await api.approveArchitecture(payload.session_id, false, "rejected by user");
      toast.info("Архитектура отклонена — генерация остановлена");
      setOpen(false);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Ошибка");
    } finally {
      setSubmitting(false);
    }
  };

  if (!payload) return null;

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogContent className="sm:max-w-3xl max-h-[85vh] flex flex-col">
        <DialogHeader>
          <DialogTitle className="text-lg font-bold flex items-center gap-2">
            <span className="text-amber-400">⏸️</span>
            Утверждение архитектуры
          </DialogTitle>
          <DialogDescription className="text-xs text-muted-foreground">
            Архитектор предложил план. Ознакомьтесь и утвердите для продолжения генерации.
          </DialogDescription>
        </DialogHeader>

        {/* Collapsible architecture draft */}
        <div className="flex-1 min-h-0 overflow-auto">
          <button
            onClick={() => setExpanded((v) => !v)}
            className="flex items-center gap-2 w-full text-left py-2 px-3 rounded-lg hover:bg-secondary/50 transition-colors text-sm font-medium text-foreground"
          >
            {expanded ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
            Предложенная архитектура
          </button>
          {expanded && (
            <div className="mt-1 px-3 py-3 rounded-lg bg-secondary/20 border border-border/30 overflow-auto max-h-[50vh]">
              <pre className="text-xs text-foreground/80 whitespace-pre-wrap font-mono leading-relaxed">
                {payload.draft_plan}
              </pre>
            </div>
          )}
        </div>

        <DialogFooter className="flex-shrink-0 gap-2 sm:gap-2 pt-3">
          <button
            onClick={handleReject}
            disabled={submitting}
            className="flex items-center gap-2 px-4 py-2 text-sm rounded-lg border border-destructive/40 text-destructive hover:bg-destructive/10 transition-colors"
          >
            <XCircle size={14} />
            Отклонить
          </button>
          <button
            onClick={handleApprove}
            disabled={submitting}
            className="flex items-center gap-2 px-5 py-2 text-sm font-semibold rounded-lg btn-gradient text-primary-foreground hover:scale-105 transition-all duration-200"
          >
            <CheckCircle2 size={14} />
            {submitting ? "Отправляю..." : "Утвердить"}
          </button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};

export default ArchitectureApprovalModal;
