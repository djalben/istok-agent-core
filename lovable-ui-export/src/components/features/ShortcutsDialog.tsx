import { Keyboard } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

interface ShortcutsDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

const groups: { title: string; items: { keys: string[]; label: string }[] }[] = [
  {
    title: "Навигация",
    items: [
      { keys: ["Ctrl", "K"], label: "Глобальный поиск" },
      { keys: ["Ctrl", "B"], label: "Скрыть панель" },
      { keys: ["Ctrl", "."], label: "Открыть настройки" },
    ],
  },
  {
    title: "Чат и генерация",
    items: [
      { keys: ["Shift", "Enter"], label: "Перенос строки в чате" },
      { keys: ["Enter"], label: "Отправить сообщение" },
      { keys: ["Esc"], label: "Отменить генерацию" },
    ],
  },
];

function Key({ children }: { children: React.ReactNode }) {
  return (
    <kbd className="inline-flex h-6 min-w-6 items-center justify-center rounded-md border border-border/80 bg-elevated/80 px-1.5 font-mono text-[11px] font-medium text-foreground/90 shadow-sm">
      {children}
    </kbd>
  );
}

export function ShortcutsDialog({ open, onOpenChange }: ShortcutsDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <div className="flex items-center gap-2">
            <div className="grid h-8 w-8 place-items-center rounded-md bg-elevated/70">
              <Keyboard className="h-4 w-4 text-primary" />
            </div>
            <div>
              <DialogTitle>Горячие клавиши</DialogTitle>
              <DialogDescription>
                Управляйте Истоком быстрее с клавиатуры.
              </DialogDescription>
            </div>
          </div>
        </DialogHeader>

        <div className="space-y-5">
          {groups.map((g) => (
            <section key={g.title}>
              <h4 className="mb-2 text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
                {g.title}
              </h4>
              <ul className="divide-y divide-border/40 rounded-lg border border-border/60 bg-elevated/30">
                {g.items.map((it) => (
                  <li
                    key={it.label}
                    className="flex items-center justify-between gap-3 px-3 py-2.5"
                  >
                    <span className="text-sm text-foreground/90">{it.label}</span>
                    <span className="flex items-center gap-1">
                      {it.keys.map((k, i) => (
                        <span key={i} className="flex items-center gap-1">
                          {i > 0 && (
                            <span className="text-[10px] text-muted-foreground">+</span>
                          )}
                          <Key>{k}</Key>
                        </span>
                      ))}
                    </span>
                  </li>
                ))}
              </ul>
            </section>
          ))}
        </div>
      </DialogContent>
    </Dialog>
  );
}
