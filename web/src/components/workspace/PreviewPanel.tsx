import WorkspacePreview, {
  type ProjectFiles,
  type SelectedElement,
} from "@/components/WorkspacePreview";
import type { AgentMilestone, StreamedFile } from "@/hooks/useGeneration";

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — PreviewPanel
//  Тонкая обёртка над WorkspacePreview: iframe, edit-mode, экспорт.
//  Изолирует Workspace от деталей рендера превью.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

export interface PreviewPanelProps {
  projectFiles: ProjectFiles;
  onFilesChange: (files: ProjectFiles) => void;
  initialLoading: boolean;
  loaderStep: number;
  loaderSteps: string[];
  editMode: boolean;
  onEditModeChange: (mode: boolean) => void;
  onElementSelect: (el: SelectedElement | null) => void;
  onTelegramExport: () => void;
  onPublish: () => Promise<string | null>;

  // Workspace v3.0
  onDeploy?: () => Promise<void>;
  deploying?: boolean;
  onSecurityAudit?: () => void;
  securityApproved?: boolean;

  // Generation state
  thinking?: boolean;

  // Live agent stream for GenerationMagic
  milestones?: AgentMilestone[];
  activeAgent?: string | null;
  streamedFiles?: StreamedFile[];
  currentFSMState?: string;
}

const PreviewPanel = (props: PreviewPanelProps) => {
  return (
    <WorkspacePreview
      projectFiles={props.projectFiles}
      onFilesChange={props.onFilesChange}
      initialLoading={props.initialLoading}
      loaderStep={props.loaderStep}
      loaderSteps={props.loaderSteps}
      editMode={props.editMode}
      onEditModeChange={props.onEditModeChange}
      onElementSelect={props.onElementSelect}
      onTelegramExport={props.onTelegramExport}
      onPublish={props.onPublish}
      onDeploy={props.onDeploy}
      deploying={props.deploying}
      onSecurityAudit={props.onSecurityAudit}
      securityApproved={props.securityApproved}
      thinking={props.thinking}
      milestones={props.milestones}
      activeAgent={props.activeAgent}
      streamedFiles={props.streamedFiles}
      currentFSMState={props.currentFSMState}
    />
  );
};

export default PreviewPanel;
