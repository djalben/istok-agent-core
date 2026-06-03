import { useState } from "react";
import { AnimatePresence, motion } from "framer-motion";
import { UserCircle2, Share2 } from "lucide-react";
import { DashboardLayout } from "@/components/features/DashboardLayout";
import { type DashboardSection } from "@/components/features/DashboardSidebar";
import { DashboardHero } from "@/components/features/DashboardHero";
import { ProjectsGrid } from "@/components/features/ProjectsGrid";
import { ProjectsView } from "@/components/features/ProjectsView";
import { StarredEmptyState } from "@/components/features/StarredEmptyState";

export function DashboardShell() {
  const [section, setSection] = useState<DashboardSection>("home");

  return (
    <DashboardLayout active={section} onSelectSection={setSection}>
      <AnimatePresence mode="wait">
        <motion.div
          key={section}
          initial={{ opacity: 0, y: 6 }}
          animate={{ opacity: 1, y: 0 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.18 }}
        >
          {section === "home" && (
            <>
              <DashboardHero />
              <div className="border-t border-border/40">
                <ProjectsGrid />
              </div>
            </>
          )}
          {section === "all" && (
            <ProjectsView
              title="Все проекты"
              subtitle="Всё в вашем рабочем пространстве, сгруппировано по последней активности."
            />
          )}
          {section === "mine" && (
            <ProjectsView
              title="Созданные мной"
              subtitle="Проекты, которые вы создали, отсортированы по последней активности."
            />
          )}
          {section === "starred" && (
            <StarredEmptyState onBrowse={() => setSection("all")} />
          )}
          {section === "shared" && (
            <EmptySection
              icon={Share2}
              title="Поделились со мной"
              copy="Здесь появятся проекты, к которым вас пригласят соавторы."
            />
          )}
        </motion.div>
      </AnimatePresence>
    </DashboardLayout>
  );
}

function EmptySection({
  icon: Icon, title, copy,
}: { icon: typeof UserCircle2; title: string; copy: string }) {
  return (
    <div className="mx-auto flex max-w-3xl flex-col items-center px-6 py-32 text-center">
      <div className="grid h-14 w-14 place-items-center rounded-2xl border border-border/60 bg-gradient-to-br from-primary/10 to-fuchsia-500/10">
        <Icon className="h-6 w-6 text-primary" />
      </div>
      <h1 className="mt-5 text-2xl font-semibold tracking-tight">{title}</h1>
      <p className="mt-1.5 text-sm text-muted-foreground">{copy}</p>
    </div>
  );
}
