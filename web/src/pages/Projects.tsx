import { useState, useEffect, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { motion, AnimatePresence } from "framer-motion";
import {
  Plus,
  Search,
  Trash2,
  Code2,
  Globe,
  Sparkles,
  ArrowLeft,
} from "lucide-react";
import { useAuth } from "@/hooks/useAuth";
import { useLanguage } from "@/hooks/useLanguage";
import { loadCloudProjects, deleteCloudProject, type CloudProject } from "@/lib/projectSync";
import { toast } from "sonner";
import HeaderBar from "@/components/HeaderBar";

const Projects = () => {
  const navigate = useNavigate();
  const { user } = useAuth();
  const { t } = useLanguage();
  const [projects, setProjects] = useState<CloudProject[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [deletingId, setDeletingId] = useState<string | null>(null);

  const fetchProjects = useCallback(async () => {
    setLoading(true);
    const data = await loadCloudProjects();
    setProjects(data);
    setLoading(false);
  }, []);

  useEffect(() => {
    fetchProjects();
  }, [fetchProjects]);

  const handleDelete = async (id: string) => {
    if (deletingId === id) {
      await deleteCloudProject(id);
      setProjects((prev) => prev.filter((p) => p.id !== id));
      setDeletingId(null);
      toast.success(t("projectsDeleted"));
    } else {
      setDeletingId(id);
      toast(t("projectsDeleteConfirm"), { duration: 3000 });
      setTimeout(() => setDeletingId(null), 3000);
    }
  };

  const filtered = projects.filter((p) =>
    p.prompt.toLowerCase().includes(search.toLowerCase())
  );

  const formatDate = (dateStr: string) => {
    const d = new Date(dateStr);
    return d.toLocaleDateString("ru-RU", { day: "2-digit", month: "2-digit", year: "numeric" });
  };

  return (
    <div className="min-h-screen bg-background">
      <HeaderBar />

      <div className="max-w-6xl mx-auto px-4 md:px-6 py-6 md:py-8">
        <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 mb-6 md:mb-8">
          <div className="flex items-center gap-4">
            <button
              onClick={() => navigate("/")}
              className="w-9 h-9 rounded-xl flex items-center justify-center text-muted-foreground hover:text-foreground hover:bg-secondary/50 transition-colors"
            >
              <ArrowLeft size={18} />
            </button>
            <div>
              <h1 className="text-xl md:text-2xl font-bold text-foreground">{t("projectsTitle")}</h1>
              <p className="text-sm text-muted-foreground mt-0.5">
                {t("projectsCount", projects.length)}
              </p>
            </div>
          </div>
          <button
            onClick={() => navigate("/project/new")}
            className="flex items-center gap-2 h-10 px-5 rounded-xl btn-gradient text-primary-foreground text-sm font-medium"
          >
            <Plus size={16} />
            {t("projectsCreate")}
          </button>
        </div>

        {projects.length > 0 && (
          <div className="relative mb-6">
            <Search size={16} className="absolute left-4 top-1/2 -translate-y-1/2 text-muted-foreground/50" />
            <input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder={t("projectsSearch")}
              className="w-full h-11 pl-11 pr-4 rounded-xl bg-secondary/30 border border-border/20 text-sm text-foreground placeholder:text-muted-foreground/40 outline-none focus:border-primary/40 focus:ring-1 focus:ring-primary/20 transition-all"
            />
          </div>
        )}

        {loading ? (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {[1, 2, 3].map((i) => (
              <div key={i} className="glass rounded-2xl border border-border/20 p-5 space-y-3">
                <div className="h-4 w-3/4 rounded-full bg-muted/40 animate-pulse" />
                <div className="h-3 w-1/2 rounded-full bg-muted/20 animate-pulse" />
                <div className="h-24 rounded-xl bg-muted/10 animate-pulse" />
              </div>
            ))}
          </div>
        ) : projects.length === 0 ? (
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            className="flex flex-col items-center justify-center py-16 md:py-24"
          >
            <div className="w-20 h-20 rounded-2xl bg-primary/10 flex items-center justify-center mb-6">
              <Sparkles size={36} className="text-primary" />
            </div>
            <h2 className="text-xl font-semibold text-foreground mb-2">{t("projectsEmpty")}</h2>
            <p className="text-sm text-muted-foreground mb-8 text-center max-w-sm">
              {t("projectsEmptyDesc")}
            </p>
            <button
              onClick={() => navigate("/project/new")}
              className="flex items-center gap-2 h-11 px-6 rounded-xl btn-gradient text-primary-foreground text-sm font-medium"
            >
              <Plus size={16} />
              {t("projectsCreateFirst")}
            </button>
          </motion.div>
        ) : filtered.length === 0 ? (
          <div className="text-center py-16">
            <Search size={32} className="mx-auto text-muted-foreground/30 mb-4" />
            <p className="text-muted-foreground">{t("projectsNotFound", search)}</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            <AnimatePresence>
              {filtered.map((project, i) => (
                <motion.div
                  key={project.id}
                  initial={{ opacity: 0, y: 12 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, scale: 0.95 }}
                  transition={{ delay: i * 0.04 }}
                  className="glass rounded-2xl border border-border/20 hover:border-primary/20 transition-all duration-300 group overflow-hidden"
                >
                  <div className="h-28 sm:h-32 bg-secondary/20 border-b border-border/10 overflow-hidden relative">
                    <iframe
                      srcDoc={project.code}
                      title={project.prompt}
                      className="w-[200%] h-[200%] border-0 pointer-events-none origin-top-left scale-50"
                      sandbox=""
                    />
                    <div className="absolute inset-0 bg-gradient-to-b from-transparent to-card/80" />
                  </div>

                  <div className="p-4 space-y-3">
                    <div className="flex items-start justify-between gap-2">
                      <h3 className="text-sm font-medium text-foreground leading-snug line-clamp-2">
                        {project.prompt.slice(0, 40)}{project.prompt.length > 40 ? "…" : ""}
                      </h3>
                      <span
                        className={`shrink-0 px-2 py-0.5 rounded-md text-[10px] font-medium ${
                          project.is_public
                            ? "bg-emerald-500/15 text-emerald-400"
                            : "bg-secondary text-muted-foreground"
                        }`}
                      >
                        {project.is_public ? t("projectsPublished") : t("projectsDraft")}
                      </span>
                    </div>

                    <p className="text-[11px] text-muted-foreground/50">
                      {formatDate(project.created_at)}
                    </p>

                    <div className="flex items-center gap-1.5 pt-1">
                      <button
                        onClick={() =>
                          navigate("/project/new", { state: { prompt: project.prompt } })
                        }
                        className="flex items-center gap-1.5 h-8 px-3 rounded-lg text-xs text-foreground bg-secondary/60 hover:bg-secondary transition-colors"
                      >
                        <Code2 size={12} />
                        {t("projectsEditor")}
                      </button>
                      {project.is_public && project.slug && (
                        <a
                          href={`/view/${project.slug}`}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="flex items-center gap-1.5 h-8 px-3 rounded-lg text-xs text-primary bg-primary/10 hover:bg-primary/20 transition-colors"
                        >
                          <Globe size={12} />
                          {t("projectsSite")}
                        </a>
                      )}
                      <div className="flex-1" />
                      <button
                        onClick={() => handleDelete(project.id)}
                        className={`w-8 h-8 rounded-lg flex items-center justify-center transition-colors ${
                          deletingId === project.id
                            ? "bg-destructive/20 text-destructive"
                            : "text-muted-foreground/40 hover:text-destructive hover:bg-destructive/10"
                        }`}
                        title={t("settingsDeleteAccount")}
                      >
                        <Trash2 size={13} />
                      </button>
                    </div>
                  </div>
                </motion.div>
              ))}
            </AnimatePresence>
          </div>
        )}
      </div>
    </div>
  );
};

export default Projects;
