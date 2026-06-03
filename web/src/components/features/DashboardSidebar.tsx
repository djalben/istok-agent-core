import { useState } from "react";
import { motion } from "framer-motion";
import { Link } from "@tanstack/react-router";
import {
  Home, Search, BookOpen, Plug, FolderKanban, Star, UserCircle2,
  Share2, Gift, Zap, ChevronsUpDown, Sparkles,
  PanelLeft, Clock, Building2, Settings, UserPlus, Check, Plus,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu, DropdownMenuContent, DropdownMenuItem,
  DropdownMenuLabel, DropdownMenuSeparator, DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { Progress } from "@/components/ui/progress";
import { cn } from "@/lib/utils";
import { UserMenu } from "@/components/features/UserMenu";
import { CreateWorkspaceModal } from "@/components/features/CreateWorkspaceModal";
import { useProjects } from "@/hooks/useProjects";

export type DashboardSection =
  | "home" | "all" | "starred" | "mine" | "shared" | "resources" | "connectors";

export interface DashboardSidebarProps {
  collapsed: boolean;
  onToggle: () => void;
  active: DashboardSection;
  onSelect: (s: DashboardSection) => void;
  onOpenSearch: () => void;
  onSelectProject: (id: string) => void;
  onShareLovable?: () => void;
  mobile?: boolean;
}

const mainNav = [
  { id: "home" as const, label: "Главная", icon: Home },
  { id: "search" as const, label: "Поиск", icon: Search, kbd: "⌘K" },
  { id: "resources" as const, label: "Ресурсы", icon: BookOpen },
  { id: "connectors" as const, label: "Коннекторы", icon: Plug },
];

const projectNav = [
  { id: "all" as const, label: "Все проекты", icon: FolderKanban },
  { id: "starred" as const, label: "Избранное", icon: Star },
  { id: "mine" as const, label: "Созданные мной", icon: UserCircle2 },
  { id: "shared" as const, label: "Поделились со мной", icon: Share2 },
];

export function DashboardSidebar({
  collapsed, onToggle, active, onSelect, onOpenSearch, onSelectProject, onShareLovable, mobile = false,
}: DashboardSidebarProps) {
  const [createWsOpen, setCreateWsOpen] = useState(false);
  const { data: projects = [] } = useProjects();
  const recentProjects = projects.slice(0, 4);
  const width = collapsed ? "w-[64px]" : "w-[260px]";

  return (
    <motion.aside
      animate={{ width: collapsed ? 64 : 260 }}
      transition={{ type: "spring", stiffness: 220, damping: 28 }}
      className={cn(
        "z-30 flex h-screen shrink-0 flex-col border-r border-border/60 bg-card/40 backdrop-blur-xl",
        !mobile && "sticky top-0 hidden md:flex",
        width,
      )}
    >
      {/* Workspace switcher row */}
      <div className="flex items-center gap-1 border-b border-border/60 p-3">
        <WorkspaceSwitcher collapsed={collapsed} onCreate={() => setCreateWsOpen(true)} />
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="ghost"
              size="icon"
              className="h-7 w-7 shrink-0 text-muted-foreground"
              onClick={onToggle}
              aria-label={collapsed ? "Открыть боковую панель" : "Свернуть боковую панель"}
            >
              <PanelLeft className="h-3.5 w-3.5" />
            </Button>
          </TooltipTrigger>
          <TooltipContent side="right" className="text-xs">
            {collapsed ? "Открыть боковую панель" : "Свернуть боковую панель"}
            <span className="ml-1.5 rounded border border-border/60 bg-background/40 px-1 py-0.5 font-mono text-[9px] text-muted-foreground">
              Ctrl+B
            </span>
          </TooltipContent>
        </Tooltip>
      </div>

      <div className="flex-1 overflow-y-auto px-2 py-3">
        <NavGroup label="Основное" collapsed={collapsed}>
          {mainNav.map((item) => (
            <SidebarItem
              key={item.id}
              collapsed={collapsed}
              icon={item.icon}
              label={item.label}
              kbd={item.kbd}
              active={active === (item.id as DashboardSection)}
              onClick={() => {
                if (item.id === "search") onOpenSearch();
                else onSelect(item.id as DashboardSection);
              }}
            />
          ))}
        </NavGroup>

        <NavGroup label="Проекты" collapsed={collapsed}>
          {projectNav.map((item) => (
            <SidebarItem
              key={item.id}
              collapsed={collapsed}
              icon={item.icon}
              label={item.label}
              active={active === item.id}
              onClick={() => onSelect(item.id)}
            />
          ))}
        </NavGroup>

        {recentProjects.length > 0 && (
          <NavGroup label="Недавние" collapsed={collapsed}>
            {recentProjects.map((p) => (
              <SidebarItem
                key={p.id}
                collapsed={collapsed}
                icon={Clock}
                label={p.name}
                onClick={() => onSelectProject(p.id)}
              />
            ))}
          </NavGroup>
        )}
      </div>

      {/* Bottom cards */}
      <div className="space-y-2 border-t border-border/60 p-3">
        {!collapsed && (
          <>
            <PromoCard
              icon={Gift}
              title="Пригласить друзей"
              subtitle="Получите 50 бесплатных кредитов"
              tint="from-fuchsia-500/30 via-violet-500/15 to-transparent"
              iconTint="bg-fuchsia-500/20 text-fuchsia-300"
              onClick={onShareLovable}
            />
            <PromoCard
              icon={Zap}
              title="Перейти на Business"
              subtitle="Безлимитные генерации"
              tint="from-amber-400/30 via-orange-500/15 to-transparent"
              iconTint="bg-amber-400/20 text-amber-300"
              to="/settings/billing"
            />
          </>
        )}
        <UserMenu collapsed={collapsed} />
      </div>
      <CreateWorkspaceModal open={createWsOpen} onOpenChange={setCreateWsOpen} />
    </motion.aside>
  );
}

function WorkspaceSwitcher({ collapsed, onCreate }: { collapsed: boolean; onCreate?: () => void }) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <button
          className={cn(
            "group flex flex-1 items-center gap-2 rounded-md p-1.5 text-left transition-colors hover:bg-muted/40",
            collapsed && "justify-center",
          )}
        >
          <div className="grid h-7 w-7 shrink-0 place-items-center rounded-md bg-gradient-primary shadow-glow">
            <Sparkles className="h-3.5 w-3.5 text-primary-foreground" />
          </div>
          {!collapsed && (
            <>
              <div className="min-w-0 flex-1">
                <p className="truncate text-xs font-medium">Рабочее пространство Александра</p>
                <p className="truncate text-[10px] text-muted-foreground">Профиль + 1 участник</p>
              </div>
              <ChevronsUpDown className="h-3.5 w-3.5 text-muted-foreground" />
            </>
          )}
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="w-72">
        <div className="flex items-center gap-2 px-2 pb-2 pt-1">
          <div className="grid h-9 w-9 place-items-center rounded-md bg-gradient-primary shadow-glow">
            <Sparkles className="h-4 w-4 text-primary-foreground" />
          </div>
          <div className="min-w-0 flex-1">
            <p className="truncate text-sm font-semibold">Рабочее пространство Александра</p>
            <p className="truncate text-[11px] text-muted-foreground">Профиль + 1 участник</p>
          </div>
        </div>
        <DropdownMenuSeparator />
        <DropdownMenuItem asChild className="gap-2">
          <Link to={"/settings/workspace" as "/"}>
            <Settings className="h-4 w-4" /> Настройки
          </Link>
        </DropdownMenuItem>
        <DropdownMenuItem asChild className="gap-2">
          <Link to={"/settings/people" as "/"}>
            <UserPlus className="h-4 w-4" /> Пригласить участников
          </Link>
        </DropdownMenuItem>
        <DropdownMenuSeparator />

        {/* Credits */}
        <div className="px-2 py-2">
          <div className="mb-1.5 flex items-center justify-between text-[11px]">
            <span className="font-medium text-foreground">Кредиты</span>
            <span className="font-mono text-muted-foreground">93,6 / 100</span>
          </div>
          <Progress value={93.6} className="h-1.5" />
          <p className="mt-1.5 text-[10px] text-muted-foreground">
            Обновится 1 числа следующего месяца
          </p>
        </div>
        <DropdownMenuSeparator />

        <DropdownMenuLabel className="text-[10px] uppercase tracking-wider text-muted-foreground">
          Все рабочие пространства
        </DropdownMenuLabel>
        <DropdownMenuItem className="gap-2">
          <div className="grid h-6 w-6 place-items-center rounded bg-gradient-primary">
            <Sparkles className="h-3 w-3 text-primary-foreground" />
          </div>
          <span className="flex-1">Рабочее пространство Александра</span>
          <Check className="h-3.5 w-3.5 text-primary" />
        </DropdownMenuItem>
        <DropdownMenuItem className="gap-2">
          <div className="grid h-6 w-6 place-items-center rounded bg-elevated">
            <Building2 className="h-3 w-3 text-muted-foreground" />
          </div>
          Acme Inc.
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem onSelect={() => onCreate?.()} className="gap-2 text-primary focus:text-primary">
          <Plus className="h-4 w-4" /> Создать новое
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

function NavGroup({
  label, collapsed, children,
}: { label: string; collapsed: boolean; children: React.ReactNode }) {
  return (
    <div className="mb-3">
      {!collapsed && (
        <p className="px-2 pb-1.5 text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
          {label}
        </p>
      )}
      <div className="space-y-0.5">{children}</div>
    </div>
  );
}

function SidebarItem({
  icon: Icon, label, kbd, active, collapsed, onClick,
}: {
  icon: typeof Home; label: string; kbd?: string;
  active?: boolean; collapsed: boolean; onClick: () => void;
}) {
  const btn = (
    <button
      onClick={onClick}
      className={cn(
        "group relative flex w-full items-center gap-2.5 rounded-md px-2 py-1.5 text-sm transition-colors",
        active
          ? "bg-primary/10 text-foreground"
          : "text-muted-foreground hover:bg-muted/40 hover:text-foreground",
        collapsed && "justify-center px-0",
      )}
    >
      {active && (
        <motion.span
          layoutId="dash-sidebar-active"
          className="absolute inset-y-1 left-0 w-0.5 rounded-r bg-primary"
        />
      )}
      <Icon className={cn("h-4 w-4 shrink-0", active && "text-primary")} />
      {!collapsed && (
        <>
          <span className="flex-1 truncate text-left">{label}</span>
          {kbd && (
            <span className="rounded border border-border/60 bg-background/40 px-1 py-0.5 font-mono text-[9px] text-muted-foreground">
              {kbd}
            </span>
          )}
        </>
      )}
    </button>
  );
  if (!collapsed) return btn;
  return (
    <Tooltip>
      <TooltipTrigger asChild>{btn}</TooltipTrigger>
      <TooltipContent side="right" className="text-xs">{label}</TooltipContent>
    </Tooltip>
  );
}

function PromoCard({
  icon: Icon, title, subtitle, tint, iconTint, onClick, to,
}: {
  icon: typeof Gift; title: string; subtitle: string;
  tint: string; iconTint: string;
  onClick?: () => void;
  to?: string;
}) {
  const inner = (
    <>
      <div className={cn("absolute inset-0 bg-gradient-to-br opacity-80", tint)} />
      <div className="relative flex items-center gap-2.5">
        <div className={cn("grid h-7 w-7 place-items-center rounded-md", iconTint)}>
          <Icon className="h-3.5 w-3.5" />
        </div>
        <div className="min-w-0 flex-1">
          <p className="truncate text-xs font-medium text-foreground">{title}</p>
          <p className="truncate text-[10px] text-muted-foreground">{subtitle}</p>
        </div>
      </div>
    </>
  );
  const className = cn(
    "group relative block w-full overflow-hidden rounded-lg border border-border/60 bg-card/60 p-2.5 text-left transition-all hover:border-primary/40",
  );
  if (to) {
    return <Link to={to as "/"} className={className}>{inner}</Link>;
  }
  return (
    <button type="button" onClick={onClick} className={className}>
      {inner}
    </button>
  );
}
