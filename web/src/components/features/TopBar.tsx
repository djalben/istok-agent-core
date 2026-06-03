import { useState } from "react";
import { Link, useRouterState } from "@tanstack/react-router";
import {
  Sparkles,
  ArrowLeft,
  Share2,
  Rocket,
  MessageCircle,
  Globe,
  FileText,
  Code2,
  Layers,
  Laptop,
  ExternalLink,
  RotateCw,
  Download,
  GitCommit,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { ShareDialog } from "@/components/features/ShareDialog";
import { PublishDialog } from "@/components/features/PublishDialog";
import { ProjectMenu } from "@/components/features/ProjectMenu";
import { toast } from "sonner";

export type BuilderView = "preview" | "files" | "code" | "layers" | "diff";

interface TopBarProps {
  showBack?: boolean;
  projectName?: string;
  view?: BuilderView;
  onViewChange?: (v: BuilderView) => void;
}

const VIEW_ITEMS: { id: BuilderView; icon: typeof Globe; label: string }[] = [
  { id: "preview", icon: Globe, label: "Предпросмотр" },
  { id: "files", icon: FileText, label: "Файлы" },
  { id: "code", icon: Code2, label: "Код" },
  { id: "diff", icon: GitCommit, label: "Изменения" },
  { id: "layers", icon: Layers, label: "Слои" },
];

export function TopBar({ showBack = false, projectName, view: viewProp, onViewChange }: TopBarProps) {
  const path = useRouterState({ select: (r) => r.location.pathname });
  const [shareOpen, setShareOpen] = useState(false);
  const [publishOpen, setPublishOpen] = useState(false);
  const [internalView, setInternalView] = useState<BuilderView>("preview");
  const view = viewProp ?? internalView;
  const setView = (v: BuilderView) => {
    onViewChange?.(v);
    if (!onViewChange) setInternalView(v);
  };
  const isBuilder = showBack;


  return (
    <>
      <header className="sticky top-0 z-40 grid h-14 grid-cols-[1fr_auto_1fr] items-center gap-3 border-b border-border/60 bg-background/80 px-4 backdrop-blur-xl">
        {/* LEFT */}
        <div className="flex items-center gap-3 min-w-0">
          {showBack && (
            <Link to="/">
              <Button variant="ghost" size="sm" className="h-8 gap-1.5 text-muted-foreground hover:text-foreground">
                <ArrowLeft className="h-4 w-4" /> К проектам
              </Button>
            </Link>
          )}
          <Link to="/" className="flex items-center gap-2">
            <div className="grid h-7 w-7 place-items-center rounded-md bg-gradient-primary shadow-glow">
              <Sparkles className="h-4 w-4 text-primary-foreground" />
            </div>
            {!showBack && <span className="text-base font-semibold tracking-tight">Исток</span>}
            {!showBack && (
              <span className="rounded-md border border-border/80 bg-muted/40 px-1.5 py-0.5 font-mono text-[10px] uppercase text-muted-foreground">
                beta
              </span>
            )}
          </Link>
          {showBack && <ProjectMenu projectName={projectName} />}
          {!showBack && (
            <nav className="hidden items-center gap-1 md:flex">
              <Link to="/">
                <Button variant="ghost" size="sm" className={path === "/" ? "text-foreground" : "text-muted-foreground"}>
                  Проекты
                </Button>
              </Link>
              <Link to="/builder">
                <Button variant="ghost" size="sm" className={path.startsWith("/builder") ? "text-foreground" : "text-muted-foreground"}>
                  Билдер
                </Button>
              </Link>
            </nav>
          )}
        </div>

        {/* CENTER */}
        {isBuilder ? (
          <div className="flex items-center gap-2">
            {/* View toggles */}
            <div className="flex items-center gap-1 rounded-full border border-border/70 bg-elevated/60 p-1">
              {VIEW_ITEMS.map((item) => {
                const Icon = item.icon;
                const active = view === item.id;
                return (
                  <button
                    key={item.id}
                    onClick={() => setView(item.id)}
                    className={cn(
                      "flex h-7 items-center gap-1.5 rounded-full px-2.5 text-xs font-medium transition-colors",
                      active
                        ? "bg-blue-600 text-white shadow-sm"
                        : "text-muted-foreground hover:bg-muted/40 hover:text-foreground",
                    )}
                    aria-label={item.label}
                  >
                    <Icon className="h-3.5 w-3.5" />
                    {active && <span>{item.label}</span>}
                  </button>
                );
              })}
            </div>

            {/* URL bar */}
            <div className="hidden items-center gap-2 rounded-full border border-border/70 bg-elevated/60 px-3 py-1.5 lg:flex">
              <Laptop className="h-3.5 w-3.5 text-muted-foreground" />
              <span className="font-mono text-xs text-muted-foreground">/</span>
              <button className="text-muted-foreground transition-colors hover:text-foreground" aria-label="Открыть">
                <ExternalLink className="h-3.5 w-3.5" />
              </button>
              <button className="text-muted-foreground transition-colors hover:text-foreground" aria-label="Обновить">
                <RotateCw className="h-3.5 w-3.5" />
              </button>
            </div>
          </div>
        ) : (
          <div />
        )}

        {/* RIGHT */}
        <div className="flex items-center justify-end gap-2">
          <Button variant="ghost" size="icon" className="h-8 w-8 text-muted-foreground hover:text-foreground" aria-label="Чат">
            <MessageCircle className="h-4 w-4" />
          </Button>

          <div className="grid h-8 w-8 place-items-center rounded-full bg-blue-600 text-xs font-semibold text-white">
            A
          </div>

          {/* Share with gradient border */}
          <div className="rounded-md bg-gradient-to-r from-fuchsia-500 via-violet-500 to-cyan-500 p-px shadow-[0_0_18px_-6px_rgba(139,92,246,0.55)]">
            <Button
              variant="ghost"
              size="sm"
              className="h-8 gap-1.5 rounded-[5px] bg-background hover:bg-elevated"
              onClick={() => setShareOpen(true)}
            >
              <Share2 className="h-3.5 w-3.5" /> Поделиться
            </Button>
          </div>

          {isBuilder && (
            <Button
              variant="outline"
              size="sm"
              onClick={() => toast.success("ZIP-архив скоро будет готов к скачиванию")}
              className="hidden h-8 gap-1.5 border-border/70 bg-elevated/60 text-xs text-muted-foreground hover:text-foreground sm:inline-flex"
            >
              <Download className="h-3.5 w-3.5" /> Скачать ZIP
            </Button>
          )}

          <Button
            size="sm"
            className="h-8 gap-1.5 bg-blue-600 text-white shadow-glow hover:bg-blue-500"
            onClick={() => setPublishOpen(true)}
          >
            <Rocket className="h-3.5 w-3.5" /> Опубликовать
          </Button>
        </div>
      </header>
      <ShareDialog open={shareOpen} onOpenChange={setShareOpen} />
      <PublishDialog open={publishOpen} onOpenChange={setPublishOpen} />
    </>
  );
}
