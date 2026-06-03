import { useMemo, useState } from "react";
import { Link } from "@tanstack/react-router";
import { motion } from "framer-motion";
import {
  ArrowUpRight, Clock, LayoutGrid, List as ListIcon, Plus, Search,
  ChevronDown, Eye, Activity, Users,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger,
  DropdownMenuLabel, DropdownMenuSeparator,
} from "@/components/ui/dropdown-menu";
import { Skeleton } from "@/components/ui/skeleton";
import { useProjects } from "@/hooks/useProjects";
import { type Project } from "@/lib/projectDisplay";
import { cn } from "@/lib/utils";

type ViewMode = "grid" | "list";

const groups = [
  { id: "14d", label: "Активны за последние 14 дней", max: 14 },
  { id: "60d", label: "Активны за последние 60 дней", max: 60 },
  { id: "older", label: "Ранее", max: Infinity },
] as const;

export function ProjectsView({ title, subtitle }: { title: string; subtitle: string }) {
  const [query, setQuery] = useState("");
  const [view, setView] = useState<ViewMode>("grid");
  const [sort, setSort] = useState("Последние изменения");
  const [visibility, setVisibility] = useState("Любая видимость");
  const [status, setStatus] = useState("Любой статус");
  const [creator, setCreator] = useState("Все авторы");

  const { data: projects = [], isLoading } = useProjects();

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    return projects.filter((p) =>
      !q || p.name.toLowerCase().includes(q) || p.description.toLowerCase().includes(q),
    );
  }, [query, projects]);

  const grouped = useMemo(() => {
    const seen = new Set<string>();
    const dayMs = 86_400_000;
    return groups.map((g) => {
      const items = filtered.filter((p) => {
        const d = p.updatedAtMs ? Math.floor((Date.now() - p.updatedAtMs) / dayMs) : 9999;
        const inGroup = d <= g.max && !seen.has(p.id);
        if (inGroup) seen.add(p.id);
        return inGroup;
      });
      return { ...g, items };
    });
  }, [filtered]);

  return (
    <div className="mx-auto max-w-7xl px-6 py-10">
      {/* Header */}
      <div className="mb-6 flex flex-col gap-2">
        <h1 className="text-2xl font-semibold tracking-tight">{title}</h1>
        <p className="text-sm text-muted-foreground">{subtitle}</p>
      </div>

      {/* Action bar */}
      <div className="mb-6 flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
        <div className="flex flex-1 items-center gap-2">
          <div className="relative max-w-sm flex-1">
            <Search className="pointer-events-none absolute left-3 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-muted-foreground" />
            <Input
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Поиск проектов..."
              className="h-9 pl-9"
            />
          </div>
          <FilterMenu icon={Clock} label={sort} options={["Последние изменения", "Недавно созданные", "По названию А-Я"]} onSelect={setSort} />
          <FilterMenu icon={Eye} label={visibility} options={["Любая видимость", "Приватные", "Публичные", "Рабочее пространство"]} onSelect={setVisibility} />
          <FilterMenu icon={Activity} label={status} options={["Любой статус", "Черновик", "Опубликован", "В архиве"]} onSelect={setStatus} />
          <FilterMenu icon={Users} label={creator} options={["Все авторы", "Я", "Участники пространства"]} onSelect={setCreator} />
        </div>

        <div className="flex items-center gap-2">
          <div className="flex items-center rounded-md border border-border/60 bg-card/40 p-0.5">
            <ToggleBtn active={view === "list"} onClick={() => setView("list")} aria="Список">
              <ListIcon className="h-3.5 w-3.5" />
            </ToggleBtn>
            <ToggleBtn active={view === "grid"} onClick={() => setView("grid")} aria="Сетка">
              <LayoutGrid className="h-3.5 w-3.5" />
            </ToggleBtn>
          </div>
          <Link to="/builder">
            <Button size="sm" className="bg-gradient-primary text-primary-foreground shadow-glow hover:opacity-90">
              <Plus className="h-4 w-4" /> Создать
            </Button>
          </Link>
        </div>
      </div>

      {/* Groups */}
      <div className="space-y-10">
        {isLoading && (
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {Array.from({ length: 6 }).map((_, i) => (
              <div key={i} className="overflow-hidden rounded-xl border border-border/60 bg-card">
                <Skeleton className="h-32 w-full rounded-none" />
                <div className="space-y-2 p-4">
                  <Skeleton className="h-4 w-2/3" />
                  <Skeleton className="h-3 w-full" />
                </div>
              </div>
            ))}
          </div>
        )}
        {!isLoading && grouped.map((g) =>
          g.items.length === 0 ? null : (
            <section key={g.id}>
              <div className="mb-3 flex items-center gap-2">
                <h2 className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
                  {g.label}
                </h2>
                <span className="rounded-full border border-border/60 bg-elevated/40 px-1.5 py-0.5 text-[10px] text-muted-foreground">
                  {g.items.length}
                </span>
              </div>
              {view === "grid" ? <GridView items={g.items} /> : <ListView items={g.items} />}
            </section>
          ),
        )}
        {!isLoading && filtered.length === 0 && (
          <div className="rounded-xl border border-dashed border-border/60 bg-card/30 px-6 py-16 text-center text-sm text-muted-foreground">
            {query ? `Нет проектов по запросу «${query}».` : "Пока нет проектов. Создайте первый, чтобы начать."}
          </div>
        )}
      </div>
    </div>
  );
}

function ToggleBtn({
  active, onClick, children, aria,
}: { active: boolean; onClick: () => void; children: React.ReactNode; aria: string }) {
  return (
    <button
      onClick={onClick}
      aria-label={aria}
      className={cn(
        "grid h-7 w-7 place-items-center rounded transition-colors",
        active ? "bg-elevated text-foreground" : "text-muted-foreground hover:text-foreground",
      )}
    >
      {children}
    </button>
  );
}

function FilterMenu({
  icon: Icon, label, options, onSelect,
}: {
  icon: typeof Clock; label: string; options: string[]; onSelect: (s: string) => void;
}) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="outline"
          size="sm"
          className="h-9 gap-1.5 border-border/60 bg-card/40 text-xs font-normal text-muted-foreground hover:text-foreground"
        >
          <Icon className="h-3.5 w-3.5" />
          {label}
          <ChevronDown className="h-3 w-3 opacity-60" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="w-48">
        <DropdownMenuLabel className="text-[10px] uppercase tracking-wider text-muted-foreground">
          {label}
        </DropdownMenuLabel>
        <DropdownMenuSeparator />
        {options.map((o) => (
          <DropdownMenuItem key={o} onClick={() => onSelect(o)}>
            {o}
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

function GridView({ items }: { items: Project[] }) {
  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {items.map((project, i) => (
        <motion.div
          key={project.id}
          initial={{ opacity: 0, y: 8 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: i * 0.03 }}
        >
          <Link to="/builder/$id" params={{ id: project.id }}>
            <article className="group relative overflow-hidden rounded-xl border border-border/80 bg-card transition-all hover:border-primary/50 hover:shadow-glow">
              <div className={`relative h-32 overflow-hidden ${project.gradient}`}>
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
        </motion.div>
      ))}
    </div>
  );
}

function ListView({ items }: { items: Project[] }) {
  return (
    <div className="overflow-hidden rounded-xl border border-border/60 bg-card/40">
      {items.map((p, i) => (
        <Link key={p.id} to="/builder/$id" params={{ id: p.id }}>
          <div
            className={cn(
              "group flex items-center gap-4 px-4 py-3 transition-colors hover:bg-muted/30",
              i !== items.length - 1 && "border-b border-border/40",
            )}
          >
            <div className={cn("h-10 w-10 shrink-0 rounded-md", p.gradient)} />
            <div className="min-w-0 flex-1">
              <div className="flex items-center gap-2">
                <h3 className="truncate text-sm font-medium">{p.name}</h3>
                <span className="rounded border border-border/60 bg-background/40 px-1.5 py-0.5 font-mono text-[9px] uppercase tracking-wider text-muted-foreground">
                  {p.framework}
                </span>
              </div>
              <p className="truncate text-xs text-muted-foreground">{p.description}</p>
            </div>
            <div className="hidden items-center gap-1 text-xs text-muted-foreground sm:flex">
              <Clock className="h-3 w-3" /> {p.updatedAt}
            </div>
            <ArrowUpRight className="h-4 w-4 text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100" />
          </div>
        </Link>
      ))}
    </div>
  );
}
