import { motion } from "framer-motion";
import { ShieldCheck, X } from "lucide-react";

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — SecurityAuditOverlay
//  Всплывающая панель с детальным разбором VerificationGate.
//  Открывается по клику [Security Audit] в PreviewPanel.
//  Извлечена из Workspace.tsx — единая ответственность.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

export interface SecurityAuditOverlayProps {
  securityApproved: boolean;
  testerApproved: boolean;
  uiReviewerApproved: boolean;
  onClose: () => void;
}

interface AuditCheck {
  id: "security" | "tester" | "ui_reviewer";
  label: string;
  passed: boolean;
  desc: string;
}

const SecurityAuditOverlay = ({
  securityApproved,
  testerApproved,
  uiReviewerApproved,
  onClose,
}: SecurityAuditOverlayProps) => {
  const checks: AuditCheck[] = [
    { id: "security", label: "Security Scan", passed: securityApproved, desc: "XSS, injection, secrets" },
    { id: "tester", label: "Tests", passed: testerApproved, desc: "Runtime smoke tests" },
    { id: "ui_reviewer", label: "UI Review", passed: uiReviewerApproved, desc: "a11y, contrast, UX" },
  ];
  const passedCount = checks.filter((c) => c.passed).length;

  return (
    <motion.div
      initial={{ opacity: 0, y: 8, scale: 0.98 }}
      animate={{ opacity: 1, y: 0, scale: 1 }}
      exit={{ opacity: 0, y: 4, scale: 0.99 }}
      transition={{ duration: 0.25, ease: [0.22, 1, 0.36, 1] }}
      className="absolute top-14 right-4 z-40 w-[320px] glass-panel rounded-xl shadow-xl border border-glass-border/40 overflow-hidden"
    >
      <header className="flex items-center justify-between px-3 py-2 border-b border-glass-border/30">
        <div className="flex items-center gap-2">
          <ShieldCheck size={14} className="text-emerald-400" />
          <h3 className="text-[12px] font-semibold">Security Audit</h3>
          <span className="text-[10px] text-muted-foreground/60">
            {passedCount}/{checks.length} passed
          </span>
        </div>
        <button
          onClick={onClose}
          className="w-6 h-6 rounded-md flex items-center justify-center text-muted-foreground hover:text-foreground hover:bg-secondary/40 transition-colors"
          aria-label="Close audit overlay"
        >
          <X size={12} />
        </button>
      </header>
      <ul className="p-2 space-y-1">
        {checks.map((c) => (
          <li
            key={c.id}
            className="flex items-center gap-2 px-2 py-1.5 rounded-lg glass-subtle"
          >
            <div
              className={`w-6 h-6 rounded-full flex items-center justify-center ${
                c.passed ? "bg-emerald-500/20" : "bg-secondary/40"
              }`}
            >
              <ShieldCheck
                size={11}
                className={c.passed ? "text-emerald-400" : "text-muted-foreground/40"}
              />
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-[11px] font-medium text-foreground">{c.label}</p>
              <p className="text-[9.5px] text-muted-foreground/55 truncate">{c.desc}</p>
            </div>
            <span
              className={`text-[9px] uppercase tracking-wider font-semibold ${
                c.passed ? "text-emerald-400" : "text-muted-foreground/40"
              }`}
            >
              {c.passed ? "Pass" : "Pending"}
            </span>
          </li>
        ))}
      </ul>
    </motion.div>
  );
};

export default SecurityAuditOverlay;
