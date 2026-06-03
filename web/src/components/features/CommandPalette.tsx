import { useEffect, useMemo, useState } from "react";
import { useNavigate } from "@tanstack/react-router";
import { AnimatePresence, motion } from "framer-motion";
import {
  Home, Plus, FileText, Newspaper, Users, Layers, CreditCard,
  Cloud, ShieldCheck, ScrollText, GitBranch, Send, Plug, Database,
  Lock, Folder, Clock, ArrowUpRight, Building2, Sparkles,
} from "lucide-react";
import {
  Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList, CommandSeparator,
} from "@/components/ui/command";
import { Dialog, DialogContent, DialogTitle } from "@/components/ui/dialog";
import { useProjects } from "@/hooks/useProjects";
import type { Project } from "@/lib/projectDisplay";
import { cn } from "@/lib/utils";

interface CommandPaletteProps {
  open: boolean;
  onOpenChange: (o: boolean) => void;
}

type Item = {
  id: string;
  label: string;
  icon: typeof Home;
  group: string;
  hint?: string;
  projectId?: string;
  onSelect?: () => void;
};

export function CommandPalette({ open, onOpenChange }: CommandPaletteProps) {
  const navigate = useNavigate();
  const [active, setActive] = useState<string>("");
  const { data: projects = [] } = useProjects();

  const projectItems: Item[] = useMemo(
    () =>
      projects.slice(0, 4).map((p) => ({
        id: `p-${p.id}`,
        label: p.name,
        icon: Folder,
        group: "Недавние проекты",
        projectId: p.id,
        hint: p.updatedAt,
      })),
    [projects],
  );

  const navItems: Item[] = [
    { id: "nav-home", label: "Главная", icon: Home, group: "Перейти", onSelect: () => navigate({ to: "/" }) },
    { id: "nav-new", label: "Создать новый проект", icon: Plus, group: "Перейти", onSelect: () => navigate({ to: "/builder" }) },
    { id: "nav-docs", label: "Документация", icon: FileText, group: "Перейти", hint: "↗" },
    { id: "nav-changelog", label: "История изменений", icon: Newspaper, group: "Перейти", hint: "↗" },
  ];

  const settingsItems: Item[] = [
    { id: "set-ws", label: "Рабочее пространство", icon: Building2, group: "Настройки" },
    { id: "set-people", label: "Участники", icon: Users, group: "Настройки" },
    { id: "set-templates", label: "Шаблоны", icon: Layers, group: "Настройки" },
    { id: "set-plans", label: "Тарифы и кредиты", icon: CreditCard, group: "Настройки" },
    { id: "set-cloud", label: "Cloud и баланс ИИ", icon: Cloud, group: "Настройки" },
    { id: "set-privacy", label: "Приватность и безопасность", icon: Lock, group: "Настройки" },
    { id: "set-sec", label: "Центр безопасности", icon: ShieldCheck, group: "Настройки" },
    { id: "set-audit", label: "Журнал аудита", icon: ScrollText, group: "Настройки" },
    { id: "set-git", label: "Git", icon: GitBranch, group: "Настройки" },
    { id: "set-tg", label: "Подключить Telegram", icon: Send, group: "Настройки" },
    { id: "set-conn", label: "Коннекторы и MCP", icon: Plug, group: "Настройки" },
    { id: "set-sb", label: "Supabase", icon: Database, group: "Настройки" },
  ];

  const all = [...projectItems, ...navItems, ...settingsItems];
  const activeItem = all.find((i) => i.id === active) ?? projectItems[0];
  const activeProject =
    (activeItem?.projectId
      ? projects.find((p) => p.id === activeItem.projectId)
      : null) ?? null;

  useEffect(() => {
    if (open) setActive(projectItems[0]?.id ?? "");
  }, [open, projectItems]);

  const handleSelect = (item: Item) => {
    if (item.projectId) {
      navigate({ to: "/builder/$id", params: { id: item.projectId } });
      onOpenChange(false);
      return;
    }
    item.onSelect?.();
    onOpenChange(false);
  };

  const groupedRender = (group: string, items: Item[]) => (
    <CommandGroup key={group} heading={group}>
      {items.map((it) => {
        const Icon = it.icon;
        return (
          <CommandItem
            key={it.id}
            value={`${it.group} ${it.label}`}
            onMouseEnter={() => setActive(it.id)}
            onSelect={() => handleSelect(it)}
            className="group/item gap-2.5 rounded-md px-2.5 py-2 text-sm aria-selected:bg-primary/10 aria-selected:text-foreground"
          >
            <Icon className="h-4 w-4 text-muted-foreground group-aria-selected/item:text-primary" />
            <span className="flex-1 truncate">{it.label}</span>
            {it.hint && (
              <span className="font-mono text-[10px] text-muted-foreground">{it.hint}</span>
            )}
          </CommandItem>
        );
      })}
    </CommandGroup>
  );

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="overflow-hidden border-border/60 bg-card/95 p-0 backdrop-blur-2xl sm:max-w-[860px]">
        <DialogTitle className="sr-only">Командная палитра</DialogTitle>
        <Command
          className="bg-transparent [&_[cmdk-input-wrapper]]:border-border/60"
          onValueChange={() => {}}
        >
          <CommandInput placeholder="Поиск проектов, настроек, переход…" className="h-12" />
          <div className="grid grid-cols-[1fr_320px]">
            <CommandList className="max-h-[460px] min-h-[460px] overflow-y-auto border-r border-border/60 px-1.5 py-2">
              <CommandEmpty>Ничего не найдено.</CommandEmpty>
              {groupedRender("Недавние проекты", projectItems)}
              <CommandSeparator className="my-1" />
              {groupedRender("Перейти", navItems)}
              <CommandSeparator className="my-1" />
              {groupedRender("Настройки", settingsItems)}
            </CommandList>
            <PreviewPane item={activeItem} project={activeProject} />
          </div>
          <div className="flex items-center justify-between border-t border-border/60 px-3 py-2 text-[10px] text-muted-foreground">
            <div className="flex items-center gap-3">
              <span className="flex items-center gap-1"><Kbd>↑↓</Kbd> навигация</span>
              <span className="flex items-center gap-1"><Kbd>↵</Kbd> выбрать</span>
              <span className="flex items-center gap-1"><Kbd>esc</Kbd> закрыть</span>
            </div>
            <span className="flex items-center gap-1">
              <Sparkles className="h-3 w-3 text-primary" /> Команды Истока
            </span>
          </div>
        </Command>
      </DialogContent>
    </Dialog>
  );
}

function Kbd({ children }: { children: React.ReactNode }) {
  return (
    <kbd className="rounded border border-border/60 bg-elevated/60 px-1 py-0.5 font-mono text-[9px] text-muted-foreground">
      {children}
    </kbd>
  );
}

function PreviewPane({
  item,
  project,
}: {
  item?: Item;
  project: Project | null;
}) {
  return (
    <div className="relative min-h-[460px] bg-elevated/20 p-4">
      <AnimatePresence mode="wait">
        <motion.div
          key={item?.id ?? "empty"}
          initial={{ opacity: 0, y: 6 }}
          animate={{ opacity: 1, y: 0 }}
          exit={{ opacity: 0, y: -6 }}
          transition={{ duration: 0.18 }}
          className="flex h-full flex-col"
        >
          {project ? (
            <>
              <div className={cn("relative h-32 w-full overflow-hidden rounded-lg", project.gradient)}>
                <div className="absolute inset-0 bg-[radial-gradient(circle_at_20%_20%,rgba(255,255,255,0.25),transparent_50%)]" />
                <div className="absolute bottom-2 left-2 rounded bg-black/40 px-1.5 py-0.5 font-mono text-[9px] uppercase tracking-wider text-white backdrop-blur">
                  {project.framework}
                </div>
              </div>
              <h3 className="mt-3 text-sm font-medium text-foreground">{project.name}</h3>
              <p className="mt-0.5 line-clamp-2 text-[11px] text-muted-foreground">
                {project.description}
              </p>
              <div className="mt-4 space-y-2 text-[11px]">
                <MetaRow label="Автор" value="Александр" />
                <MetaRow label="Статус" value="Приватный" />
                <MetaRow label="Последнее изменение" value={project.updatedAt} icon={Clock} />
                <MetaRow label="Фреймворк" value={project.framework} />
              </div>
              <div className="mt-auto flex items-center justify-between rounded-md border border-border/60 bg-card/60 px-2.5 py-1.5 text-[11px] text-muted-foreground">
                <span>Нажмите ↵, чтобы открыть</span>
                <ArrowUpRight className="h-3.5 w-3.5" />
              </div>
            </>
          ) : item ? (
            <>
              <div className="grid h-32 w-full place-items-center rounded-lg border border-dashed border-border/60 bg-gradient-to-br from-primary/10 via-fuchsia-500/5 to-cyan-500/10">
                <item.icon className="h-10 w-10 text-primary" />
              </div>
              <h3 className="mt-3 text-sm font-medium text-foreground">{item.label}</h3>
              <p className="mt-0.5 text-[11px] text-muted-foreground">{item.group}</p>
              <div className="mt-auto rounded-md border border-border/60 bg-card/60 px-2.5 py-1.5 text-[11px] text-muted-foreground">
                Нажмите ↵, чтобы открыть
              </div>
            </>
          ) : null}
        </motion.div>
      </AnimatePresence>
    </div>
  );
}

function MetaRow({
  label, value, icon: Icon,
}: { label: string; value: string; icon?: typeof Clock }) {
  return (
    <div className="flex items-center justify-between rounded-md border border-border/40 bg-background/40 px-2 py-1.5">
      <span className="text-muted-foreground">{label}</span>
      <span className="flex items-center gap-1 font-mono text-foreground">
        {Icon && <Icon className="h-3 w-3 text-muted-foreground" />}
        {value}
      </span>
    </div>
  );
}
