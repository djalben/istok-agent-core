import { useState } from "react";
import { TopBar, type BuilderView } from "@/components/features/TopBar";
import { BuilderWorkspace } from "@/components/features/BuilderWorkspace";
import { HistorySidebar } from "@/components/features/HistorySidebar";

interface BuilderShellProps {
  mode: "clean" | "project";
  projectName?: string;
}

export function BuilderShell({ mode, projectName }: BuilderShellProps) {
  const [view, setView] = useState<BuilderView>("preview");
  return (
    <div className="flex h-screen flex-col bg-background">
      <TopBar showBack projectName={projectName} view={view} onViewChange={setView} />
      <div className="flex min-h-0 flex-1">
        <div className="min-w-0 flex-1">
          <BuilderWorkspace mode={mode} projectName={projectName} view={view} />
        </div>
        <HistorySidebar />
      </div>
    </div>
  );
}
