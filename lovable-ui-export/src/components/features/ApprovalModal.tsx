import { motion } from "framer-motion";
import { Check, FileText, Sparkles, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { businessPlanMarkdown } from "@/lib/mockData";

interface ApprovalModalProps {
  open: boolean;
  onApprove: () => void;
  onReject: () => void;
}

function renderMarkdown(md: string) {
  const lines = md.split("\n");
  const out: React.ReactNode[] = [];
  let list: string[] = [];
  let key = 0;
  const flushList = () => {
    if (list.length) {
      out.push(
        <ul key={key++} className="my-3 space-y-1.5 pl-4">
          {list.map((li, i) => (
            <li key={i} className="flex gap-2 text-sm text-foreground/85">
              <span className="mt-2 h-1 w-1 shrink-0 rounded-full bg-primary" />
              <span dangerouslySetInnerHTML={{ __html: inline(li) }} />
            </li>
          ))}
        </ul>,
      );
      list = [];
    }
  };
  const inline = (s: string) =>
    s
      .replace(/\*\*(.+?)\*\*/g, '<strong class="text-foreground font-semibold">$1</strong>')
      .replace(/`([^`]+)`/g, '<code class="rounded bg-elevated px-1 py-0.5 font-mono text-[12px] text-primary">$1</code>');

  for (const raw of lines) {
    const line = raw.trimEnd();
    if (!line.trim()) { flushList(); continue; }
    if (line.startsWith("# ")) { flushList(); out.push(<h1 key={key++} className="mt-2 text-2xl font-semibold tracking-tight">{line.slice(2)}</h1>); }
    else if (line.startsWith("## ")) { flushList(); out.push(<h2 key={key++} className="mt-5 text-base font-semibold text-foreground">{line.slice(3)}</h2>); }
    else if (line.startsWith("> ")) {
      flushList();
      out.push(
        <blockquote key={key++} className="my-3 rounded-md border-l-2 border-primary bg-primary/10 px-3 py-2 text-sm text-foreground/90">
          <span dangerouslySetInnerHTML={{ __html: inline(line.slice(2)) }} />
        </blockquote>,
      );
    }
    else if (/^\d+\.\s/.test(line) || line.startsWith("- ")) list.push(line.replace(/^(\d+\.|\-)\s/, ""));
    else { flushList(); out.push(<p key={key++} className="text-sm leading-relaxed text-foreground/85" dangerouslySetInnerHTML={{ __html: inline(line) }} />); }
  }
  flushList();
  return out;
}

export function ApprovalModal({ open, onApprove, onReject }: ApprovalModalProps) {
  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 grid place-items-center bg-background/70 p-4 backdrop-blur-md">
      <motion.div
        initial={{ opacity: 0, y: 16, scale: 0.98 }}
        animate={{ opacity: 1, y: 0, scale: 1 }}
        className="relative w-full max-w-2xl overflow-hidden rounded-2xl border border-border bg-card shadow-glow"
      >
        <div className="absolute inset-x-0 top-0 h-px bg-gradient-primary" />
        <div className="flex items-start justify-between gap-3 border-b border-border/60 p-5">
          <div className="flex gap-3">
            <div className="grid h-10 w-10 place-items-center rounded-lg bg-gradient-primary shadow-glow">
              <Sparkles className="h-5 w-5 text-primary-foreground" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h2 className="text-base font-semibold">Требуется подтверждение</h2>
                <span className="rounded-md border border-primary/40 bg-primary/15 px-1.5 py-0.5 font-mono text-[10px] uppercase text-primary">
                  HITL
                </span>
              </div>
              <p className="mt-0.5 text-xs text-muted-foreground">
                Агент-директор сделал паузу перед написанием кода. Изучите бизнес-план и подтвердите, чтобы продолжить.
              </p>
            </div>
          </div>
          <Button variant="ghost" size="icon" className="h-8 w-8" onClick={onReject} aria-label="Закрыть">
            <X className="h-4 w-4" />
          </Button>
        </div>

        <div className="flex items-center gap-2 border-b border-border/60 bg-elevated/40 px-5 py-2 font-mono text-[11px] text-muted-foreground">
          <FileText className="h-3.5 w-3.5 text-primary" /> business-plan.md
          <span className="ml-auto">подготовил Исследователь · 1.4k токенов</span>
        </div>

        <ScrollArea className="max-h-[50vh]">
          <div className="space-y-1 p-6">{renderMarkdown(businessPlanMarkdown)}</div>
        </ScrollArea>

        <div className="flex items-center justify-between gap-3 border-t border-border/60 bg-panel/50 p-4">
          <p className="text-xs text-muted-foreground">
            Стоимость <span className="font-mono text-foreground">$0.42</span> · ~24с работы агентов
          </p>
          <div className="flex gap-2">
            <Button variant="ghost" onClick={onReject}>
              Запросить изменения
            </Button>
            <Button
              onClick={onApprove}
              className="gap-1.5 bg-gradient-primary text-primary-foreground shadow-glow hover:opacity-90"
            >
              <Check className="h-4 w-4" /> Подтвердить и сгенерировать
            </Button>
          </div>
        </div>
      </motion.div>
    </div>
  );
}
