import { useEffect, useState } from "react";
import { Link } from "@tanstack/react-router";
import { motion } from "framer-motion";
import { ArrowUpRight, Plus, Sparkles, Clock } from "lucide-react";
import { mockProjects } from "@/lib/mockData";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { ProjectCardMenu } from "@/components/features/ProjectCardMenu";


export function ProjectsGrid() {
  const [loading, setLoading] = useState(true);
  useEffect(() => {
    const t = setTimeout(() => setLoading(false), 700);
    return () => clearTimeout(t);
  }, []);

  return (
    <div className="mx-auto max-w-7xl px-6 py-12">
      <div className="mb-10 flex flex-col gap-6 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <div className="inline-flex items-center gap-2 rounded-full border border-border/80 bg-elevated px-3 py-1 text-xs text-muted-foreground">
            <Sparkles className="h-3 w-3 text-primary" /> 6 проектов · 24 генерации за неделю
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
          : mockProjects.map((project, i) => (
              <motion.div
                key={project.id}
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: i * 0.04 }}
              >
                <div className="relative">
                  <ProjectCardMenu project={project} />
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
    </div>
  );
}
