import { useState, useRef } from "react";
import { Zap, Sparkles, Wand2, Loader2, Link2, Rocket } from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";
import { toast } from "sonner";
import { useLanguage } from "@/hooks/useLanguage";
import NeuralBackground from "@/components/NeuralBackground";
import { api } from "@/lib/api";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog";

interface HeroSectionProps {
  onGenerate: (prompt: string, referenceUrl?: string) => void;
}

const HeroSection = ({ onGenerate }: HeroSectionProps) => {
  const [prompt, setPrompt] = useState("");
  const [referenceUrl, setReferenceUrl] = useState("");
  const [showUrlInput, setShowUrlInput] = useState(false);
  const [focused, setFocused] = useState(false);
  const [firing, setFiring] = useState(false);
  const [enhancing, setEnhancing] = useState(false);
  const [briefOpen, setBriefOpen] = useState(false);
  const [briefText, setBriefText] = useState("");
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const { t } = useLanguage();

  const quickPrompts = [t("qpCRM"), t("qpCafe"), t("qpBot"), t("qpDashboard"), t("qpPortfolio")];

  const handleGenerate = () => {
    if (!prompt.trim()) return;
    setFiring(true);
    setTimeout(() => {
      setFiring(false);
      onGenerate(prompt, referenceUrl.trim() || undefined);
    }, 500);
  };

  const handleEnhance = async () => {
    if (!prompt.trim()) return;
    setEnhancing(true);
    try {
      const enhanced = await api.enhancePrompt(prompt.trim(), referenceUrl.trim() || undefined);
      if (enhanced) {
        setBriefText(enhanced);
        setBriefOpen(true);
        toast.success("✨ Бизнес-бриф готов!");
      }
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Ошибка связи с ИИ-помощником");
    } finally {
      setEnhancing(false);
    }
  };

  const handleBriefLaunch = () => {
    if (!briefText.trim()) return;
    setBriefOpen(false);
    setPrompt(briefText);
    setFiring(true);
    setTimeout(() => {
      setFiring(false);
      onGenerate(briefText, referenceUrl.trim() || undefined);
    }, 500);
  };

  const handleQuickPrompt = (text: string) => {
    setPrompt(text);
    textareaRef.current?.focus();
  };

  return (
    <section className="relative h-[calc(100vh-80px)] flex flex-col items-center justify-center px-4 md:px-6 py-6 overflow-hidden">
      <NeuralBackground />

      <div className="absolute inset-0 flex items-center justify-center pointer-events-none overflow-hidden">
        <div className="hero-glow" />
      </div>
      <div className="absolute top-1/4 -left-32 w-[500px] h-[500px] rounded-full pointer-events-none floating-blob" />

      <motion.div
        initial={{ opacity: 0, scale: 0.9 }}
        animate={{ opacity: 1, scale: 1 }}
        transition={{ duration: 0.5 }}
        className="mb-6 flex items-center gap-2 px-4 py-1.5 rounded-full border border-border/50 glass-subtle relative z-10"
      >
        <Sparkles size={14} className="text-primary" />
        <span className="text-xs text-muted-foreground">{t("heroBadge")}</span>
      </motion.div>

      <motion.h1
        initial={{ opacity: 0, y: 24 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.6, ease: "easeOut" }}
        className="text-3xl sm:text-4xl md:text-5xl lg:text-6xl font-extrabold text-foreground text-center tracking-tight text-glow max-w-4xl leading-[1.1] relative z-10 hero-text-contrast"
      >
        {t("heroTitle")}
      </motion.h1>

      <motion.p
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.6, delay: 0.1, ease: "easeOut" }}
        className="text-muted-foreground text-sm md:text-base lg:text-lg mt-4 mb-6 text-center max-w-2xl leading-relaxed px-4 relative z-10 hero-subtitle-contrast"
      >
        {t("heroSubtitle")}
      </motion.p>

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.6, delay: 0.2, ease: "easeOut" }}
        className="w-full max-w-2xl px-2 relative z-10"
      >
        <div className={`glass-subtle rounded-2xl p-1 transition-all duration-300 ${
          focused
            ? "shadow-[0_0_0_1px_hsla(243,76%,58%,0.4),0_0_30px_hsla(243,76%,58%,0.12),0_0_60px_hsla(243,76%,58%,0.04)]"
            : "shadow-none"
        }`}>
          <textarea
            ref={textareaRef}
            value={prompt}
            onChange={(e) => setPrompt(e.target.value)}
            onFocus={() => setFocused(true)}
            onBlur={() => setFocused(false)}
            placeholder={t("heroPlaceholder")}
            rows={3}
            className="w-full bg-transparent text-foreground text-sm md:text-base resize-none outline-none placeholder:text-muted-foreground/50 px-4 md:px-5 py-3 md:py-4 rounded-2xl"
          />

          {/* Reference URL input (collapsible) */}
          <AnimatePresence>
            {showUrlInput && (
              <motion.div
                initial={{ height: 0, opacity: 0 }}
                animate={{ height: "auto", opacity: 1 }}
                exit={{ height: 0, opacity: 0 }}
                transition={{ duration: 0.2 }}
                className="overflow-hidden px-4 md:px-5"
              >
                <div className="flex items-center gap-2 py-2 border-t border-border/20">
                  <Link2 size={14} className="text-muted-foreground/50 shrink-0" />
                  <input
                    type="url"
                    value={referenceUrl}
                    onChange={(e) => setReferenceUrl(e.target.value)}
                    placeholder="https://competitor-site.com (опционально)"
                    className="w-full bg-transparent text-foreground text-xs md:text-sm outline-none placeholder:text-muted-foreground/40"
                  />
                </div>
              </motion.div>
            )}
          </AnimatePresence>

          {/* Bottom toolbar: reference URL toggle + Magic Wand */}
          <div className="flex items-center justify-between px-3 md:px-4 pb-2 pt-1">
            <button
              onClick={() => setShowUrlInput((v) => !v)}
              className={`flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-[11px] font-medium transition-all duration-200 ${
                showUrlInput
                  ? "bg-primary/15 text-primary"
                  : "text-muted-foreground/50 hover:text-muted-foreground hover:bg-secondary/50"
              }`}
            >
              <Link2 size={12} />
              <span>URL-референс</span>
            </button>

            <button
              onClick={handleEnhance}
              disabled={!prompt.trim() || enhancing}
              className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition-all duration-200 ${
                prompt.trim() && !enhancing
                  ? "bg-amber-500/15 text-amber-400 hover:bg-amber-500/25 hover:shadow-[0_0_12px_hsla(38,92%,50%,0.15)]"
                  : "text-muted-foreground/40 cursor-not-allowed"
              }`}
              title="ИИ-помощник: улучшить промт"
            >
              {enhancing ? (
                <Loader2 size={13} className="animate-spin" />
              ) : (
                <Wand2 size={13} />
              )}
              <span>{enhancing ? "Улучшаю..." : "✨ Улучшить промт"}</span>
            </button>
          </div>
        </div>
      </motion.div>

      <motion.div
        initial={{ opacity: 0, y: 12 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.6, delay: 0.3, ease: "easeOut" }}
        className="flex flex-wrap justify-center gap-2 mt-4 max-w-2xl px-2 relative z-10"
      >
        {quickPrompts.map((qp) => (
          <button
            key={qp}
            onClick={() => handleQuickPrompt(qp)}
            className="px-3 py-1.5 text-xs text-muted-foreground hover:text-foreground border border-border/60 hover:border-primary/40 rounded-lg bg-secondary/50 hover:bg-primary/10 transition-all duration-200"
          >
            {qp}
          </button>
        ))}
      </motion.div>

      <motion.div
        initial={{ opacity: 0, y: 16 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.6, delay: 0.4, ease: "easeOut" }}
        className="mt-6 relative z-10"
      >
        <button
          onClick={handleGenerate}
          disabled={!prompt.trim()}
          className={`flex items-center gap-2.5 px-8 md:px-10 py-3.5 md:py-4 font-semibold text-sm md:text-base rounded-xl transition-all duration-300 ${
            prompt.trim()
              ? "btn-gradient text-primary-foreground hover:scale-110 hover:shadow-[0_8px_40px_hsla(243,76%,50%,0.5)]"
              : "bg-secondary text-muted-foreground cursor-not-allowed"
          }`}
        >
          <Zap size={16} className={firing ? "lightning-flash" : ""} />
          {t("heroGenerate")}
        </button>
        {firing && (
          <div className="absolute top-1/2 left-full -translate-y-1/2 h-px bg-primary energy-crack" />
        )}
      </motion.div>

      <motion.p
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ duration: 0.8, delay: 0.8 }}
        className="mt-12 text-xs text-muted-foreground/40 relative z-10"
      >
        {t("trustedBy")}
      </motion.p>

      {/* Presentation modal for enhanced brief */}
      <Dialog open={briefOpen} onOpenChange={setBriefOpen}>
        <DialogContent className="sm:max-w-2xl max-h-[90vh] flex flex-col">
          <DialogHeader>
            <DialogTitle className="text-lg font-bold">
              Утверждение бизнес-концепции
            </DialogTitle>
            <DialogDescription className="text-xs text-muted-foreground">
              Отредактируйте бриф при необходимости, затем запустите генерацию проекта.
            </DialogDescription>
          </DialogHeader>
          <div className="flex-1 min-h-0 overflow-auto py-2">
            <textarea
              value={briefText}
              onChange={(e) => setBriefText(e.target.value)}
              rows={15}
              className="w-full bg-secondary/30 text-foreground text-sm leading-relaxed resize-y outline-none rounded-lg border border-border/40 focus:border-primary/50 px-4 py-3 transition-colors"
            />
          </div>
          <DialogFooter className="flex-shrink-0 gap-2 sm:gap-2">
            <button
              onClick={() => setBriefOpen(false)}
              className="px-4 py-2 text-sm rounded-lg border border-border/60 text-muted-foreground hover:bg-secondary/50 transition-colors"
            >
              Отмена
            </button>
            <button
              onClick={handleBriefLaunch}
              disabled={!briefText.trim()}
              className="flex items-center gap-2 px-5 py-2 text-sm font-semibold rounded-lg btn-gradient text-primary-foreground hover:scale-105 hover:shadow-[0_4px_24px_hsla(243,76%,50%,0.4)] transition-all duration-200"
            >
              <Rocket size={14} />
              Запустить генерацию
            </button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </section>
  );
};

export default HeroSection;
