import { useState } from "react";
import { motion } from "framer-motion";
import { Search, ArrowRight, Sparkles } from "lucide-react";

export function IntelligenceBar() {
  const [focused, setFocused] = useState(false);
  const [value, setValue] = useState("");
  const [submitted, setSubmitted] = useState(false);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!value.trim()) return;
    setSubmitted(true);
    setTimeout(() => setSubmitted(false), 2000);
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: -10 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5, ease: [0.25, 0.1, 0.25, 1] }}
      className="w-full"
    >
      <div className="flex items-center gap-3 mb-3">
        <div className="flex items-center gap-2">
          <Sparkles className="w-4 h-4 text-indigo-400" />
          <span className="text-[10px] uppercase tracking-[0.2em] text-zinc-500 font-medium">
            Слой интеллекта
          </span>
        </div>
        <div className="flex-1 h-px bg-gradient-to-r from-white/[0.06] to-transparent" />
        <span className="text-[10px] text-zinc-600 font-mono">Deep Research v2.4</span>
      </div>

      <form onSubmit={handleSubmit} className="relative">
        <div
          className={`relative flex items-center glass rounded-2xl transition-all duration-300 ${
            focused
              ? "border-indigo-500/30 glow-indigo-sm"
              : "border-white/[0.06]"
          }`}
        >
          <div className="flex items-center pl-4 pr-2">
            <Search className="w-4 h-4 text-zinc-500" />
          </div>

          <input
            type="text"
            value={value}
            onChange={(e) => setValue(e.target.value)}
            onFocus={() => setFocused(true)}
            onBlur={() => setFocused(false)}
            placeholder="URL конкурента или спецификация проекта"
            className="flex-1 bg-transparent py-3.5 px-2 text-sm text-white placeholder:text-zinc-600 outline-none font-mono"
          />

          <div className="pr-2">
            <motion.button
              type="submit"
              whileHover={{ scale: 1.03 }}
              whileTap={{ scale: 0.97 }}
              className={`flex items-center gap-2 px-4 py-2 rounded-xl text-[13px] font-medium transition-all duration-300 ${
                submitted
                  ? "bg-green-500/20 text-green-400 border border-green-500/20"
                  : "bg-gradient-to-r from-indigo-500 to-violet-500 text-white glow-indigo-sm hover:glow-indigo"
              }`}
            >
              {submitted ? (
                "Анализ..."
              ) : (
                <>
                  <span className="hidden sm:inline">Проанализировать и Начать</span>
                  <span className="sm:hidden">Начать</span>
                  <ArrowRight className="w-3.5 h-3.5" />
                </>
              )}
            </motion.button>
          </div>
        </div>

        {/* Subtle gradient line under input */}
        <div className="absolute -bottom-px left-1/4 right-1/4 h-px bg-gradient-to-r from-transparent via-indigo-500/20 to-transparent" />
      </form>
    </motion.div>
  );
}
