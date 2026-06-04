import { useState, useEffect, useCallback, useMemo } from "react";
import { useNavigate } from "react-router-dom";
import { motion, AnimatePresence } from "framer-motion";
import { toast } from "sonner";
import { ArrowLeft, Code2, Eye } from "lucide-react";
import { ResizablePanelGroup, ResizablePanel, ResizableHandle } from "@/components/ui/resizable";
import WorkspacePreview, { type SelectedElement } from "@/components/WorkspacePreview";
import { BuilderChatPanel } from "@/components/builder/BuilderChatPanel";
import { BuilderCodePanel } from "@/components/builder/BuilderCodePanel";
import { AgentPulse } from "@/components/builder/AgentPulse";
import SecurityAuditOverlay from "@/components/workspace/SecurityAuditOverlay";
import FeatureApprovalModal from "@/components/workspace/ArchitectureApprovalModal";
import MediaApprovalModal from "@/components/workspace/MediaApprovalModal";
import InsufficientFundsOverlay from "@/components/workspace/InsufficientFundsOverlay";
import { useGeneration } from "@/hooks/useGeneration";
import { milestonesToAgents } from "@/lib/builderTypes";
import { api } from "@/lib/api";

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — Workspace (тонкий shell)
//  Композиция четырёх независимых модулей:
//    • useGeneration  — хук SSE + cloud sync + activeAgent
//    • ChatPanel      — левая панель: чат + ввод
//    • PreviewPanel   — центр: live preview + Deploy / Audit
//    • MilestonesPanel — правая панель: 10-агентный таймлайн + Verified
//  Никакой бизнес-логики здесь — только связывание данных и UI-состояний.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

/**
 * Resolve target file for surgical editing.
 * If componentName is provided, tries common paths. Otherwise falls back to index.html.
 */
function resolveTargetFile(files: Record<string, string>, componentName: string | null): string | null {
  if (!componentName) {
    // Single-file projects: edit index.html directly
    if (files["index.html"]) return "index.html";
    return null;
  }
  // Try common React file paths
  const candidates = [
    `src/components/${componentName}.tsx`,
    `src/components/${componentName}.jsx`,
    `src/components/ui/${componentName}.tsx`,
    `src/components/ui/${componentName}.jsx`,
    `src/${componentName}.tsx`,
    `components/${componentName}.tsx`,
    `${componentName}.tsx`,
    `${componentName}.jsx`,
  ];
  for (const c of candidates) {
    if (files[c]) return c;
  }
  // Fuzzy match: find any file whose name contains the componentName (case-insensitive)
  const lower = componentName.toLowerCase();
  const match = Object.keys(files).find((f) => f.toLowerCase().includes(lower));
  if (match) return match;
  // Last resort: if only index.html exists
  if (files["index.html"]) return "index.html";
  return null;
}

const Workspace = () => {
  const navigate = useNavigate();
  const {
    messages,
    thinking,
    initialLoading,
    hydrating,
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
    currentFSMState,
    securityApproved,
    testerApproved,
    uiReviewerApproved,
    activeAgent,
    streamedFiles,
    canResume,
    resumeGeneration,
    send,
    applyTelegramExport,
  } = useGeneration();

  // ── UI state (purely presentational) ─────────────────
  const [chatInput, setChatInput] = useState("");
  const [editMode, setEditMode] = useState(false);
  const [selectedElement, setSelectedElement] = useState<SelectedElement | null>(null);
  const [deploying, setDeploying] = useState(false);
  const [securityAuditOpen, setSecurityAuditOpen] = useState(false);
  const [rightView, setRightView] = useState<"preview" | "code">("preview");

  // Live agents derived from SSE milestones (Director → Researcher → … pulse)
  const agents = useMemo(
    () => milestonesToAgents(milestones, activeAgent),
    [milestones, activeAgent],
  );

  useEffect(() => {
    if (!editMode) setSelectedElement(null);
  }, [editMode]);

  // Listen for Inspector floating panel "Apply" button — surgical single-file edit
  useEffect(() => {
    const handler = async (e: Event) => {
      const detail = (e as CustomEvent).detail;
      if (!detail?.instruction) return;
      const el = detail.element as SelectedElement | null;

      // Resolve target file from projectFiles
      const targetFile = resolveTargetFile(projectFiles, el?.componentName || null);
      if (!targetFile) {
        // Fallback: send via full generation pipeline
        setChatInput(detail.instruction);
        setTimeout(() => {
          setChatInput("");
          setSelectedElement(null);
          setEditMode(false);
          send(detail.instruction, { selectedElement: el });
        }, 50);
        return;
      }

      // Surgical edit via dedicated endpoint
      setSelectedElement(null);
      setEditMode(false);
      try {
        const result = await api.editComponent(targetFile, projectFiles[targetFile], detail.instruction);
        if (result.newCode) {
          setProjectFiles({ ...projectFiles, [result.filePath]: result.newCode });
          toast.success(`Файл ${result.filePath} обновлён`);
        }
      } catch (err: unknown) {
        const msg = err instanceof Error ? err.message : "Unknown error";
        toast.error(`Ошибка редактирования: ${msg}`);
      }
    };
    window.addEventListener("istok-inspector-apply", handler);
    return () => window.removeEventListener("istok-inspector-apply", handler);
  }, [send, projectFiles, setProjectFiles]);

  const handleSend = async () => {
    if (!chatInput.trim() || thinking) return;
    const opts = { selectedElement };
    setChatInput("");
    if (selectedElement) {
      setSelectedElement(null);
      setEditMode(false);
    }
    await send(chatInput, opts);
  };

  // ── Deploy → Railway ──────────────────────────
  const handleDeploy = useCallback(async () => {
    if (deploying) return;
    if (Object.keys(projectFiles).length === 0) {
      toast.error("Нечего деплоить — сгенерируйте проект сначала.");
      return;
    }
    setDeploying(true);
    try {
      const files = Object.entries(projectFiles).map(([path, content]) => ({
        path,
        content: String(content ?? ""),
      }));
      const projectName =
        (currentPrompt || "istok-app")
          .toLowerCase()
          .replace(/[^a-z0-9-]+/g, "-")
          .replace(/^-+|-+$/g, "")
          .slice(0, 48) || "istok-app";

      const res = await api.deployToRailway({ project_name: projectName, files });

      if (res.status === "unavailable") {
        toast.warning(res.message || "Railway API токен не настроен", { duration: 6000 });
      } else if (res.status === "failed") {
        toast.error(`Deploy failed: ${res.error || "unknown"}`);
      } else {
        toast.success(res.message || "Deploy queued on Railway", { duration: 5000 });
        if (res.deploy_url) window.open(res.deploy_url, "_blank");
      }
    } catch (err) {
      toast.error(`Deploy error: ${err instanceof Error ? err.message : String(err)}`);
    } finally {
      setDeploying(false);
    }
  }, [deploying, projectFiles, currentPrompt]);

  const handleSecurityAudit = useCallback(() => {
    setSecurityAuditOpen((v) => !v);
  }, []);

  const preview = (
    <WorkspacePreview
      projectFiles={projectFiles}
      onFilesChange={setProjectFiles}
      initialLoading={initialLoading || hydrating}
      loaderStep={loaderStep}
      loaderSteps={loaderSteps}
      editMode={editMode}
      onEditModeChange={setEditMode}
      onElementSelect={setSelectedElement}
      onTelegramExport={applyTelegramExport}
      onPublish={publishCurrent}
      onDeploy={handleDeploy}
      deploying={deploying}
      onSecurityAudit={handleSecurityAudit}
      securityApproved={securityApproved}
      milestones={milestones}
      activeAgent={activeAgent}
      thinking={thinking}
      streamedFiles={streamedFiles}
      currentFSMState={currentFSMState}
      canResume={canResume}
      onResume={resumeGeneration}
    />
  );

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.4 }}
      className="h-dvh flex flex-col overflow-hidden bg-background"
    >
      {/* ── Top bar ── */}
      <header className="flex h-11 shrink-0 items-center justify-between border-b border-border/60 bg-panel px-3">
        <div className="flex items-center gap-2">
          <button
            onClick={() => navigate("/")}
            className="flex h-7 items-center gap-1.5 rounded-md px-2 text-xs text-muted-foreground transition-colors hover:bg-elevated hover:text-foreground"
          >
            <ArrowLeft className="h-3.5 w-3.5" /> Назад
          </button>
          <div className="h-4 w-px bg-border/60" />
          <div className="grid h-6 w-6 place-items-center rounded-md bg-gradient-primary text-[10px] font-bold text-primary-foreground">
            И
          </div>
          <span className="text-sm font-semibold tracking-tight">Исток</span>
          {currentPrompt && (
            <span className="ml-1 max-w-[280px] truncate font-mono text-[11px] text-muted-foreground">
              · {currentPrompt}
            </span>
          )}
        </div>
        <div className="flex h-7 items-center overflow-hidden rounded-md border border-border/70 bg-elevated/60 text-xs">
          <button
            onClick={() => setRightView("code")}
            className={`flex h-full items-center gap-1.5 px-2.5 transition-colors ${
              rightView === "code" ? "bg-primary/15 text-primary" : "text-muted-foreground hover:text-foreground"
            }`}
          >
            <Code2 className="h-3.5 w-3.5" /> Код
          </button>
          <button
            onClick={() => setRightView("preview")}
            className={`flex h-full items-center gap-1.5 border-l border-border/70 px-2.5 transition-colors ${
              rightView === "preview" ? "bg-primary/15 text-primary" : "text-muted-foreground hover:text-foreground"
            }`}
          >
            <Eye className="h-3.5 w-3.5" /> Предпросмотр
          </button>
        </div>
      </header>

      {/* ── 2-panel IDE: Chat | (Preview ⇄ Code) ── */}
      <div className="min-h-0 flex-1">
        {/* Desktop: resizable dual-panel; right side toggles Preview/Code via header */}
        <ResizablePanelGroup direction="horizontal" className="hidden h-full w-full md:flex">
          <ResizablePanel defaultSize={30} minSize={22} maxSize={45}>
            <BuilderChatPanel
              messages={messages}
              thinking={thinking}
              input={chatInput}
              onInputChange={setChatInput}
              onSend={handleSend}
              agentMode={agentMode}
              onModeChange={setAgentMode}
              projectName={currentPrompt}
              editMode={editMode}
              onEditModeChange={setEditMode}
            />
          </ResizablePanel>
          <ResizableHandle />
          <ResizablePanel defaultSize={70} minSize={40}>
            {rightView === "code" ? (
              <BuilderCodePanel projectFiles={projectFiles} />
            ) : (
              <div className="relative flex h-full flex-col bg-panel">
                <div className="min-h-0 flex-1">{preview}</div>
                <AgentPulse agents={agents} />
                <AnimatePresence>
                  {securityAuditOpen && (
                    <SecurityAuditOverlay
                      securityApproved={securityApproved}
                      testerApproved={testerApproved}
                      uiReviewerApproved={uiReviewerApproved}
                      onClose={() => setSecurityAuditOpen(false)}
                    />
                  )}
                </AnimatePresence>
              </div>
            )}
          </ResizablePanel>
        </ResizablePanelGroup>

        {/* Mobile: chat + toggled code/preview */}
        <div className="flex h-full flex-col md:hidden">
          <div className="h-1/2 min-h-0 border-b border-border/60">
            <BuilderChatPanel
              messages={messages}
              thinking={thinking}
              input={chatInput}
              onInputChange={setChatInput}
              onSend={handleSend}
              agentMode={agentMode}
              onModeChange={setAgentMode}
              projectName={currentPrompt}
              editMode={editMode}
              onEditModeChange={setEditMode}
            />
          </div>
          <div className="min-h-0 flex-1">
            {rightView === "code" ? (
              <BuilderCodePanel projectFiles={projectFiles} />
            ) : (
              <div className="flex h-full flex-col bg-panel">
                <div className="min-h-0 flex-1">{preview}</div>
                <AgentPulse agents={agents} />
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Human-in-the-Loop: Business feature approval modal */}
      <FeatureApprovalModal />
      {/* Human-in-the-Loop: Media prompt approval modal (design review) */}
      <MediaApprovalModal />
      {/* Pause & Resume: insufficient funds overlay */}
      <InsufficientFundsOverlay />
    </motion.div>
  );
};

export default Workspace;
