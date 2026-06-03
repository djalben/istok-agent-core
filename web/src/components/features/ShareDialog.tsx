import { useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { Copy, Check, Link2, Mail, Trash2, Globe, Lock } from "lucide-react";
import { toast } from "sonner";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

interface Member {
  id: string;
  email: string;
  role: "owner" | "editor" | "viewer";
  initials: string;
}

interface ShareDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

const initialMembers: Member[] = [
  { id: "1", email: "you@istok.app", role: "owner", initials: "YO" },
  { id: "2", email: "alex@istok.app", role: "editor", initials: "AL" },
  { id: "3", email: "mira@design.co", role: "viewer", initials: "MI" },
];

const roleLabel = (r: Member["role"]) =>
  r === "owner" ? "Владелец" : r === "editor" ? "Редактор" : "Наблюдатель";

export function ShareDialog({ open, onOpenChange }: ShareDialogProps) {
  const [members, setMembers] = useState<Member[]>(initialMembers);
  const [invite, setInvite] = useState("");
  const [role, setRole] = useState<"editor" | "viewer">("editor");
  const [linkAccess, setLinkAccess] = useState<"restricted" | "anyone">("restricted");
  const [copied, setCopied] = useState(false);

  const inviteUrl = "https://istok.app/i/n3xt-vibe-9821";

  const addMember = () => {
    if (!invite.includes("@")) {
      toast.error("Введите корректный email");
      return;
    }
    setMembers((m) => [
      ...m,
      {
        id: crypto.randomUUID(),
        email: invite,
        role,
        initials: invite.slice(0, 2).toUpperCase(),
      },
    ]);
    toast.success(`Приглашён ${invite} как ${roleLabel(role)}`);
    setInvite("");
  };

  const copyLink = async () => {
    await navigator.clipboard.writeText(inviteUrl);
    setCopied(true);
    toast.success("Ссылка-приглашение скопирована");
    setTimeout(() => setCopied(false), 1800);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg border-border/80 bg-card p-0">
        <DialogHeader className="border-b border-border/60 p-6 pb-4">
          <DialogTitle className="text-base font-semibold">Поделиться проектом</DialogTitle>
          <DialogDescription className="text-xs text-muted-foreground">
            Пригласите коллег к совместной работе в рабочем пространстве Истока.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-5 p-6">
          <div className="flex gap-2">
            <div className="relative flex-1">
              <Mail className="pointer-events-none absolute left-3 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-muted-foreground" />
              <Input
                placeholder="name@company.com"
                value={invite}
                onChange={(e) => setInvite(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && addMember()}
                className="h-9 pl-9 text-sm"
              />
            </div>
            <Select value={role} onValueChange={(v: "editor" | "viewer") => setRole(v)}>
              <SelectTrigger className="h-9 w-32 text-xs">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="editor">Редактор</SelectItem>
                <SelectItem value="viewer">Наблюдатель</SelectItem>
              </SelectContent>
            </Select>
            <Button onClick={addMember} size="sm" className="h-9 bg-gradient-primary text-primary-foreground">
              Пригласить
            </Button>
          </div>

          <div className="space-y-1.5">
            <p className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
              Участники · {members.length}
            </p>
            <div className="max-h-56 space-y-1 overflow-y-auto rounded-lg border border-border/60 bg-elevated/40 p-1.5">
              <AnimatePresence initial={false}>
                {members.map((m) => (
                  <motion.div
                    key={m.id}
                    initial={{ opacity: 0, y: -6 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0, x: 12 }}
                    className="flex items-center gap-3 rounded-md px-2 py-2 hover:bg-muted/40"
                  >
                    <div className="grid h-7 w-7 place-items-center rounded-full bg-gradient-primary text-[10px] font-semibold text-primary-foreground">
                      {m.initials}
                    </div>
                    <span className="flex-1 truncate text-sm">{m.email}</span>
                    {m.role === "owner" ? (
                      <span className="font-mono text-[10px] uppercase text-muted-foreground">Владелец</span>
                    ) : (
                      <>
                        <Select
                          value={m.role}
                          onValueChange={(v: "editor" | "viewer") =>
                            setMembers((arr) => arr.map((x) => (x.id === m.id ? { ...x, role: v } : x)))
                          }
                        >
                          <SelectTrigger className="h-7 w-28 text-xs">
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="editor">Редактор</SelectItem>
                            <SelectItem value="viewer">Наблюдатель</SelectItem>
                          </SelectContent>
                        </Select>
                        <Button
                          variant="ghost"
                          size="icon"
                          className="h-7 w-7 text-muted-foreground hover:text-destructive"
                          onClick={() => setMembers((arr) => arr.filter((x) => x.id !== m.id))}
                          aria-label="Удалить участника"
                        >
                          <Trash2 className="h-3.5 w-3.5" />
                        </Button>
                      </>
                    )}
                  </motion.div>
                ))}
              </AnimatePresence>
            </div>
          </div>

          <div className="rounded-lg border border-border/60 bg-elevated/40 p-3">
            <div className="mb-2 flex items-center gap-2">
              {linkAccess === "anyone" ? (
                <Globe className="h-3.5 w-3.5 text-primary" />
              ) : (
                <Lock className="h-3.5 w-3.5 text-muted-foreground" />
              )}
              <span className="text-xs font-medium">Ссылка-приглашение</span>
              <Select value={linkAccess} onValueChange={(v: "restricted" | "anyone") => setLinkAccess(v)}>
                <SelectTrigger className="ml-auto h-7 w-44 text-xs">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="restricted">Ограниченный доступ</SelectItem>
                  <SelectItem value="anyone">Любой по ссылке</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="flex items-center gap-2">
              <div className="flex flex-1 items-center gap-2 truncate rounded-md border border-border/60 bg-background px-3 py-1.5 font-mono text-xs text-muted-foreground">
                <Link2 className="h-3 w-3 shrink-0" />
                <span className="truncate">{inviteUrl}</span>
              </div>
              <Button variant="outline" size="sm" className="h-8 gap-1.5" onClick={copyLink}>
                {copied ? <Check className="h-3.5 w-3.5 text-primary" /> : <Copy className="h-3.5 w-3.5" />}
                {copied ? "Скопировано" : "Копировать"}
              </Button>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
