import { useState, useCallback } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { X, Sparkles, MousePointer2 } from "lucide-react";
import type { SelectedElement } from "@/components/WorkspacePreview";

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — InspectorEditPanel
//  Плавающая модалка для Point-and-Click визуального
//  редактирования. Показывается при выборе элемента.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

interface InspectorEditPanelProps {
  selectedElement: SelectedElement | null;
  onClose: () => void;
  onApply: (instruction: string) => void;
  thinking?: boolean;
}

const InspectorEditPanel = ({
  selectedElement,
  onClose,
  onApply,
  thinking = false,
}: InspectorEditPanelProps) => {
  const [instruction, setInstruction] = useState("");

  const handleApply = useCallback(() => {
    if (!instruction.trim()) return;
    onApply(instruction.trim());
    setInstruction("");
  }, [instruction, onApply]);

  const label = selectedElement?.componentName
    ? selectedElement.componentName
    : selectedElement?.tag || "element";

  return (
    <AnimatePresence>
      {selectedElement && (
        <motion.div
          initial={{ opacity: 0, y: 20, scale: 0.95 }}
          animate={{ opacity: 1, y: 0, scale: 1 }}
          exit={{ opacity: 0, y: 20, scale: 0.95 }}
          transition={{ duration: 0.2, ease: [0.22, 1, 0.36, 1] }}
          className="absolute bottom-4 left-1/2 -translate-x-1/2 z-50 w-[420px] max-w-[calc(100vw-2rem)]"
        >
          <div className="rounded-xl border border-primary/20 bg-background/95 backdrop-blur-xl shadow-2xl shadow-primary/5 overflow-hidden">
            {/* Header */}
            <div className="flex items-center justify-between px-4 py-2.5 border-b border-border/20 bg-primary/5">
              <div className="flex items-center gap-2">
                <MousePointer2 size={13} className="text-primary" />
                <span className="text-xs font-medium text-primary">
                  Редактирование: &lt;{label}&gt;
                </span>
              </div>
              <button
                onClick={onClose}
                className="w-6 h-6 rounded-md flex items-center justify-center text-muted-foreground hover:text-foreground hover:bg-secondary/50 transition-colors"
              >
                <X size={12} />
              </button>
            </div>

            {/* Element info */}
            <div className="px-4 py-2 border-b border-border/10">
              <div className="flex flex-wrap gap-1.5">
                {selectedElement?.componentName && (
                  <span className="px-2 py-0.5 rounded-md bg-primary/10 text-primary text-[10px] font-medium">
                    {selectedElement.componentName}
                  </span>
                )}
                <span className="px-2 py-0.5 rounded-md bg-secondary/60 text-muted-foreground text-[10px]">
                  {selectedElement?.tag}
                </span>
                {selectedElement?.id && (
                  <span className="px-2 py-0.5 rounded-md bg-secondary/60 text-muted-foreground text-[10px]">
                    #{selectedElement.id}
                  </span>
                )}
                {selectedElement?.classes && (
                  <span className="px-2 py-0.5 rounded-md bg-secondary/60 text-muted-foreground text-[10px] max-w-[200px] truncate">
                    .{selectedElement.classes.split(" ")[0]}
                  </span>
                )}
              </div>
              {selectedElement?.text && (
                <p className="mt-1 text-[10px] text-muted-foreground/70 truncate italic">
                  "{selectedElement.text.slice(0, 60)}"
                </p>
              )}
            </div>

            {/* Input */}
            <div className="p-3">
              <textarea
                value={instruction}
                onChange={(e) => setInstruction(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === "Enter" && !e.shiftKey) {
                    e.preventDefault();
                    handleApply();
                  }
                }}
                placeholder="Что нужно изменить в этом элементе?"
                className="w-full h-16 px-3 py-2 rounded-lg bg-secondary/30 border border-border/20 text-sm text-foreground placeholder:text-muted-foreground/50 resize-none focus:outline-none focus:ring-1 focus:ring-primary/30 transition-shadow"
                autoFocus
                disabled={thinking}
              />
              <div className="flex items-center justify-between mt-2">
                <span className="text-[10px] text-muted-foreground/50">
                  Enter — отправить, Shift+Enter — перенос
                </span>
                <div className="flex items-center gap-2">
                  <button
                    onClick={onClose}
                    className="h-7 px-3 rounded-md text-[11px] text-muted-foreground hover:text-foreground hover:bg-secondary/50 transition-colors"
                  >
                    Отмена
                  </button>
                  <button
                    onClick={handleApply}
                    disabled={!instruction.trim() || thinking}
                    className="h-7 px-3 rounded-md text-[11px] font-medium bg-primary/15 text-primary hover:bg-primary/25 transition-colors disabled:opacity-40 disabled:pointer-events-none flex items-center gap-1.5"
                  >
                    <Sparkles size={11} />
                    Применить
                  </button>
                </div>
              </div>
            </div>
          </div>
        </motion.div>
      )}
    </AnimatePresence>
  );
};

export default InspectorEditPanel;
