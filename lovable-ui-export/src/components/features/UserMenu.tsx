import { useNavigate } from "@tanstack/react-router";
import {
  User, Settings, Palette, LifeBuoy, BookOpen, Users, Home, LogOut,
  Sun, Moon, Monitor, HelpCircle, Flag, Activity, FileText, Sparkles,
  Shield, ScrollText, ChevronUp, Keyboard,
} from "lucide-react";
import {
  DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuLabel,
  DropdownMenuSeparator, DropdownMenuShortcut, DropdownMenuSub,
  DropdownMenuSubContent, DropdownMenuSubTrigger, DropdownMenuTrigger,
  DropdownMenuRadioGroup, DropdownMenuRadioItem,
} from "@/components/ui/dropdown-menu";
import { cn } from "@/lib/utils";
import { useState } from "react";
import { ShortcutsDialog } from "@/components/features/ShortcutsDialog";

interface UserMenuProps {
  collapsed?: boolean;
}

const gradients = [
  { id: "aurora", label: "Аврора", css: "from-fuchsia-500 via-violet-500 to-indigo-500" },
  { id: "sunset", label: "Закат", css: "from-amber-400 via-rose-500 to-fuchsia-600" },
  { id: "ocean", label: "Океан", css: "from-cyan-400 via-sky-500 to-indigo-600" },
  { id: "forest", label: "Лес", css: "from-emerald-400 via-teal-500 to-cyan-600" },
];

export function UserMenu({ collapsed }: UserMenuProps) {
  const navigate = useNavigate();
  const [theme, setTheme] = useState("dark");
  const [gradient, setGradient] = useState("aurora");
  const [shortcutsOpen, setShortcutsOpen] = useState(false);

  const go = (to: string) => navigate({ to: to as "/" });

  return (
    <>
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <button
          className={cn(
            "group flex w-full items-center gap-2 rounded-md border border-border/60 bg-elevated/40 p-2 text-left transition-colors hover:bg-muted/40",
            collapsed && "justify-center border-transparent bg-transparent p-1",
          )}
        >
          <div className="grid h-7 w-7 shrink-0 place-items-center rounded-full bg-gradient-primary text-[10px] font-semibold text-primary-foreground">
            AX
          </div>
          {!collapsed && (
            <>
              <div className="min-w-0 flex-1">
                <p className="truncate text-xs font-medium">Александр</p>
                <p className="truncate text-[10px] text-muted-foreground">alex@istok.app</p>
              </div>
              <ChevronUp className="h-3.5 w-3.5 text-muted-foreground transition-transform group-data-[state=open]:rotate-180" />
            </>
          )}
        </button>
      </DropdownMenuTrigger>

      <DropdownMenuContent side="top" align="start" className="w-64">
        <DropdownMenuLabel className="flex items-center gap-2 py-2">
          <div className="grid h-8 w-8 place-items-center rounded-full bg-gradient-primary text-[11px] font-semibold text-primary-foreground">
            AX
          </div>
          <div className="min-w-0 flex-1">
            <p className="truncate text-sm font-medium">Александр</p>
            <p className="truncate text-[11px] font-normal text-muted-foreground">alex@istok.app</p>
          </div>
        </DropdownMenuLabel>
        <DropdownMenuSeparator />

        <DropdownMenuItem onSelect={() => go("/profile")}>
          <User /> Профиль
        </DropdownMenuItem>
        <DropdownMenuItem onSelect={() => go("/settings/account")}>
          <Settings /> Настройки
          <DropdownMenuShortcut>Ctrl .</DropdownMenuShortcut>
        </DropdownMenuItem>

        <DropdownMenuSub>
          <DropdownMenuSubTrigger>
            <Palette /> Внешний вид
          </DropdownMenuSubTrigger>
          <DropdownMenuSubContent className="w-60">
            <DropdownMenuLabel className="text-[10px] uppercase tracking-wider text-muted-foreground">
              Фоновый градиент
            </DropdownMenuLabel>
            <div className="grid grid-cols-4 gap-1.5 p-1.5">
              {gradients.map((g) => (
                <button
                  key={g.id}
                  onClick={() => setGradient(g.id)}
                  title={g.label}
                  className={cn(
                    "h-9 rounded-md bg-gradient-to-br ring-offset-2 ring-offset-popover transition",
                    g.css,
                    gradient === g.id && "ring-2 ring-primary",
                  )}
                />
              ))}
            </div>
            <DropdownMenuSeparator />
            <DropdownMenuLabel className="text-[10px] uppercase tracking-wider text-muted-foreground">
              Тема
            </DropdownMenuLabel>
            <DropdownMenuRadioGroup value={theme} onValueChange={setTheme}>
              <DropdownMenuRadioItem value="light"><Sun className="mr-2 h-4 w-4" />Светлая</DropdownMenuRadioItem>
              <DropdownMenuRadioItem value="dark"><Moon className="mr-2 h-4 w-4" />Тёмная</DropdownMenuRadioItem>
              <DropdownMenuRadioItem value="system"><Monitor className="mr-2 h-4 w-4" />Системная</DropdownMenuRadioItem>
            </DropdownMenuRadioGroup>
          </DropdownMenuSubContent>
        </DropdownMenuSub>

        <DropdownMenuSub>
          <DropdownMenuSubTrigger>
            <LifeBuoy /> Поддержка
          </DropdownMenuSubTrigger>
          <DropdownMenuSubContent className="w-52">
            <DropdownMenuItem onSelect={() => go("/help")}>
              <HelpCircle /> Центр помощи
            </DropdownMenuItem>
            <DropdownMenuItem>
              <Flag /> Пожаловаться
            </DropdownMenuItem>
            <DropdownMenuItem onSelect={() => go("/status")}>
              <Activity /> Статус
            </DropdownMenuItem>
          </DropdownMenuSubContent>
        </DropdownMenuSub>

        <DropdownMenuSub>
          <DropdownMenuSubTrigger>
            <BookOpen /> Документация
          </DropdownMenuSubTrigger>
          <DropdownMenuSubContent className="w-52">
            <DropdownMenuItem onSelect={() => go("/docs")}>
              <FileText /> Документация
            </DropdownMenuItem>
            <DropdownMenuItem onSelect={() => go("/prompts")}>
              <Sparkles /> Промты
            </DropdownMenuItem>
            <DropdownMenuItem onSelect={() => go("/terms")}>
              <ScrollText /> Условия и конфиденциальность
            </DropdownMenuItem>
            <DropdownMenuItem>
              <FileText /> История изменений
            </DropdownMenuItem>
            <DropdownMenuItem>
              <Shield /> Безопасность
            </DropdownMenuItem>
          </DropdownMenuSubContent>
        </DropdownMenuSub>

        <DropdownMenuSeparator />
        <DropdownMenuItem onSelect={() => setShortcutsOpen(true)}>
          <Keyboard /> Горячие клавиши
          <DropdownMenuShortcut>?</DropdownMenuShortcut>
        </DropdownMenuItem>
        <DropdownMenuItem>
          <Users /> Сообщество
        </DropdownMenuItem>
        <DropdownMenuItem onSelect={() => go("/")}>
          <Home /> Главная
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem className="text-destructive focus:text-destructive">
          <LogOut /> Выйти
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
    <ShortcutsDialog open={shortcutsOpen} onOpenChange={setShortcutsOpen} />
    </>
  );
}
