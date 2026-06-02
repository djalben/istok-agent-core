import { useState, useEffect, useRef } from "react";
import { CheckCircle2, XCircle, SendHorizonal, Loader2 } from "lucide-react";
import { toast } from "sonner";
import { api, ApiError } from "@/lib/api";
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

interface ChatMessage {
  role: "user" | "assistant";
  text: string;
}

// Stub: feedback loop is now in-place (no replan event needed).
// Kept for backward compatibility with useGeneration.ts replan listener.
export function getLastReplanFeedback(): string { return ""; }

const FeatureApprovalModal = () => {
  const [open, setOpen] = useState(false);
  const [payload, setPayload] = useState<ApprovalPayload | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [replanning, setReplanning] = useState(false);
  const [feedback, setFeedback] = useState("");
  const [chatHistory, setChatHistory] = useState<ChatMessage[]>([]);
  const chatEndRef = useRef<HTMLDivElement>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    const handler = (e: Event) => {
      const detail = (e as CustomEvent<ApprovalPayload>).detail;
      if (detail?.session_id) {
        if (replanning) {
          // Updated plan arrived after feedback — update plan, keep chat history
          setPayload(detail);
          setChatHistory((prev) => [
            ...prev,
            { role: "assistant", text: detail.draft_plan },
          ]);
          setReplanning(false);
        } else {
          // First time or new session — reset everything
          setPayload(detail);
          setChatHistory([{ role: "assistant", text: detail.draft_plan }]);
          setOpen(true);
        }
      }
    };
    window.addEventListener("istok:user_action", handler);
    return () => window.removeEventListener("istok:user_action", handler);
  }, [replanning]);

  // Auto-scroll chat to bottom
  useEffect(() => {
    chatEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [chatHistory, replanning]);

  const handleApprove = async () => {
    if (!payload) return;
    setSubmitting(true);
    try {
      await api.approveArchitecture(payload.session_id, true);
      toast.success("✅ Функционал утверждён — начинаем разработку!");
      setOpen(false);
      resetState();
    } catch (err) {
      if (err instanceof ApiError && err.status === 404) {
        toast.info("✅ Уже утверждено автоматически — генерация продолжается");
        setOpen(false);
        resetState();
      } else {
        toast.error(err instanceof Error ? err.message : "Ошибка утверждения");
      }
    } finally {
      setSubmitting(false);
    }
  };

  const handleReject = async () => {
    if (!payload) return;
    setSubmitting(true);
    try {
      await api.approveArchitecture(payload.session_id, false, "rejected by user");
      toast.info("Функционал отклонён — генерация остановлена");
      setOpen(false);
      resetState();
    } catch (err) {
      if (err instanceof ApiError && err.status === 404) {
        toast.info("Время ожидания вышло — система автоматически продолжила работу");
        setOpen(false);
        resetState();
      } else {
        toast.error(err instanceof Error ? err.message : "Ошибка");
      }
    } finally {
      setSubmitting(false);
    }
  };

  const handleSendFeedback = async () => {
    if (!payload || !feedback.trim()) return;
    const text = feedback.trim();
    setSubmitting(true);
    try {
      // Add user message to chat
      setChatHistory((prev) => [...prev, { role: "user", text }]);
      setFeedback("");
      setReplanning(true);

      // Send feedback — backend will replan and send new user_action event
      await api.approveArchitecture(payload.session_id, false, text);
      toast.info("🔄 Правки отправлены — перепланирование...");
    } catch (err) {
      if (err instanceof ApiError && err.status === 404) {
        toast.info("Время ожидания вышло — система автоматически продолжила работу");
        setOpen(false);
        resetState();
      } else {
        setReplanning(false);
        toast.error(err instanceof Error ? err.message : "Ошибка отправки правок");
      }
    } finally {
      setSubmitting(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter" && !e.shiftKey && feedback.trim()) {
      e.preventDefault();
      handleSendFeedback();
    }
  };

  const resetState = () => {
    setFeedback("");
    setChatHistory([]);
    setReplanning(false);
  };

  if (!payload) return null;

  // Latest plan is the last assistant message
  const latestPlan = chatHistory.filter((m) => m.role === "assistant").pop()?.text ?? payload.draft_plan;
  // Previous feedback messages (for mini-chat display)
  const feedbackMessages = chatHistory.filter((m) => m.role === "user");
  const iteration = feedbackMessages.length;

  return (
    <Dialog open={open} onOpenChange={(v) => { if (!replanning) setOpen(v); }}>
      <DialogContent className="sm:max-w-3xl max-h-[85vh] flex flex-col">
        <DialogHeader>
          <DialogTitle className="text-lg font-bold flex items-center gap-2">
            <span className="text-amber-400">📋</span>
            Утверждение функционала
            {iteration > 0 && (
              <span className="text-xs font-normal text-muted-foreground ml-2">
                (итерация {iteration + 1})
              </span>
            )}
          </DialogTitle>
          <DialogDescription className="text-xs text-muted-foreground">
            Ознакомьтесь с планом и утвердите, или напишите правки — ИИ перепланирует с учётом ваших пожеланий.
          </DialogDescription>
        </DialogHeader>

        {/* Chat area: plan + feedback history */}
        <div className="flex-1 min-h-0 overflow-auto space-y-3 pr-1">
          {/* Current plan */}
          <div className="px-3 py-3 rounded-lg bg-secondary/20 border border-border/30 overflow-auto max-h-[40vh]">
            <div className="text-sm text-foreground/90 whitespace-pre-wrap leading-relaxed">
              {latestPlan}
            </div>
          </div>

          {/* Mini-chat: user feedback history */}
          {feedbackMessages.length > 0 && (
            <div className="space-y-2 px-1">
              <div className="text-xs font-medium text-muted-foreground">💬 История правок:</div>
              {chatHistory.slice(1).map((msg, i) => (
                <div
                  key={i}
                  className={`text-sm rounded-lg px-3 py-2 ${
                    msg.role === "user"
                      ? "bg-primary/10 border border-primary/20 text-foreground ml-8"
                      : "bg-secondary/30 border border-border/20 text-foreground/80 mr-8"
                  }`}
                >
                  <span className="text-xs font-medium text-muted-foreground block mb-1">
                    {msg.role === "user" ? "Вы:" : "Обновлённый план:"}
                  </span>
                  <span className="whitespace-pre-wrap">
                    {msg.role === "assistant" && msg.text.length > 300
                      ? msg.text.slice(0, 300) + "... (см. выше)"
                      : msg.text}
                  </span>
                </div>
              ))}
            </div>
          )}

          {/* Loading indicator during replanning */}
          {replanning && (
            <div className="flex items-center gap-2 px-3 py-3 rounded-lg bg-primary/5 border border-primary/20 animate-pulse">
              <Loader2 size={16} className="animate-spin text-primary" />
              <span className="text-sm text-muted-foreground">
                ИИ-Планировщик перерабатывает план с учётом ваших правок...
              </span>
            </div>
          )}

          <div ref={chatEndRef} />
        </div>

        {/* Feedback input */}
        <div className="px-1 pt-2">
          <div className="flex gap-2">
            <textarea
              ref={textareaRef}
              value={feedback}
              onChange={(e) => setFeedback(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="Что-то упустили? Напишите правки здесь..."
              rows={2}
              disabled={replanning || submitting}
              className="flex-1 text-sm rounded-lg border border-border/40 bg-secondary/10 px-3 py-2 placeholder:text-muted-foreground/50 focus:outline-none focus:ring-1 focus:ring-primary/40 resize-none disabled:opacity-50"
            />
            <button
              onClick={handleSendFeedback}
              disabled={!feedback.trim() || replanning || submitting}
              className="self-end px-3 py-2 rounded-lg bg-primary/10 border border-primary/30 text-primary hover:bg-primary/20 transition-colors disabled:opacity-30 disabled:cursor-not-allowed"
              title="Отправить правки (Enter)"
            >
              <SendHorizonal size={16} />
            </button>
          </div>
        </div>

        <DialogFooter className="flex-shrink-0 gap-2 sm:gap-2 pt-2">
          <button
            onClick={handleReject}
            disabled={submitting || replanning}
            className="flex items-center gap-2 px-4 py-2 text-sm rounded-lg border border-destructive/40 text-destructive hover:bg-destructive/10 transition-colors disabled:opacity-50"
          >
            <XCircle size={14} />
            Отклонить
          </button>
          <button
            onClick={handleApprove}
            disabled={submitting || replanning}
            className="flex items-center gap-2 px-5 py-2 text-sm font-semibold rounded-lg btn-gradient text-primary-foreground hover:scale-105 transition-all duration-200 disabled:opacity-50 disabled:hover:scale-100"
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
