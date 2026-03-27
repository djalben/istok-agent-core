import { useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { X, Copy, Check, ExternalLink, Globe } from "lucide-react";

interface PublishModalProps {
  open: boolean;
  onClose: () => void;
  projectUrl: string;
}

const PublishModal = ({ open, onClose, projectUrl }: PublishModalProps) => {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    await navigator.clipboard.writeText(projectUrl);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <AnimatePresence>
      {open && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm"
          onClick={onClose}
        >
          <motion.div
            initial={{ opacity: 0, scale: 0.95, y: 10 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.95, y: 10 }}
            transition={{ duration: 0.2 }}
            onClick={(e) => e.stopPropagation()}
            className="w-full max-w-md glass rounded-2xl border border-border/30 p-6"
          >
            <div className="flex items-center justify-between mb-6">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-xl bg-primary/20 flex items-center justify-center">
                  <Globe size={20} className="text-primary" />
                </div>
                <div>
                  <h2 className="text-lg font-semibold text-foreground">Проект опубликован!</h2>
                  <p className="text-xs text-muted-foreground">Доступен всем по ссылке</p>
                </div>
              </div>
              <button
                onClick={onClose}
                className="w-8 h-8 rounded-lg flex items-center justify-center text-muted-foreground hover:text-foreground hover:bg-secondary/50 transition-colors"
              >
                <X size={16} />
              </button>
            </div>

            <div className="flex items-center gap-2 bg-secondary/50 rounded-xl p-3 border border-border/20">
              <span className="flex-1 text-sm text-foreground truncate font-mono">
                {projectUrl}
              </span>
              <button
                onClick={handleCopy}
                className="shrink-0 flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-primary/15 text-primary hover:bg-primary/25 transition-colors text-xs font-medium"
              >
                {copied ? <Check size={13} /> : <Copy size={13} />}
                {copied ? "Скопировано" : "Копировать"}
              </button>
            </div>

            <a
              href={projectUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="mt-4 w-full flex items-center justify-center gap-2 h-10 rounded-xl btn-gradient text-primary-foreground text-sm font-medium"
            >
              <ExternalLink size={14} />
              Открыть проект
            </a>
          </motion.div>
        </motion.div>
      )}
    </AnimatePresence>
  );
};

export default PublishModal;
