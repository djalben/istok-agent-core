import { useState, useEffect, useCallback } from "react";
import { CheckCircle2, XCircle, Sparkles, Loader2, Image as ImageIcon } from "lucide-react";
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

interface MediaApprovalPayload {
  media_prompts: string[];
  session_id: string;
}

const MediaApprovalModal = () => {
  const [open, setOpen] = useState(false);
  const [payload, setPayload] = useState<MediaApprovalPayload | null>(null);
  const [prompts, setPrompts] = useState<string[]>([]);
  const [submitting, setSubmitting] = useState(false);
  const [enhancingIdx, setEnhancingIdx] = useState<number | null>(null);

  useEffect(() => {
    const handler = (e: Event) => {
      const detail = (e as CustomEvent<MediaApprovalPayload>).detail;
      if (detail?.session_id && Array.isArray(detail.media_prompts)) {
        setPayload(detail);
        setPrompts([...detail.media_prompts]);
        setOpen(true);
      }
    };
    window.addEventListener("istok:media_approval", handler);
    return () => window.removeEventListener("istok:media_approval", handler);
  }, []);

  const updatePrompt = useCallback((idx: number, value: string) => {
    setPrompts((prev) => {
      const next = [...prev];
      next[idx] = value;
      return next;
    });
  }, []);

  const enhancePrompt = useCallback(async (idx: number) => {
    const current = prompts[idx];
    if (!current?.trim()) return;
    setEnhancingIdx(idx);
    try {
      const enhanced = await api.enhancePrompt(
        `Improve this image generation prompt. Add professional photography/design tags (cinematic, 8k, photorealistic, studio lighting, etc). Keep the original subject. Return ONLY the improved prompt, nothing else:\n\n${current}`
      );
      if (enhanced?.trim()) {
        updatePrompt(idx, enhanced.trim());
        toast.success("Промпт улучшен");
      }
    } catch {
      toast.error("Не удалось улучшить промпт");
    } finally {
      setEnhancingIdx(null);
    }
  }, [prompts, updatePrompt]);

  const handleApprove = async () => {
    if (!payload) return;
    setSubmitting(true);
    try {
      await api.approveMedia(payload.session_id, true, prompts);
      toast.success("🎨 Промпты утверждены — генерация медиа запущена!");
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
      await api.approveMedia(payload.session_id, false, []);
      toast.info("Генерация медиа пропущена");
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
            <span className="text-purple-400"><ImageIcon size={20} /></span>
            Дизайн-ревью: медиа-промпты
          </DialogTitle>
          <DialogDescription className="text-xs text-muted-foreground">
            Проверьте и при необходимости отредактируйте описания для генерации изображений и видео.
            Используйте ✨ для автоматического улучшения промпта профессиональными тегами.
          </DialogDescription>
        </DialogHeader>

        <div className="flex-1 min-h-0 overflow-auto space-y-3 py-2">
          {prompts.map((prompt, idx) => (
            <div key={idx} className="rounded-lg border border-border/40 bg-secondary/10 p-3">
              <div className="flex items-center gap-2 mb-2">
                <span className="text-xs font-medium text-muted-foreground">
                  Промпт #{idx + 1}
                </span>
                <button
                  onClick={() => enhancePrompt(idx)}
                  disabled={enhancingIdx !== null}
                  className="ml-auto flex items-center gap-1.5 px-2.5 py-1 text-xs rounded-md border border-purple-500/30 text-purple-400 hover:bg-purple-500/10 transition-colors disabled:opacity-50"
                >
                  {enhancingIdx === idx ? (
                    <Loader2 size={12} className="animate-spin" />
                  ) : (
                    <Sparkles size={12} />
                  )}
                  Улучшить (AI)
                </button>
              </div>
              <textarea
                value={prompt}
                onChange={(e) => updatePrompt(idx, e.target.value)}
                rows={3}
                className="w-full text-sm rounded-lg border border-border/30 bg-background/50 px-3 py-2 placeholder:text-muted-foreground/50 focus:outline-none focus:ring-1 focus:ring-purple-500/40 resize-none"
              />
            </div>
          ))}
        </div>

        <DialogFooter className="flex-shrink-0 gap-2 sm:gap-2 pt-2">
          <button
            onClick={handleReject}
            disabled={submitting}
            className="flex items-center gap-2 px-4 py-2 text-sm rounded-lg border border-destructive/40 text-destructive hover:bg-destructive/10 transition-colors"
          >
            <XCircle size={14} />
            Пропустить медиа
          </button>
          <button
            onClick={handleApprove}
            disabled={submitting || enhancingIdx !== null}
            className="flex items-center gap-2 px-5 py-2 text-sm font-semibold rounded-lg btn-gradient text-primary-foreground hover:scale-105 transition-all duration-200"
          >
            <CheckCircle2 size={14} />
            {submitting ? "Отправляю..." : "Утвердить и сгенерировать"}
          </button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};

export default MediaApprovalModal;
