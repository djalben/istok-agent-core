import { useState, useEffect } from "react";
import { AlertTriangle, RefreshCw, Loader2 } from "lucide-react";
import { toast } from "sonner";
import { api, ApiError } from "@/lib/api";

const InsufficientFundsOverlay = () => {
  const [open, setOpen] = useState(false);
  const [sessionId, setSessionId] = useState("");
  const [resuming, setResuming] = useState(false);

  useEffect(() => {
    const handler = (e: Event) => {
      const detail = (e as CustomEvent<{ session_id: string }>).detail;
      if (detail?.session_id) {
        setSessionId(detail.session_id);
        setOpen(true);
      }
    };
    window.addEventListener("istok:insufficient_funds", handler);
    return () => window.removeEventListener("istok:insufficient_funds", handler);
  }, []);

  const handleResume = async () => {
    if (!sessionId) return;
    setResuming(true);
    try {
      await api.resumeFunds(sessionId);
      toast.success("Генерация возобновлена!");
      setOpen(false);
    } catch (err) {
      if (err instanceof ApiError && err.status === 404) {
        toast.info("Сессия уже завершена или возобновлена");
        setOpen(false);
      } else {
        toast.error(err instanceof Error ? err.message : "Ошибка возобновления");
      }
    } finally {
      setResuming(false);
    }
  };

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/60 backdrop-blur-sm">
      <div className="mx-4 w-full max-w-md rounded-2xl border border-amber-500/30 bg-background/95 p-6 shadow-2xl">
        <div className="flex flex-col items-center text-center gap-4">
          <div className="flex h-14 w-14 items-center justify-center rounded-full bg-amber-500/10 border border-amber-500/30">
            <AlertTriangle size={28} className="text-amber-400" />
          </div>

          <div className="space-y-2">
            <h2 className="text-lg font-bold text-foreground">
              Генерация приостановлена
            </h2>
            <p className="text-sm text-muted-foreground leading-relaxed">
              Недостаточно средств для продолжения генерации. Пополните кошелек и нажмите кнопку ниже.
            </p>
          </div>

          <button
            onClick={handleResume}
            disabled={resuming}
            className="mt-2 flex items-center gap-2 px-6 py-3 text-sm font-semibold rounded-xl btn-gradient text-primary-foreground hover:scale-105 transition-all duration-200 disabled:opacity-50 disabled:hover:scale-100"
          >
            {resuming ? (
              <Loader2 size={16} className="animate-spin" />
            ) : (
              <RefreshCw size={16} />
            )}
            {resuming ? "Возобновляю..." : "Возобновить генерацию"}
          </button>
        </div>
      </div>
    </div>
  );
};

export default InsufficientFundsOverlay;
