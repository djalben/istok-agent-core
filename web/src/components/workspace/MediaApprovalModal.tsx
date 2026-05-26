import { useState, useEffect, useCallback } from "react";
import { CheckCircle2, XCircle, Sparkles, Loader2, Image as ImageIcon, Eye, Film } from "lucide-react";
import { toast } from "sonner";
import { api, type MediaAsset } from "@/lib/api";
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

type VideoOption = "variant1" | "variant2" | "variant3" | "custom" | "none";

const MediaApprovalModal = () => {
  const [open, setOpen] = useState(false);
  const [sessionId, setSessionId] = useState("");
  const [assets, setAssets] = useState<MediaAsset[]>([]);
  const [previews, setPreviews] = useState<Record<string, string>>({});
  const [submitting, setSubmitting] = useState(false);
  const [enhancingId, setEnhancingId] = useState<string | null>(null);
  const [previewingId, setPreviewingId] = useState<string | null>(null);

  // Video options
  const [videoOption, setVideoOption] = useState<VideoOption>("none");
  const [customVideoPrompt, setCustomVideoPrompt] = useState("");

  useEffect(() => {
    const handler = (e: Event) => {
      const detail = (e as CustomEvent<MediaApprovalPayload>).detail;
      if (detail?.session_id && Array.isArray(detail.media_assets)) {
        setSessionId(detail.session_id);
        setAssets(detail.media_assets.map(a => ({ ...a })));
        setPreviews({});
        // Auto-select first video variant if present
        const videoAssets = detail.media_assets.filter(a => a.type === "video");
        if (videoAssets.length > 0) setVideoOption("variant1");
        else setVideoOption("none");
        setCustomVideoPrompt("");
        setOpen(true);
      }
    };
    window.addEventListener("istok:media_approval", handler);
    return () => window.removeEventListener("istok:media_approval", handler);
  }, []);

  const imageAssets = assets.filter(a => a.type === "image");
  const videoAssets = assets.filter(a => a.type === "video");

  const updateAssetPrompt = useCallback((id: string, prompt: string) => {
    setAssets(prev => prev.map(a => a.id === id ? { ...a, prompt } : a));
  }, []);

  const enhancePrompt = useCallback(async (id: string) => {
    const asset = assets.find(a => a.id === id);
    if (!asset?.prompt?.trim()) return;
    setEnhancingId(id);
    try {
      const enhanced = await api.enhancePrompt(
        `Improve this image generation prompt. Add professional photography/design tags (cinematic, 8k, photorealistic, studio lighting, etc). Keep the original subject. Return ONLY the improved prompt, nothing else:\n\n${asset.prompt}`
      );
      if (enhanced?.trim()) {
        updateAssetPrompt(id, enhanced.trim());
        toast.success("Промпт улучшен");
      }
    } catch {
      toast.error("Не удалось улучшить промпт");
    } finally {
      setEnhancingId(null);
    }
  }, [assets, updateAssetPrompt]);

  const generatePreview = useCallback(async (id: string) => {
    const asset = assets.find(a => a.id === id);
    if (!asset?.prompt?.trim()) return;
    setPreviewingId(id);
    try {
      const url = await api.generateMediaPreview(asset.prompt);
      if (url) {
        setPreviews(prev => ({ ...prev, [id]: url }));
        toast.success("Превью сгенерировано");
      }
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Ошибка генерации превью");
    } finally {
      setPreviewingId(null);
    }
  }, [assets]);

  const handleApprove = async () => {
    if (!sessionId) return;
    setSubmitting(true);
    try {
      // Build final assets array with selected video (if any)
      const finalAssets: MediaAsset[] = [...imageAssets];
      if (videoOption !== "none" && videoAssets.length > 0) {
        let selectedPrompt = "";
        if (videoOption === "variant1" && videoAssets[0]) selectedPrompt = videoAssets[0].prompt;
        else if (videoOption === "variant2" && videoAssets[1]) selectedPrompt = videoAssets[1].prompt;
        else if (videoOption === "variant3" && videoAssets[2]) selectedPrompt = videoAssets[2].prompt;
        else if (videoOption === "custom") selectedPrompt = customVideoPrompt;
        if (selectedPrompt) {
          finalAssets.push({
            id: "promo_video", type: "video", placement: "promo_video",
            label: "Промо-ролик", prompt: selectedPrompt,
          });
        }
      }
      await api.approveMedia(sessionId, true, finalAssets);
      toast.success("🎨 Медиа утверждено — генерация запущена!");
      setOpen(false);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Ошибка утверждения");
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
      toast.error(err instanceof Error ? err.message : "Ошибка");
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
            Отредактируйте описания и нажмите «Показать превью» для предварительной генерации.
          </DialogDescription>
        </DialogHeader>

        <div className="flex-1 min-h-0 overflow-auto space-y-4 py-2">
          {/* ── Image Assets ── */}
          {imageAssets.length > 0 && (
            <div className="space-y-3">
              <h3 className="text-sm font-semibold text-foreground/80 flex items-center gap-2">
                <ImageIcon size={14} /> Изображения
              </h3>
              {imageAssets.map((asset) => (
                <div key={asset.id} className="rounded-lg border border-border/40 bg-secondary/10 p-3">
                  <div className="flex items-center gap-2 mb-2">
                    <span className="text-xs font-semibold text-purple-300 bg-purple-500/10 px-2 py-0.5 rounded">
                      {asset.label}
                    </span>
                    <div className="ml-auto flex items-center gap-1.5">
                      <button
                        onClick={() => enhancePrompt(asset.id)}
                        disabled={enhancingId !== null || previewingId !== null}
                        className="flex items-center gap-1 px-2 py-1 text-xs rounded-md border border-purple-500/30 text-purple-400 hover:bg-purple-500/10 transition-colors disabled:opacity-50"
                      >
                        {enhancingId === asset.id ? <Loader2 size={11} className="animate-spin" /> : <Sparkles size={11} />}
                        Улучшить (AI)
                      </button>
                      <button
                        onClick={() => generatePreview(asset.id)}
                        disabled={previewingId !== null || enhancingId !== null}
                        className="flex items-center gap-1 px-2 py-1 text-xs rounded-md border border-sky-500/30 text-sky-400 hover:bg-sky-500/10 transition-colors disabled:opacity-50"
                      >
                        {previewingId === asset.id ? <Loader2 size={11} className="animate-spin" /> : <Eye size={11} />}
                        Превью
                      </button>
                    </div>
                  </div>
                  <textarea
                    value={asset.prompt}
                    onChange={(e) => updateAssetPrompt(asset.id, e.target.value)}
                    rows={2}
                    className="w-full text-sm rounded-lg border border-border/30 bg-background/50 px-3 py-2 placeholder:text-muted-foreground/50 focus:outline-none focus:ring-1 focus:ring-purple-500/40 resize-none"
                  />
                  {previews[asset.id] && (
                    <div className="mt-2 rounded-lg overflow-hidden border border-border/30">
                      <img
                        src={previews[asset.id]}
                        alt={asset.label}
                        className="w-full max-h-48 object-cover"
                      />
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}

          {/* ── Video Section ── */}
          <div className="space-y-3">
            <h3 className="text-sm font-semibold text-foreground/80 flex items-center gap-2">
              <Film size={14} /> Промо-ролик
            </h3>
            <div className="rounded-lg border border-border/40 bg-secondary/10 p-3 space-y-2">
              {videoAssets.map((va, idx) => (
                <label
                  key={va.id}
                  className={`flex items-start gap-2.5 p-2 rounded-md cursor-pointer transition-colors ${videoOption === `variant${idx + 1}` ? "bg-purple-500/10 border border-purple-500/30" : "hover:bg-secondary/30 border border-transparent"}`}
                >
                  <input
                    type="radio"
                    name="video_option"
                    checked={videoOption === `variant${idx + 1}`}
                    onChange={() => setVideoOption(`variant${idx + 1}` as VideoOption)}
                    className="mt-1 accent-purple-500"
                  />
                  <div className="flex-1 min-w-0">
                    <span className="text-xs font-medium text-muted-foreground">Вариант {idx + 1}</span>
                    <p className="text-sm text-foreground/80 mt-0.5 line-clamp-2">{va.prompt}</p>
                  </div>
                </label>
              ))}

              {/* Custom video prompt */}
              <label
                className={`flex items-start gap-2.5 p-2 rounded-md cursor-pointer transition-colors ${videoOption === "custom" ? "bg-purple-500/10 border border-purple-500/30" : "hover:bg-secondary/30 border border-transparent"}`}
              >
                <input
                  type="radio"
                  name="video_option"
                  checked={videoOption === "custom"}
                  onChange={() => setVideoOption("custom")}
                  className="mt-1 accent-purple-500"
                />
                <div className="flex-1 min-w-0">
                  <span className="text-xs font-medium text-muted-foreground">Свой промпт</span>
                  {videoOption === "custom" && (
                    <textarea
                      value={customVideoPrompt}
                      onChange={(e) => setCustomVideoPrompt(e.target.value)}
                      placeholder="Опишите желаемый промо-ролик..."
                      rows={2}
                      className="mt-1 w-full text-sm rounded-lg border border-border/30 bg-background/50 px-3 py-2 placeholder:text-muted-foreground/50 focus:outline-none focus:ring-1 focus:ring-purple-500/40 resize-none"
                    />
                  )}
                </div>
              </label>

              {/* No video */}
              <label
                className={`flex items-center gap-2.5 p-2 rounded-md cursor-pointer transition-colors ${videoOption === "none" ? "bg-secondary/30 border border-border/40" : "hover:bg-secondary/30 border border-transparent"}`}
              >
                <input
                  type="radio"
                  name="video_option"
                  checked={videoOption === "none"}
                  onChange={() => setVideoOption("none")}
                  className="accent-purple-500"
                />
                <span className="text-sm text-muted-foreground">Без ролика</span>
              </label>
            </div>
          </div>
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
            disabled={submitting || enhancingId !== null || previewingId !== null}
            className="flex items-center gap-2 px-5 py-2 text-sm font-semibold rounded-lg btn-gradient text-primary-foreground hover:scale-105 transition-all duration-200"
          >
            <CheckCircle2 size={14} />
            {submitting ? "Отправляю..." : "Утвердить и продолжить"}
          </button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};

export default MediaApprovalModal;
