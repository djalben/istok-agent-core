import { useState, useEffect, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { motion, AnimatePresence } from "framer-motion";
import { toast } from "sonner";
import { SidebarProvider } from "@/components/ui/sidebar";
import ChatPanel from "@/components/workspace/ChatPanel";
import PreviewPanel from "@/components/workspace/PreviewPanel";
import MilestonesPanel from "@/components/workspace/MilestonesPanel";
import SecurityAuditOverlay from "@/components/workspace/SecurityAuditOverlay";
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
                streamedFiles={streamedFiles}
                currentFSMState={currentFSMState}
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
    </motion.div>
  );
};

export default Workspace;
