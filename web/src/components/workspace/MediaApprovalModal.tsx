import { useState, useEffect, useCallback, useRef } from "react";
import { CheckCircle2, XCircle, Loader2, Image as ImageIcon, Film, Lock, Wand2 } from "lucide-react";
import { toast } from "sonner";
import { api, ApiError, type MediaAsset } from "@/lib/api";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog";

interface MediaApprovalPayload {
  media_assets: MediaAsset[];
  session_id: string;
}

type VideoOption = "variant1" | "variant2" | "variant3" | "none";

const MAX_AI_GENERATIONS = 3;

const MediaApprovalModal = () => {
  const [open, setOpen] = useState(false);
  const [sessionId, setSessionId] = useState("");
  const [assets, setAssets] = useState<MediaAsset[]>([]);
  const [submitting, setSubmitting] = useState(false);
  const [generatingId, setGeneratingId] = useState<string | null>(null);
  const [aiUsed, setAiUsed] = useState(0);
  const [videoOption, setVideoOption] = useState<VideoOption>("none");
  const scrollRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handler = (e: Event) => {
      const detail = (e as CustomEvent<MediaApprovalPayload>).detail;
      if (detail?.session_id && Array.isArray(detail.media_assets)) {
        setSessionId(detail.session_id);
        setAssets(detail.media_assets.map(a => ({ ...a })));
        setAiUsed(0);
        const hasVideo = detail.media_assets.some(a => a.type === "video");
        setVideoOption(hasVideo ? "variant1" : "none");
        setOpen(true);
      }
    };
    window.addEventListener("istok:media_approval", handler);
    return () => window.removeEventListener("istok:media_approval", handler);
  }, []);

  const imageAssets = assets.filter(a => a.type === "image");
  const videoAssets = assets.filter(a => a.type === "video");
  const aiRemaining = MAX_AI_GENERATIONS - aiUsed;

  const updateAssetPrompt = useCallback((id: string, prompt: string) => {
    setAssets(prev => prev.map(a => a.id === id ? { ...a, prompt } : a));
  }, []);

  const handleGenerateAI = useCallback(async (id: string) => {
    if (aiRemaining <= 0) {
      toast.error("Лимит AI-генераций исчерпан");
      return;
    }
    const asset = assets.find(a => a.id === id);
    if (!asset?.prompt?.trim()) return;

    setGeneratingId(id);
    try {
      const { url, source } = await api.generateMediaPreview(id, asset.prompt);
      if (url) {
        setAssets(prev => prev.map(a => a.id === id ? { ...a, preview_url: url } : a));
        setAiUsed(prev => prev + 1);
        toast.success(source === "ai" ? "AI-превью сгенерировано" : "Стоковое фото загружено");
      }
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Ошибка генерации");
    } finally {
      setGeneratingId(null);
    }
  }, [assets, aiRemaining]);

  const handleApprove = async () => {
    if (!sessionId) return;
    setSubmitting(true);
    try {
      const finalAssets: MediaAsset[] = imageAssets.map(a => ({ ...a }));
      if (videoOption !== "none" && videoAssets.length > 0) {
        const idx = parseInt(videoOption.replace("variant", "")) - 1;
        const va = videoAssets[idx];
        if (va) {
          finalAssets.push({
            id: "promo_video", type: "video", placement: "promo_video",
            label: "Промо-ролик", prompt: va.prompt,
          });
        }
      }
      await api.approveMedia(sessionId, true, finalAssets);
      toast.success("Media Studio: ассеты утверждены!");
      setOpen(false);
    } catch (err) {
      if (err instanceof ApiError && err.status === 404) {
        toast.info("✅ Медиа уже утверждены автоматически — генерация продолжается");
        setOpen(false);
      } else {
        toast.error(err instanceof Error ? err.message : "Ошибка утверждения");
      }
    } finally {
      setSubmitting(false);
    }
  };

  const handleReject = async () => {
    if (!sessionId) return;
    setSubmitting(true);
    try {
      await api.approveMedia(sessionId, false, []);
      toast.info("Генерация медиа пропущена");
      setOpen(false);
    } catch (err) {
      if (err instanceof ApiError && err.status === 404) {
        toast.info("Время ожидания вышло — система автоматически продолжила работу");
        setOpen(false);
      } else {
        toast.error(err instanceof Error ? err.message : "Ошибка");
      }
    } finally {
      setSubmitting(false);
    }
  };

  if (!sessionId) return null;

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogContent className="sm:max-w-4xl max-h-[90vh] flex flex-col">
        <DialogHeader>
          <DialogTitle className="text-lg font-bold flex items-center gap-2">
            <ImageIcon size={20} className="text-purple-400" />
            Media Studio — Дизайн-ревью
          </DialogTitle>
          <DialogDescription className="text-xs text-muted-foreground">
            Выберите изображения для проекта. Стоковые фото включены бесплатно, AI-генерация — {aiRemaining}/{MAX_AI_GENERATIONS} осталось.
          </DialogDescription>
        </DialogHeader>

        <div ref={scrollRef} className="flex-1 min-h-0 overflow-auto space-y-5 py-2">
          {/* ── Image Assets ── */}
          {imageAssets.length > 0 && (
            <div className="space-y-3">
              <h3 className="text-sm font-semibold text-foreground/80 flex items-center gap-2">
                <ImageIcon size={14} className="text-purple-400" /> Изображения
              </h3>
              {imageAssets.map((asset) => (
                <div key={asset.id} className="rounded-xl border border-border/40 bg-secondary/10 overflow-hidden">
                  {/* Image preview */}
                  <div className="relative bg-black/20 flex items-center justify-center min-h-[180px]">
                    {asset.preview_url ? (
                      <img
                        src={asset.preview_url}
                        alt={asset.label}
                        className="w-full max-h-[220px] object-cover"
                        onError={(e) => {
                          (e.target as HTMLImageElement).src = `https://placehold.co/800x400/1a1a2e/8b5cf6?text=${encodeURIComponent(asset.label)}`;
                        }}
                      />
                    ) : (
                      <div className="flex flex-col items-center gap-2 py-8 text-muted-foreground/50">
                        <ImageIcon size={32} />
                        <span className="text-xs">Превью не загружено</span>
                      </div>
                    )}
                    <span className="absolute top-2 left-2 text-xs font-semibold text-white bg-purple-600/80 backdrop-blur-sm px-2.5 py-1 rounded-full">
                      {asset.label}
                    </span>
                  </div>

                  {/* Prompt + AI button */}
                  <div className="p-3 space-y-2">
                    <textarea
                      value={asset.prompt}
                      onChange={(e) => updateAssetPrompt(asset.id, e.target.value)}
                      rows={2}
                      placeholder="Опишите желаемое изображение..."
                      className="w-full text-sm rounded-lg border border-border/30 bg-background/50 px-3 py-2 placeholder:text-muted-foreground/50 focus:outline-none focus:ring-1 focus:ring-purple-500/40 resize-none"
                    />
                    <button
                      onClick={() => handleGenerateAI(asset.id)}
                      disabled={generatingId !== null || aiRemaining <= 0}
                      className="flex items-center gap-2 px-3 py-1.5 text-xs font-medium rounded-lg border border-purple-500/30 text-purple-300 hover:bg-purple-500/10 transition-all disabled:opacity-40 disabled:cursor-not-allowed"
                    >
                      {generatingId === asset.id ? (
                        <Loader2 size={13} className="animate-spin" />
                      ) : (
                        <Wand2 size={13} />
                      )}
                      {generatingId === asset.id
                        ? "Генерация..."
                        : `Сгенерировать AI-версию (осталось ${aiRemaining}/${MAX_AI_GENERATIONS})`
                      }
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}

          {/* ── Video Section ── */}
          <div className="space-y-3">
            <h3 className="text-sm font-semibold text-foreground/80 flex items-center gap-2">
              <Film size={14} className="text-sky-400" />
              Промо-ролик (AI Video)
              <span className="ml-auto flex items-center gap-1 text-[10px] font-medium text-amber-400 bg-amber-500/10 px-2 py-0.5 rounded-full">
                <Lock size={10} /> Premium
              </span>
            </h3>
            <div className="rounded-xl border border-border/40 bg-secondary/10 p-3 space-y-2">
              {videoAssets.map((va, idx) => (
                <label
                  key={va.id}
                  className={`flex items-start gap-2.5 p-2.5 rounded-lg cursor-pointer transition-all ${
                    videoOption === `variant${idx + 1}`
                      ? "bg-sky-500/10 border border-sky-500/30 shadow-sm"
                      : "hover:bg-secondary/30 border border-transparent"
                  }`}
                >
                  <input
                    type="radio"
                    name="video_option"
                    checked={videoOption === `variant${idx + 1}`}
                    onChange={() => setVideoOption(`variant${idx + 1}` as VideoOption)}
                    className="mt-1 accent-sky-500"
                  />
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="text-xs font-semibold text-sky-300">Сценарий {idx + 1}</span>
                      <Lock size={10} className="text-amber-400/60" />
                    </div>
                    <p className="text-sm text-foreground/70 mt-1 leading-relaxed">{va.prompt}</p>
                  </div>
                </label>
              ))}

              {/* No video */}
              <label
                className={`flex items-center gap-2.5 p-2.5 rounded-lg cursor-pointer transition-all ${
                  videoOption === "none"
                    ? "bg-secondary/30 border border-border/40"
                    : "hover:bg-secondary/30 border border-transparent"
                }`}
              >
                <input
                  type="radio"
                  name="video_option"
                  checked={videoOption === "none"}
                  onChange={() => setVideoOption("none")}
                  className="accent-sky-500"
                />
                <span className="text-sm text-muted-foreground">Без видео</span>
              </label>
            </div>
          </div>
        </div>

        <DialogFooter className="flex-shrink-0 gap-2 sm:gap-2 pt-3 border-t border-border/20">
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
            disabled={submitting || generatingId !== null}
            className="flex items-center gap-2 px-5 py-2.5 text-sm font-semibold rounded-lg btn-gradient text-primary-foreground hover:scale-105 transition-all duration-200"
          >
            {submitting ? (
              <Loader2 size={14} className="animate-spin" />
            ) : (
              <CheckCircle2 size={14} />
            )}
            {submitting ? "Отправляю..." : "Утвердить и продолжить"}
          </button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};

export default MediaApprovalModal;
