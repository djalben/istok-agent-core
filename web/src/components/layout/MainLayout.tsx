import { useState, type ReactNode } from "react";
import HeaderBar from "@/components/HeaderBar";
import ProjectSidebar from "@/components/ProjectSidebar";

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — MainLayout
//  Единый layout-wrapper для всех страниц с HeaderBar + ProjectSidebar.
//  Workspace.tsx сохраняет свой собственный chat-sidebar и не оборачивается.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

export interface MainLayoutProps {
  children: ReactNode;
  /** Show ProjectSidebar slot (default: false — header-only pages). */
  withSidebar?: boolean;
  /** Apply mesh-gradient + grid background to content area (default: false). */
  decorated?: boolean;
  /** Override max-width container styles for content area. */
  contentClassName?: string;
}

/**
 * MainLayout — общий каркас страниц.
 *
 * Варианты:
 *   • header-only:        <MainLayout>...</MainLayout>          (Settings, Projects, Pricing)
 *   • с sidebar:          <MainLayout withSidebar>...</MainLayout>  (Index)
 *   • с фоном Premium:    <MainLayout decorated>...</MainLayout>
 */
const MainLayout = ({
  children,
  withSidebar = false,
  decorated = false,
  contentClassName,
}: MainLayoutProps) => {
  const [sidebarCollapsed, setSidebarCollapsed] = useState(true);

  if (withSidebar) {
    return (
      <div className="h-screen flex flex-col overflow-hidden">
        <div className="flex-1 flex overflow-hidden relative">
          <ProjectSidebar
            collapsed={sidebarCollapsed}
            onToggle={() => setSidebarCollapsed(!sidebarCollapsed)}
          />
          <div className="flex-1 min-w-0 flex flex-col">
            <HeaderBar />
            <div
              className={
                contentClassName ??
                (decorated
                  ? "flex-1 overflow-y-auto mesh-gradient-bg grid-pattern"
                  : "flex-1 overflow-y-auto")
              }
            >
              {children}
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background">
      <HeaderBar />
      <main
        className={
          contentClassName ??
          (decorated ? "mesh-gradient-bg grid-pattern" : undefined)
        }
      >
        {children}
      </main>
    </div>
  );
};

export default MainLayout;
