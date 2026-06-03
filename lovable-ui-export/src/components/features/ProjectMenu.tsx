import { useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import {
  ChevronDown,
  Sparkles,
  Settings,
  Plug,
  RefreshCw,
  UploadCloud,
  Pencil,
  Star,
  FolderInput,
  Info,
  Moon,
  Sun,
  Monitor,
  HelpCircle,
  ChevronRight,
  Gift,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
  DropdownMenuTrigger,
  DropdownMenuPortal,
} from "@/components/ui/dropdown-menu";
import { toast } from "sonner";

interface ProjectMenuProps {
  projectName?: string;
}

const credits = { used: 6.4, total: 100 };

export function ProjectMenu({ projectName }: ProjectMenuProps) {
  const [open, setOpen] = useState(false);
  const remaining = (credits.total - credits.used).toFixed(1);
  const pct = ((credits.total - credits.used) / credits.total) * 100;

  return (
    <DropdownMenu open={open} onOpenChange={setOpen}>
      <DropdownMenuTrigger asChild>
        <Button
          variant="ghost"
          size="sm"
          className="h-8 gap-1.5 px-2 text-sm font-medium text-foreground/90 hover:bg-elevated"
        >
          <span className="text-muted-foreground">Исток</span>
          {projectName && (
            <>
              <span className="text-muted-foreground/60">/</span>
              <span className="max-w-[160px] truncate">{projectName}</span>
            </>
          )}
          <ChevronDown
            className={`h-3.5 w-3.5 text-muted-foreground transition-transform ${
              open ? "rotate-180" : ""
            }`}
          />
        </Button>
      </DropdownMenuTrigger>
      <AnimatePresence>
        {open && (
          <DropdownMenuContent
            asChild
            align="start"
            sideOffset={8}
            className="w-72 border-border/70 bg-popover/95 p-1.5 backdrop-blur-xl"
          >
            <motion.div
              initial={{ opacity: 0, y: -6, scale: 0.97 }}
              animate={{ opacity: 1, y: 0, scale: 1 }}
              exit={{ opacity: 0, y: -6, scale: 0.97 }}
              transition={{ duration: 0.14, ease: "easeOut" }}
            >
              {/* Credits */}
              <div className="px-2 pb-2 pt-1.5">
                <div className="mb-1.5 flex items-center justify-between">
                  <span className="text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
                    Кредиты
                  </span>
                  <span className="font-mono text-[11px] font-medium text-foreground">
                    {remaining} <span className="text-muted-foreground">осталось</span>
                  </span>
                </div>
                <div className="h-1.5 overflow-hidden rounded-full bg-elevated">
                  <motion.div
                    initial={{ width: 0 }}
                    animate={{ width: `${pct}%` }}
                    transition={{ duration: 0.6, ease: "easeOut" }}
                    className="h-full rounded-full bg-gradient-primary shadow-glow"
                  />
                </div>
                <div className="mt-1 flex justify-between font-mono text-[10px] text-muted-foreground">
                  <span>использовано: {credits.used}</span>
                  <span>всего: {credits.total}</span>
                </div>
              </div>

              <DropdownMenuItem
                onSelect={(e) => {
                  e.preventDefault();
                  toast.success("Акция с бесплатными кредитами активирована");
                  setOpen(false);
                }}
                className="gap-2 rounded-md px-2 py-1.5 text-xs font-medium text-primary focus:bg-primary/10 focus:text-primary"
              >
                <Gift className="h-3.5 w-3.5" /> Получить бесплатные кредиты
              </DropdownMenuItem>

              <DropdownMenuSeparator className="my-1" />

              <Row icon={<Settings className="h-3.5 w-3.5" />} label="Настройки" />
              <Row icon={<Plug className="h-3.5 w-3.5" />} label="Коннекторы" />
              <Row icon={<RefreshCw className="h-3.5 w-3.5" />} label="Пересоздать проект" />

              <DropdownMenuSeparator className="my-1" />

              <Row
                icon={<UploadCloud className="h-3.5 w-3.5" />}
                label="Опубликовать в профиле"
                badge="Новое"
              />
              <Row icon={<Pencil className="h-3.5 w-3.5" />} label="Переименовать проект" />
              <Row icon={<Star className="h-3.5 w-3.5" />} label="В избранное" />
              <Row icon={<FolderInput className="h-3.5 w-3.5" />} label="Переместить в папку" />
              <Row icon={<Info className="h-3.5 w-3.5" />} label="Подробности" shortcut="⌘I" />

              <DropdownMenuSeparator className="my-1" />

              <DropdownMenuSub>
                <DropdownMenuSubTrigger className="flex cursor-pointer items-center gap-2 rounded-md px-2 py-1.5 text-xs focus:bg-elevated data-[state=open]:bg-elevated">
                  <Moon className="h-3.5 w-3.5 text-muted-foreground" />
                  <span className="font-medium">Внешний вид</span>
                  <span className="ml-auto font-mono text-[10px] text-muted-foreground">
                    Системная
                  </span>
                  <ChevronRight className="h-3.5 w-3.5 text-muted-foreground" />
                </DropdownMenuSubTrigger>
                <DropdownMenuPortal>
                  <DropdownMenuSubContent className="w-44 border-border/70 bg-popover/95 p-1.5 backdrop-blur-xl">
                    <SubRow icon={<Sun className="h-3.5 w-3.5" />} label="Светлая" />
                    <SubRow icon={<Moon className="h-3.5 w-3.5" />} label="Тёмная" active />
                    <SubRow icon={<Monitor className="h-3.5 w-3.5" />} label="Системная" />
                  </DropdownMenuSubContent>
                </DropdownMenuPortal>
              </DropdownMenuSub>

              <Row icon={<HelpCircle className="h-3.5 w-3.5" />} label="Помощь" shortcut="?" />
            </motion.div>
          </DropdownMenuContent>
        )}
      </AnimatePresence>
    </DropdownMenu>
  );
}

function Row({
  icon,
  label,
  badge,
  shortcut,
}: {
  icon: React.ReactNode;
  label: string;
  badge?: string;
  shortcut?: string;
}) {
  return (
    <DropdownMenuItem
      onSelect={(e) => {
        e.preventDefault();
        toast(label);
      }}
      className="flex cursor-pointer items-center gap-2 rounded-md px-2 py-1.5 text-xs focus:bg-elevated"
    >
      <span className="text-muted-foreground">{icon}</span>
      <span className="font-medium">{label}</span>
      {badge && (
        <Badge className="ml-auto h-4 border-0 bg-gradient-primary px-1.5 text-[9px] font-medium text-primary-foreground">
          {badge}
        </Badge>
      )}
      {shortcut && !badge && (
        <span className="ml-auto font-mono text-[10px] text-muted-foreground">{shortcut}</span>
      )}
    </DropdownMenuItem>
  );
}

function SubRow({
  icon,
  label,
  active,
}: {
  icon: React.ReactNode;
  label: string;
  active?: boolean;
}) {
  return (
    <DropdownMenuItem
      onSelect={(e) => {
        e.preventDefault();
        toast(`Внешний вид: ${label}`);
      }}
      className="flex cursor-pointer items-center gap-2 rounded-md px-2 py-1.5 text-xs focus:bg-elevated"
    >
      <span className={active ? "text-primary" : "text-muted-foreground"}>{icon}</span>
      <span className="font-medium">{label}</span>
      {active && <Sparkles className="ml-auto h-3 w-3 text-primary" />}
    </DropdownMenuItem>
  );
}
