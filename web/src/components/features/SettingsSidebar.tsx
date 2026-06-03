import { Link, useRouterState } from "@tanstack/react-router";
import { ArrowLeft, User, Smartphone, Settings as Cog, CreditCard, Cloud, Users, BookOpen, Sparkles, LayoutTemplate, Palette, GitBranch, Globe, Lock, ShieldCheck, ClipboardList, FolderKanban } from "lucide-react";
import { cn } from "@/lib/utils";

type Item = { to: string; label: string; icon: React.ComponentType<{ className?: string }>; badge?: "Business" | "Enterprise" };
type Group = { title: string; items: Item[] };

const groups: Group[] = [
  {
    title: "Проект",
    items: [
      { to: "/settings/project", label: "Обзор", icon: FolderKanban },
    ],
  },

  {
    title: "Счёт",
    items: [
      { to: "/settings/account", label: "Профиль", icon: User },
      { to: "/settings/apps", label: "Устройства и приложения", icon: Smartphone },
    ],
  },
  {
    title: "Рабочее пространство",
    items: [
      { to: "/settings/workspace", label: "Настройки", icon: Cog },
      { to: "/settings/billing", label: "Планы и кредиты", icon: CreditCard },
      { to: "/settings/cloud", label: "Баланс Облака и ИИ", icon: Cloud },
    ],
  },
  {
    title: "Членство и доступ",
    items: [{ to: "/settings/people", label: "Люди", icon: Users }],
  },
  {
    title: "Настройка",
    items: [
      { to: "/settings/knowledge", label: "База знаний", icon: BookOpen },
      { to: "/settings/skills", label: "Навыки", icon: Sparkles },
      { to: "/settings/templates", label: "Шаблоны", icon: LayoutTemplate, badge: "Business" },
      { to: "/settings/design-systems", label: "Системы проектирования", icon: Palette, badge: "Enterprise" },
    ],
  },
  {
    title: "Сборка и развёртывание",
    items: [
      { to: "/settings/git", label: "Git", icon: GitBranch },
      { to: "/settings/domains", label: "Домены рабочей области", icon: Globe },
    ],
  },
  {
    title: "Безопасность",
    items: [
      { to: "/settings/privacy", label: "Конфиденциальность и безопасность", icon: Lock },
      { to: "/settings/security-center", label: "Центр безопасности", icon: ShieldCheck, badge: "Business" },
      { to: "/settings/audit-logs", label: "Журналы аудита", icon: ClipboardList, badge: "Enterprise" },
    ],
  },
];

function Badge({ tier }: { tier: "Business" | "Enterprise" }) {
  return (
    <span
      className={cn(
        "ml-auto rounded-md px-1.5 py-0.5 text-[9px] font-medium uppercase tracking-wider",
        tier === "Business"
          ? "bg-violet-500/15 text-violet-300"
          : "bg-amber-500/15 text-amber-300",
      )}
    >
      {tier}
    </span>
  );
}

export function SettingsSidebar() {
  const path = useRouterState({ select: (s) => s.location.pathname });
  return (
    <aside className="hidden w-64 shrink-0 border-r border-border/60 bg-card/20 lg:flex lg:flex-col">
      <div className="border-b border-border/60 p-4">
        <Link
          to="/"
          className="inline-flex items-center gap-1.5 text-xs text-muted-foreground transition-colors hover:text-foreground"
        >
          <ArrowLeft className="h-3.5 w-3.5" /> Назад
        </Link>
        <h2 className="mt-3 text-lg font-semibold">Настройки</h2>
      </div>
      <nav className="flex-1 overflow-y-auto px-2 py-3">
        {groups.map((g) => (
          <div key={g.title} className="mb-4">
            <p className="px-2 pb-1.5 text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
              {g.title}
            </p>
            <div className="space-y-0.5">
              {g.items.map((it) => {
                const active = path === it.to;
                const Icon = it.icon;
                return (
                  <Link
                    key={it.to}
                    to={it.to}
                    className={cn(
                      "group flex items-center gap-2 rounded-md px-2 py-1.5 text-xs transition-colors",
                      active
                        ? "bg-muted/60 text-foreground"
                        : "text-muted-foreground hover:bg-muted/30 hover:text-foreground",
                    )}
                  >
                    <Icon className="h-3.5 w-3.5 shrink-0" />
                    <span className="truncate">{it.label}</span>
                    {it.badge && <Badge tier={it.badge} />}
                  </Link>
                );
              })}
            </div>
          </div>
        ))}
      </nav>
    </aside>
  );
}
