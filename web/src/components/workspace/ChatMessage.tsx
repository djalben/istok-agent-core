import { useMemo, useState } from "react";
import { motion } from "framer-motion";
import { Bot, User, Copy, Pencil, Check } from "lucide-react";
import type { ChatMessage as ChatMessageType } from "@/hooks/useGeneration";

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — ChatMessage (with code-block [Edit Prompt])
//  Парсит сообщение на текст + ```...``` code blocks. Каждый блок
//  получает кнопку [Edit Prompt] — при клике вызывает onEditPrompt
//  с extracted code, который ChatPanel подставляет в input.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

export interface ChatMessageProps {
  message: ChatMessageType;
  /** Called when the user clicks [Edit Prompt] on a code block. */
  onEditPrompt?: (codeSnippet: string, messageId: string) => void;
}

interface Segment {
  type: "text" | "code";
  content: string;
  lang?: string;
}

/** Split raw string into alternating text / fenced code segments. */
function splitSegments(raw: string): Segment[] {
  const src = raw ?? "";
  const regex = /```([a-zA-Z0-9_+-]*)\n?([\s\S]*?)```/g;
  const out: Segment[] = [];
  let lastIdx = 0;
  let m: RegExpExecArray | null;
  while ((m = regex.exec(src)) !== null) {
    if (m.index > lastIdx) {
      out.push({ type: "text", content: src.slice(lastIdx, m.index) });
    }
    out.push({ type: "code", content: m[2].trim(), lang: m[1] || "plain" });
    lastIdx = m.index + m[0].length;
  }
  if (lastIdx < src.length) {
    out.push({ type: "text", content: src.slice(lastIdx) });
  }
  if (out.length === 0) out.push({ type: "text", content: src });
  return out;
}

const ChatMessage = ({ message, onEditPrompt }: ChatMessageProps) => {
  const raw = typeof message.content === "string" ? message.content : JSON.stringify(message.content, null, 2);
  const segments = useMemo(() => splitSegments(raw), [raw]);
  const isUser = message.role === "user";

  return (
    <motion.div
      initial={{ opacity: 0, y: 6 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.2 }}
      className={`flex items-end gap-1.5 ${isUser ? "justify-end" : "justify-start"}`}
    >
      {!isUser && (
        <div className="w-5 h-5 rounded-full bg-primary/20 flex items-center justify-center shrink-0 mb-0.5">
          <Bot size={10} className="text-primary" />
        </div>
      )}

      <div
        className={`max-w-[85%] text-xs leading-relaxed flex flex-col gap-1.5 ${
          isUser ? "items-end" : "items-start"
        }`}
      >
        {segments.map((seg, i) =>
          seg.type === "text" ? (
            seg.content.trim() ? (
              <TextBubble key={i} isUser={isUser} content={seg.content} />
            ) : null
          ) : (
            <CodeBlock
              key={i}
              lang={seg.lang ?? "plain"}
              code={seg.content}
              onEditPrompt={
                onEditPrompt ? (snippet) => onEditPrompt(snippet, message.id) : undefined
              }
            />
          ),
        )}
      </div>

      {isUser && (
        <div className="w-5 h-5 rounded-full bg-secondary/80 flex items-center justify-center shrink-0 mb-0.5">
          <User size={10} className="text-muted-foreground" />
        </div>
      )}
    </motion.div>
  );
};

// ─────────────────────────────────────────────────────────────────
//  Subcomponents
// ─────────────────────────────────────────────────────────────────

const TextBubble = ({ isUser, content }: { isUser: boolean; content: string }) => (
  <div
    className={`px-3 py-2 whitespace-pre-wrap ${
      isUser
        ? "bg-primary/15 text-foreground rounded-2xl rounded-br-sm"
        : "bg-secondary/60 text-foreground rounded-2xl rounded-bl-sm"
    }`}
  >
    {content}
  </div>
);

interface CodeBlockProps {
  lang: string;
  code: string;
  onEditPrompt?: (snippet: string) => void;
}

const CodeBlock = ({ lang, code, onEditPrompt }: CodeBlockProps) => {
  const [copied, setCopied] = useState(false);

  const handleCopy = () => {
    navigator.clipboard.writeText(code).catch(() => {});
    setCopied(true);
    setTimeout(() => setCopied(false), 1400);
  };

  return (
    <div className="w-full glass-subtle rounded-xl overflow-hidden border border-glass-border/30">
      {/* Toolbar */}
      <div className="flex items-center justify-between px-2.5 py-1.5 border-b border-glass-border/30 bg-secondary/20">
        <span className="text-[9px] font-mono uppercase tracking-wider text-muted-foreground/70">
          {lang}
        </span>
        <div className="flex items-center gap-1">
          <button
            onClick={handleCopy}
            className="flex items-center gap-1 h-5 px-1.5 rounded text-[9px] text-muted-foreground hover:text-foreground hover:bg-secondary/40 transition-colors"
            title="Copy"
          >
            {copied ? <Check size={9} className="text-emerald-400" /> : <Copy size={9} />}
            <span>{copied ? "Copied" : "Copy"}</span>
          </button>
          {onEditPrompt && (
            <button
              onClick={() => onEditPrompt(code)}
              className="flex items-center gap-1 h-5 px-1.5 rounded text-[9px] text-primary bg-primary/10 hover:bg-primary/20 transition-colors"
              title="Edit Prompt — use this block as base for next message"
            >
              <Pencil size={9} />
              <span>Edit Prompt</span>
            </button>
          )}
        </div>
      </div>

      {/* Code */}
      <pre className="px-3 py-2 text-[10.5px] leading-relaxed font-mono text-foreground/90 overflow-x-auto max-h-[260px]">
        <code>{code}</code>
      </pre>
    </div>
  );
};

export default ChatMessage;
