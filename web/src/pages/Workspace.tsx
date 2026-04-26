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
      className="h-screen flex flex-col overflow-hidden bg-background"
    >
      <SidebarProvider defaultOpen={true}>
        <div className="flex-1 flex w-full overflow-hidden">
          {/* ── LEFT: ChatPanel ─────────────────────── */}
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
          <div className="flex-1 flex min-w-0 p-3 gap-3 mesh-gradient-bg">
            {/* Center: PreviewPanel */}
            <motion.div
              className="flex-1 min-w-0 floating-canvas relative"
              initial={{ opacity: 0, y: 8, scale: 0.99 }}
              animate={{ opacity: 1, y: 0, scale: 1 }}
              transition={{ duration: 0.5, ease: [0.22, 1, 0.36, 1] }}
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

            {/* Right rail: MilestonesPanel — 10 agents + Verified */}
            <motion.aside
              initial={{ opacity: 0, x: 8 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ duration: 0.5, delay: 0.1, ease: [0.22, 1, 0.36, 1] }}
              className="hidden xl:flex flex-col w-[280px] shrink-0 glass-panel rounded-xl p-3 overflow-hidden"
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
