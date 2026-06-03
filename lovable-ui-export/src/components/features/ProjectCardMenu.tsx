import { useState } from "react";
import { useNavigate } from "@tanstack/react-router";
import {
  Link2,
  MoreHorizontal,
  ExternalLink,
  Globe,
  BarChart3,
  FolderInput,
  GitFork,
  Pencil,
  ArrowRightLeft,
  Settings,
  UploadCloud,
  Trash2,
  Search,
  FolderPlus,
  AlertTriangle,
} from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { cn } from "@/lib/utils";
import type { Project } from "@/lib/mockData";

type ActiveModal = "move" | "remix" | "rename" | "transfer" | null;

interface ProjectCardMenuProps {
  project: Project;
}

const stop = (e: React.SyntheticEvent) => {
  e.preventDefault();
  e.stopPropagation();
};

export function ProjectCardMenu({ project }: ProjectCardMenuProps) {
  const navigate = useNavigate();
  const [modal, setModal] = useState<ActiveModal>(null);

  const copy = (text: string, msg: string) => {
    navigator.clipboard?.writeText(text).catch(() => {});
    toast.success(msg);
  };

  return (
    <>
      <div
        className="absolute right-2 top-2 z-10 flex items-center gap-1 opacity-0 transition-opacity duration-150 group-hover:opacity-100 focus-within:opacity-100"
        onClick={stop}
      >
        {/* Link icon dropdown */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild onClick={stop}>
            <button
              type="button"
              aria-label="Скопировать ссылку"
              className="grid h-7 w-7 place-items-center rounded-md border border-border/60 bg-background/80 text-muted-foreground backdrop-blur-md transition-colors hover:bg-elevated hover:text-foreground"
            >
              <Link2 className="h-3.5 w-3.5" />
            </button>
          </DropdownMenuTrigger>
          <DropdownMenuContent
            align="end"
            sideOffset={6}
            className="w-64 border-border/70 bg-popover/95 p-1 backdrop-blur-xl"
            onClick={stop}
          >
            <DropdownMenuItem
              onSelect={(e) => {
                e.preventDefault();
                copy(
                  `${window.location.origin}/builder/${project.id}`,
                  "Ссылка на проект скопирована",
                );
              }}
              className="cursor-pointer rounded-md px-2 py-1.5 text-xs"
            >
              Скопировать ссылку на проект
            </DropdownMenuItem>
            <DropdownMenuItem
              onSelect={(e) => {
                e.preventDefault();
                copy(
                  `https://${project.id}.istok.app`,
                  "Ссылка на опубликованное приложение скопирована",
                );
              }}
              className="cursor-pointer rounded-md px-2 py-1.5 text-xs"
            >
              Скопировать ссылку на опубликованное приложение
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>

        {/* 3-dots dropdown */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild onClick={stop}>
            <button
              type="button"
              aria-label="Действия с проектом"
              className="grid h-7 w-7 place-items-center rounded-md border border-border/60 bg-background/80 text-muted-foreground backdrop-blur-md transition-colors hover:bg-elevated hover:text-foreground"
            >
              <MoreHorizontal className="h-3.5 w-3.5" />
            </button>
          </DropdownMenuTrigger>
          <DropdownMenuContent
            align="end"
            sideOffset={6}
            className="w-60 border-border/70 bg-popover/95 p-1 backdrop-blur-xl"
            onClick={stop}
          >
            <MenuRow
              icon={<ExternalLink className="h-3.5 w-3.5" />}
              label="Открыть в новой вкладке"
              onSelect={() => window.open(`/builder/${project.id}`, "_blank")}
            />
            <MenuRow
              icon={<Globe className="h-3.5 w-3.5" />}
              label="Посмотреть опубликованный сайт"
              onSelect={() => toast(`Открываю ${project.id}.istok.app`)}
            />
            <MenuRow
              icon={<BarChart3 className="h-3.5 w-3.5" />}
              label="Аналитика"
              onSelect={() => toast("Аналитика проекта")}
            />
            <MenuRow
              icon={<FolderInput className="h-3.5 w-3.5" />}
              label="Переместить в папку"
              onSelect={() => setModal("move")}
            />
            <MenuRow
              icon={<GitFork className="h-3.5 w-3.5" />}
              label="Ремикс"
              onSelect={() => setModal("remix")}
            />
            <MenuRow
              icon={<Pencil className="h-3.5 w-3.5" />}
              label="Переименовать"
              onSelect={() => setModal("rename")}
            />
            <MenuRow
              icon={<ArrowRightLeft className="h-3.5 w-3.5" />}
              label="Перемещение на рабочее место"
              onSelect={() => setModal("transfer")}
            />
            <MenuRow
              icon={<Settings className="h-3.5 w-3.5" />}
              label="Настройки"
              onSelect={() => navigate({ to: "/settings/project" })}
            />
            <MenuRow
              icon={<UploadCloud className="h-3.5 w-3.5" />}
              label="Опубликовать в профиле"
              onSelect={() => toast.success("Опубликовано в профиле")}
            />
            <DropdownMenuSeparator className="my-1" />
            <DropdownMenuItem
              onSelect={(e) => {
                e.preventDefault();
                toast.error(`«${project.name}» удалён`);
              }}
              className="flex cursor-pointer items-center gap-2 rounded-md px-2 py-1.5 text-xs text-destructive focus:bg-destructive/10 focus:text-destructive"
            >
              <Trash2 className="h-3.5 w-3.5" /> Удалить
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>

      <MoveDialog
        open={modal === "move"}
        onOpenChange={(o) => !o && setModal(null)}
        project={project}
      />
      <RemixDialog
        open={modal === "remix"}
        onOpenChange={(o) => !o && setModal(null)}
        project={project}
      />
      <RenameDialog
        open={modal === "rename"}
        onOpenChange={(o) => !o && setModal(null)}
        project={project}
      />
      <TransferDialog
        open={modal === "transfer"}
        onOpenChange={(o) => !o && setModal(null)}
        project={project}
      />
    </>
  );
}

function MenuRow({
  icon,
  label,
  onSelect,
}: {
  icon: React.ReactNode;
  label: string;
  onSelect: () => void;
}) {
  return (
    <DropdownMenuItem
      onSelect={(e) => {
        e.preventDefault();
        onSelect();
      }}
      className="flex cursor-pointer items-center gap-2 rounded-md px-2 py-1.5 text-xs focus:bg-elevated"
    >
      <span className="text-muted-foreground">{icon}</span>
      <span className="font-medium">{label}</span>
    </DropdownMenuItem>
  );
}

/* ------------------------------- Modals ------------------------------- */

const mockFolders = [
  { id: "none", name: "Нет папки", current: true },
  { id: "personal", name: "Личные эксперименты" },
  { id: "clients", name: "Клиентские проекты" },
  { id: "archive", name: "Архив 2025" },
];

function MoveDialog({
  open,
  onOpenChange,
  project,
}: {
  open: boolean;
  onOpenChange: (o: boolean) => void;
  project: Project;
}) {
  const [query, setQuery] = useState("");
  const [value, setValue] = useState("none");
  const filtered = mockFolders.filter((f) =>
    f.name.toLowerCase().includes(query.toLowerCase()),
  );

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Переместить в папку</DialogTitle>
          <DialogDescription>
            Выберите папку для проекта «{project.name}».
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-3">
          <div className="relative">
            <Search className="pointer-events-none absolute left-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-muted-foreground" />
            <Input
              placeholder="Поиск по папкам..."
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              className="h-9 pl-8 text-sm"
            />
          </div>

          <Button
            variant="outline"
            size="sm"
            className="w-full justify-start gap-2 text-xs"
            onClick={() => toast("Создание новой папки")}
          >
            <FolderPlus className="h-3.5 w-3.5" />
            Создать новую папку
          </Button>

          <RadioGroup value={value} onValueChange={setValue} className="space-y-1">
            {filtered.map((f) => (
              <label
                key={f.id}
                htmlFor={`folder-${f.id}`}
                className={cn(
                  "flex cursor-pointer items-center gap-3 rounded-md border border-border/60 px-3 py-2 text-sm transition-colors hover:bg-elevated",
                  value === f.id && "border-primary/60 bg-primary/5",
                )}
              >
                <RadioGroupItem id={`folder-${f.id}`} value={f.id} />
                <span className="font-medium">{f.name}</span>
                {f.current && (
                  <span className="ml-auto rounded-md bg-muted px-1.5 py-0.5 text-[10px] uppercase tracking-wider text-muted-foreground">
                    Текущий
                  </span>
                )}
              </label>
            ))}
            {filtered.length === 0 && (
              <p className="px-1 py-4 text-center text-xs text-muted-foreground">
                Папки не найдены
              </p>
            )}
          </RadioGroup>
        </div>

        <DialogFooter>
          <Button variant="ghost" onClick={() => onOpenChange(false)}>
            Отмена
          </Button>
          <Button
            onClick={() => {
              toast.success("Изменения сохранены");
              onOpenChange(false);
            }}
          >
            Сохранить изменения
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function RemixDialog({
  open,
  onOpenChange,
  project,
}: {
  open: boolean;
  onOpenChange: (o: boolean) => void;
  project: Project;
}) {
  const [name, setName] = useState(`${project.name} (ремикс)`);
  const [history, setHistory] = useState(true);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Проект ремикса</DialogTitle>
          <DialogDescription>
            Создайте копию проекта с собственной историей правок.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <div className="space-y-1.5">
            <Label htmlFor="remix-name">Название проекта</Label>
            <Input
              id="remix-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="h-9 text-sm"
            />
          </div>

          <div className="flex items-start justify-between gap-4 rounded-md border border-border/60 p-3">
            <div className="space-y-0.5">
              <p className="text-sm font-medium">Включите историю проекта</p>
              <p className="text-xs text-muted-foreground">
                Скопировать все сообщения чата и шаги агентов.
              </p>
            </div>
            <Switch checked={history} onCheckedChange={setHistory} />
          </div>
        </div>

        <DialogFooter>
          <Button variant="ghost" onClick={() => onOpenChange(false)}>
            Отмена
          </Button>
          <Button
            onClick={() => {
              toast.success(`«${name}» создан`);
              onOpenChange(false);
            }}
          >
            Ремикс
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function RenameDialog({
  open,
  onOpenChange,
  project,
}: {
  open: boolean;
  onOpenChange: (o: boolean) => void;
  project: Project;
}) {
  const [name, setName] = useState(project.name);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Переименовать проект</DialogTitle>
          <DialogDescription>
            Новое название будет видно вам и соавторам.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-1.5">
          <Label htmlFor="rename-name">Название</Label>
          <Input
            id="rename-name"
            value={name}
            maxLength={100}
            onChange={(e) => setName(e.target.value)}
            className="h-9 text-sm"
          />
          <p className="text-xs text-muted-foreground">
            Используйте до 100 символов. Сейчас: {name.length}/100.
          </p>
        </div>

        <DialogFooter>
          <Button variant="ghost" onClick={() => onOpenChange(false)}>
            Отмена
          </Button>
          <Button
            onClick={() => {
              toast.success("Название обновлено");
              onOpenChange(false);
            }}
          >
            Сохранять
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

const mockWorkspaces = [
  { id: "personal", name: "Личное пространство" },
  { id: "acme", name: "Acme Studio" },
  { id: "nebula", name: "Nebula Labs" },
];

function TransferDialog({
  open,
  onOpenChange,
  project,
}: {
  open: boolean;
  onOpenChange: (o: boolean) => void;
  project: Project;
}) {
  const [target, setTarget] = useState<string | undefined>();

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Перемещение рабочего пространства</DialogTitle>
          <DialogDescription>
            Перенесите «{project.name}» в другое пространство.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <div className="flex items-start gap-2 rounded-md border border-destructive/40 bg-destructive/10 p-3 text-xs text-destructive">
            <AlertTriangle className="mt-0.5 h-4 w-4 shrink-0" />
            <p>
              Все сообщники будут удалены из проекта. После переноса доступ
              получат только участники нового рабочего пространства.
            </p>
          </div>

          <div className="space-y-1.5">
            <Label>От</Label>
            <Select value="personal" disabled>
              <SelectTrigger className="h-9 text-sm">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="personal">Личное пространство</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-1.5">
            <Label>К</Label>
            <Select value={target} onValueChange={setTarget}>
              <SelectTrigger className="h-9 text-sm">
                <SelectValue placeholder="Выберите пространство" />
              </SelectTrigger>
              <SelectContent>
                {mockWorkspaces
                  .filter((w) => w.id !== "personal")
                  .map((w) => (
                    <SelectItem key={w.id} value={w.id}>
                      {w.name}
                    </SelectItem>
                  ))}
              </SelectContent>
            </Select>
          </div>
        </div>

        <DialogFooter>
          <Button variant="ghost" onClick={() => onOpenChange(false)}>
            Отмена
          </Button>
          <Button
            variant="destructive"
            disabled={!target}
            onClick={() => {
              toast.success("Перевод подтверждён");
              onOpenChange(false);
            }}
          >
            Подтвердить перевод
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
