import { FileText, FileCode, Palette, ChevronRight, FolderOpen } from "lucide-react";
import { useLanguage } from "@/hooks/useLanguage";

interface FileExplorerProps {
  files: Record<string, string>;
  activeFile: string;
  onSelectFile: (filename: string) => void;
}

const getFileIcon = (filename: string) => {
  if (filename.endsWith(".css")) return <Palette size={13} className="text-blue-400 shrink-0" />;
  if (filename.endsWith(".js") || filename.endsWith(".ts")) return <FileCode size={13} className="text-yellow-400 shrink-0" />;
  return <FileText size={13} className="text-orange-400 shrink-0" />;
};

const FileExplorer = ({ files, activeFile, onSelectFile }: FileExplorerProps) => {
  const { t } = useLanguage();
  const filenames = Object.keys(files);

  return (
    <div className="w-[200px] min-w-[200px] border-r border-[hsl(var(--border))]/10 bg-background/50 flex flex-col">
      <div className="h-8 flex items-center gap-1.5 px-3 border-b border-[hsl(var(--border))]/10 shrink-0">
        <FolderOpen size={12} className="text-muted-foreground" />
        <span className="text-[10px] uppercase tracking-widest text-muted-foreground font-medium">{t("wsFiles")}</span>
      </div>
      <div className="flex-1 overflow-y-auto py-1">
        {filenames.map((name) => (
          <button
            key={name}
            onClick={() => onSelectFile(name)}
            className={`w-full flex items-center gap-2 px-3 py-1.5 text-xs transition-colors ${
              activeFile === name
                ? "bg-primary/10 text-primary"
                : "text-muted-foreground hover:text-foreground hover:bg-secondary/30"
            }`}
          >
            {activeFile === name && <ChevronRight size={10} className="shrink-0" />}
            {getFileIcon(name)}
            <span className="truncate">{name}</span>
          </button>
        ))}
      </div>
    </div>
  );
};

export default FileExplorer;
