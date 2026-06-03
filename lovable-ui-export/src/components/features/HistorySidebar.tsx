import { useState } from "react";
import { History, ChevronRight, RotateCcw, GitCommit } from "lucide-react";
import { cn } from "@/lib/utils";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Button } from "@/components/ui/button";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { toast } from "sonner";

interface HistoryItem {
  id: string;
  title: string;
  time: string;
  hash: string;
}

const items: HistoryItem[] = [
  { id: "1", title: "Создание Hero-блока", time: "только что", hash: "a1f3c2" },
  { id: "2", title: "Добавление кнопок CTA", time: "2 мин назад", hash: "9c2e1d" },
  { id: "3", title: "Настройка маршрутов", time: "5 мин назад", hash: "73b88a" },
  { id: "4", title: "Подключение базы данных", time: "12 мин назад", hash: "55da90" },
  { id: "5", title: "Инициализация проекта", time: "18 мин назад", hash: "0011ff" },
];

export function HistorySidebar() {
  const [open, setOpen] = useState(true);
  const [activeId, setActiveId] = useState("1");

  return (
    <aside
      className={cn(
        "hidden shrink-0 flex-col border-l border-border/60 bg-panel/40 transition-[width] duration-200 md:flex",
        open ? "w-64" : "w-10",
      )}
    >
      <button
        onClick={() => setOpen((v) => !v)}
        className="flex h-10 items-center gap-2 border-b border-border/60 px-3 text-xs font-medium text-muted-foreground hover:text-foreground"
        aria-label={open ? "Скрыть историю" : "Показать историю"}
      >
        <History className="h-3.5 w-3.5 text-primary" />
        {open && <span className="flex-1 text-left">История</span>}
        <ChevronRight
          className={cn(
            "h-3.5 w-3.5 transition-transform",
            open && "rotate-180",
          )}
        />
      </button>

      {open && (
        <ScrollArea className="flex-1">
          <ol className="relative space-y-px p-2">
            <span className="absolute bottom-2 left-[18px] top-2 w-px bg-border/60" />
            {items.map((it, i) => {
              const isActive = it.id === activeId;
              return (
                <li
                  key={it.id}
                  className={cn(
                    "group relative flex items-start gap-2 rounded-md px-2 py-2 transition-colors",
                    isActive ? "bg-elevated/70" : "hover:bg-elevated/40",
                  )}
                >
                  <button
                    onClick={() => setActiveId(it.id)}
                    className="absolute inset-0 rounded-md"
                    aria-label={it.title}
                  />
                  <span
                    className={cn(
                      "relative z-10 mt-0.5 grid h-5 w-5 shrink-0 place-items-center rounded-full border bg-background",
                      isActive
                        ? "border-primary text-primary shadow-glow"
                        : "border-border/80 text-muted-foreground",
                    )}
                  >
                    <GitCommit className="h-3 w-3" />
                  </span>
                  <div className="relative z-10 min-w-0 flex-1">
                    <p
                      className={cn(
                        "truncate text-xs font-medium",
                        isActive ? "text-foreground" : "text-foreground/80",
                      )}
                    >
                      {it.title}
                    </p>
                    <div className="mt-0.5 flex items-center gap-1.5 text-[10px] text-muted-foreground">
                      <span className="font-mono">{it.hash}</span>
                      <span>·</span>
                      <span>{it.time}</span>
                    </div>
                  </div>
                  {i > 0 && (
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <Button
                          variant="ghost"
                          size="icon"
                          className="relative z-10 h-6 w-6 opacity-0 transition-opacity group-hover:opacity-100"
                          onClick={(e) => {
                            e.stopPropagation();
                            toast.success(`Откат к версии ${it.hash}`);
                          }}
                        >
                          <RotateCcw className="h-3 w-3" />
                        </Button>
                      </TooltipTrigger>
                      <TooltipContent side="left">Откатить</TooltipContent>
                    </Tooltip>
                  )}
                </li>
              );
            })}
          </ol>
        </ScrollArea>
      )}
    </aside>
  );
}
