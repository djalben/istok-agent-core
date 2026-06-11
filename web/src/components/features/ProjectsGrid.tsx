import { Link } from "@tanstack/react-router";
import { motion } from "framer-motion";
import { ArrowUpRight, Plus, Sparkles, Clock, FolderPlus, Trash2 } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { ProjectCardMenu } from "@/components/features/ProjectCardMenu";
import type { Project } from "@/lib/projectDisplay";
import { useProjects, useDeleteProject } from "@/hooks/useProjects";


export function ProjectsGrid() {
  const { data: projects = [], isLoading, isError } = useProjects();
  const deleteProject = useDeleteProject();
  const loading = isLoading;
  const isEmpty = !loading && !isError && projects.length === 0;

  // Inline удаление проекта из истории: native confirm защищает от случайного клика,
  // затем React Query инвалидирует кэш списка (useDeleteProject).
  const handleDelete = (e: React.MouseEvent, project: Project) => {
    e.preventDefault();
    e.stopPropagation();
    if (!window.confirm("Вы уверены, что хотите удалить этот проект?")) return;
    deleteProject.mutate(project.id, {
      onSuccess: () => toast.success(`«${project.name}» удалён`),
      onError: (err) =>
        toast.error(err instanceof Error ? err.message : "Не удалось удалить проект"),
    });
  };

  return (
    <div className="mx-auto max-w-7xl px-6 py-12">
      <div className="mb-10 flex flex-col gap-6 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <div className="inline-flex items-center gap-2 rounded-full border border-border/80 bg-elevated px-3 py-1 text-xs text-muted-foreground">
            <Sparkles className="h-3 w-3 text-primary" /> {loading ? "Загрузка…" : `${projects.length} ${projects.length === 1 ? "проект" : "проектов"}`}
          </div>
          <h1 className="mt-4 text-4xl font-semibold tracking-tight">
            С возвращением. <span className="text-gradient">Запустите что-нибудь сегодня.</span>
          </h1>
          <p className="mt-2 max-w-xl text-muted-foreground">
            Исток управляет командой ИИ-агентов, которые исследуют, проектируют и собирают full-stack приложения из одного запроса.
          </p>
        </div>
        <Link to="/builder">
          <Button size="lg" className="bg-gradient-primary text-primary-foreground shadow-glow hover:opacity-90">
            <Plus className="h-4 w-4" /> Новый проект
          </Button>
        </Link>
      </div>

      {isEmpty ? (
        <div className="grid place-items-center rounded-xl border border-dashed border-border/60 bg-card/20 py-20 text-center">
          <div className="grid h-12 w-12 place-items-center rounded-full bg-muted/40 text-muted-foreground">
            <FolderPlus className="h-5 w-5" />
          </div>
          <p className="mt-3 text-sm font-medium">Пока нет проектов</p>
          <p className="mt-1 max-w-xs text-xs text-muted-foreground">
            Опишите идею — команда агентов Истока соберёт первое приложение.
          </p>
          <Link to="/builder" className="mt-4">
            <Button size="sm" className="bg-gradient-primary text-primary-foreground shadow-glow hover:opacity-90">
              <Plus className="h-4 w-4" /> Новый проект
            </Button>
          </Link>
        </div>
      ) : (
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {loading
          ? Array.from({ length: 6 }).map((_, i) => (
              <div
                key={i}
                className="overflow-hidden rounded-xl border border-border/60 bg-card"
              >
                <Skeleton className="h-36 w-full rounded-none" />
                <div className="space-y-2 p-4">
                  <Skeleton className="h-4 w-2/3" />
                  <Skeleton className="h-3 w-full" />
                  <Skeleton className="h-3 w-4/5" />
                  <Skeleton className="mt-2 h-3 w-20" />
                </div>
              </div>
            ))
          : projects.map((project, i) => (
              <motion.div
                key={project.id}
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: i * 0.04 }}
              >
                <div className="group/card relative">
                  <ProjectCardMenu project={project} />
                  <button
                    type="button"
                    aria-label="Удалить проект"
                    onClick={(e) => handleDelete(e, project)}
                    disabled={deleteProject.isPending}
                    className="absolute bottom-3 right-3 z-10 grid place-items-center rounded-md p-1.5 text-zinc-500 opacity-0 transition-all duration-300 hover:bg-red-500/10 hover:text-red-400 group-hover/card:opacity-100 disabled:pointer-events-none disabled:opacity-40"
                  >
                    <Trash2 className="h-4 w-4" />
                  </button>
                  <Link to="/builder/$id" params={{ id: project.id }}>
                    <article className="group relative overflow-hidden rounded-xl border border-border/80 bg-card transition-all hover:border-primary/50 hover:shadow-glow">
                      <div className={`relative h-36 overflow-hidden ${project.gradient}`}>
                        <div className="absolute inset-0 bg-black/10 transition-transform duration-500 group-hover:scale-105" />
                        <div className="absolute bottom-3 left-3 rounded-md bg-black/40 px-2 py-1 font-mono text-[10px] uppercase tracking-wider text-white backdrop-blur-sm">
                          {project.framework}
                        </div>
                        <ArrowUpRight className="absolute right-3 top-3 h-5 w-5 text-white/80 opacity-0 transition-opacity group-hover:opacity-100" />
                      </div>
                      <div className="p-4">
                        <h3 className="font-medium text-foreground">{project.name}</h3>
                        <p className="mt-1 line-clamp-2 text-sm text-muted-foreground">{project.description}</p>
                        <div className="mt-3 flex items-center gap-1.5 text-xs text-muted-foreground">
                          <Clock className="h-3 w-3" /> {project.updatedAt}
                        </div>
                      </div>
                    </article>
                  </Link>
                </div>

              </motion.div>
            ))}
      </div>
      )}
    </div>
  );
}
