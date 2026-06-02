import { useState, useCallback, useMemo, useEffect, useRef } from "react";
import { motion, AnimatePresence } from "framer-motion";
import {
  Monitor,
  Smartphone,
  Tablet,
  RotateCcw,
  Download,
  FolderDown,
  Code2,
  Eye,
  Upload,
  Lock,
  Globe,
  X,
  FileText,
  FileCode,
  Palette,
  MousePointer2,
  Send as SendIcon,
  Rocket,
  ShieldCheck,
  Loader2,
} from "lucide-react";
import { SandpackProvider, SandpackPreview as SandpackLivePreview, useSandpack } from "@codesandbox/sandpack-react";
import JSZip from "jszip";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { SidebarTrigger } from "@/components/ui/sidebar";
import { toast } from "sonner";
import CodeEditor, { getLanguage } from "@/components/CodeEditor";
import PublishModal from "@/components/PublishModal";
import FileExplorer from "@/components/FileExplorer";
import GenerationMagic from "@/components/workspace/GenerationMagic";
import InspectorEditPanel from "@/components/workspace/InspectorEditPanel";
import { useLanguage } from "@/hooks/useLanguage";

export interface ProjectFiles {
  [filename: string]: string;
}

export interface SelectedElement {
  tag: string;
  classes: string;
  text: string;
  id: string;
  componentName?: string | null;
}

/** Detect if project uses React/JSX (needs Sandpack bundler, not raw iframe) */
function isReactProject(files: ProjectFiles): boolean {
  return Object.keys(files).some((f) => /\.(tsx|jsx)$/.test(f));
}

/** Known LLM hallucinations: fake package → real replacement (or null to just remove) */
const HALLUCINATED_DEPS: Record<string, string | null> = {
  "@radix-ui/react-sheet": "@radix-ui/react-dialog",
  "@radix-ui/react-toaster": "@radix-ui/react-toast",
  "@radix-ui/react-drawer": "vaul",
  "@radix-ui/react-sonner": "sonner",
  "@radix-ui/react-carousel": "embla-carousel-react",
  "@radix-ui/react-input-otp": "input-otp",
  "@radix-ui/react-chart": "recharts",
  "@radix-ui/react-calendar": "react-day-picker",
  "@radix-ui/react-form": null,
  "@radix-ui/react-table": null,
  "@radix-ui/react-card": null,
  "@radix-ui/react-badge": null,
  "@radix-ui/react-sidebar": null,
  "@radix-ui/react-breadcrumb": null,
  "@radix-ui/react-skeleton": null,
  "@radix-ui/react-textarea": null,
};

/** Convert projectFiles to Sandpack format (paths prefixed with /) */
function toSandpackFiles(files: ProjectFiles): Record<string, string> {
  const result: Record<string, string> = {};
  for (const [path, content] of Object.entries(files)) {
    const key = path.startsWith("/") ? path : `/${path}`;
    let value = content == null ? "" : String(content);
    // Sandpack's bundler installs straight from package.json (incl. devDependencies),
    // bypassing customSetup. Sanitize it so a single hallucinated package (e.g.
    // @radix-ui/react-sheet) can't 404 the entire dependency install.
    if (path === "package.json" || path === "/package.json") {
      value = sanitizePackageJson(value);
    }
    result[key] = value;
  }
  return result;
}

/** Rewrite package.json: remap/remove hallucinated deps and drop devDependencies. */
function sanitizePackageJson(raw: string): string {
  try {
    const pkg = JSON.parse(raw);
    const deps: Record<string, string> = { ...(pkg.dependencies || {}) };
    for (const [fake, real] of Object.entries(HALLUCINATED_DEPS)) {
      if (fake in deps) {
        const ver = deps[fake];
        delete deps[fake];
        if (real && !(real in deps)) deps[real] = real === "latest" ? "latest" : ver;
      }
    }
    // Sandpack bundles internally — most devDependencies (eslint, vite, typescript...)
    // are unneeded and frequently include version/registry issues that break install.
    // BUT a few are required at bundle time: Sandpack runs PostCSS/Tailwind when these
    // are present, and tailwindcss-animate declares tailwindcss as a non-optional peer.
    // Promote that whitelist into dependencies instead of dropping it.
    const KEEP_FROM_DEV = [
      "tailwindcss", "tailwindcss-animate", "postcss", "autoprefixer",
      "vite", "@vitejs/plugin-react",
      "typescript", "@types/react", "@types/react-dom", "@types/node",
    ];
    const dev: Record<string, string> = pkg.devDependencies || {};
    for (const name of KEEP_FROM_DEV) {
      if (name in dev && !(name in deps)) deps[name] = dev[name];
    }
    pkg.dependencies = deps;
    delete pkg.devDependencies;
    return JSON.stringify(pkg, null, 2);
  } catch {
    return raw;
  }
}

/** Extract dependencies from generated package.json for Sandpack customSetup */
function extractSandpackDeps(files: ProjectFiles): Record<string, string> | undefined {
  const raw = files["package.json"];
  if (!raw) return undefined;
  try {
    const pkg = JSON.parse(String(raw));
    const deps = { ...(pkg.dependencies || {}) };
    // Remove deps already provided by vite-react-ts template
    delete deps["react"];
    delete deps["react-dom"];
    // Sanitize hallucinated npm packages
    for (const [fake, real] of Object.entries(HALLUCINATED_DEPS)) {
      if (fake in deps) {
        delete deps[fake];
        if (real && !(real in deps)) deps[real] = "latest";
      }
    }
    return Object.keys(deps).length > 0 ? deps : undefined;
  } catch {
    return undefined;
  }
}

/**
 * Detects Sandpack nodebox "Failed to get shell by ID" crashes (a known race in the
 * vite-react-ts runtime) and asks the parent to remount the provider with a fresh key.
 */
function SandpackCrashGuard({ onCrash }: { onCrash: () => void }) {
  const { listen } = useSandpack();
  useEffect(() => {
    const unsub = listen((msg) => {
      const m = msg as { type?: string; action?: string; message?: string; title?: string };
      const isError =
        (m.type === "action" && m.action === "show-error") || m.type === "error";
      const text = `${m.message ?? ""} ${m.title ?? ""}`;
      if (isError && /shell/i.test(text)) onCrash();
    });
    return unsub;
  }, [listen, onCrash]);
  return null;
}

interface WorkspacePreviewProps {
  projectFiles: ProjectFiles;
  onFilesChange: (files: ProjectFiles) => void;
  initialLoading: boolean;
  loaderStep: number;
  loaderSteps: string[];
  onPublish?: () => Promise<string | null>;
  editMode?: boolean;
  onEditModeChange?: (v: boolean) => void;
  onElementSelect?: (el: SelectedElement | null) => void;
  onTelegramExport?: () => void;

  // ── Workspace v3.0: Runable-parity buttons ──
  /** Триггер Railway deploy. Показывает "Deploying..." + opens logs_url on success. */
  onDeploy?: () => Promise<void>;
  /** Deploy in flight indicator. */
  deploying?: boolean;
  /** Открывает Security Audit панель. */
  onSecurityAudit?: () => void;
  /** Security gate пройден — галочка на кнопке аудита. */
  securityApproved?: boolean;

  // Generation state — keeps GenerationMagic visible throughout all tiers
  thinking?: boolean;

  // Live agent stream for GenerationMagic
  milestones?: { agent: string; message: string; progress: number; status: string }[];
  activeAgent?: string | null;
  streamedFiles?: { name: string; size: number; receivedAt: Date }[];
  currentFSMState?: string;

  // Resume support
  canResume?: boolean;
  onResume?: () => void;
}

/** Edit-mode script injected into the iframe */
const EDIT_MODE_SCRIPT = `
<script data-istok-inspector>
(function() {
  var editMode = false;
  var lastHovered = null;

  function getSelector(el) {
    var tag = el.tagName.toLowerCase();
    var classes = el.className && typeof el.className === 'string' ? '.' + el.className.trim().split(/\\s+/).join('.') : '';
    var id = el.id ? '#' + el.id : '';
    return tag + id + classes;
  }

  function onHover(e) {
    if (!editMode) return;
    if (lastHovered) lastHovered.removeAttribute('data-istok-selected');
    e.target.setAttribute('data-istok-hover', '');
    lastHovered = e.target;
  }

  function onLeave(e) {
    if (!editMode) return;
    e.target.removeAttribute('data-istok-hover');
  }

  function onClick(e) {
    if (!editMode) return;
    e.preventDefault();
    e.stopPropagation();
    e.stopImmediatePropagation();
    
    if (lastHovered) lastHovered.removeAttribute('data-istok-selected');
    
    var el = e.target;
    el.setAttribute('data-istok-selected', '');
    el.removeAttribute('data-istok-hover');
    lastHovered = el;

    var text = (el.textContent || '').trim().slice(0, 80);
    window.parent.postMessage({
      type: 'istok-element-select',
      payload: {
        tag: el.tagName.toLowerCase(),
        classes: (el.className && typeof el.className === 'string') ? el.className.trim() : '',
        text: text,
        id: el.id || ''
      }
    }, '*');
  }

  window.addEventListener('message', function(e) {
    if (e.data && e.data.type === 'istok-edit-mode') {
      editMode = e.data.enabled;
      if (!editMode) {
        document.querySelectorAll('[data-istok-hover],[data-istok-selected]').forEach(function(el) {
          el.removeAttribute('data-istok-hover');
          el.removeAttribute('data-istok-selected');
        });
      }
    }
  });

  document.addEventListener('mouseover', onHover, true);
  document.addEventListener('mouseout', onLeave, true);
  document.addEventListener('click', onClick, true);
})();
</script>
<style data-istok-inspector>
[data-istok-hover] {
  outline: 2px dashed hsla(263, 70%, 58%, 0.6) !important;
  outline-offset: 2px !important;
  cursor: pointer !important;
  box-shadow: 0 0 12px hsla(263, 70%, 58%, 0.15) !important;
  transition: outline 0.15s ease, box-shadow 0.15s ease !important;
}
[data-istok-selected] {
  outline: 2px solid hsl(263, 70%, 58%) !important;
  outline-offset: 2px !important;
  cursor: pointer !important;
  box-shadow: 0 0 20px hsla(263, 70%, 58%, 0.25), 0 0 40px hsla(263, 70%, 58%, 0.08) !important;
  transition: outline 0.15s ease, box-shadow 0.15s ease !important;
}
</style>
`;

/** Build a single HTML document from multi-file project for iframe preview */
function buildPreviewHtml(files: ProjectFiles, injectEditMode: boolean): string {
  // CRITICAL: always coerce to string — API can return objects, numbers, null
  const safeStr = (v: unknown): string => (v == null ? "" : String(v));
  let html = safeStr(files["index.html"]);
  if (Object.keys(files).length === 1 && html && !injectEditMode) return html;

  let result = html;

  // Inline CSS files
  for (const [name, raw] of Object.entries(files)) {
    const content = safeStr(raw);
    if (name.endsWith(".css")) {
      const linkRegex = new RegExp(`<link[^>]*href=["']${name.replace(".", "\\.")}["'][^>]*/?>`, "gi");
      if (linkRegex.test(result)) {
        result = result.replace(linkRegex, `<style>/* ${name} */\n${content}\n</style>`);
      } else {
        result = result.replace("</head>", `<style>/* ${name} */\n${content}\n</style>\n</head>`);
      }
    }
  }

  // Inline JS files
  for (const [name, raw] of Object.entries(files)) {
    const content = safeStr(raw);
    if (name.endsWith(".js") || name.endsWith(".ts")) {
      const scriptRegex = new RegExp(`<script[^>]*src=["']${name.replace(".", "\\.")}["'][^>]*>\\s*</script>`, "gi");
      if (scriptRegex.test(result)) {
        result = result.replace(scriptRegex, `<script>/* ${name} */\n${content}\n</script>`);
      } else {
        result = result.replace("</body>", `<script>/* ${name} */\n${content}\n</script>\n</body>`);
      }
    }
  }

  // Inject edit mode script
  if (injectEditMode) {
    if (result.includes("</head>")) {
      result = result.replace("</head>", `${EDIT_MODE_SCRIPT}\n</head>`);
    } else if (result.includes("</body>")) {
      result = result.replace("</body>", `${EDIT_MODE_SCRIPT}\n</body>`);
    } else {
      result += EDIT_MODE_SCRIPT;
    }
  }

  return result;
}

/**
 * Canonical helpers moved to @/lib/sse-parsers to eliminate duplication.
 * Re-exported here only to preserve existing import paths (call sites like
 * useGeneration + ViewProject use these names).
 */
export { stripMarkdownFences, codeToFiles, filesToCode } from "@/lib/sse-parsers";

const WorkspacePreview = ({
  projectFiles,
  onFilesChange,
  initialLoading,
  loaderStep,
  loaderSteps,
  onPublish,
  editMode = false,
  onEditModeChange,
  onElementSelect,
  onTelegramExport,
  onDeploy,
  deploying = false,
  onSecurityAudit,
  securityApproved = false,
  thinking = false,
  milestones = [],
  activeAgent = null,
  streamedFiles = [],
  currentFSMState = "idle",
  canResume = false,
  onResume,
}: WorkspacePreviewProps) => {
  const { t } = useLanguage();
  const [viewMode, setViewMode] = useState<"desktop" | "tablet" | "mobile">("desktop");
  const [activeTab, setActiveTab] = useState<string>("preview");
  const [activeFile, setActiveFile] = useState<string>("index.html");
  const [openTabs, setOpenTabs] = useState<string[]>(["index.html"]);
  const [publishModalOpen, setPublishModalOpen] = useState(false);
  const [publishedUrl, setPublishedUrl] = useState("");
  const [publishing, setPublishing] = useState(false);
  const [iframeRef, setIframeRef] = useState<HTMLIFrameElement | null>(null);
  const [iframeReady, setIframeReady] = useState(true);
  const [inspectorElement, setInspectorElement] = useState<SelectedElement | null>(null);

  // Reset iframeReady when new generation starts; restore when generation ends
  // (AnimatePresence mode="wait" prevents iframe from mounting while loading shows,
  //  so onLoad can't fire — we must explicitly allow iframe to render)
  useEffect(() => {
    if (thinking) setIframeReady(false);
    else setIframeReady(true);
  }, [thinking]);

  // Always inject edit mode script so it's ready
  const previewHtml = useMemo(() => buildPreviewHtml(projectFiles, true), [projectFiles]);

  // Sandpack mode: detect multi-file React/Vite projects
  const reactProject = useMemo(() => isReactProject(projectFiles), [projectFiles]);
  const sandpackFiles = useMemo(
    () => (reactProject ? toSandpackFiles(projectFiles) : {}),
    [projectFiles, reactProject],
  );
  const sandpackDeps = useMemo(
    () => (reactProject ? extractSandpackDeps(projectFiles) : undefined),
    [projectFiles, reactProject],
  );

  // Auto-recovery for nodebox shell crashes: remount SandpackProvider with a fresh key
  // (capped) instead of leaving the user stuck on Sandpack's red error overlay.
  const sandpackRetries = useRef(0);
  const [sandpackKey, setSandpackKey] = useState(0);
  const handleSandpackCrash = useCallback(() => {
    if (sandpackRetries.current >= 2) return;
    sandpackRetries.current += 1;
    setSandpackKey((k) => k + 1);
  }, []);
  useEffect(() => {
    sandpackRetries.current = 0;
  }, [sandpackFiles]);

  // Send edit mode state to iframe (both legacy and InspectorProvider protocols)
  useEffect(() => {
    if (iframeRef?.contentWindow) {
      // Legacy single-file protocol
      iframeRef.contentWindow.postMessage({ type: "istok-edit-mode", enabled: editMode }, "*");
      // React InspectorProvider protocol (field name must match: provider reads `enabled`)
      iframeRef.contentWindow.postMessage({ type: "ISTOK_SET_INSPECT", enabled: editMode }, "*");
    }
    // React/Sandpack projects don't use iframeRef — Sandpack owns its iframe. Broadcast
    // directly to its preview iframe. Retry briefly since it loads asynchronously.
    if (reactProject) {
      const notify = () => {
        document.querySelectorAll<HTMLIFrameElement>("iframe.sp-preview-iframe").forEach((f) => {
          f.contentWindow?.postMessage({ type: "ISTOK_SET_INSPECT", enabled: editMode }, "*");
        });
      };
      notify();
      const t1 = setTimeout(notify, 400);
      const t2 = setTimeout(notify, 1200);
      return () => { clearTimeout(t1); clearTimeout(t2); };
    }
  }, [editMode, iframeRef, reactProject]);

  // Listen for element selection from iframe (both legacy script and InspectorProvider)
  useEffect(() => {
    const handler = (e: MessageEvent) => {
      // Legacy single-file HTML inspector
      if (e.data?.type === "istok-element-select") {
        const el: SelectedElement = e.data.payload;
        setInspectorElement(el);
        onElementSelect?.(el);
      }
      // React InspectorProvider (multi-file projects)
      if (e.data?.type === "ISTOK_ELEMENT_CLICKED" && e.data.elementData) {
        const d = e.data.elementData;
        const el: SelectedElement = {
          tag: d.tagName || "div",
          classes: d.className || "",
          text: (d.textContent || "").slice(0, 80),
          id: d.id || "",
          componentName: d.componentName || null,
        };
        setInspectorElement(el);
        onElementSelect?.(el);
      }
    };
    window.addEventListener("message", handler);
    return () => window.removeEventListener("message", handler);
  }, [onElementSelect]);

  // Send edit mode on iframe load + mark iframe as ready
  const handleIframeLoad = useCallback(() => {
    setIframeReady(true);
    if (iframeRef?.contentWindow) {
      iframeRef.contentWindow.postMessage({ type: "istok-edit-mode", enabled: editMode }, "*");
      iframeRef.contentWindow.postMessage({ type: "ISTOK_SET_INSPECT", value: editMode }, "*");
    }
  }, [editMode, iframeRef]);

  const handlePublish = useCallback(async () => {
    if (!onPublish) {
      const blob = new Blob([previewHtml], { type: "text/html" });
      window.open(URL.createObjectURL(blob), "_blank");
      return;
    }
    setPublishing(true);
    const slug = await onPublish();
    setPublishing(false);
    if (slug) {
      const url = `${window.location.origin}/view/${slug}`;
      setPublishedUrl(url);
      setPublishModalOpen(true);
      toast.success("Проект опубликован!");
    } else {
      toast.error("Сначала сгенерируйте код проекта");
    }
  }, [onPublish, previewHtml]);

  const handleDownload = useCallback(() => {
    const blob = new Blob([projectFiles["index.html"] || ""], { type: "text/html;charset=utf-8" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "index.html";
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    toast.success("Файл скачан!");
  }, [projectFiles]);

  const handleDownloadZip = useCallback(async () => {
    const zip = new JSZip();
    for (const [name, content] of Object.entries(projectFiles)) {
      zip.file(name, content);
    }
    const blob = await zip.generateAsync({ type: "blob" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "project.zip";
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    toast.success("Архив проекта готов к загрузке");
  }, [projectFiles]);

  const handleDownloadProject = useCallback(async () => {
    try {
      const apiBase = import.meta.env.VITE_API_URL || "http://localhost:8080";
      const resp = await fetch(`${apiBase}/api/v1/project/export`);
      if (!resp.ok) {
        const err = await resp.json().catch(() => ({ error: "Download failed" }));
        toast.error(err.error || "Download failed");
        return;
      }
      const blob = await resp.blob();
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = resp.headers.get("Content-Disposition")?.match(/filename="?([^"]+)"?/)?.[1] || "istok-project.zip";
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
      toast.success("Полный проект скачан с сервера!");
    } catch (e) {
      toast.error("Ошибка загрузки проекта");
    }
  }, []);

  const handleSelectFile = useCallback((filename: string) => {
    setActiveFile(filename);
    if (!openTabs.includes(filename)) {
      setOpenTabs(prev => [...prev, filename]);
    }
    setActiveTab("code");
  }, [openTabs]);

  const handleCloseTab = useCallback((filename: string, e: React.MouseEvent) => {
    e.stopPropagation();
    setOpenTabs(prev => {
      const next = prev.filter(f => f !== filename);
      if (next.length === 0) return ["index.html"];
      if (activeFile === filename) {
        const idx = prev.indexOf(filename);
        setActiveFile(next[Math.min(idx, next.length - 1)]);
      }
      return next;
    });
  }, [activeFile]);

  const handleCodeChange = useCallback(
    (newCode: string) => {
      onFilesChange({ ...projectFiles, [activeFile]: newCode });
    },
    [projectFiles, activeFile, onFilesChange]
  );

  const getTabIcon = (name: string) => {
    if (name.endsWith(".css")) return <Palette size={11} className="text-blue-400 shrink-0" />;
    if (name.endsWith(".js") || name.endsWith(".ts")) return <FileCode size={11} className="text-yellow-400 shrink-0" />;
    return <FileText size={11} className="text-orange-400 shrink-0" />;
  };

  const currentFileContent = projectFiles[activeFile] || projectFiles["index.html"] || "";

  return (
    <div className="flex-1 min-w-0 min-h-0 h-full flex flex-col overflow-hidden relative">
      {/* Toolbar — sticky within flex column */}
      <header
        className="z-20 h-14 shrink-0 border-b border-[hsl(var(--border))]/10 flex items-center justify-between px-2 sm:px-3 glass bg-background/95 backdrop-blur-md"
      >
        <div className="flex items-center gap-2">
          <SidebarTrigger className="text-muted-foreground hover:text-foreground" />
          <div className="w-px h-5 bg-border/20" />
          <Tabs value={activeTab} onValueChange={setActiveTab}>
            <TabsList className="h-7 bg-secondary/40 p-0.5">
              <TabsTrigger value="preview" className="h-6 px-2.5 text-[11px] gap-1 data-[state=active]:bg-background">
                <Eye size={11} /> {t("wsPreview") || "Превью"}
              </TabsTrigger>
              <TabsTrigger value="code" className="h-6 px-2.5 text-[11px] gap-1 data-[state=active]:bg-background">
                <Code2 size={11} /> {t("wsCode") || "Код"}
              </TabsTrigger>
            </TabsList>
          </Tabs>
          <div className="w-px h-5 bg-border/20" />

          {/* Inspector (point-and-click visual editor) toggle */}
          {activeTab === "preview" && (
            <button
              onClick={() => onEditModeChange?.(!editMode)}
              className={`flex items-center gap-1.5 h-7 px-2.5 rounded-md text-[11px] font-medium transition-all duration-200 ${
                editMode
                  ? "bg-primary/20 text-primary shadow-[0_0_12px_hsla(263,70%,58%,0.15)]"
                  : "text-muted-foreground hover:text-foreground hover:bg-secondary/50"
              }`}
              title="Инспектор — точечное редактирование элементов"
            >
              <MousePointer2 size={13} className={editMode ? "animate-pulse" : ""} />
              <span className="hidden sm:inline">🔍 Инспектор</span>
            </button>
          )}

          {activeTab === "preview" && (
            <div className="hidden md:flex items-center gap-1.5 bg-secondary/40 rounded-lg px-3 py-1 min-w-[140px] max-w-[300px] lg:max-w-[400px]">
              <Lock size={10} className="text-muted-foreground/50 shrink-0" />
              <Globe size={10} className="text-muted-foreground/50 shrink-0" />
              <span className="text-[11px] text-muted-foreground/70 truncate">preview.istok.app/project/new</span>
            </div>
          )}
        </div>

        <div className="flex items-center gap-1">
          {activeTab === "preview" && (
            <>
              {[
                { mode: "desktop" as const, icon: Monitor },
                { mode: "tablet" as const, icon: Tablet },
                { mode: "mobile" as const, icon: Smartphone },
              ].map(({ mode, icon: Icon }) => (
                <button
                  key={mode}
                  onClick={() => setViewMode(mode)}
                  className={`w-7 h-7 rounded-md flex items-center justify-center transition-colors ${
                    viewMode === mode ? "bg-primary/15 text-primary" : "text-muted-foreground hover:text-foreground"
                  }`}
                >
                  <Icon size={14} />
                </button>
              ))}
              <div className="w-px h-5 bg-border/20 mx-1" />
            </>
          )}
          <button onClick={() => onFilesChange({ ...projectFiles })} className="w-7 h-7 rounded-md flex items-center justify-center text-muted-foreground hover:text-foreground transition-colors" title="Обновить">
            <RotateCcw size={14} />
          </button>
          <button onClick={handleDownload} className="flex items-center gap-1.5 h-7 px-2.5 rounded-md text-muted-foreground hover:text-foreground hover:bg-secondary/50 transition-colors text-[11px]" title="Скачать HTML">
            <Download size={13} />
            <span className="hidden sm:inline">.html</span>
          </button>
          <button onClick={handleDownloadZip} className="flex items-center gap-1.5 h-7 px-2.5 rounded-md text-muted-foreground hover:text-foreground hover:bg-secondary/50 transition-colors text-[11px]" title="Скачать ZIP">
            <FolderDown size={13} />
            <span className="hidden sm:inline">.zip</span>
          </button>
          <button onClick={handleDownloadProject} className="flex items-center gap-1.5 h-7 px-2.5 rounded-md bg-primary/10 text-primary hover:bg-primary/20 transition-colors text-[11px] font-medium" title="Download Project (full server-side ZIP)">
            <Rocket size={13} />
            <span className="hidden sm:inline">Download Project</span>
          </button>
          {onTelegramExport && (
            <button onClick={onTelegramExport} className="flex items-center gap-1.5 h-7 px-2.5 rounded-md text-muted-foreground hover:text-foreground hover:bg-secondary/50 transition-colors text-[11px]" title="Экспорт в Telegram Web App">
              <SendIcon size={13} />
              <span className="hidden sm:inline">TWA</span>
            </button>
          )}
          {/* Security Audit — показывает галочку если Verification Gate пройден */}
          {onSecurityAudit && (
            <button
              onClick={onSecurityAudit}
              className={`flex items-center gap-1.5 h-7 px-2.5 rounded-md text-[11px] font-medium transition-colors ${
                securityApproved
                  ? "bg-emerald-500/15 text-emerald-400 hover:bg-emerald-500/25"
                  : "bg-secondary/40 text-muted-foreground hover:text-foreground hover:bg-secondary/60"
              }`}
              title={securityApproved ? "Security Gate: PASSED" : "Run Security Audit"}
            >
              <ShieldCheck size={12} />
              <span className="hidden sm:inline">
                {securityApproved ? "Secure" : "Audit"}
              </span>
            </button>
          )}

          {/* Deploy (Railway) */}
          {onDeploy && (
            <button
              onClick={() => onDeploy()}
              disabled={deploying}
              className="flex items-center gap-1.5 h-7 px-2.5 rounded-md bg-amber-500/15 text-amber-400 hover:bg-amber-500/25 transition-colors text-[11px] font-medium disabled:opacity-50"
              title="Deploy to Railway"
            >
              {deploying ? (
                <Loader2 size={12} className="animate-spin" />
              ) : (
                <Rocket size={12} />
              )}
              <span className="hidden sm:inline">
                {deploying ? "Deploying…" : "Deploy"}
              </span>
            </button>
          )}

          {/* Publish (preview URL) */}
          <button
            onClick={handlePublish}
            disabled={publishing}
            className="flex items-center gap-1.5 h-7 px-3 rounded-md bg-primary/15 text-primary hover:bg-primary/25 transition-colors text-[11px] font-medium disabled:opacity-50"
          >
            <Upload size={12} />
            <span className="hidden sm:inline">
              {publishing ? "Публикация..." : "Опубликовать"}
            </span>
          </button>
        </div>
      </header>

      {/* Content area — fills all remaining space below toolbar */}
      <div className="flex-1 min-h-0 relative overflow-hidden">
        <AnimatePresence mode="wait">
          {(initialLoading || thinking || !iframeReady || canResume) ? (
            <motion.div
              key="generation-magic"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0, scale: 0.98 }}
              transition={{ duration: 0.5 }}
              className="absolute inset-0 z-[20]"
            >
              <GenerationMagic
                logs={milestones.length > 0 ? milestones.map(m => m.message) : loaderSteps.slice(0, loaderStep + 1)}
                progress={milestones.length > 0 ? Math.max(...milestones.map(m => m.progress), 0) : (loaderSteps.length > 0 ? Math.round(((loaderStep + 1) / loaderSteps.length) * 100) : 0)}
                streamedFiles={streamedFiles}
                milestones={milestones}
                currentFSMState={currentFSMState}
                canResume={canResume}
                onResume={onResume}
              />
            </motion.div>
          ) : activeTab === "preview" ? (
            <motion.div
              key="preview-tab"
              initial={{ opacity: 0, y: 4 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -4 }}
              transition={{ duration: 0.25, ease: [0.22, 1, 0.36, 1] }}
              className="absolute inset-0 flex items-center justify-center bg-transparent"
            >
              {editMode && (
                <div className="absolute top-2 left-1/2 -translate-x-1/2 z-10 px-3 py-1 rounded-full glass border border-primary/20 text-[10px] text-primary font-medium flex items-center gap-1.5">
                  <MousePointer2 size={10} />
                  Нажмите на элемент для выбора
                </div>
              )}
              <div className={`w-full h-full overflow-hidden transition-all duration-300 ${
                editMode ? "shadow-glow-primary" : ""
              } ${
                viewMode === "desktop" ? "" : viewMode === "tablet" ? "max-w-[768px] mx-auto" : "max-w-[375px] mx-auto"
              }`}>
                {reactProject ? (
                  <SandpackProvider
                    key={sandpackKey}
                    template="vite-react-ts"
                    files={sandpackFiles}
                    theme="dark"
                    customSetup={sandpackDeps ? { dependencies: sandpackDeps } : undefined}
                    options={{ externalResources: ["https://cdn.tailwindcss.com"] }}
                  >
                    <SandpackCrashGuard onCrash={handleSandpackCrash} />
                    <SandpackLivePreview
                      showNavigator={false}
                      style={{ height: "100%", width: "100%" }}
                    />
                  </SandpackProvider>
                ) : (
                  <iframe
                    ref={setIframeRef}
                    onLoad={handleIframeLoad}
                    key={previewHtml}
                    title="preview"
                    className="w-full h-full border-0"
                    srcDoc={previewHtml}
                    sandbox="allow-scripts allow-same-origin allow-forms"
                  />
                )}
              </div>
              {/* Floating Inspector Edit Panel */}
              {editMode && (
                <InspectorEditPanel
                  selectedElement={inspectorElement}
                  onClose={() => {
                    setInspectorElement(null);
                    onElementSelect?.(null);
                  }}
                  onApply={(instruction) => {
                    // Forward to parent chat with element context
                    onElementSelect?.(inspectorElement);
                    setInspectorElement(null);
                    // Dispatch custom event for Workspace to pick up
                    window.dispatchEvent(new CustomEvent("istok-inspector-apply", {
                      detail: { instruction, element: inspectorElement },
                    }));
                  }}
                  thinking={thinking}
                />
              )}
            </motion.div>
          ) : (
            <motion.div
              key="code-tab"
              initial={{ opacity: 0, y: 4 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -4 }}
              transition={{ duration: 0.25, ease: [0.22, 1, 0.36, 1] }}
              className="absolute inset-0 flex flex-col bg-[hsl(240,6%,7%)]"
            >
              {/* VS Code-style tabs */}
              <div className="flex items-center border-b border-glass-border/30 bg-[hsl(240,6%,9%)] shrink-0 overflow-x-auto">
                <AnimatePresence mode="popLayout">
                  {openTabs.map((tab) => (
                    <motion.button
                      key={tab}
                      initial={{ opacity: 0, width: 0 }}
                      animate={{ opacity: 1, width: "auto" }}
                      exit={{ opacity: 0, width: 0 }}
                      transition={{ duration: 0.15 }}
                      onClick={() => setActiveFile(tab)}
                      className={`group flex items-center gap-1.5 h-8 px-3 text-[11px] border-r border-[hsl(var(--border))]/10 whitespace-nowrap transition-colors ${
                        activeFile === tab
                          ? "bg-[hsl(240,6%,7%)] text-foreground border-t-2 border-t-primary"
                          : "text-muted-foreground hover:text-foreground hover:bg-[hsl(240,6%,8%)]"
                      }`}
                    >
                      {getTabIcon(tab)}
                      <span>{tab}</span>
                      <span
                        onClick={(e) => handleCloseTab(tab, e)}
                        className="ml-1 w-4 h-4 rounded flex items-center justify-center opacity-0 group-hover:opacity-100 hover:bg-secondary/50 transition-all"
                      >
                        <X size={9} />
                      </span>
                    </motion.button>
                  ))}
                </AnimatePresence>
              </div>
              {/* Editor area with file explorer */}
              <div className="flex-1 flex min-h-0">
                <FileExplorer
                  files={projectFiles}
                  activeFile={activeFile}
                  onSelectFile={handleSelectFile}
                />
                <div className="flex-1 min-w-0">
                  <CodeEditor code={currentFileContent} onChange={handleCodeChange} language={getLanguage(activeFile)} />
                </div>
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </div>

      <PublishModal open={publishModalOpen} onClose={() => setPublishModalOpen(false)} projectUrl={publishedUrl} />
    </div>
  );
};

export default WorkspacePreview;
