import { useState, useRef, useEffect } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { Send, Bot, User, Terminal } from "lucide-react";
import { MOCK_MESSAGES, type AgentMessage, api } from "@lib/api";

function ThinkingIndicator() {
  return (
    <motion.div
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, y: -5 }}
      className="flex items-start gap-3 px-4"
    >
      <div className="w-7 h-7 rounded-lg bg-indigo-500/10 border border-indigo-500/20 flex items-center justify-center flex-shrink-0 mt-0.5">
        <Bot className="w-3.5 h-3.5 text-indigo-400 thinking-pulse" />
      </div>
      <div className="glass rounded-2xl rounded-tl-md px-4 py-3">
        <div className="dot-pulse flex gap-1.5">
          <span className="w-1.5 h-1.5 rounded-full bg-indigo-400" />
          <span className="w-1.5 h-1.5 rounded-full bg-indigo-400" />
          <span className="w-1.5 h-1.5 rounded-full bg-indigo-400" />
        </div>
      </div>
    </motion.div>
  );
}

function MessageBubble({ message, index }: { message: AgentMessage; index: number }) {
  const isAgent = message.role === "agent" || message.role === "system";

  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{
        duration: 0.35,
        delay: index * 0.08,
        ease: [0.25, 0.1, 0.25, 1],
      }}
      className={`flex items-start gap-3 px-4 ${isAgent ? "" : "flex-row-reverse"}`}
    >
      <div
        className={`w-7 h-7 rounded-lg flex items-center justify-center flex-shrink-0 mt-0.5 ${
          isAgent
            ? "bg-indigo-500/10 border border-indigo-500/20"
            : "bg-white/[0.06] border border-white/[0.08]"
        }`}
      >
        {isAgent ? (
          <Bot className="w-3.5 h-3.5 text-indigo-400" />
        ) : (
          <User className="w-3.5 h-3.5 text-zinc-400" />
        )}
      </div>

      <div
        className={`max-w-[80%] rounded-2xl px-4 py-3 ${
          isAgent
            ? "glass rounded-tl-md"
            : "bg-indigo-500/10 border border-indigo-500/15 rounded-tr-md"
        }`}
      >
        <p className="text-[13px] leading-relaxed text-zinc-200">
          {message.content}
        </p>
        <span className="text-[10px] text-zinc-600 mt-1.5 block font-mono">
          {(() => {
            const d = new Date(message.timestamp);
            return isNaN(d.getTime()) ? "" : d.toLocaleTimeString("ru-RU", { hour: "2-digit", minute: "2-digit" });
          })()}
        </span>
      </div>
    </motion.div>
  );
}

export function AgentTerminal() {
  const [messages, setMessages] = useState<AgentMessage[]>(MOCK_MESSAGES);
  const [input, setInput] = useState("");
  const [isThinking, setIsThinking] = useState(false);
  const scrollRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [messages, isThinking]);

  const handleSend = (e: React.FormEvent) => {
    e.preventDefault();
    if (!input.trim()) return;

    const userMsg: AgentMessage = {
      id: String(Date.now()),
      projectId: "proj_a1b2c3d4e5",
      role: "user",
      content: input,
      timestamp: new Date().toISOString(),
      status: "complete",
    };
    setMessages((prev) => [...prev, userMsg]);
    setInput("");
    setIsThinking(true);

    api
      .sendMessage("proj_a1b2c3d4e5", { content: input })
      .then((response) => {
        setIsThinking(false);
        const agentMsg: AgentMessage = {
          id: response.id,
          projectId: response.projectId,
          role: response.role,
          content: response.content,
          timestamp: response.timestamp,
          status: response.status,
        };
        setMessages((prev) => [...prev, agentMsg]);
      })
      .catch(() => {
        setIsThinking(false);
        const errMsg: AgentMessage = {
          id: String(Date.now() + 1),
          projectId: "proj_a1b2c3d4e5",
          role: "system",
          content: "Ошибка соединения с бэкендом. Повторите попытку.",
          timestamp: new Date().toISOString(),
          status: "error",
        };
        setMessages((prev) => [...prev, errMsg]);
      });
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5, delay: 0.1, ease: [0.25, 0.1, 0.25, 1] }}
      className="glass rounded-2xl flex flex-col h-full overflow-hidden relative noise"
    >
      {/* Header */}
      <div className="flex items-center gap-2.5 px-4 h-11 border-b border-white/[0.06] flex-shrink-0">
        <Terminal className="w-3.5 h-3.5 text-indigo-400" />
        <span className="text-[11px] uppercase tracking-[0.15em] text-zinc-500 font-medium">
          Терминал Агента
        </span>
        <div className="ml-auto flex items-center gap-1.5">
          <div className="w-1.5 h-1.5 rounded-full bg-green-500 green-pulse" />
          <span className="text-[10px] text-zinc-600">онлайн</span>
        </div>
      </div>

      {/* Messages */}
      <div
        ref={scrollRef}
        className="flex-1 overflow-y-auto py-4 space-y-4 relative z-10"
      >
        <AnimatePresence mode="popLayout">
          {messages.map((msg, i) => (
            <MessageBubble key={msg.id} message={msg} index={i} />
          ))}
          {isThinking && <ThinkingIndicator key="thinking" />}
        </AnimatePresence>
      </div>

      {/* Input */}
      <form onSubmit={handleSend} className="p-3 border-t border-white/[0.06] flex-shrink-0 relative z-10">
        <div className="flex items-center gap-2 glass rounded-xl px-3">
          <input
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            placeholder="Введите команду..."
            className="flex-1 bg-transparent py-2.5 text-[13px] text-white placeholder:text-zinc-600 outline-none font-mono"
          />
          <motion.button
            type="submit"
            whileHover={{ scale: 1.08 }}
            whileTap={{ scale: 0.92 }}
            className="w-8 h-8 rounded-lg bg-indigo-500/15 border border-indigo-500/20 flex items-center justify-center hover:bg-indigo-500/25 transition-colors"
          >
            <Send className="w-3.5 h-3.5 text-indigo-400" />
          </motion.button>
        </div>
      </form>
    </motion.div>
  );
}
