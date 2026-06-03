import { useState } from "react";
import { Check, Copy, Gift, Link as LinkIcon, Sparkles, UserPlus, Zap } from "lucide-react";
import { toast } from "sonner";
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

const REFERRAL_LINK = "https://istok.app/?ref=alexandr-9f3a";

export function ReferralModal({
  open, onOpenChange,
}: { open: boolean; onOpenChange: (o: boolean) => void }) {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(REFERRAL_LINK);
      setCopied(true);
      toast.success("Реферальная ссылка скопирована");
      setTimeout(() => setCopied(false), 1800);
    } catch {
      toast.error("Не удалось скопировать ссылку");
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md overflow-hidden border-border/60 bg-card p-0">
        <div className="relative overflow-hidden">
          <div
            aria-hidden
            className="absolute inset-0"
            style={{
              background:
                "radial-gradient(120% 80% at 20% 0%, oklch(0.65 0.28 320 / 0.55), transparent 60%), radial-gradient(100% 80% at 90% 20%, oklch(0.7 0.2 250 / 0.45), transparent 60%), linear-gradient(180deg, oklch(0.25 0.05 290), oklch(0.15 0.03 290))",
            }}
          />
          <div className="relative flex h-40 items-center justify-center">
            <div className="relative">
              <div className="absolute inset-0 -z-10 rounded-full bg-fuchsia-500/40 blur-2xl" />
              <div className="grid h-16 w-16 place-items-center rounded-2xl bg-gradient-to-br from-fuchsia-400 to-violet-600 shadow-2xl ring-1 ring-white/30">
                <Gift className="h-7 w-7 text-white" />
              </div>
            </div>
            <Sparkles className="absolute left-10 top-6 h-4 w-4 text-white/70" />
            <Sparkles className="absolute right-12 bottom-6 h-3 w-3 text-white/60" />
          </div>
        </div>

        <div className="space-y-5 px-6 pb-6">
          <DialogHeader className="space-y-1.5">
            <DialogTitle className="text-xl">Поделитесь любовью</DialogTitle>
            <DialogDescription>
              Приглашайте друзей в Исток и получайте бесплатные кредиты, когда они присоединяются.
            </DialogDescription>
          </DialogHeader>

          <div className="rounded-lg border border-border/60 bg-elevated/40 p-3">
            <p className="mb-2 text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
              Как это работает
            </p>
            <ol className="space-y-2">
              <Step icon={LinkIcon} text="Поделитесь персональной ссылкой" />
              <Step icon={UserPlus} text="Они регистрируются в Истоке" />
              <Step icon={Zap} text="Вы получаете 100 кредитов — мгновенно" />
            </ol>
          </div>

          <div className="flex items-center gap-3 rounded-lg border border-border/60 bg-background/40 p-3">
            <QrMock />
            <div className="min-w-0 flex-1 space-y-2">
              <p className="text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
                Ваша реферальная ссылка
              </p>
              <div className="flex items-center gap-2">
                <Input
                  readOnly
                  disabled
                  value={REFERRAL_LINK}
                  className="h-8 flex-1 truncate bg-card/60 font-mono text-[11px]"
                />
                <Button size="sm" variant="secondary" className="h-8 shrink-0" onClick={handleCopy}>
                  {copied ? <Check className="h-3.5 w-3.5" /> : <Copy className="h-3.5 w-3.5" />}
                  Копировать
                </Button>
              </div>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}

function Step({ icon: Icon, text }: { icon: typeof LinkIcon; text: string }) {
  return (
    <li className="flex items-center gap-2.5 text-sm">
      <span className="grid h-6 w-6 place-items-center rounded-md bg-primary/15 text-primary">
        <Icon className="h-3 w-3" />
      </span>
      <span className="text-foreground/90">{text}</span>
    </li>
  );
}

function QrMock() {
  const cells = Array.from({ length: 49 }, (_, i) => {
    const x = i % 7, y = Math.floor(i / 7);
    const corner = (x < 2 && y < 2) || (x > 4 && y < 2) || (x < 2 && y > 4);
    return corner || (x * 13 + y * 7 + 3) % 3 === 0;
  });
  return (
    <div className="grid h-20 w-20 shrink-0 grid-cols-7 gap-[2px] rounded-md bg-white p-1.5">
      {cells.map((on, i) => (
        <div key={i} className={on ? "rounded-[1px] bg-black" : "bg-white"} />
      ))}
    </div>
  );
}
