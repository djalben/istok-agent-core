import { useEffect, useRef, useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import {
  Send,
  History,
  ArrowLeft,
  Bot,
  Loader2,
  Trash2,
  Layout,
  X,
  MousePointer2,
  Brain,
  Zap,
} from "lucide-react";
import {
  Sidebar,
  SidebarContent,
  SidebarHeader,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupLabel,
  SidebarGroupContent,
  SidebarMenu,
  SidebarMenuItem,
  SidebarMenuButton,
} from "@/components/ui/sidebar";
import type { GenerationMode } from "@/lib/api";
import type { SelectedElement } from "@/components/WorkspacePreview";
import type { CloudProject } from "@/lib/projectSync";
import type { ChatMessage as ChatMessageType } from "@/hooks/useGeneration";
import { useLanguage } from "@/hooks/useLanguage";
import ChatMessage from "./ChatMessage";
import VersionHistoryDialog from "./VersionHistoryDialog";

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — ChatPanel
//  Чисто презентационный компонент: чат + ввод + переключатель режимов.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

export interface ChatPanelProps {
  // Chat
  messages: ChatMessageType[];
  thinking: boolean;
  chatInput: string;
  onChatInputChange: (value: string) => void;
  onSend: () => void;

  // Mode
  agentMode: GenerationMode;
  onModeChange: (mode: GenerationMode) => void;

  // Saved projects
  savedProjects: CloudProject[];
  onLoadProject: (project: CloudProject) => void;
  onDeleteProject: (id: string) => void;

  // Selected element (edit mode)
  selectedElement: SelectedElement | null;
  onClearSelectedElement: () => void;

  // Header
  currentPrompt: string;
  onNavigateBack: () => void;
  onNavigateTemplates: () => void;
}

/** Build a polished prompt template from a code snippet for [Edit Prompt]. */
function buildEditPromptTemplate(snippet: string): string {
  const lines = snippet.split(/\r?\n/);
  const head = lines.slice(0, 12).join("\n");
  const truncated = lines.length > 12 ? `${head}\n// ... (усечёно ${lines.length - 12} строк)` : head;
  return `Отрефакторь следующий блок, сохранив семантику:\n\n\`\`\`\n${truncated}\n\`\`\`\n\nИзменения: `;
}

const MODE_LABELS: Record<GenerationMode, string> = {
  agent: "🧠 Инновационное проектирование · Gemini 3 Pro · Reasoning",
  synthesis: "🔍 Адаптивный синтез конкурентов · DeepSeek V3.2 + Gemini 3 Pro",
  code: "⚡ Быстрая генерация · Gemini 3 Pro",
};

const MODE_COSTS: Record<GenerationMode, string> = {
  agent: "~70 кр",
  synthesis: "~100 кр",
  code: "~20 кр",
};

const ChatPanel = ({
  messages,
  thinking,
  chatInput,
  onChatInputChange,
  onSend,
  agentMode,
  onModeChange,
  savedProjects,
  onLoadProject,
  onDeleteProject,
  selectedElement,
  onClearSelectedElement,
  currentPrompt,
  onNavigateBack,
  onNavigateTemplates,
}: ChatPanelProps) => {
  const { t } = useLanguage();
  const chatEndRef = useRef<HTMLDivElement>(null);
  const [versionHistoryOpen, setVersionHistoryOpen] = useState(false);

  useEffect(() => {
    chatEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages, thinking]);

  const handleEditPrompt = (snippet: string) => {
    onChatInputChange(buildEditPromptTemplate(snippet));
  };

  return (
    <Sidebar
      className="glass-panel border-r border-glass-border/40"
      collapsible="offcanvas"
    >
      <SidebarHeader className="border-b border-glass-border/30 px-3 py-3 space-y-3 bg-transparent">
        <div className="flex items-center gap-2">
          <button
            onClick={onNavigateBack}
            className="flex items-center gap-1 text-muted-foreground hover:text-foreground transition-colors text-xs"
          >
            <ArrowLeft size={13} />
            <span>{t("back")}</span>
          </button>
          <div className="w-px h-4 bg-border/30" />
          <span className="text-xs font-medium text-foreground truncate flex-1">
            {currentPrompt
              ? currentPrompt.slice(0, 30) + (currentPrompt.length > 30 ? "…" : "")
              : t("wsNewProject")}
          </span>
          <button
            onClick={() => setVersionHistoryOpen(true)}
            className="flex items-center gap-1 h-5 px-1.5 rounded text-[10px] text-muted-foreground hover:text-foreground hover:bg-secondary/40 transition-colors"
            title="Version History"
          >
            <History size={10} />
            <span className="hidden lg:inline">History</span>
          </button>
        </div>

        {/* ── Mode switcher ──────────────── */}
        <div className="space-y-1.5">
          <div className="flex items-center gap-1 p-0.5 rounded-lg glass-subtle">
            <ModeButton
              active={agentMode === "agent"}
              accent="violet"
              icon={<Brain size={10} />}
              label="ИННОВ."
              onClick={() => onModeChange("agent")}
            />
            <ModeButton
              active={agentMode === "synthesis"}
              accent="amber"
              icon={<Layout size={10} />}
              label="СИНТЕЗ"
              onClick={() => onModeChange("synthesis")}
            />
            <ModeButton
              active={agentMode === "code"}
              accent="sky"
              icon={<Zap size={10} />}
              label="КОД"
              onClick={() => onModeChange("code")}
            />
          </div>
          <p className="text-[10px] text-muted-foreground/60 px-0.5 leading-relaxed">
            {MODE_LABELS[agentMode]}
          </p>

          {/* Credit calc */}
          <div className="glass-subtle rounded-md p-1.5">
            <p className="text-[9px] font-medium text-muted-foreground/70 mb-1">
              Стоимость генерации:
            </p>
            <div className="flex flex-wrap gap-x-3 gap-y-0.5 text-[9px] text-muted-foreground/50">
              <span>🎬 Видео <span className="text-amber-400/80 font-semibold">+50</span></span>
              <span>🎨 Картинки <span className="text-amber-400/80 font-semibold">+20</span></span>
              <span>🔍 Синтез <span className="text-amber-400/80 font-semibold">+30</span></span>
            </div>
            <p className="text-[9px] text-muted-foreground/40 mt-1">
              Итого:{" "}
              <span className="text-white/70 font-semibold">{MODE_COSTS[agentMode]}</span>
            </p>
          </div>
        </div>
      </SidebarHeader>

      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupLabel className="text-[10px] tracking-widest uppercase text-muted-foreground/50">
            {t("wsMyProjects")}
          </SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {savedProjects.length === 0 ? (
                <div className="px-3 py-2 text-[11px] text-muted-foreground/40">
                  {t("wsNoSaved")}
                </div>
              ) : (
                savedProjects.map((project) => (
                  <SidebarMenuItem key={project.id}>
                    <div className="flex items-center group">
                      <SidebarMenuButton
                        className="text-xs flex-1"
                        onClick={() => onLoadProject(project)}
                      >
                        <History size={12} className="text-primary shrink-0" />
                        <span className="truncate">{project.prompt}</span>
                      </SidebarMenuButton>
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          onDeleteProject(project.id);
                        }}
                        className="opacity-0 group-hover:opacity-100 w-6 h-6 flex items-center justify-center text-muted-foreground hover:text-destructive transition-all shrink-0"
                      >
                        <Trash2 size={11} />
                      </button>
                    </div>
                  </SidebarMenuItem>
                ))
              )}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>

        <div className="flex-1 overflow-y-auto px-3 py-2 space-y-3">
          {messages.map((msg) => (
            <ChatMessage
              key={msg.id}
              message={msg}
              onEditPrompt={(snippet) => handleEditPrompt(snippet)}
            />
          ))}
          {thinking && (
            <motion.div
              initial={{ opacity: 0, y: 6 }}
              animate={{ opacity: 1, y: 0 }}
              className="flex items-end gap-1.5"
            >
              <div className="w-5 h-5 rounded-full bg-primary/20 flex items-center justify-center shrink-0 mb-0.5">
                <Bot size={10} className="text-primary" />
              </div>
              <div className="bg-secondary/60 rounded-2xl rounded-bl-sm px-3 py-2.5 flex items-center gap-2">
                <Loader2 size={12} className="text-primary animate-spin" />
                <span className="text-xs text-muted-foreground">{t("wsThinking")}</span>
              </div>
            </motion.div>
          )}
          <div ref={chatEndRef} />
        </div>
      </SidebarContent>

      <SidebarFooter className="border-t border-glass-border/30 p-3 bg-transparent">
        <AnimatePresence>
          {selectedElement && (
            <motion.div
              initial={{ opacity: 0, height: 0 }}
              animate={{ opacity: 1, height: "auto" }}
              exit={{ opacity: 0, height: 0 }}
              className="mb-2 px-3 py-2 rounded-lg bg-primary/10 border border-primary/20 flex items-center gap-2"
            >
              <MousePointer2 size={12} className="text-primary shrink-0" />
              <div className="flex-1 min-w-0">
                <p className="text-[10px] text-primary font-medium">Выбранный элемент</p>
                <p className="text-[11px] text-muted-foreground truncate">
                  &lt;{selectedElement.tag}&gt;{" "}
                  {selectedElement.text && `"${selectedElement.text.slice(0, 30)}..."`}
                </p>
              </div>
              <button
                onClick={onClearSelectedElement}
                className="w-5 h-5 rounded flex items-center justify-center text-muted-foreground hover:text-foreground transition-colors"
              >
                <X size={10} />
              </button>
            </motion.div>
          )}
        </AnimatePresence>

        <div className="flex items-center gap-2 glass rounded-xl px-3 py-2 transition-shadow focus-within:shadow-glow">
          <input
            value={chatInput}
            onChange={(e) => onChatInputChange(String(e.target.value ?? ""))}
            onKeyDown={(e) => e.key === "Enter" && !e.shiftKey && onSend()}
            placeholder={
              selectedElement
                ? "Опишите изменение для элемента..."
                : thinking
                ? t("wsGenerating")
                : t("wsPlaceholder")
            }
            disabled={thinking}
            className="flex-1 bg-transparent text-xs text-foreground outline-none placeholder:text-muted-foreground/50 disabled:opacity-50"
          />
          <button
            onClick={onSend}
            disabled={!chatInput.trim() || thinking}
            className={`w-6 h-6 rounded-lg flex items-center justify-center transition-colors ${
              chatInput.trim() && !thinking
                ? "bg-primary text-primary-foreground"
                : "text-muted-foreground/30"
            }`}
          >
            <Send size={11} />
          </button>
        </div>
        <button
          onClick={onNavigateTemplates}
          className="w-full flex items-center gap-2 px-3 py-2 text-xs text-muted-foreground hover:text-foreground rounded-lg hover:bg-secondary/50 transition-colors mt-1"
        >
          <Layout size={12} />
          <span>{t("wsTemplates")}</span>
        </button>
      </SidebarFooter>

      <VersionHistoryDialog
        open={versionHistoryOpen}
        projects={savedProjects}
        onClose={() => setVersionHistoryOpen(false)}
        onRestore={onLoadProject}
        onDelete={onDeleteProject}
      />
    </Sidebar>
  );
};

interface ModeButtonProps {
  active: boolean;
  accent: "violet" | "amber" | "sky";
  icon: React.ReactNode;
  label: string;
  onClick: () => void;
}

const MODE_ACCENT_CLASSES: Record<ModeButtonProps["accent"], string> = {
  violet: "bg-violet-600/90 text-white shadow-sm shadow-violet-900/40",
  amber: "bg-amber-600/90 text-white shadow-sm shadow-amber-900/40",
  sky: "bg-sky-600/90 text-white shadow-sm shadow-sky-900/40",
};

const ModeButton = ({ active, accent, icon, label, onClick }: ModeButtonProps) => (
  <button
    onClick={onClick}
    className={`flex-1 flex items-center justify-center gap-1.5 px-1.5 py-1.5 rounded-md text-[10px] font-medium transition-all ${
      active ? MODE_ACCENT_CLASSES[accent] : "text-muted-foreground hover:text-foreground"
    }`}
  >
    {icon}
    {label}
  </button>
);

export default ChatPanel;
