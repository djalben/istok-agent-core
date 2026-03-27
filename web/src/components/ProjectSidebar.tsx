import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { PanelLeftClose, PanelLeft, Plus, FileText, LogIn, Menu, X } from "lucide-react";
import { useAuth } from "@/hooks/useAuth";
import { useLanguage } from "@/hooks/useLanguage";
import { loadCloudProjects, type CloudProject } from "@/lib/projectSync";
import { useIsMobile } from "@/hooks/use-mobile";

interface ProjectSidebarProps {
  collapsed: boolean;
  onToggle: () => void;
}

const ProjectSidebar = ({ collapsed, onToggle }: ProjectSidebarProps) => {
  const { user } = useAuth();
  const navigate = useNavigate();
  const { t } = useLanguage();
  const isMobile = useIsMobile();
  const [projects, setProjects] = useState<CloudProject[]>([]);
  const [mobileOpen, setMobileOpen] = useState(false);

  useEffect(() => {
    if (!user) {
      setProjects([]);
      return;
    }
    loadCloudProjects().then(setProjects);
  }, [user]);

  // On mobile, show burger button
  if (isMobile) {
    return (
      <>
        {/* Burger button - fixed */}
        <button
          onClick={() => setMobileOpen(true)}
          className="fixed top-3 left-3 z-[60] w-10 h-10 rounded-xl glass flex items-center justify-center text-muted-foreground hover:text-foreground transition-colors"
        >
          <Menu size={18} />
        </button>

        {/* Mobile overlay sidebar */}
        {mobileOpen && (
          <>
            <div className="fixed inset-0 bg-background/80 backdrop-blur-sm z-[70]" onClick={() => setMobileOpen(false)} />
            <div className="fixed left-0 top-0 h-full w-72 bg-background border-r border-border/50 z-[80] flex flex-col animate-in slide-in-from-left duration-300">
              <div className="flex items-center justify-between px-4 py-4 border-b border-border/50">
                <h2 className="text-[10px] font-semibold text-muted-foreground/60 tracking-widest uppercase">
                  {t("sidebarProjects")}
                </h2>
                <button onClick={() => setMobileOpen(false)} className="text-muted-foreground hover:text-foreground transition-colors">
                  <X size={18} />
                </button>
              </div>
              <div className="px-2 py-2">
                <button
                  onClick={() => { navigate("/"); setMobileOpen(false); }}
                  className="w-full flex items-center gap-2 px-3 py-2 text-xs text-muted-foreground hover:text-foreground border border-border/50 hover:border-foreground/20 rounded-lg transition-all duration-200"
                >
                  <Plus size={12} />
                  <span>{t("newProject")}</span>
                </button>
              </div>
              <div className="flex-1 overflow-y-auto px-2">
                {!user ? (
                  <button
                    onClick={() => { navigate("/auth"); setMobileOpen(false); }}
                    className="w-full flex items-center gap-2 px-3 py-4 text-xs text-muted-foreground hover:text-foreground transition-colors"
                  >
                    <LogIn size={12} />
                    <span>{t("sidebarLoginPrompt")}</span>
                  </button>
                ) : projects.length === 0 ? (
                  <div className="px-3 py-4 text-[11px] text-muted-foreground/40 text-center">
                    {t("sidebarNoProjects")}
                  </div>
                ) : (
                  projects.map((project) => (
                    <button
                      key={project.id}
                      onClick={() => { navigate("/project/new", { state: { prompt: project.prompt } }); setMobileOpen(false); }}
                      className="w-full text-left flex items-start gap-2 px-3 py-2.5 mb-0.5 rounded-lg text-muted-foreground hover:text-foreground hover:bg-secondary/50 transition-all duration-200"
                    >
                      <FileText size={12} className="shrink-0 mt-0.5 text-primary/60" />
                      <div className="min-w-0">
                        <div className="text-xs font-medium truncate">{project.prompt.slice(0, 40)}</div>
                        <div className="text-[10px] mt-0.5 opacity-40">
                          {new Date(project.updated_at).toLocaleDateString("ru-RU")}
                        </div>
                      </div>
                    </button>
                  ))
                )}
              </div>
            </div>
          </>
        )}
      </>
    );
  }

  // Desktop: collapsed state
  if (collapsed) {
    return (
      <div className="h-full w-12 bg-background/80 backdrop-blur-xl border-r border-border/50 flex flex-col items-center pt-4">
        <button onClick={onToggle} className="text-muted-foreground hover:text-foreground transition-colors duration-200">
          <PanelLeft size={16} />
        </button>
      </div>
    );
  }

  // Desktop: expanded state
  return (
    <div className="h-full w-56 bg-background/80 backdrop-blur-xl border-r border-border/50 flex flex-col">
      <div className="flex items-center justify-between px-4 py-4 border-b border-border/50">
        <h2 className="text-[10px] font-semibold text-muted-foreground/60 tracking-widest uppercase">
          {t("sidebarProjects")}
        </h2>
        <button onClick={onToggle} className="text-muted-foreground hover:text-foreground transition-colors duration-200">
          <PanelLeftClose size={16} />
        </button>
      </div>

      <div className="px-2 py-2">
        <button
          onClick={() => navigate("/")}
          className="w-full flex items-center gap-2 px-3 py-2 text-xs text-muted-foreground hover:text-foreground border border-border/50 hover:border-foreground/20 rounded-lg transition-all duration-200"
        >
          <Plus size={12} />
          <span>{t("newProject")}</span>
        </button>
      </div>

      <div className="flex-1 overflow-y-auto px-2">
        {!user ? (
          <button
            onClick={() => navigate("/auth")}
            className="w-full flex items-center gap-2 px-3 py-4 text-xs text-muted-foreground hover:text-foreground transition-colors"
          >
            <LogIn size={12} />
            <span>{t("sidebarLoginPrompt")}</span>
          </button>
        ) : projects.length === 0 ? (
          <div className="px-3 py-4 text-[11px] text-muted-foreground/40 text-center">
            {t("sidebarNoProjects")}
          </div>
        ) : (
          projects.map((project) => (
            <button
              key={project.id}
              onClick={() => navigate("/project/new", { state: { prompt: project.prompt } })}
              className="w-full text-left flex items-start gap-2 px-3 py-2.5 mb-0.5 rounded-lg text-muted-foreground hover:text-foreground hover:bg-secondary/50 transition-all duration-200"
            >
              <FileText size={12} className="shrink-0 mt-0.5 text-primary/60" />
              <div className="min-w-0">
                <div className="text-xs font-medium truncate">{project.prompt.slice(0, 40)}</div>
                <div className="text-[10px] mt-0.5 opacity-40">
                  {new Date(project.updated_at).toLocaleDateString("ru-RU")}
                </div>
              </div>
            </button>
          ))
        )}
      </div>
    </div>
  );
};

export default ProjectSidebar;
