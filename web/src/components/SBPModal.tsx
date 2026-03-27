import { motion, AnimatePresence } from "framer-motion";
import { X, QrCode, Loader2 } from "lucide-react";
import { useLanguage } from "@/hooks/useLanguage";

interface SBPModalProps {
  open: boolean;
  onClose: () => void;
  packageInfo?: { name: string; amount: number; credits: number } | null;
}

const SBPModal = ({ open, onClose, packageInfo }: SBPModalProps) => {
  const { t } = useLanguage();

  if (!open) return null;

  const displayAmount = packageInfo ? `${packageInfo.amount.toLocaleString("ru-RU")} ₽` : "2 990 ₽/мес";
  const displayTitle = packageInfo ? packageInfo.name : t("sbpTitle");

  return (
    <AnimatePresence>
      {open && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          className="fixed inset-0 z-50 flex items-center justify-center p-4"
          onClick={onClose}
        >
          <div className="absolute inset-0 bg-background/80 backdrop-blur-sm" />

          <motion.div
            initial={{ opacity: 0, scale: 0.9, y: 20 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.9, y: 20 }}
            transition={{ type: "spring", damping: 25, stiffness: 300 }}
            onClick={(e) => e.stopPropagation()}
            className="relative glass-subtle rounded-2xl p-8 max-w-sm w-full border border-primary/20 shadow-[0_0_60px_hsla(243,76%,58%,0.1)]"
          >
            <button
              onClick={onClose}
              className="absolute top-4 right-4 w-8 h-8 rounded-lg flex items-center justify-center text-muted-foreground hover:text-foreground hover:bg-secondary/50 transition-colors"
            >
              <X size={16} />
            </button>

            <div className="flex items-center justify-center mb-6">
              <div className="w-12 h-12 rounded-xl bg-primary/10 flex items-center justify-center">
                <QrCode size={28} className="text-primary" />
              </div>
            </div>

            <h3 className="text-lg font-bold text-foreground text-center mb-1">
              {displayTitle}
            </h3>
            <p className="text-sm text-muted-foreground text-center mb-6">
              {t("sbpSubtitle")}
            </p>

            {/* QR Code placeholder */}
            <div className="mx-auto w-48 h-48 rounded-xl bg-foreground/5 border-2 border-dashed border-border/50 flex flex-col items-center justify-center mb-4">
              <div className="w-36 h-36 bg-foreground/[0.03] rounded-lg flex items-center justify-center relative">
                <div className="grid grid-cols-8 gap-[2px] w-28 h-28 opacity-20">
                  {Array.from({ length: 64 }).map((_, i) => (
                    <div
                      key={i}
                      className={`rounded-[1px] ${Math.random() > 0.4 ? "bg-foreground" : "bg-transparent"}`}
                    />
                  ))}
                </div>
                <div className="absolute inset-0 flex items-center justify-center">
                  <div className="w-8 h-8 rounded-md bg-primary/20 flex items-center justify-center">
                    <span className="text-[8px] font-bold text-primary">СБП</span>
                  </div>
                </div>
              </div>
            </div>

            <p className="text-center text-sm font-semibold text-foreground mb-1">
              {t("sbpAmountLabel")}: {displayAmount}
            </p>
            {packageInfo && (
              <p className="text-center text-xs text-muted-foreground mb-3">
                {t("creditTokens", packageInfo.credits)}
              </p>
            )}

            <div className="flex items-center justify-center gap-2 mb-4">
              <Loader2 size={14} className="text-primary animate-spin" />
              <span className="text-xs text-muted-foreground">{t("sbpWaiting")}</span>
            </div>

            <p className="text-[10px] text-muted-foreground/50 text-center mb-6">
              {t("sbpNote")}
            </p>

            <button
              onClick={onClose}
              className="w-full py-2.5 rounded-xl text-sm font-medium bg-secondary text-foreground hover:bg-secondary/80 transition-colors"
            >
              {t("sbpClose")}
            </button>
          </motion.div>
        </motion.div>
      )}
    </AnimatePresence>
  );
};

export default SBPModal;
