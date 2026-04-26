import { useState, useEffect, useRef, useCallback } from "react";
import { useLocation } from "react-router-dom";
import { toast } from "sonner";
import { api, type GenerationMode, type GenerateResponse } from "@/lib/api";
import { parseAgentText, detectAndUnpackProject } from "@/lib/sse-parsers";
import {
  filesToCode,
  codeToFiles,
  stripMarkdownFences,
  type ProjectFiles,
  type SelectedElement,
} from "@/components/WorkspacePreview";
import {
  loadCloudProjects,
  saveCloudProject,
  deleteCloudProject,
  syncLocalToCloud,
  publishProject,
  getProjectByPrompt,
  type CloudProject,
} from "@/lib/projectSync";
import { useAuth } from "@/hooks/useAuth";
import { useLanguage } from "@/hooks/useLanguage";

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — useGeneration
//  Бизнес-логика Workspace: SSE-стриминг, чат, файлы проекта,
//  cloud sync, milestone-агрегация по агентам.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

export interface ChatMessage {
  id: string;
  role: "user" | "assistant";
  content: string;
  timestamp: Date;
}

export type MilestoneStatus = "running" | "completed" | "error";

export interface AgentMilestone {
  agent: string;
  status: MilestoneStatus;
  state?: string;
  message: string;
  progress: number;
  startedAt: Date;
  updatedAt: Date;
}

export interface FSMTransition {
  from?: string;
  to?: string;
  state?: string;
  reason?: string;
  agent?: string;
  message?: string;
  at: Date;
}

/** Canonical pipeline order of all backend agents. */
export const AGENT_PIPELINE = [
  "director",
  "researcher",
  "brain",
  "architect",
  "planner",
  "coder",
  "designer",
  "validator",
  "security",
  "tester",
  "ui_reviewer",
  "videographer",
] as const;

export type AgentPipelineId = (typeof AGENT_PIPELINE)[number];

export interface SendOptions {
  selectedElement?: SelectedElement | null;
}

export interface UseGenerationReturn {
  // Chat state
  messages: ChatMessage[];
  thinking: boolean;
  initialLoading: boolean;
  loaderStep: number;
  loaderSteps: string[];

  // Project state (single source of truth)
  projectFiles: ProjectFiles;
  setProjectFiles: (files: ProjectFiles) => void;
  currentPrompt: string;

  // Cloud projects
  savedProjects: CloudProject[];
  loadProject: (project: CloudProject) => void;
  deleteProject: (id: string) => Promise<void>;
  publishCurrent: () => Promise<string | null>;

  // Mode
  agentMode: GenerationMode;
  setAgentMode: (mode: GenerationMode) => void;

  // Milestones (agent timeline)
  milestones: AgentMilestone[];

  // FSM stream (state transitions from backend)
  fsmHistory: FSMTransition[];
  currentFSMState: string;

  // Verification Layer outcomes
  securityApproved: boolean;
  testerApproved: boolean;
  uiReviewerApproved: boolean;

  // Active agent (currently running). null when idle.
  activeAgent: AgentPipelineId | null;

  // Actions
  send: (input: string, opts?: SendOptions) => Promise<void>;
  applyTelegramExport: () => void;
}

// Russian + English action verbs for auto-detecting "agent" intent
const ACTION_VERBS = [
  "создай", "создайте", "сделай", "сделайте", "разработай", "разработайте",
  "проанализируй", "проанализируйте", "исследуй", "исследуйте",
  "напиши", "напишите", "построй", "постройте", "реализуй", "реализуйте",
  "сгенерируй", "сгенерируйте", "придумай", "оптимизируй", "улучши",
  "добавь", "внедри", "разбери", "объясни", "спроектируй",
  "create", "build", "make", "develop", "analyze", "analyse",
  "research", "write", "design", "implement", "generate",
  "optimize", "improve", "add", "fix", "explain", "refactor",
];

function autoDetectIntent(text: string, current: GenerationMode): GenerationMode {
  if (!text) return "agent";
  const lower = text.toLowerCase();
  return ACTION_VERBS.some((v) => lower.includes(v)) ? "agent" : current;
}

function buildSelectedElementContext(
  raw: string,
  selected: SelectedElement,
): string {
  const selector =
    `${selected.tag}` +
    (selected.id ? `#${selected.id}` : "") +
    (selected.classes ? `.${selected.classes.split(" ").join(".")}` : "");
  const textSnippet = selected.text ? ` с текстом "${selected.text}"` : "";
  return `В текущем коде найди элемент '${selector}'${textSnippet} и примени к нему следующее изменение: ${raw}`;
}

export function useGeneration(): UseGenerationReturn {
  const location = useLocation();
  const { user } = useAuth();
  const { t } = useLanguage();
  const initialPrompt = (location.state as { prompt?: string })?.prompt || "";

  const loaderSteps = [
    t("loader1"),
    t("loader2"),
    t("loader3"),
    t("loader4"),
    t("loader5"),
  ];

  const DEFAULT_FILES: ProjectFiles = {
    "index.html": `<html><body style='background:hsl(240,6%,7%);color:hsl(240,5%,92%);font-family:Inter,sans-serif;display:flex;align-items:center;justify-content:center;height:100vh;margin:0;'><div style='text-align:center'><p style='font-size:15px;opacity:0.7'>${t("wsDefaultPreviewTitle")}</p><p style='font-size:12px;opacity:0.35;margin-top:8px'>${t("wsDefaultPreviewSub")}</p></div></body></html>`,
  };

  // ── State ────────────────────────────────────────────────
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [thinking, setThinking] = useState(false);
  const [initialLoading, setInitialLoading] = useState(!!initialPrompt);
  const [loaderStep, setLoaderStep] = useState(0);
  const [projectFiles, setProjectFiles] = useState<ProjectFiles>(DEFAULT_FILES);
  const [savedProjects, setSavedProjects] = useState<CloudProject[]>([]);
  const [currentPrompt, setCurrentPrompt] = useState(initialPrompt);
  const [agentMode, setAgentMode] = useState<GenerationMode>("agent");
  const [milestones, setMilestones] = useState<AgentMilestone[]>([]);
  const [fsmHistory, setFSMHistory] = useState<FSMTransition[]>([]);
  const [currentFSMState, setCurrentFSMState] = useState<string>("Created");
  const [securityApproved, setSecurityApproved] = useState(false);
  const [testerApproved, setTesterApproved] = useState(false);
  const [uiReviewerApproved, setUIReviewerApproved] = useState(false);
  const [activeAgent, setActiveAgent] = useState<AgentPipelineId | null>(null);

  // ── Refs (avoid stale closures + double-init) ──────────
  const hasInitialized = useRef(false);
  const hasSynced = useRef(false);
  const generateRef = useRef<(msgs: ChatMessage[]) => Promise<void>>();
  const generateCalled = useRef(false);

  // ── Sync local → cloud on user login ───────────────────
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

  // ── Save current project to cloud ──────────────────────
  const saveCurrentProject = useCallback(
    async (files: ProjectFiles) => {
      if (!user || !currentPrompt) return;
      const code = filesToCode(files);
      if (code.includes(t("wsDefaultPreviewTitle"))) return;
      await saveCloudProject(user.id, currentPrompt, code);
      const projects = await loadCloudProjects();
      setSavedProjects(projects);
    },
    [currentPrompt, user, t],
  );

  // ── Milestone tracker: updates on every SSE status ─────
  const upsertMilestone = useCallback(
    (agent: string, status: MilestoneStatus, message: string, progress: number, state?: string) => {
      const cleanAgent = agent || "system";
      setMilestones((prev) => {
        const idx = prev.findIndex((m) => m.agent === cleanAgent);
        const now = new Date();
        if (idx === -1) {
          return [
            ...prev,
            { agent: cleanAgent, status, state, message, progress, startedAt: now, updatedAt: now },
          ];
        }
        const next = [...prev];
        next[idx] = {
          ...next[idx],
          status,
          state: state ?? next[idx].state,
          message,
          progress: Math.max(next[idx].progress, progress),
          updatedAt: now,
        };
        return next;
      });
    },
    [],
  );

  // ── Core: generate code (SSE for agent/synthesis, POST for code) ──
  const generate = useCallback(
    async (allMessages: ChatMessage[]) => {
      setThinking(true);
      const lastUser = [...allMessages].reverse().find((m) => m.role === "user");
      const userRequest =
        typeof lastUser?.content === "string" ? lastUser.content : String(lastUser?.content ?? "");

      // Context-awareness: pass existing HTML if editing real project
      const defaultHtml = DEFAULT_FILES["index.html"] ?? "";
      const currentHtml =
        typeof projectFiles["index.html"] === "string" ? projectFiles["index.html"] : "";
      const hasRealProject = currentHtml.length > 200 && currentHtml !== defaultHtml;
      const specification =
        (agentMode === "agent" || agentMode === "synthesis") &&
        hasRealProject &&
        currentHtml.length < 12000
          ? `${userRequest}\n\n--- EXISTING CODE (MODIFY this, do not rebuild from scratch) ---\n${currentHtml}`
          : userRequest;

      if (agentMode === "agent" || agentMode === "synthesis") {
        const streamStatusId = `stream-${Date.now()}`;
        const modeLabel =
          agentMode === "synthesis"
            ? "🔍 Запускаю адаптивный синтез конкурентов..."
            : "🧠 Запускаю инновационное проектирование...";
        setMessages((prev) => [
          ...prev,
          { id: streamStatusId, role: "assistant", content: modeLabel, timestamp: new Date() },
        ]);
        // Reset all per-run state
        setMilestones([]);
        setFSMHistory([]);
        setCurrentFSMState("Created");
        setSecurityApproved(false);
        setTesterApproved(false);
        setUIReviewerApproved(false);
        setActiveAgent(null);

        await Promise.race([
          new Promise<void>((resolve) => {
            api.generateProjectStream(
              { specification, mode: agentMode },
              // onStatus
              (status) => {
                const safeMsg = parseAgentText(status?.message, true);
                const normalizedStatus =
                  status.status === "completed"
                    ? "completed"
                    : status.status === "error"
                      ? "error"
                      : "running";
                upsertMilestone(
                  status.agent,
                  normalizedStatus,
                  safeMsg,
                  status.progress ?? 0,
                  status.state,
                );
                // Map SSE agent field → canonical pipeline id → activeAgent state.
                // Это реализует чёткий маппинг из требования °2. Проверка SSE-канала»:
                // event.Agent с бэкенда прямо обновляет activeAgent, а не просто падает в чат.
                const agentKey = (status.agent || "").toLowerCase().replace(/\s+/g, "_");
                if (normalizedStatus === "running") {
                  const canonical = (AGENT_PIPELINE as readonly string[]).includes(agentKey)
                    ? (agentKey as AgentPipelineId)
                    : null;
                  if (canonical) setActiveAgent(canonical);
                } else if (normalizedStatus === "completed" || normalizedStatus === "error") {
                  // Clear active only if it matches the current one — иначе следующий агент уже стартовал.
                  setActiveAgent((prev) => (prev === agentKey ? null : prev));
                }
                // Verification Layer detection: agent name + completed status
                if (normalizedStatus === "completed") {
                  if (agentKey === "security" || agentKey === "validator") {
                    setSecurityApproved(true);
                  } else if (agentKey === "tester") {
                    setTesterApproved(true);
                  } else if (agentKey === "ui_reviewer" || agentKey === "uireviewer") {
                    setUIReviewerApproved(true);
                  }
                }
                if (!safeMsg) return;
                setMessages((prev) => {
                  const idx = prev.findIndex((m) => m.id === streamStatusId);
                  if (idx === -1) return prev;
                  const updated = [...prev];
                  updated[idx] = { ...updated[idx], content: safeMsg };
                  return updated;
                });
              },
              // onResult
              async (result: GenerateResponse) => {
                setThinking(false);
                setActiveAgent(null);
                const rawFiles = result.files ?? (result.code ? { "index.html": result.code } : {});
                let files: ProjectFiles = Object.fromEntries(
                  Object.entries(rawFiles).map(([k, v]) => [k, parseAgentText(v, true)]),
                );
                if (Object.keys(files).length === 0) {
                  const codeStr = parseAgentText(result.code, true);
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
              (err: Error) => {
                setThinking(false);
                setActiveAgent(null);
                toast.error(t("wsGenError"));
                const errContent = `❌ ${parseAgentText(err?.message ?? err, false)}`;
                setMessages((prev) =>
                  prev
                    .filter((m) => m.id !== streamStatusId)
                    .concat([
                      {
                        id: Date.now().toString(),
                        role: "assistant",
                        content: errContent,
                        timestamp: new Date(),
                      },
                    ]),
                );
                resolve();
              },
              // onFSM — FSM transitions (Researching → Planning → Coding → Verified → Completed)
              (transition) => {
                const at = new Date();
                const nextState = transition.to || transition.state;
                if (nextState) setCurrentFSMState(nextState);
                setFSMHistory((prev) => [...prev, { ...transition, at }]);
              },
            );
          }),
          new Promise<void>((_, reject) =>
            setTimeout(() => reject(new Error("SSE_TIMEOUT")), 10 * 60 * 1000),
          ),
        ]).catch(() => {
          setThinking(false);
          toast.error("⏱️ Таймаут генерации (10 мин). Попробуйте еще.");
          setMessages((prev) =>
            prev
              .filter((m) => m.id !== streamStatusId)
              .concat([
                {
                  id: Date.now().toString(),
                  role: "assistant",
                  content: "⏱️ Таймаут генерации. Попробуйте ещё раз.",
                  timestamp: new Date(),
                },
              ]),
          );
        });
      } else {
        // ── CODE MODE: fast POST ────────────────────────────
        try {
          const apiMessages = allMessages.map((m) => ({ role: m.role, content: m.content }));
          const response = await api.generateFromChat(apiMessages, "code");

          if (response.files) {
            const files: ProjectFiles = Object.fromEntries(
              Object.entries(response.files).map(([k, v]) => [k, stripMarkdownFences(String(v))]),
            );
            setProjectFiles(files);
            await saveCurrentProject(files);
            toast.success(t("wsSaved"));
            const fileCount = Object.keys(files).length;
            setMessages((prev) => [
              ...prev,
              {
                id: Date.now().toString(),
                role: "assistant",
                content: `${t("wsCodeUpdated")} (${fileCount} ${fileCount === 1 ? "файл" : "файлов"})`,
                timestamp: new Date(),
              },
            ]);
          } else if (response.code) {
            const files = { "index.html": stripMarkdownFences(response.code) };
            setProjectFiles(files);
            await saveCurrentProject(files);
            toast.success(t("wsSaved"));
            setMessages((prev) => [
              ...prev,
              {
                id: Date.now().toString(),
                role: "assistant",
                content: t("wsCodeUpdated"),
                timestamp: new Date(),
              },
            ]);
          } else if (response.message) {
            const respMsg = parseAgentText(response.message, false);
            setMessages((prev) => [
              ...prev,
              {
                id: Date.now().toString(),
                role: "assistant",
                content: respMsg,
                timestamp: new Date(),
              },
            ]);
          }
        } catch (err: unknown) {
          console.error("generate-code error:", err);
          toast.error(t("wsGenError"));
          setMessages((prev) => [
            ...prev,
            {
              id: Date.now().toString(),
              role: "assistant",
              content: t("wsGenErrorRetry"),
              timestamp: new Date(),
            },
          ]);
        } finally {
          setThinking(false);
        }
      }
    },
    [agentMode, projectFiles, saveCurrentProject, t, upsertMilestone, DEFAULT_FILES],
  );

  // Keep ref in sync to avoid effect re-runs on identity change
  generateRef.current = generate;

  // ── Initial-prompt loader effect ──────────────────────
  useEffect(() => {
    if (!initialPrompt || hasInitialized.current) return;
    hasInitialized.current = true;
    generateCalled.current = false;
    const firstMsg: ChatMessage = {
      id: "1",
      role: "user",
      content: initialPrompt,
      timestamp: new Date(),
    };
    setMessages([firstMsg]);
    let step = 0;
    const interval = setInterval(() => {
      step++;
      setLoaderStep(step);
      if (step >= loaderSteps.length - 1) {
        clearInterval(interval);
        setTimeout(() => {
          setInitialLoading(false);
          if (generateCalled.current) return;
          generateCalled.current = true;
          generateRef.current?.([firstMsg]);
        }, 400);
      }
    }, 1200);
    return () => clearInterval(interval);
  }, [initialPrompt, loaderSteps.length]);

  // ── Public actions ───────────────────────────────────
  const send = useCallback(
    async (input: string, opts?: SendOptions) => {
      const trimmed = String(input || "").trim();
      if (!trimmed || thinking) return;

      // Auto-detect: maybe upgrade to agent mode
      const detected = autoDetectIntent(trimmed, agentMode);
      if (detected === "agent" && agentMode !== "agent") {
        setAgentMode("agent");
      }

      // Selected-element context
      const finalContent = opts?.selectedElement
        ? buildSelectedElementContext(trimmed, opts.selectedElement)
        : trimmed;

      const userMsg: ChatMessage = {
        id: Date.now().toString(),
        role: "user",
        content: finalContent,
        timestamp: new Date(),
      };
      if (!currentPrompt) setCurrentPrompt(trimmed);
      const updated = [...messages, userMsg];
      setMessages(updated);
      await generate(updated);
    },
    [agentMode, currentPrompt, messages, thinking, generate],
  );

  const loadProject = useCallback(
    (project: CloudProject) => {
      const files = codeToFiles(project.code);
      setProjectFiles(files);
      setCurrentPrompt(project.prompt);
      setMessages([
        {
          id: "loaded",
          role: "user",
          content: project.prompt,
          timestamp: new Date(project.created_at),
        },
        {
          id: "loaded-reply",
          role: "assistant",
          content: t("wsLoadedFromCloud"),
          timestamp: new Date(),
        },
      ]);
      setInitialLoading(false);
      toast.success(t("wsLoaded"));
    },
    [t],
  );

  const deleteProject = useCallback(
    async (id: string) => {
      await deleteCloudProject(id);
      const projects = await loadCloudProjects();
      setSavedProjects(projects);
      toast.success(t("wsDeleted"));
    },
    [t],
  );

  const publishCurrent = useCallback(async (): Promise<string | null> => {
    if (!user || !currentPrompt) return null;
    const project = await getProjectByPrompt(user.id, currentPrompt);
    if (!project) return null;
    return await publishProject(project.id);
  }, [user, currentPrompt]);

  const applyTelegramExport = useCallback(() => {
    const TWA_SCRIPT = '<script src="https://telegram.org/js/telegram-web-app.js"></script>';
    const TWA_META =
      '<meta name="viewport" content="width=device-width, initial-scale=1.0, minimum-scale=1.0, maximum-scale=1.0, user-scalable=no" />';

    const updated = { ...projectFiles };
    const html =
      typeof updated["index.html"] === "string"
        ? updated["index.html"]
        : String(updated["index.html"] ?? "");

    if (html.includes("telegram-web-app.js")) {
      toast.info("Telegram Web App скрипт уже добавлен");
      return;
    }
    if (html.includes("</head>")) {
      updated["index.html"] = html.replace(
        "</head>",
        `  ${TWA_META}\n  ${TWA_SCRIPT}\n</head>`,
      );
    } else if (html.includes("<html")) {
      updated["index.html"] = html.replace(
        "<html",
        `<html>\n<head>\n  ${TWA_META}\n  ${TWA_SCRIPT}\n</head>\n<html`,
      );
    } else {
      updated["index.html"] = `<head>\n  ${TWA_META}\n  ${TWA_SCRIPT}\n</head>\n${html}`;
    }
    setProjectFiles(updated);
    toast.success(t("wsTelegramDone"));
  }, [projectFiles, t]);

  return {
    messages,
    thinking,
    initialLoading,
    loaderStep,
    loaderSteps,
    projectFiles,
    setProjectFiles,
    currentPrompt,
    savedProjects,
    loadProject,
    deleteProject,
    publishCurrent,
    agentMode,
    setAgentMode,
    milestones,
    fsmHistory,
    currentFSMState,
    securityApproved,
    testerApproved,
    uiReviewerApproved,
    activeAgent,
    send,
    applyTelegramExport,
  };
}
