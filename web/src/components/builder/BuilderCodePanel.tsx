import { useMemo, useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { ChevronRight, File, Folder, FolderOpen, X } from "lucide-react";
import { ScrollArea } from "@/components/ui/scroll-area";
import { cn } from "@/lib/utils";
import { filesToTree, flattenFiles, type FileNode } from "@/lib/builderTypes";

const langColor: Record<string, string> = {
  tsx: "text-sky-400",
  ts: "text-sky-400",
  jsx: "text-sky-400",
  js: "text-sky-400",
  css: "text-pink-400",
  json: "text-amber-400",
  md: "text-emerald-400",
  html: "text-orange-400",
};

function FileTreeNode({
  node,
  depth,
  activePath,
  onSelect,
}: {
  node: FileNode;
  depth: number;
  activePath: string;
  onSelect: (n: FileNode) => void;
}) {
  const [open, setOpen] = useState(true);
  const isActive = node.path === activePath;
  if (node.type === "folder") {
    return (
      <div>
        <button
          onClick={() => setOpen((o) => !o)}
          className="flex w-full items-center gap-1 rounded px-1.5 py-1 text-left text-xs text-foreground/80 hover:bg-elevated"
          style={{ paddingLeft: 6 + depth * 12 }}
        >
          <ChevronRight className={cn("h-3 w-3 transition-transform", open && "rotate-90")} />
          {open ? <FolderOpen className="h-3.5 w-3.5 text-primary" /> : <Folder className="h-3.5 w-3.5 text-primary" />}
          <span className="font-medium">{node.name}</span>
        </button>
        <AnimatePresence initial={false}>
          {open && (
            <motion.div initial={{ height: 0, opacity: 0 }} animate={{ height: "auto", opacity: 1 }} exit={{ height: 0, opacity: 0 }} className="overflow-hidden">
              {node.children?.map((c) => (
                <FileTreeNode key={c.path} node={c} depth={depth + 1} activePath={activePath} onSelect={onSelect} />
              ))}
            </motion.div>
          )}
        </AnimatePresence>
      </div>
    );
  }
  const ext = node.name.split(".").pop() ?? "";
  return (
    <button
      onClick={() => onSelect(node)}
      className={cn(
        "flex w-full items-center gap-1.5 rounded px-1.5 py-1 text-left text-xs",
        isActive ? "bg-primary/15 text-foreground" : "text-foreground/70 hover:bg-elevated",
      )}
      style={{ paddingLeft: 6 + depth * 12 }}
    >
      <File className={cn("h-3.5 w-3.5", langColor[ext] ?? "text-muted-foreground")} />
      <span>{node.name}</span>
    </button>
  );
}

function highlight(code: string, lang?: string): string {
  let html = code
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;");
  if (lang === "md") return html;
  html = html.replace(/(\/\/[^\n]*)/g, '<span class="text-muted-foreground/70">$1</span>');
  html = html.replace(/("[^"\n]*"|'[^'\n]*'|`[^`\n]*`)/g, '<span class="text-emerald-400">$1</span>');
  html = html.replace(
    /\b(import|from|export|default|return|const|let|var|function|if|else|for|of|in|new|async|await|class|extends|interface|type)\b/g,
    '<span class="text-violet-400">$1</span>',
  );
  html = html.replace(/\b([A-Z][A-Za-z0-9_]+)\b/g, '<span class="text-sky-300">$1</span>');
  html = html.replace(/(&lt;\/?)([a-zA-Z][\w-]*)/g, '$1<span class="text-pink-400">$2</span>');
  return html;
}

interface BuilderCodePanelProps {
  projectFiles: Record<string, string>;
}

export function BuilderCodePanel({ projectFiles }: BuilderCodePanelProps) {
  const tree = useMemo(() => filesToTree(projectFiles), [projectFiles]);
  const allFiles = useMemo(() => flattenFiles(tree), [tree]);
  const [activePath, setActivePath] = useState<string>("");
  const [openTabs, setOpenTabs] = useState<string[]>([]);

  const resolvedActive = activePath || allFiles[0]?.path || "";
  const active = allFiles.find((f) => f.path === resolvedActive);
  const isEmpty = allFiles.length === 0;
  const tabs = openTabs.length ? openTabs : allFiles.slice(0, 2).map((f) => f.path);

  const openFile = (n: FileNode) => {
    setActivePath(n.path);
    setOpenTabs((t) => {
      const base = t.length ? t : allFiles.slice(0, 2).map((f) => f.path);
      return base.includes(n.path) ? base : [...base, n.path];
    });
  };
  const closeTab = (p: string, e: React.MouseEvent) => {
    e.stopPropagation();
    setOpenTabs((t) => {
      const base = t.length ? t : allFiles.slice(0, 2).map((f) => f.path);
      const next = base.filter((x) => x !== p);
      if (resolvedActive === p && next.length) setActivePath(next[next.length - 1]);
      return next;
    });
  };

  const lineCount = active?.content?.split("\n").length ?? 0;

  return (
    <div className="flex h-full flex-col bg-background">
      <div className="flex h-10 items-center justify-between border-b border-border/60 bg-panel px-3">
        <div className="flex items-center gap-2">
          <div className="h-2 w-2 rounded-full bg-primary animate-pulse" />
          <span className="text-xs font-medium text-muted-foreground">Рабочая область</span>
        </div>
        <span className="font-mono text-[10px] text-muted-foreground">
          {isEmpty ? "файлов ещё нет" : `${allFiles.length} файлов`}
        </span>
      </div>

      <div className="flex min-h-0 flex-1">
        {/* Tree */}
        <div className="w-56 shrink-0 border-r border-border/60 bg-panel/50">
          <div className="px-3 py-2 font-mono text-[10px] uppercase tracking-wider text-muted-foreground">
            Проводник
          </div>
          <ScrollArea className="h-[calc(100%-32px)]">
            <div className="px-2 pb-4">
              {isEmpty ? (
                <div className="px-2 py-6 text-center text-[11px] text-muted-foreground">
                  Файлы появятся здесь, как только агенты их сгенерируют.
                </div>
              ) : (
                tree.map((f) => (
                  <FileTreeNode key={f.path} node={f} depth={0} activePath={resolvedActive} onSelect={openFile} />
                ))
              )}
            </div>
          </ScrollArea>
        </div>

        {/* Editor */}
        <div className="flex min-w-0 flex-1 flex-col">
          <div className="flex h-9 items-center gap-px overflow-x-auto border-b border-border/60 bg-panel/30 scrollbar-thin">
            {tabs.map((p) => {
              const file = allFiles.find((f) => f.path === p);
              if (!file) return null;
              const isActive = p === resolvedActive;
              return (
                <button
                  key={p}
                  onClick={() => setActivePath(p)}
                  className={cn(
                    "group flex h-full items-center gap-2 border-r border-border/60 px-3 text-xs",
                    isActive ? "bg-background text-foreground" : "text-muted-foreground hover:bg-elevated",
                  )}
                >
                  <File className={cn("h-3 w-3", langColor[file.language ?? ""] ?? "text-muted-foreground")} />
                  <span>{file.name}</span>
                  <X
                    className="h-3 w-3 opacity-0 transition-opacity hover:text-foreground group-hover:opacity-60"
                    onClick={(e) => closeTab(p, e)}
                  />
                </button>
              );
            })}
          </div>

          <ScrollArea className="flex-1">
            {active ? (
              <div className="flex min-h-full font-mono text-[13px] leading-6">
                <div className="select-none border-r border-border/60 px-3 py-4 text-right text-muted-foreground/50">
                  {Array.from({ length: lineCount }).map((_, i) => (
                    <div key={i}>{i + 1}</div>
                  ))}
                </div>
                <pre className="flex-1 overflow-x-auto px-4 py-4">
                  <code
                    className="text-foreground/90"
                    dangerouslySetInnerHTML={{ __html: highlight(active.content ?? "", active.language) }}
                  />
                </pre>
              </div>
            ) : (
              <div className="flex h-full items-center justify-center px-6 py-16 text-center">
                <div>
                  <div className="mx-auto h-10 w-10 rounded-lg border border-dashed border-border/80 bg-elevated/40" />
                  <p className="mt-3 text-sm font-medium text-foreground/80">Файл не выбран</p>
                  <p className="mt-1 text-xs text-muted-foreground">
                    Отправьте запрос — сгенерированные файлы появятся здесь.
                  </p>
                </div>
              </div>
            )}
          </ScrollArea>
        </div>
      </div>
    </div>
  );
}
