import { useRef, useState, type ChangeEvent, type DragEvent } from "react";
import { motion, AnimatePresence } from "framer-motion";
import {
  Plus,
  Settings as SettingsIcon,
  BookOpen,
  Github,
  Plug,
  FilePlus2,
  Sparkles,
  Upload,
  Check,
  X,
  Key,
  Globe,
  Cpu,
  FileText,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
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
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { toast } from "sonner";

type ModalKey = "settings" | "knowledge" | "github" | "connectors" | "skill" | null;

const menuMotion = {
  initial: { opacity: 0, y: 6, scale: 0.97 },
  animate: { opacity: 1, y: 0, scale: 1 },
  exit: { opacity: 0, y: 6, scale: 0.97 },
  transition: { duration: 0.14, ease: "easeOut" as const },
} as const;

export function AttachmentMenu() {
  const [open, setOpen] = useState(false);
  const [modal, setModal] = useState<ModalKey>(null);
  const [uploadOpen, setUploadOpen] = useState(false);
  const [files, setFiles] = useState<{ name: string; size: number }[]>([]);
  const [dragging, setDragging] = useState(false);
  const fileRef = useRef<HTMLInputElement>(null);
  const silentFileRef = useRef<HTMLInputElement>(null);

  const openModal = (key: ModalKey) => {
    setOpen(false);
    setTimeout(() => setModal(key), 80);
  };

  const handleFiles = (list: FileList | null) => {
    if (!list) return;
    const next = Array.from(list).map((f) => ({ name: f.name, size: f.size }));
    setFiles((prev) => [...prev, ...next]);
  };

  const onDrop = (e: DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    setDragging(false);
    handleFiles(e.dataTransfer.files);
  };

  return (
    <>
      <DropdownMenu open={open} onOpenChange={setOpen}>
        <DropdownMenuTrigger asChild>
          <Button
            variant="ghost"
            size="icon"
            className="h-7 w-7 text-muted-foreground hover:text-primary"
          >
            <Plus className="h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>
        <AnimatePresence>
          {open && (
            <DropdownMenuContent
              asChild
              align="start"
              side="top"
              sideOffset={8}
              className="w-64 border-border/70 bg-popover/95 p-1.5 backdrop-blur-xl"
            >
              <motion.div {...menuMotion}>
                <DropdownMenuLabel className="px-2 py-1.5 text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
                  Рабочее пространство
                </DropdownMenuLabel>
                <MenuRow
                  icon={<SettingsIcon className="h-3.5 w-3.5" />}
                  label="Настройки"
                  hint="Модель · ключи"
                  onClick={() => openModal("settings")}
                />
                <MenuRow
                  icon={<BookOpen className="h-3.5 w-3.5" />}
                  label="База знаний"
                  hint="Документы · контекст"
                  onClick={() => openModal("knowledge")}
                />
                <MenuRow
                  icon={<Github className="h-3.5 w-3.5" />}
                  label="GitHub"
                  hint="Синхронизация"
                  onClick={() => openModal("github")}
                />
                <MenuRow
                  icon={<Plug className="h-3.5 w-3.5" />}
                  label="Коннекторы"
                  hint="Интеграции"
                  onClick={() => openModal("connectors")}
                />

                <DropdownMenuSeparator className="my-1" />
                <DropdownMenuLabel className="px-2 py-1.5 text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
                  Прикрепить к запросу
                </DropdownMenuLabel>
                <MenuRow
                  icon={<FilePlus2 className="h-3.5 w-3.5" />}
                  label="Добавить ссылку"
                  hint="Изображение · URL · файл"
                  onClick={() => {
                    setOpen(false);
                    setTimeout(() => setUploadOpen(true), 80);
                  }}
                />
                <MenuRow
                  icon={<Sparkles className="h-3.5 w-3.5" />}
                  label="Добавить навык"
                  hint="Повторно используемый промт"
                  onClick={() => openModal("skill")}
                />
                <DropdownMenuSeparator className="my-1" />
                <DropdownMenuItem
                  onSelect={(e) => {
                    e.preventDefault();
                    setOpen(false);
                    silentFileRef.current?.click();
                  }}
                  className="gap-2 rounded-md px-2 py-1.5 text-xs"
                >
                  <Upload className="h-3.5 w-3.5 text-muted-foreground" />
                  Быстрое вложение…
                </DropdownMenuItem>
              </motion.div>
            </DropdownMenuContent>
          )}
        </AnimatePresence>
      </DropdownMenu>

      <input
        ref={silentFileRef}
        type="file"
        multiple
        className="hidden"
        onChange={(e: ChangeEvent<HTMLInputElement>) => {
          handleFiles(e.target.files);
          if (e.target.files?.length) {
            toast.success(`Прикреплено файлов: ${e.target.files.length}`);
          }
          e.target.value = "";
        }}
      />

      <SettingsModal open={modal === "settings"} onClose={() => setModal(null)} />
      <KnowledgeModal open={modal === "knowledge"} onClose={() => setModal(null)} />
      <GithubModal open={modal === "github"} onClose={() => setModal(null)} />
      <ConnectorsModal open={modal === "connectors"} onClose={() => setModal(null)} />
      <SkillModal open={modal === "skill"} onClose={() => setModal(null)} />

      <Dialog open={uploadOpen} onOpenChange={setUploadOpen}>
        <DialogContent className="border-border/70 bg-popover/95 backdrop-blur-xl sm:max-w-lg">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <FilePlus2 className="h-4 w-4 text-primary" /> Добавить ссылку
            </DialogTitle>
            <DialogDescription>
              Перетащите файлы, изображения или вставьте URL, чтобы расширить контекст следующей генерации.
            </DialogDescription>
          </DialogHeader>

          <div
            onDragOver={(e) => {
              e.preventDefault();
              setDragging(true);
            }}
            onDragLeave={() => setDragging(false)}
            onDrop={onDrop}
            onClick={() => fileRef.current?.click()}
            className={`group relative cursor-pointer overflow-hidden rounded-xl border-2 border-dashed p-8 text-center transition-all ${
              dragging
                ? "border-primary bg-primary/10 shadow-glow"
                : "border-border bg-elevated/40 hover:border-primary/50 hover:bg-elevated"
            }`}
          >
            <motion.div
              animate={dragging ? { scale: 1.05 } : { scale: 1 }}
              className="flex flex-col items-center gap-2"
            >
              <div className="grid h-12 w-12 place-items-center rounded-xl bg-gradient-primary shadow-glow">
                <Upload className="h-5 w-5 text-primary-foreground" />
              </div>
              <p className="text-sm font-medium">
                {dragging ? "Отпустите для загрузки" : "Перетащите файлы или нажмите, чтобы выбрать"}
              </p>
              <p className="font-mono text-[11px] text-muted-foreground">
                PNG · JPG · PDF · TXT · MD до 20 МБ
              </p>
            </motion.div>
            <input
              ref={fileRef}
              type="file"
              multiple
              className="hidden"
              onChange={(e) => handleFiles(e.target.files)}
            />
          </div>

          <div className="space-y-1.5">
            <Label className="text-xs text-muted-foreground">Или вставьте URL</Label>
            <Input placeholder="https://docs.example.com/api" className="bg-elevated" />
          </div>

          <AnimatePresence>
            {files.length > 0 && (
              <motion.div
                initial={{ opacity: 0, height: 0 }}
                animate={{ opacity: 1, height: "auto" }}
                exit={{ opacity: 0, height: 0 }}
                className="space-y-1.5 overflow-hidden"
              >
                <Label className="text-xs text-muted-foreground">
                  Вложения ({files.length})
                </Label>
                <div className="max-h-32 space-y-1 overflow-y-auto">
                  {files.map((f, i) => (
                    <motion.div
                      key={`${f.name}-${i}`}
                      initial={{ opacity: 0, x: -6 }}
                      animate={{ opacity: 1, x: 0 }}
                      className="flex items-center justify-between rounded-md border border-border bg-elevated px-2 py-1.5 text-xs"
                    >
                      <div className="flex min-w-0 items-center gap-2">
                        <FileText className="h-3.5 w-3.5 shrink-0 text-primary" />
                        <span className="truncate">{f.name}</span>
                      </div>
                      <button
                        onClick={() => setFiles((p) => p.filter((_, idx) => idx !== i))}
                        className="text-muted-foreground hover:text-destructive"
                      >
                        <X className="h-3.5 w-3.5" />
                      </button>
                    </motion.div>
                  ))}
                </div>
              </motion.div>
            )}
          </AnimatePresence>

          <DialogFooter>
            <Button variant="ghost" onClick={() => setUploadOpen(false)}>
              Отмена
            </Button>
            <Button
              className="bg-gradient-primary text-primary-foreground"
              onClick={() => {
                toast.success(
                  files.length
                    ? `Прикреплено вложений: ${files.length}`
                    : "Ссылка добавлена",
                );
                setUploadOpen(false);
              }}
            >
              <Check className="mr-1.5 h-3.5 w-3.5" /> Прикрепить
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}

function MenuRow({
  icon,
  label,
  hint,
  onClick,
}: {
  icon: React.ReactNode;
  label: string;
  hint: string;
  onClick: () => void;
}) {
  return (
    <DropdownMenuItem
      onSelect={(e) => {
        e.preventDefault();
        onClick();
      }}
      className="group flex cursor-pointer items-center justify-between gap-2 rounded-md px-2 py-1.5 text-xs focus:bg-elevated"
    >
      <div className="flex items-center gap-2">
        <span className="grid h-6 w-6 place-items-center rounded-md border border-border/70 bg-elevated text-muted-foreground group-hover:border-primary/50 group-hover:text-primary">
          {icon}
        </span>
        <span className="font-medium">{label}</span>
      </div>
      <span className="font-mono text-[10px] text-muted-foreground">{hint}</span>
    </DropdownMenuItem>
  );
}

function ModalShell({
  open,
  onClose,
  icon,
  title,
  description,
  children,
  onSave,
  saveLabel = "Сохранить",
}: {
  open: boolean;
  onClose: () => void;
  icon: React.ReactNode;
  title: string;
  description: string;
  children: React.ReactNode;
  onSave?: () => void;
  saveLabel?: string;
}) {
  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="max-h-[85vh] overflow-y-auto border-border/70 bg-popover/95 backdrop-blur-xl sm:max-w-xl">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <span className="grid h-7 w-7 place-items-center rounded-md bg-gradient-primary text-primary-foreground shadow-glow">
              {icon}
            </span>
            {title}
          </DialogTitle>
          <DialogDescription>{description}</DialogDescription>
        </DialogHeader>
        <div className="space-y-4 py-1">{children}</div>
        <DialogFooter>
          <Button variant="ghost" onClick={onClose}>
            Отмена
          </Button>
          <Button
            className="bg-gradient-primary text-primary-foreground"
            onClick={() => {
              onSave?.();
              toast.success(`${title} — сохранено`);
              onClose();
            }}
          >
            <Check className="mr-1.5 h-3.5 w-3.5" /> {saveLabel}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function Field({
  label,
  hint,
  children,
}: {
  label: string;
  hint?: string;
  children: React.ReactNode;
}) {
  return (
    <div className="space-y-1.5">
      <div className="flex items-center justify-between">
        <Label className="text-xs font-medium">{label}</Label>
        {hint && <span className="font-mono text-[10px] text-muted-foreground">{hint}</span>}
      </div>
      {children}
    </div>
  );
}

function SettingsModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  return (
    <ModalShell
      open={open}
      onClose={onClose}
      icon={<SettingsIcon className="h-3.5 w-3.5" />}
      title="Настройки рабочего пространства"
      description="Настройте команду агентов, выбор модели и параметры по умолчанию."
    >
      <Field label="Модель по умолчанию" hint="стриминг">
        <Input defaultValue="istok-coder-large" className="bg-elevated font-mono text-xs" />
      </Field>
      <Field label="API-ключ" hint="ротирован 2 дн. назад">
        <div className="relative">
          <Key className="pointer-events-none absolute left-2.5 top-2.5 h-3.5 w-3.5 text-muted-foreground" />
          <Input
            defaultValue="sk_live_••••••••••••••3f21"
            className="bg-elevated pl-8 font-mono text-xs"
          />
        </div>
      </Field>
      <Separator />
      <Toggle label="Авто-подтверждение планов" desc="Пропускать чекпоинт с участием человека." />
      <Toggle label="Транслировать ход мысли агентов" desc="Показывать рассуждения в панели Pulse." defaultChecked />
      <Toggle label="Телеметрия" desc="Помогает улучшать качество промтов." defaultChecked />
    </ModalShell>
  );
}

function KnowledgeModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  return (
    <ModalShell
      open={open}
      onClose={onClose}
      icon={<BookOpen className="h-3.5 w-3.5" />}
      title="База знаний"
      description="Постоянный контекст, к которому имеет доступ каждая генерация."
    >
      <Field label="Бриф проекта">
        <Textarea
          defaultValue="Исток — AI-билдер для разработчиков. Голос лаконичный, технический и уверенный."
          className="min-h-[96px] bg-elevated text-xs"
        />
      </Field>
      <Field label="Индексированные источники" hint="активно: 3">
        <div className="space-y-1.5">
          {[
            { name: "docs.istok.dev", icon: <Globe className="h-3.5 w-3.5" /> },
            { name: "design-system.md", icon: <FileText className="h-3.5 w-3.5" /> },
            { name: "api-spec.openapi.yaml", icon: <Cpu className="h-3.5 w-3.5" /> },
          ].map((s) => (
            <div
              key={s.name}
              className="flex items-center justify-between rounded-md border border-border bg-elevated px-2.5 py-1.5"
            >
              <div className="flex items-center gap-2 text-xs">
                <span className="text-primary">{s.icon}</span>
                <span className="font-mono">{s.name}</span>
              </div>
              <Badge variant="secondary" className="font-mono text-[10px]">
                синхр.
              </Badge>
            </div>
          ))}
        </div>
      </Field>
    </ModalShell>
  );
}

function GithubModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  return (
    <ModalShell
      open={open}
      onClose={onClose}
      icon={<Github className="h-3.5 w-3.5" />}
      title="Синхронизация с GitHub"
      description="Отправляйте сгенерированный код в репозиторий и открывайте PR от агента."
      saveLabel="Подключить репозиторий"
    >
      <Field label="Репозиторий">
        <Input defaultValue="istok-labs/web-app" className="bg-elevated font-mono text-xs" />
      </Field>
      <Field label="Основная ветка">
        <Input defaultValue="main" className="bg-elevated font-mono text-xs" />
      </Field>
      <Toggle label="Создавать PR автоматически" desc="Каждая генерация создаёт черновик PR." defaultChecked />
      <Toggle label="Подписывать коммиты" desc="Использовать GPG-ключ рабочего пространства." />
    </ModalShell>
  );
}

function ConnectorsModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const items = [
    { name: "Stripe", status: "Подключено", on: true },
    { name: "Supabase", status: "Подключено", on: true },
    { name: "Resend", status: "Доступно", on: false },
    { name: "OpenAI", status: "Подключено", on: true },
    { name: "Linear", status: "Доступно", on: false },
  ];
  return (
    <ModalShell
      open={open}
      onClose={onClose}
      icon={<Plug className="h-3.5 w-3.5" />}
      title="Коннекторы"
      description="Внешние сервисы, к которым агент обращается во время генерации."
    >
      <div className="space-y-1.5">
        {items.map((it) => (
          <div
            key={it.name}
            className="flex items-center justify-between rounded-md border border-border bg-elevated px-3 py-2"
          >
            <div>
              <p className="text-sm font-medium">{it.name}</p>
              <p className="font-mono text-[10px] text-muted-foreground">{it.status}</p>
            </div>
            <Switch defaultChecked={it.on} />
          </div>
        ))}
      </div>
    </ModalShell>
  );
}

function SkillModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  return (
    <ModalShell
      open={open}
      onClose={onClose}
      icon={<Sparkles className="h-3.5 w-3.5" />}
      title="Добавить навык"
      description="Сохраните повторно используемый промт, к которому агент сможет обратиться."
      saveLabel="Сохранить навык"
    >
      <Field label="Название навыка">
        <Input placeholder="например: Аудит лендинга" className="bg-elevated text-xs" />
      </Field>
      <Field label="Триггер-фраза" hint="необязательно">
        <Input placeholder="когда пользователь просит «провести аудит» страницы" className="bg-elevated text-xs" />
      </Field>
      <Field label="Инструкции">
        <Textarea
          placeholder="Шаги, тон, примеры…"
          className="min-h-[100px] bg-elevated text-xs"
        />
      </Field>
    </ModalShell>
  );
}

function Toggle({
  label,
  desc,
  defaultChecked,
}: {
  label: string;
  desc: string;
  defaultChecked?: boolean;
}) {
  return (
    <div className="flex items-center justify-between rounded-md border border-border bg-elevated px-3 py-2">
      <div>
        <p className="text-sm font-medium">{label}</p>
        <p className="text-[11px] text-muted-foreground">{desc}</p>
      </div>
      <Switch defaultChecked={defaultChecked} />
    </div>
  );
}
