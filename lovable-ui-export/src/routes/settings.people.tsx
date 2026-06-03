import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { UserPlus, MoreHorizontal } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from "@/components/ui/select";
import { BackButton } from "@/components/features/BackButton";
import { toast } from "sonner";

export const Route = createFileRoute("/settings/people")({
  component: PeopleSettings,
  head: () => ({ meta: [{ title: "Участники — Исток" }] }),
});

const members = [
  { name: "Александр Иванов", email: "alex@istok.app", role: "Владелец", initials: "АИ" },
  { name: "Мария Петрова", email: "maria@istok.app", role: "Редактор", initials: "МП" },
  { name: "Иван Смирнов", email: "ivan@istok.app", role: "Наблюдатель", initials: "ИС" },
];

function InviteDialog({ open, onOpenChange }: { open: boolean; onOpenChange: (v: boolean) => void }) {
  const [emails, setEmails] = useState("");
  const [role, setRole] = useState("editor");
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Пригласить участников</DialogTitle>
          <DialogDescription>Введите email через запятую и выберите роль.</DialogDescription>
        </DialogHeader>
        <div className="space-y-3 py-2">
          <div>
            <Label htmlFor="emails">Email-адреса</Label>
            <Input
              id="emails"
              placeholder="name@company.com, other@company.com"
              value={emails}
              onChange={(e) => setEmails(e.target.value)}
              className="mt-1.5"
            />
          </div>
          <div>
            <Label>Роль</Label>
            <Select value={role} onValueChange={setRole}>
              <SelectTrigger className="mt-1.5"><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="admin">Администратор</SelectItem>
                <SelectItem value="editor">Редактор</SelectItem>
                <SelectItem value="viewer">Наблюдатель</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
        <DialogFooter>
          <Button variant="ghost" onClick={() => onOpenChange(false)}>Отмена</Button>
          <Button
            disabled={!emails.trim()}
            onClick={() => { toast.success("Приглашения отправлены"); setEmails(""); onOpenChange(false); }}
            className="bg-gradient-primary text-primary-foreground"
          >
            Отправить приглашения
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function PeopleSettings() {
  const [inviteOpen, setInviteOpen] = useState(false);
  return (
    <div className="mx-auto max-w-3xl px-4 py-8 sm:px-6">
      <div className="mb-6 flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold">Люди</h1>
          <p className="mt-1 text-sm text-muted-foreground">Управляйте участниками и их уровнем доступа.</p>
        </div>
        <Button onClick={() => setInviteOpen(true)} className="bg-gradient-primary text-primary-foreground">
          <UserPlus /> Пригласить участников
        </Button>
      </div>

      <div className="overflow-hidden rounded-xl border border-border/60 bg-card/40">
        {members.map((m, i) => (
          <div
            key={m.email}
            className={`flex items-center gap-3 p-4 ${i !== members.length - 1 ? "border-b border-border/60" : ""}`}
          >
            <div className="grid h-9 w-9 place-items-center rounded-full bg-gradient-primary text-xs font-semibold text-primary-foreground">
              {m.initials}
            </div>
            <div className="min-w-0 flex-1">
              <p className="truncate text-sm font-medium">{m.name}</p>
              <p className="truncate text-xs text-muted-foreground">{m.email}</p>
            </div>
            <span className="rounded-full border border-border/60 bg-elevated/60 px-2.5 py-0.5 text-xs text-muted-foreground">
              {m.role}
            </span>
            <Button variant="ghost" size="icon" className="h-8 w-8 text-muted-foreground">
              <MoreHorizontal className="h-4 w-4" />
            </Button>
          </div>
        ))}
      </div>
      <InviteDialog open={inviteOpen} onOpenChange={setInviteOpen} />
    </div>
  );
}

