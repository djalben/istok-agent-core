import { useState, useRef } from "react";
import { Zap, Sparkles } from "lucide-react";
import { motion } from "framer-motion";
import { useLanguage } from "@/hooks/useLanguage";
import NeuralBackground from "@/components/NeuralBackground";

interface HeroSectionProps {
  onGenerate: (prompt: string) => void;
}

const HeroSection = ({ onGenerate }: HeroSectionProps) => {
  const [prompt, setPrompt] = useState("");
  const [focused, setFocused] = useState(false);
  const [firing, setFiring] = useState(false);
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const { t } = useLanguage();

  const quickPrompts = [t("qpCRM"), t("qpCafe"), t("qpBot"), t("qpDashboard"), t("qpPortfolio")];

  const handleGenerate = () => {
    if (!prompt.trim()) return;
    setFiring(true);
    setTimeout(() => {
      setFiring(false);
      onGenerate(prompt);
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
    </section>
  );
};

export default HeroSection;
