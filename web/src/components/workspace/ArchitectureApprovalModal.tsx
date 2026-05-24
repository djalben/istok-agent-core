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

const FeatureApprovalModal = () => {
  const [open, setOpen] = useState(false);
  const [payload, setPayload] = useState<ApprovalPayload | null>(null);
  const [expanded, setExpanded] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [feedback, setFeedback] = useState("");

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
      await api.approveArchitecture(payload.session_id, true, feedback || undefined);
      toast.success("✅ Функционал утверждён — начинаем разработку!");
      setOpen(false);
      setFeedback("");
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
      await api.approveArchitecture(payload.session_id, false, feedback || "rejected by user");
      toast.info("Функционал отклонён — генерация остановлена");
      setOpen(false);
      setFeedback("");
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
            <span className="text-amber-400">📋</span>
            Утверждение функционала
          </DialogTitle>
          <DialogDescription className="text-xs text-muted-foreground">
            Мы подготовили список функций вашего будущего приложения. Ознакомьтесь и утвердите, чтобы начать разработку.
          </DialogDescription>
        </DialogHeader>

        {/* Collapsible business features draft */}
        <div className="flex-1 min-h-0 overflow-auto">
          <button
            onClick={() => setExpanded((v) => !v)}
            className="flex items-center gap-2 w-full text-left py-2 px-3 rounded-lg hover:bg-secondary/50 transition-colors text-sm font-medium text-foreground"
          >
            {expanded ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
            Бизнес-план приложения
          </button>
          {expanded && (
            <div className="mt-1 px-3 py-3 rounded-lg bg-secondary/20 border border-border/30 overflow-auto max-h-[40vh]">
              <div className="text-sm text-foreground/90 whitespace-pre-wrap leading-relaxed">
                {payload.draft_plan}
              </div>
            </div>
          )}
        </div>

        {/* Feedback textarea */}
        <div className="px-1 pt-2">
          <textarea
            value={feedback}
            onChange={(e) => setFeedback(e.target.value)}
            placeholder="Например: Добавь функцию авторизации через Telegram или убери блок с отзывами..."
            rows={2}
            className="w-full text-sm rounded-lg border border-border/40 bg-secondary/10 px-3 py-2 placeholder:text-muted-foreground/50 focus:outline-none focus:ring-1 focus:ring-primary/40 resize-none"
          />
        </div>

        <DialogFooter className="flex-shrink-0 gap-2 sm:gap-2 pt-2">
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
            {submitting ? "Отправляю..." : "Утвердить и начать разработку"}
          </button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};

export default FeatureApprovalModal;
