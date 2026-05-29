import { useState, useEffect, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { motion, AnimatePresence } from "framer-motion";
import { toast } from "sonner";
import { SidebarProvider } from "@/components/ui/sidebar";
import ChatPanel from "@/components/workspace/ChatPanel";
import PreviewPanel from "@/components/workspace/PreviewPanel";
import MilestonesPanel from "@/components/workspace/MilestonesPanel";
import SecurityAuditOverlay from "@/components/workspace/SecurityAuditOverlay";
import FeatureApprovalModal from "@/components/workspace/ArchitectureApprovalModal";
import MediaApprovalModal from "@/components/workspace/MediaApprovalModal";
import { useGeneration } from "@/hooks/useGeneration";
import type { SelectedElement } from "@/components/WorkspacePreview";
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

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.4 }}
      className="h-dvh flex flex-col overflow-hidden bg-background"
    >
      <SidebarProvider defaultOpen={true} className="min-h-0 h-full">
        <div className="flex-1 flex w-full min-h-0 overflow-hidden">
          {/* ── LEFT: ChatPanel (offcanvas on mobile, visible on lg+) ── */}
          <ChatPanel
            messages={messages}
            thinking={thinking}
            chatInput={chatInput}
            onChatInputChange={setChatInput}
            onSend={handleSend}
            agentMode={agentMode}
            onModeChange={setAgentMode}
            savedProjects={savedProjects}
            onLoadProject={loadProject}
            onDeleteProject={deleteProject}
            selectedElement={selectedElement}
            onClearSelectedElement={() => setSelectedElement(null)}
            currentPrompt={currentPrompt}
            onNavigateBack={() => navigate("/")}
            onNavigateTemplates={() => navigate("/")}
          />

          {/* ── CENTER + RIGHT ───────────────────────── */}
          <div className="flex-1 flex flex-col lg:flex-row min-w-0 min-h-0 mesh-gradient-bg">
            {/* Center: PreviewPanel — 100% on mobile */}
            <motion.div
              className="flex-1 min-w-0 min-h-0 flex flex-col relative"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              transition={{ duration: 0.4, ease: [0.22, 1, 0.36, 1] }}
            >
              <PreviewPanel
                projectFiles={projectFiles}
                onFilesChange={setProjectFiles}
                initialLoading={initialLoading}
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
            </motion.div>

            {/* Right rail: MilestonesPanel — hidden on mobile/tablet, visible on xl+ */}
            <motion.aside
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              transition={{ duration: 0.4 }}
              className="hidden xl:flex flex-col w-[280px] shrink-0 glass-panel border-l border-glass-border/30 p-3 overflow-hidden"
            >
              <MilestonesPanel
                activeAgent={activeAgent}
                milestones={milestones}
                currentFSMState={currentFSMState}
                securityApproved={securityApproved}
                testerApproved={testerApproved}
                uiReviewerApproved={uiReviewerApproved}
              />
            </motion.aside>
          </div>
        </div>
      </SidebarProvider>

      {/* Human-in-the-Loop: Business feature approval modal */}
      <FeatureApprovalModal />
      {/* Human-in-the-Loop: Media prompt approval modal (design review) */}
      <MediaApprovalModal />
    </motion.div>
  );
};

export default Workspace;
