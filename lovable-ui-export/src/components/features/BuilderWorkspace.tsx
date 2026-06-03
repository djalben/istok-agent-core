import { useState } from "react";
import { Panel, Group as PanelGroup, Separator as PanelResizeHandle } from "react-resizable-panels";
import { ChatPanel } from "@/components/features/ChatPanel";
import { CodePanel } from "@/components/features/CodePanel";
import { PreviewPanel } from "@/components/features/PreviewPanel";
import { DiffPanel } from "@/components/features/DiffPanel";
import { ApprovalModal } from "@/components/features/ApprovalModal";
import { useGenerationStream } from "@/hooks/useGenerationStream";
import { initialChat, mockFiles, type ChatMessage, type FileNode } from "@/lib/mockData";
import type { BuilderView } from "@/components/features/TopBar";

interface BuilderWorkspaceProps {
  mode: "clean" | "project";
  projectName?: string;
  view?: BuilderView;
}

function ResizeHandle() {
  return (
    <PanelResizeHandle className="group relative w-1 cursor-col-resize bg-border/40 transition-colors hover:bg-primary/60">
      <div className="absolute inset-y-0 -left-1 -right-1" />
    </PanelResizeHandle>
  );
}

export function BuilderWorkspace({ mode, projectName, view = "layers" }: BuilderWorkspaceProps) {
  const isClean = mode === "clean";
  const { agents, start } = useGenerationStream();
  const [modalOpen, setModalOpen] = useState(!isClean);
  const [files, setFiles] = useState<FileNode[]>(isClean ? [] : mockFiles);
  const [messages, setMessages] = useState<ChatMessage[]>(isClean ? [] : initialChat);
  const [started, setStarted] = useState(!isClean);

  const handleChatSubmit = () => {
    start();
    if (isClean && !started) {
      setStarted(true);
      setFiles(mockFiles);
    }
  };

  const chat = (
    <ChatPanel
      initialMessages={messages}
      onMessagesChange={setMessages}
      onSubmit={handleChatSubmit}
      projectName={projectName}
    />
  );

  const renderRight = () => {
    if (view === "preview") return <PreviewPanel agents={agents} clean={isClean && !started} />;
    if (view === "code" || view === "files") return <CodePanel files={files} />;
    if (view === "diff") return <DiffPanel />;
    return null;
  };

  return (
    <>
      <div className="min-h-0 flex-1">
        {view === "layers" ? (
          <PanelGroup orientation="horizontal" className="hidden h-full w-full md:flex">
            <Panel defaultSize={26} minSize={18}>{chat}</Panel>
            <ResizeHandle />
            <Panel defaultSize={44} minSize={28}>
              <CodePanel files={files} />
            </Panel>
            <ResizeHandle />
            <Panel defaultSize={30} minSize={20}>
              <PreviewPanel agents={agents} clean={isClean && !started} />
            </Panel>
          </PanelGroup>
        ) : (
          <PanelGroup orientation="horizontal" className="hidden h-full w-full md:flex">
            <Panel defaultSize={30} minSize={20}>{chat}</Panel>
            <ResizeHandle />
            <Panel defaultSize={70} minSize={40}>
              {renderRight()}
            </Panel>
          </PanelGroup>
        )}
        {/* Mobile: chat only */}
        <div className="flex h-full md:hidden">
          {view === "preview" ? <PreviewPanel agents={agents} clean={isClean && !started} /> :
           view === "code" || view === "files" ? <CodePanel files={files} /> :
           view === "diff" ? <DiffPanel /> :
           chat}
        </div>
      </div>

      <ApprovalModal
        open={modalOpen}
        onApprove={() => { setModalOpen(false); start(); }}
        onReject={() => setModalOpen(false)}
      />
    </>
  );
}
