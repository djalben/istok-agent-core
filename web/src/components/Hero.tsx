import { useState, useRef } from "react";
import { motion } from "framer-motion";
import { Zap, ArrowRight, Sparkles } from "lucide-react";
import { useLanguage } from "@/hooks/useLanguage";

interface HeroProps {
  onGenerate: (prompt: string) => void;
}

const QUICK_PROMPTS = [
  "Landing page для SaaS стартапа",
  "Dashboard с аналитикой",
  "Интернет-магазин одежды",
  "Портфолио дизайнера",
  "CRM система",
];

const STATS = [
  { value: "12 000+", label: "проектов создано" },
  { value: "< 60 сек", label: "время генерации" },
  { value: "6 агентов", label: "работают параллельно" },
];

const Hero = ({ onGenerate }: HeroProps) => {
  const [prompt, setPrompt] = useState("");
  const [focused, setFocused] = useState(false);
  const inputRef = useRef<HTMLTextAreaElement>(null);
  const { t } = useLanguage();

  const handleSubmit = () => {
    if (prompt.trim()) onGenerate(prompt.trim());
  };

  const handleKey = (e: React.KeyboardEvent) => {
    if (e.key === "Enter" && (e.metaKey || e.ctrlKey)) handleSubmit();
  };

  return (
    <section className="relative min-h-screen flex flex-col items-center justify-center overflow-hidden bg-[#08080a] px-4">

      {/* ── Background Radial Glow ── */}
      <div className="pointer-events-none absolute inset-0 overflow-hidden">
        <div className="absolute top-[-10%] left-1/2 -translate-x-1/2 w-[900px] h-[600px] rounded-full bg-violet-700/10 blur-[120px]" />
        <div className="absolute top-[20%] left-[10%] w-[500px] h-[400px] rounded-full bg-indigo-800/8 blur-[100px]" />
        <div className="absolute bottom-[10%] right-[5%] w-[400px] h-[300px] rounded-full bg-violet-900/8 blur-[80px]" />
        {/* Grid pattern */}
        <div
          className="absolute inset-0 opacity-[0.025]"
          style={{
            backgroundImage:
              "linear-gradient(rgba(255,255,255,0.3) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,0.3) 1px, transparent 1px)",
            backgroundSize: "60px 60px",
          }}
        />
      </div>

      <div className="relative z-10 w-full max-w-5xl mx-auto text-center flex flex-col items-center">

        {/* Badge */}
        <motion.div
          initial={{ opacity: 0, y: 16 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5 }}
          className="mb-8 inline-flex items-center gap-2 px-4 py-1.5 rounded-full border border-violet-500/20 bg-violet-500/5 text-violet-300 text-xs font-medium"
        >
          <Sparkles size={12} className="text-violet-400" />
          Мультимодальный AI — Gemini + Claude + DeepSeek
        </motion.div>

        {/* H1 */}
        <motion.h1
          initial={{ opacity: 0, y: 24 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, delay: 0.05 }}
          className="text-5xl sm:text-6xl md:text-7xl lg:text-8xl font-extrabold text-white tracking-tight leading-[0.95] mb-6"
        >
          {t("heroTitle") || (
            <>
              Опишите идею.
              <br />
              <span className="bg-gradient-to-r from-violet-400 via-indigo-400 to-violet-400 bg-clip-text text-transparent">
                ИСТОК создаст код.
              </span>
            </>
          )}
        </motion.h1>

        {/* Subtitle */}
        <motion.p
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, delay: 0.12 }}
          className="text-lg md:text-xl text-white/50 max-w-2xl leading-relaxed mb-12"
        >
          {t("heroSubtitle") ||
            "6 AI-агентов параллельно исследуют, проектируют и пишут production-ready код. Без шаблонов — только ваша идея."}
        </motion.p>

        {/* Input */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, delay: 0.2 }}
          className="w-full max-w-2xl mb-6"
        >
          <div
            className={`relative rounded-2xl transition-all duration-300 ${
              focused
                ? "shadow-[0_0_0_1px_rgba(139,92,246,0.5),0_0_40px_rgba(139,92,246,0.12)]"
                : "shadow-[0_0_0_1px_rgba(255,255,255,0.06)]"
            } bg-white/[0.04]`}
          >
            <textarea
              ref={inputRef}
              value={prompt}
              onChange={(e) => setPrompt(e.target.value)}
              onFocus={() => setFocused(true)}
              onBlur={() => setFocused(false)}
              onKeyDown={handleKey}
              placeholder={t("heroPlaceholder") || "Опишите проект... (Ctrl+Enter для запуска)"}
              rows={3}
              className="w-full bg-transparent text-white text-base resize-none outline-none placeholder:text-white/25 px-5 py-4 pr-14 rounded-2xl"
            />
            <button
              onClick={handleSubmit}
              disabled={!prompt.trim()}
              className={`absolute bottom-3 right-3 w-10 h-10 rounded-xl flex items-center justify-center transition-all duration-200 ${
                prompt.trim()
                  ? "bg-gradient-to-br from-violet-600 to-indigo-600 text-white shadow-[0_0_20px_rgba(124,58,237,0.4)] hover:scale-105 hover:shadow-[0_0_30px_rgba(124,58,237,0.6)]"
                  : "bg-white/5 text-white/20 cursor-not-allowed"
              }`}
            >
              <ArrowRight size={18} />
            </button>
          </div>
        </motion.div>

        {/* Quick prompts */}
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ duration: 0.6, delay: 0.3 }}
          className="flex flex-wrap justify-center gap-2 mb-12"
        >
          {QUICK_PROMPTS.map((qp) => (
            <button
              key={qp}
              onClick={() => {
                setPrompt(qp);
                inputRef.current?.focus();
              }}
              className="px-3 py-1.5 text-xs text-white/40 hover:text-white/80 border border-white/8 hover:border-white/20 rounded-lg bg-white/[0.02] hover:bg-white/[0.05] transition-all duration-200"
            >
              {qp}
            </button>
          ))}
        </motion.div>

        {/* Generate button (full) */}
        <motion.button
          initial={{ opacity: 0, y: 12 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, delay: 0.35 }}
          onClick={handleSubmit}
          disabled={!prompt.trim()}
          className={`flex items-center gap-3 px-10 py-4 rounded-2xl font-semibold text-base transition-all duration-300 ${
            prompt.trim()
              ? "bg-gradient-to-r from-violet-600 to-indigo-600 text-white shadow-[0_0_30px_rgba(124,58,237,0.4)] hover:shadow-[0_0_50px_rgba(124,58,237,0.6)] hover:scale-105"
              : "bg-white/5 text-white/20 cursor-not-allowed border border-white/10"
          }`}
        >
          <Zap size={18} fill={prompt.trim() ? "currentColor" : "none"} />
          {t("heroGenerate") || "Запустить генерацию"}
        </motion.button>

        {/* Stats */}
        <motion.div
          initial={{ opacity: 0, y: 16 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.7, delay: 0.5 }}
          className="mt-20 grid grid-cols-3 gap-8 max-w-lg w-full"
        >
          {STATS.map(({ value, label }) => (
            <div key={label} className="text-center">
              <div className="text-2xl font-bold text-white mb-1">{value}</div>
              <div className="text-xs text-white/35">{label}</div>
            </div>
          ))}
        </motion.div>
      </div>
    </section>
  );
};

export default Hero;
