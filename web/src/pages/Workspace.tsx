import { useState, useEffect, useRef, useCallback } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { motion, AnimatePresence } from "framer-motion";
import {
  Send,
  History,
  ArrowLeft,
  Bot,
  User,
  Loader2,
  Trash2,
  Layout,
  X,
  MousePointer2,
  Brain,
  Zap,
} from "lucide-react";
import type { GenerationMode } from "@/lib/api";
import { stripMarkdownFences } from "@/components/WorkspacePreview";
import {
  SidebarProvider,
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
// import { supabase } from "@/integrations/supabase/client"; // Не используется - переход на Go Auth
import { toast } from "sonner";
import { useAuth } from "@/hooks/useAuth";
import { useLanguage } from "@/hooks/useLanguage";
import { useCredits } from "@/hooks/useCredits";
import WorkspacePreview, { ProjectFiles, filesToCode, codeToFiles, SelectedElement } from "@/components/WorkspacePreview";
import {
  loadCloudProjects,
  saveCloudProject,
  deleteCloudProject,
  syncLocalToCloud,
  publishProject,
  getProjectByPrompt,
  type CloudProject,
} from "@/lib/projectSync";

interface ChatMessage {
  id: string;
  role: "user" | "assistant";
  content: string;
  timestamp: Date;
}

/** Strip Claude 3.7 <thinking>...</thinking> blocks from any string */
function stripThinking(s: string): string {
  return s.replace(/<thinking>[\s\S]*?<\/thinking>/gi, "").trim();
}

/**
 * If content looks like a JSON project dump (keys ending in .html/.tsx/.ts/.css/.js)
 * return the parsed ProjectFiles, otherwise return null.
 */
function detectAndUnpackProject(content: string): Record<string, string> | null {
  const s = typeof content === "string" ? content.trim() : "";
  if (!s.startsWith("{")) return null;
  try {
    const parsed = JSON.parse(s) as Record<string, unknown>;
    const fileKeys = Object.keys(parsed).filter((k) =>
      /\.(html|tsx|ts|jsx|js|css|md)$/i.test(k)
    );
    if (fileKeys.length === 0) return null;
    const files: Record<string, string> = {};
    for (const k of fileKeys) {
      files[k] = safeContent(parsed[k]);
    }
    return files;
  } catch {
    return null;
  }
}

/** Normalize any value to a display string — mirrors api.ts extractMessage */
function safeContent(raw: unknown): string {
  if (raw == null) return "";
  if (typeof raw === "string") return raw;
  if (typeof raw === "number" || typeof raw === "boolean") return String(raw);
  if (typeof raw === "object") {
    const obj = raw as Record<string, unknown>;
    const pick =
      obj.text ?? obj.content ?? obj.reasoning_content ??
      obj.thinking ?? obj.message ?? obj.description ?? obj.output;
    if (pick != null && typeof pick !== "object") return String(pick);
    if (typeof pick === "object") return safeContent(pick);
    return JSON.stringify(raw);
  }
  return String(raw);
}

/** safeContent + strip Claude 3.7 <thinking> blocks in one pass */
function safeContentClean(raw: unknown): string {
  if (raw == null) return "";
  if (typeof raw === "string") return stripThinking(raw);
  if (typeof raw === "number" || typeof raw === "boolean") return String(raw);
  if (typeof raw === "object") {
    const obj = raw as Record<string, unknown>;
    const pick =
      obj.text ?? obj.content ?? obj.reasoning_content ??
      obj.thinking ?? obj.message ?? obj.description ?? obj.output;
    if (pick != null && typeof pick !== "object") return stripThinking(String(pick));
    if (typeof pick === "object") return safeContentClean(pick);
    return JSON.stringify(raw, null, 2);
  }
  return stripThinking(String(raw));
}

const Workspace = () => {
  const location = useLocation();
  const navigate = useNavigate();
  const { user } = useAuth();
  const { t } = useLanguage();
  const { setCredits } = useCredits();
  const initialPrompt = (location.state as { prompt?: string })?.prompt || "";

  const loaderSteps = [t("loader1"), t("loader2"), t("loader3"), t("loader4"), t("loader5")];

  const DEFAULT_FILES: ProjectFiles = {
    "index.html": `<html><body style='background:hsl(240,6%,7%);color:hsl(240,5%,92%);font-family:Inter,sans-serif;display:flex;align-items:center;justify-content:center;height:100vh;margin:0;'><div style='text-align:center'><p style='font-size:15px;opacity:0.7'>${t("wsDefaultPreviewTitle")}</p><p style='font-size:12px;opacity:0.35;margin-top:8px'>${t("wsDefaultPreviewSub")}</p></div></body></html>`
  };

  const [chatInput, setChatInput] = useState("");
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [thinking, setThinking] = useState(false);
  const [initialLoading, setInitialLoading] = useState(!!initialPrompt);
  const [loaderStep, setLoaderStep] = useState(0);
  const [projectFiles, setProjectFiles] = useState<ProjectFiles>(DEFAULT_FILES);
  const [savedProjects, setSavedProjects] = useState<CloudProject[]>([]);
  const [currentPrompt, setCurrentPrompt] = useState(initialPrompt);
  const [editMode, setEditMode] = useState(false);
  const [selectedElement, setSelectedElement] = useState<SelectedElement | null>(null);
  const [agentMode, setAgentMode] = useState<GenerationMode>("agent");
  const chatEndRef = useRef<HTMLDivElement>(null);
  const hasInitialized = useRef(false);
  const hasSynced = useRef(false);

  useEffect(() => {
    if (!user || hasSynced.current) return;
    hasSynced.current = true;
    (async () => {
      const synced = await syncLocalToCloud(user.id);
      if (synced > 0) toast.success(t("wsSynced", synced));
      const projects = await loadCloudProjects();
      setSavedProjects(projects);
    })();
  }, [user, t]);

  const saveCurrentProject = useCallback(
    async (files: ProjectFiles) => {
      if (!user || !currentPrompt) return;
      const code = filesToCode(files);
      if (code.includes(t("wsDefaultPreviewTitle"))) return;
      await saveCloudProject(user.id, currentPrompt, code);
      const projects = await loadCloudProjects();
      setSavedProjects(projects);
    },
    [currentPrompt, user, t]
  );

  const generateCode = useCallback(
    async (allMessages: ChatMessage[]) => {
      console.log("\u2699\ufe0f generateCode called, agentMode=", agentMode, "msgs=", allMessages.length);
      setThinking(true);
      const { api } = await import("@/lib/api");
      const lastUser = [...allMessages].reverse().find((m) => m.role === "user");
      const userRequest = typeof lastUser?.content === "string" ? lastUser.content : String(lastUser?.content ?? "");

      // Context awareness: when editing an existing project, pass current code so Coder modifies rather than rebuilds
      const defaultHtml = DEFAULT_FILES["index.html"] ?? "";
      const currentHtml = typeof projectFiles["index.html"] === "string" ? projectFiles["index.html"] : "";
      const hasRealProject = currentHtml.length > 200 && currentHtml !== defaultHtml;
      const specification = ((agentMode === "agent" || agentMode === "synthesis") && hasRealProject && currentHtml.length < 12000)
        ? `${userRequest}\n\n--- EXISTING CODE (MODIFY this, do not rebuild from scratch) ---\n${currentHtml}`
        : userRequest;

      if (agentMode === "agent" || agentMode === "synthesis") {
        // ── AGENT / SYNTHESIS MODE: SSE streaming с мультимодальными статусами ──
        const streamStatusId = `stream-${Date.now()}`;
        const modeLabel = agentMode === "synthesis" ? "🔍 Запускаю адаптивный синтез конкурентов..." : "🧠 Запускаю инновационное проектирование...";
        setMessages((prev) => [
          ...prev,
          { id: streamStatusId, role: "assistant", content: modeLabel, timestamp: new Date() },
        ]);

        console.log("🚀 SSE: запуск generateProjectStream, mode=", agentMode, "spec_len=", specification.length);
        console.log("DEBUG 2: Данные готовы к отправке", { specification: specification.substring(0, 100), mode: agentMode, baseURL: (api as any).baseURL });
        await new Promise<void>((resolve) => {
          api.generateProjectStream(
            { specification, mode: agentMode },
            // onStatus — обновляем последнее сообщение агента
            (status) => {
              console.log("📡 SSE onStatus:", status?.agent, status?.status, status?.message, "progress=", status?.progress);
              const safeMsg = safeContentClean(status?.message);
              if (!safeMsg) return;
              setMessages((prev) => {
                const idx = prev.findIndex((m) => m.id === streamStatusId);
                if (idx === -1) return prev;
                const updated = [...prev];
                updated[idx] = { ...updated[idx], content: safeMsg };
                return updated;
              });
            },
            // onResult — финальный результат
            async (result) => {
              console.log("🎉 SSE onResult:", Object.keys(result?.files ?? {}), "duration=", result?.duration);
              setThinking(false);
              // Coerce every file value to string, strip thinking blocks
              const rawFiles = result.files ?? (result.code ? { "index.html": result.code } : {});
              let files: ProjectFiles = Object.fromEntries(
                Object.entries(rawFiles).map(([k, v]) => [k, safeContentClean(v)])
              );
              // Safety net: if files map is empty, check if code is a JSON project dump
              if (Object.keys(files).length === 0) {
                const codeStr = safeContentClean(result.code);
                const unpacked = detectAndUnpackProject(codeStr);
                if (unpacked) files = unpacked;
              }
              if (Object.keys(files).length > 0) {
                setProjectFiles(files);
                await saveCurrentProject(files);
                toast.success(t("wsSaved"));
              }
              const doneContent = `🎉 Мультимодальный проект готов! (${Object.keys(files).length} файлов)`;
              setMessages((prev) => [
                ...prev.filter((m) => m.id !== streamStatusId),
                {
                  id: Date.now().toString(),
                  role: "assistant",
                  content: doneContent,
                  timestamp: new Date(),
                },
              ]);
              resolve();
            },
            // onError
            (err) => {
              console.error("🚨 SSE onError:", err?.message || err);
              setThinking(false);
              toast.error(t("wsGenError"));
              const errContent = `❌ ${safeContent(err?.message ?? err)}`;
              setMessages((prev) => prev.filter((m) => m.id !== streamStatusId).concat([
                { id: Date.now().toString(), role: "assistant", content: errContent, timestamp: new Date() },
              ]));
              resolve();
            }
          );
        });
      } else {
        // ── CODE MODE: быстрый POST → DeepSeek-V3 ──
        try {
          const apiMessages = allMessages.map((m) => ({ role: m.role, content: m.content }));
          const response = await api.generateFromChat(apiMessages, "code");

          if (response.files) {
            // Strip markdown fences from every file value
            const files: ProjectFiles = Object.fromEntries(
              Object.entries(response.files).map(([k, v]) => [k, stripMarkdownFences(String(v))])
            );
            setProjectFiles(files);
            await saveCurrentProject(files);
            toast.success(t("wsSaved"));
            const fileCount = Object.keys(files).length;
            setMessages((prev) => [
              ...prev,
              { id: Date.now().toString(), role: "assistant", content: `${t("wsCodeUpdated")} (${fileCount} ${fileCount === 1 ? "файл" : "файлов"})`, timestamp: new Date() },
            ]);
          } else if (response.code) {
            const files = { "index.html": stripMarkdownFences(response.code) };
            setProjectFiles(files);
            await saveCurrentProject(files);
            toast.success(t("wsSaved"));
            setMessages((prev) => [
              ...prev,
              { id: Date.now().toString(), role: "assistant", content: t("wsCodeUpdated"), timestamp: new Date() },
            ]);
          } else if (response.message) {
            const respMsg = safeContent(response.message);
            setMessages((prev) => [
              ...prev,
              { id: Date.now().toString(), role: "assistant", content: respMsg, timestamp: new Date() },
            ]);
          }
        } catch (err: any) {
          console.error("generate-code error:", err);
          toast.error(t("wsGenError"));
          setMessages((prev) => [
            ...prev,
            { id: Date.now().toString(), role: "assistant", content: t("wsGenErrorRetry"), timestamp: new Date() },
          ]);
        } finally {
          setThinking(false);
        }
      }
    },
    [saveCurrentProject, t, setCredits, agentMode, projectFiles, DEFAULT_FILES]
  );

  useEffect(() => {
    if (!initialPrompt || hasInitialized.current) return;
    hasInitialized.current = true;
    const firstMsg: ChatMessage = { id: "1", role: "user", content: initialPrompt, timestamp: new Date() };
    setMessages([firstMsg]);
    const interval = setInterval(() => {
      setLoaderStep((prev) => {
        if (prev >= loaderSteps.length - 1) {
          clearInterval(interval);
          setTimeout(() => {
            setInitialLoading(false);
            generateCode([firstMsg]);
          }, 400);
          return prev;
        }
        return prev + 1;
      });
    }, 1200);
    return () => clearInterval(interval);
  }, [initialPrompt, generateCode, loaderSteps.length]);

  useEffect(() => {
    chatEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages, thinking]);

  // Clear selected element when edit mode is turned off
  useEffect(() => {
    if (!editMode) setSelectedElement(null);
  }, [editMode]);

  const autoDetectIntent = (text: unknown): GenerationMode => {
    const safeText = String(text || "");
    if (!safeText) return "agent";
    const actionVerbs = [
      // Russian
      "создай", "создайте", "сделай", "сделайте", "разработай", "разработайте",
      "проанализируй", "проанализируйте", "исследуй", "исследуйте",
      "напиши", "напишите", "построй", "постройте", "реализуй", "реализуйте",
      "сгенерируй", "сгенерируйте", "придумай", "оптимизируй", "улучши",
      "добавь", "внедри", "разбери", "объясни", "спроектируй",
      // English
      "create", "build", "make", "develop", "analyze", "analyse",
      "research", "write", "design", "implement", "generate",
      "optimize", "improve", "add", "fix", "explain", "refactor",
    ];
    const lower = safeText.toLowerCase();
    return actionVerbs.some((v) => lower.includes(v)) ? "agent" : agentMode;
  };

  const handleSend = async () => {
    const safeInput = typeof chatInput === "string" ? chatInput : "";
    console.log("\u270f\ufe0f handleSend:", { input: safeInput.substring(0, 50), thinking, agentMode });
    if (!safeInput.trim() || thinking) return;

    // Auto-detect intent and upgrade to agent mode if needed
    const detectedMode = autoDetectIntent(safeInput);
    if (detectedMode === "agent" && agentMode !== "agent") {
      setAgentMode("agent");
    }

    // If element is selected, prepend context
    let finalContent = safeInput;
    if (selectedElement) {
      const selector = `${selectedElement.tag}${selectedElement.id ? '#' + selectedElement.id : ''}${selectedElement.classes ? '.' + selectedElement.classes.split(' ').join('.') : ''}`;
      const textSnippet = selectedElement.text ? ` с текстом "${selectedElement.text}"` : '';
      finalContent = `В текущем коде найди элемент '${selector}'${textSnippet} и примени к нему следующее изменение: ${safeInput}`;
      setSelectedElement(null);
      setEditMode(false);
    }
    
    const userMsg: ChatMessage = { id: Date.now().toString(), role: "user", content: finalContent, timestamp: new Date() };
    if (!currentPrompt) setCurrentPrompt(safeInput);
    const updated = [...messages, userMsg];
    setMessages(updated);
    setChatInput("");
    await generateCode(updated);
  };

  const handleLoadProject = (project: CloudProject) => {
    const files = codeToFiles(project.code);
    setProjectFiles(files);
    setCurrentPrompt(project.prompt);
    setMessages([
      { id: "loaded", role: "user", content: project.prompt, timestamp: new Date(project.created_at) },
      { id: "loaded-reply", role: "assistant", content: t("wsLoadedFromCloud"), timestamp: new Date() },
    ]);
    setInitialLoading(false);
    toast.success(t("wsLoaded"));
  };

  const handleDeleteProject = async (id: string) => {
    await deleteCloudProject(id);
    const projects = await loadCloudProjects();
    setSavedProjects(projects);
    toast.success(t("wsDeleted"));
  };

  const handleTelegramExport = useCallback(() => {
    const TWA_SCRIPT = '<script src="https://telegram.org/js/telegram-web-app.js"></script>';
    const TWA_META = '<meta name="viewport" content="width=device-width, initial-scale=1.0, minimum-scale=1.0, maximum-scale=1.0, user-scalable=no" />';
    
    const updated = { ...projectFiles };
    const html = typeof updated["index.html"] === "string" ? updated["index.html"] : String(updated["index.html"] ?? "");
    
    if (html.includes("telegram-web-app.js")) {
      toast.info("Telegram Web App скрипт уже добавлен");
      return;
    }

    if (html.includes("</head>")) {
      updated["index.html"] = html.replace("</head>", `  ${TWA_META}\n  ${TWA_SCRIPT}\n</head>`);
    } else if (html.includes("<html")) {
      updated["index.html"] = html.replace("<html", `<html>\n<head>\n  ${TWA_META}\n  ${TWA_SCRIPT}\n</head>\n<html`);
    } else {
      updated["index.html"] = `<head>\n  ${TWA_META}\n  ${TWA_SCRIPT}\n</head>\n${html}`;
    }
    
    setProjectFiles(updated);
    toast.success(t("wsTelegramDone"));
  }, [projectFiles, t]);

  return (
    <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ duration: 0.4 }} className="h-screen flex flex-col overflow-hidden bg-background">
      <SidebarProvider defaultOpen={true}>
        <div className="flex-1 flex w-full overflow-hidden">
          <Sidebar className="border-r border-[hsl(var(--border))]/10" collapsible="offcanvas">
            <SidebarHeader className="border-b border-[hsl(var(--border))]/10 px-3 py-3 space-y-3">
              <div className="flex items-center gap-2">
                <button onClick={() => navigate("/")} className="flex items-center gap-1 text-muted-foreground hover:text-foreground transition-colors text-xs">
                  <ArrowLeft size={13} />
                  <span>{t("back")}</span>
                </button>
                <div className="w-px h-4 bg-border/30" />
                <span className="text-xs font-medium text-foreground truncate">
                  {currentPrompt ? currentPrompt.slice(0, 30) + (currentPrompt.length > 30 ? "…" : "") : t("wsNewProject")}
                </span>
              </div>

              {/* ── Режим работы ─────────────────── */}
              <div className="space-y-1.5">
                <div className="flex items-center gap-1 p-0.5 rounded-lg bg-secondary/40 border border-border/20">
                  <button
                    onClick={() => setAgentMode("agent")}
                    className={`flex-1 flex items-center justify-center gap-1.5 px-1.5 py-1.5 rounded-md text-[10px] font-medium transition-all ${
                      agentMode === "agent"
                        ? "bg-violet-600/90 text-white shadow-sm shadow-violet-900/40"
                        : "text-muted-foreground hover:text-foreground"
                    }`}
                  >
                    <Brain size={10} />
                    ИННОВ.
                  </button>
                  <button
                    onClick={() => setAgentMode("synthesis")}
                    className={`flex-1 flex items-center justify-center gap-1.5 px-1.5 py-1.5 rounded-md text-[10px] font-medium transition-all ${
                      agentMode === "synthesis"
                        ? "bg-amber-600/90 text-white shadow-sm shadow-amber-900/40"
                        : "text-muted-foreground hover:text-foreground"
                    }`}
                  >
                    <Layout size={10} />
                    СИНТЕЗ
                  </button>
                  <button
                    onClick={() => setAgentMode("code")}
                    className={`flex-1 flex items-center justify-center gap-1.5 px-1.5 py-1.5 rounded-md text-[10px] font-medium transition-all ${
                      agentMode === "code"
                        ? "bg-sky-600/90 text-white shadow-sm shadow-sky-900/40"
                        : "text-muted-foreground hover:text-foreground"
                    }`}
                  >
                    <Zap size={10} />
                    КОД
                  </button>
                </div>
                <p className="text-[10px] text-muted-foreground/60 px-0.5 leading-relaxed">
                  {agentMode === "agent"
                    ? "🧠 Инновационное проектирование · Claude Opus 4.6 · Reasoning"
                    : agentMode === "synthesis"
                    ? "🔍 Адаптивный синтез конкурентов · DeepSeek V3.2 + Claude Opus"
                    : "⚡ Быстрая генерация · Claude Opus 4.6"}
                </p>
                {/* ── Калькулятор кредитов ── */}
                <div className="bg-secondary/30 rounded-md p-1.5 border border-border/10">
                  <p className="text-[9px] font-medium text-muted-foreground/70 mb-1">Стоимость генерации:</p>
                  <div className="flex flex-wrap gap-x-3 gap-y-0.5 text-[9px] text-muted-foreground/50">
                    <span>🎬 Видео <span className="text-amber-400/80 font-semibold">+50</span></span>
                    <span>🎨 Картинки <span className="text-amber-400/80 font-semibold">+20</span></span>
                    <span>🔍 Синтез <span className="text-amber-400/80 font-semibold">+30</span></span>
                  </div>
                  <p className="text-[9px] text-muted-foreground/40 mt-1">
                    Итого: <span className="text-white/70 font-semibold">
                      {agentMode === "agent" ? "~70 кр" : agentMode === "synthesis" ? "~100 кр" : "~20 кр"}
                    </span>
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
                      <div className="px-3 py-2 text-[11px] text-muted-foreground/40">{t("wsNoSaved")}</div>
                    ) : (
                      savedProjects.map((project) => (
                        <SidebarMenuItem key={project.id}>
                          <div className="flex items-center group">
                            <SidebarMenuButton className="text-xs flex-1" onClick={() => handleLoadProject(project)}>
                              <History size={12} className="text-primary shrink-0" />
                              <span className="truncate">{project.prompt}</span>
                            </SidebarMenuButton>
                            <button
                              onClick={(e) => { e.stopPropagation(); handleDeleteProject(project.id); }}
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
                  <motion.div key={msg.id} initial={{ opacity: 0, y: 6 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.2 }}
                    className={`flex items-end gap-1.5 ${msg.role === "user" ? "justify-end" : "justify-start"}`}>
                    {msg.role === "assistant" && (
                      <div className="w-5 h-5 rounded-full bg-primary/20 flex items-center justify-center shrink-0 mb-0.5"><Bot size={10} className="text-primary" /></div>
                    )}
                    <div className={`max-w-[85%] px-3 py-2 text-xs leading-relaxed ${
                      msg.role === "user" ? "bg-primary/15 text-foreground rounded-2xl rounded-br-sm" : "bg-secondary/60 text-foreground rounded-2xl rounded-bl-sm"
                    }` }>{typeof msg.content === "string" ? msg.content : JSON.stringify(msg.content, null, 2)}</div>
                    {msg.role === "user" && (
                      <div className="w-5 h-5 rounded-full bg-secondary/80 flex items-center justify-center shrink-0 mb-0.5"><User size={10} className="text-muted-foreground" /></div>
                    )}
                  </motion.div>
                ))}
                {thinking && (
                  <motion.div initial={{ opacity: 0, y: 6 }} animate={{ opacity: 1, y: 0 }} className="flex items-end gap-1.5">
                    <div className="w-5 h-5 rounded-full bg-primary/20 flex items-center justify-center shrink-0 mb-0.5"><Bot size={10} className="text-primary" /></div>
                    <div className="bg-secondary/60 rounded-2xl rounded-bl-sm px-3 py-2.5 flex items-center gap-2">
                      <Loader2 size={12} className="text-primary animate-spin" />
                      <span className="text-xs text-muted-foreground">{t("wsThinking")}</span>
                    </div>
                  </motion.div>
                )}
                <div ref={chatEndRef} />
              </div>
            </SidebarContent>

            <SidebarFooter className="border-t border-[hsl(var(--border))]/10 p-3">
              {/* Selected element indicator */}
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
                        &lt;{selectedElement.tag}&gt; {selectedElement.text && `"${selectedElement.text.slice(0, 30)}..."`}
                      </p>
                    </div>
                    <button
                      onClick={() => setSelectedElement(null)}
                      className="w-5 h-5 rounded flex items-center justify-center text-muted-foreground hover:text-foreground transition-colors"
                    >
                      <X size={10} />
                    </button>
                  </motion.div>
                )}
              </AnimatePresence>

              <div className="flex items-center gap-2 glass-subtle rounded-xl px-3 py-2">
                <input
                  value={chatInput}
                  onChange={(e) => setChatInput(String(e.target.value ?? ""))}
                  onKeyDown={(e) => e.key === "Enter" && !e.shiftKey && handleSend()}
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
                  onClick={handleSend}
                  disabled={!chatInput.trim() || thinking}
                  className={`w-6 h-6 rounded-lg flex items-center justify-center transition-colors ${
                    chatInput.trim() && !thinking ? "bg-primary text-primary-foreground" : "text-muted-foreground/30"
                  }`}
                >
                  <Send size={11} />
                </button>
              </div>
              <button
                onClick={() => navigate("/")}
                className="w-full flex items-center gap-2 px-3 py-2 text-xs text-muted-foreground hover:text-foreground rounded-lg hover:bg-secondary/50 transition-colors mt-1"
              >
                <Layout size={12} />
                <span>{t("wsTemplates")}</span>
              </button>
            </SidebarFooter>
          </Sidebar>

          <WorkspacePreview
            projectFiles={projectFiles}
            onFilesChange={setProjectFiles}
            initialLoading={initialLoading}
            loaderStep={loaderStep}
            loaderSteps={loaderSteps}
            editMode={editMode}
            onEditModeChange={setEditMode}
            onElementSelect={setSelectedElement}
            onTelegramExport={handleTelegramExport}
            onPublish={async () => {
              if (!user || !currentPrompt) return null;
              const project = await getProjectByPrompt(user.id, currentPrompt);
              if (!project) return null;
              return await publishProject(project.id);
            }}
          />
        </div>
      </SidebarProvider>
    </motion.div>
  );
};

export default Workspace;
