import { useState, useRef } from "react";
import { Zap, Wand2, Loader2 } from "lucide-react";
import { toast } from "sonner";

interface GenerationInputProps {
  onGenerate: (prompt: string) => void;
}

const GenerationInput = ({ onGenerate }: GenerationInputProps) => {
  const [prompt, setPrompt] = useState("");
  const [focused, setFocused] = useState(false);
  const [firing, setFiring] = useState(false);
  const [enhancing, setEnhancing] = useState(false);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const handleGenerate = () => {
    if (!prompt.trim()) return;
    setFiring(true);
    setTimeout(() => {
      setFiring(false);
      onGenerate(prompt);
    }, 500);
  };

  const handleEnhance = async () => {
    if (!prompt.trim()) return;
    setEnhancing(true);
    try {
      const apiBase = import.meta.env.VITE_API_URL || "http://localhost:8080";
      const resp = await fetch(`${apiBase}/api/v1/prompt/enhance`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ prompt: prompt.trim() }),
      });
      if (!resp.ok) {
        const err = await resp.json().catch(() => ({ error: "Enhancement failed" }));
        toast.error(err.error || "Enhancement failed");
        return;
      }
      const data = await resp.json();
      if (data.enhanced) {
        setPrompt(data.enhanced);
        toast.success("Промт улучшен ИИ-помощником!");
        textareaRef.current?.focus();
      }
    } catch {
      toast.error("Ошибка связи с ИИ-помощником");
    } finally {
      setEnhancing(false);
    }
  };

  return (
    <div className="flex flex-col items-center justify-center h-full px-8">
      <h1 className="text-4xl font-extrabold text-foreground mb-3 text-center tracking-tight text-glow animate-fade-in-up">
        Создайте приложение из идеи
      </h1>
      <p className="text-muted-foreground text-sm mb-12 text-center max-w-md leading-relaxed animate-fade-in-up animate-delay-100">
        Опишите, что вы хотите построить, и Исток сгенерирует его для вас
      </p>

      <div className="w-full max-w-2xl animate-fade-in-up animate-delay-200">
        <div className={`glass-subtle rounded-xl p-1 transition-all duration-300 ${
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
            placeholder="Опишите ваше приложение на русском языке..."
            rows={5}
            className="w-full bg-transparent text-foreground text-base resize-none outline-none placeholder:text-muted-foreground/60 px-4 py-3 rounded-xl"
          />
          <div className="flex items-center justify-end px-3 pb-2">
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
              <span>{enhancing ? "Улучшаю..." : "Magic Wand"}</span>
            </button>
          </div>
        </div>
      </div>

      <div className="mt-8 relative animate-fade-in-up animate-delay-300">
        <button
          onClick={handleGenerate}
          disabled={!prompt.trim()}
          className={`flex items-center gap-2 px-7 py-3 font-semibold text-sm rounded-xl transition-all duration-200 ${
            prompt.trim()
              ? "btn-gradient text-primary-foreground"
              : "bg-secondary text-muted-foreground cursor-not-allowed"
          }`}
        >
          <Zap size={16} className={firing ? "lightning-flash" : ""} />
          Запустить генерацию
        </button>

        {firing && (
          <div className="absolute top-1/2 left-full -translate-y-1/2 h-px bg-primary energy-crack" />
        )}
      </div>
    </div>
  );
};

export default GenerationInput;
