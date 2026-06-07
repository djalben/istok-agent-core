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

// ──────────────────────────────────────────────────────────────────────────
// IMMUTABLE FOUNDATION (sandbox isolation approach)
// ──────────────────────────────────────────────────────────────────────────
// LLM-generated infra files are unstable and crash Sandpack's Vite/Rollup.
// We fully isolate infra from the AI: only the AI's /src code is used, while
// package.json + all build configs are hardcoded to a proven Vite 4 setup.

/** AI-generated infrastructure files — filtered out, replaced by hardcoded versions. */
const INFRA_FILES = new Set([
  "package.json", "/package.json",
  "vite.config.ts", "/vite.config.ts",
  "vite.config.js", "/vite.config.js",
  "tsconfig.json", "/tsconfig.json",
  "tsconfig.node.json", "/tsconfig.node.json",
  "postcss.config.js", "/postcss.config.js",
  "postcss.config.cjs", "/postcss.config.cjs",
  "tailwind.config.js", "/tailwind.config.js",
  "tailwind.config.ts", "/tailwind.config.ts",
  "index.html", "/index.html",
]);

/** Hardcoded package.json — proven-stable Vite 4 environment. */
const HARDCODED_PACKAGE_JSON = JSON.stringify({
  name: "istok-project",
  private: true,
  version: "0.0.0",
  type: "module",
  scripts: {
    dev: "vite",
    build: "tsc && vite build",
    preview: "vite preview",
  },
  dependencies: {
    "react": "^18.2.0",
    "react-dom": "^18.2.0",
    "lucide-react": "^0.263.1",
    "framer-motion": "^10.12.16",
    "clsx": "^1.2.1",
    "tailwind-merge": "^1.13.2",
    "react-router-dom": "^6.14.1",
    "class-variance-authority": "^0.7.0",
    "@radix-ui/react-slot": "^1.0.2",
    "@radix-ui/react-dialog": "^1.0.5",
    "@radix-ui/react-dropdown-menu": "^2.0.6",
    "@radix-ui/react-tabs": "^1.0.4",
    "@radix-ui/react-toast": "^1.1.5",
    "@radix-ui/react-label": "^2.0.2",
    "@radix-ui/react-select": "^2.0.0",
    "@radix-ui/react-separator": "^1.0.3",
    "@radix-ui/react-scroll-area": "^1.0.5",
    "@radix-ui/react-accordion": "^1.1.2",
    "@radix-ui/react-avatar": "^1.0.4",
    "@radix-ui/react-checkbox": "^1.0.4",
    "@radix-ui/react-popover": "^1.0.7",
    "@radix-ui/react-tooltip": "^1.0.7",
    "@radix-ui/react-switch": "^1.0.3",
    "sonner": "^1.4.0",
  },
  devDependencies: {
    "@types/react": "^18.2.15",
    "@types/react-dom": "^18.2.7",
    "@vitejs/plugin-react": "^4.0.3",
    "autoprefixer": "^10.4.14",
    "postcss": "^8.4.24",
    "tailwindcss": "^3.3.2",
    "typescript": "^5.0.2",
    "vite": "^4.4.5",
    "esbuild-wasm": "^0.18.20",
  },
}, null, 2);

/** Hardcoded vite.config.ts — Vite 4 + React plugin + @ alias. */
const HARDCODED_VITE_CONFIG = `import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@": "/src",
    },
  },
});
`;

/** Hardcoded tailwind.config.js — scans src for classes. */
const HARDCODED_TAILWIND_CONFIG = `/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{js,ts,jsx,tsx}"],
  theme: {
    extend: {},
  },
  plugins: [],
};
`;

/** Hardcoded postcss.config.js — Tailwind + Autoprefixer. */
const HARDCODED_POSTCSS_CONFIG = `export default {
  plugins: {
    tailwindcss: {},
    autoprefixer: {},
  },
};
`;

/** Hardcoded index.html — standard Vite entrypoint. */
const HARDCODED_INDEX_HTML = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Istok Project</title>
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/src/main.tsx"></script>
  </body>
</html>
`;

// Binary/asset extensions that can't be text-bundled and only bloat the worker
// payload (they trigger "DataCloneError: out of memory" when postMessage'd to the
// Nodebox worker on heavy projects). Skipped from the live preview.
const BINARY_EXT = /\.(png|jpe?g|gif|webp|avif|bmp|ico|woff2?|ttf|otf|eot|mp4|webm|mov|mp3|wav|ogg|pdf|zip|gz|tar)$/i;
// Guards against the Nodebox worker OOM: drop any single oversized file and stop
// adding files once the cumulative payload gets too large for postMessage cloning.
const MAX_FILE_BYTES = 256 * 1024; // 256KB per file
const MAX_TOTAL_BYTES = 4 * 1024 * 1024; // ~4MB total

/**
 * Find the React entry point so the classic bundler starts from the generated app
 * instead of the template's default ("Hello world") /index.tsx. Order matters.
 */
function detectSandpackEntry(files: Record<string, string>): string {
  const candidates = [
    "/src/main.tsx", "/src/main.jsx", "/src/main.ts",
    "/src/index.tsx", "/src/index.jsx",
    "/src/App.tsx", "/src/App.jsx",
    "/index.tsx", "/index.jsx", "/App.tsx",
  ];
  return candidates.find((c) => c in files) ?? "/src/main.tsx";
}

/** Relative import specifier (with extension) from one sandbox path to another. */
function relativeImport(fromPath: string, toPath: string): string {
  const from = fromPath.split("/").slice(0, -1);
  const to = toPath.split("/");
  let i = 0;
  while (i < from.length && i < to.length - 1 && from[i] === to[i]) i++;
  const ups = from.slice(i).map(() => "..");
  const rel = [...ups, ...to.slice(i)].join("/");
  return rel.startsWith(".") ? rel : `./${rel}`;
}

/**
 * A file that ACTUALLY mounts React (createRoot / ReactDOM.render) AND exists in the
 * file set, preferring conventional entry names. Returns null when the generated
 * project ships no mount file — the caller then synthesizes one.
 */
function findRenderEntry(files: Record<string, string>): string | null {
  const mounts = Object.keys(files).filter(
    (f) => /\.(tsx|jsx|ts|js)$/.test(f) && /createRoot|ReactDOM\.render/.test(files[f]),
  );
  const prefer = [
    "/src/main.tsx", "/src/main.jsx", "/src/index.tsx", "/src/index.jsx",
    "/index.tsx", "/index.jsx",
  ];
  return prefer.find((p) => mounts.includes(p)) ?? mounts[0] ?? null;
}

/** The root component to render when no mount file exists: prefer App, else first default-exporting component. */
function findRootComponent(files: Record<string, string>): string | null {
  const prefer = ["/src/App.tsx", "/src/App.jsx", "/App.tsx", "/App.jsx"];
  const hit = prefer.find((p) => p in files);
  if (hit) return hit;
  const comps = Object.keys(files).filter((f) => /\.(tsx|jsx)$/.test(f)).sort();
  return comps.find((f) => /export\s+default/.test(files[f])) ?? comps[0] ?? null;
}

/** Synthesize a React 18 entry (/src/main.tsx) that mounts the given root component. */
function buildMainShim(componentPath: string): string {
  const rel = relativeImport("/src/main.tsx", componentPath);
  return `import React from "react";
import { createRoot } from "react-dom/client";
import App from "${rel}";

const rootEl = document.getElementById("root");
if (rootEl) createRoot(rootEl).render(React.createElement(App));
`;
}

/**
 * Convert projectFiles to Sandpack format. Filters out AI-generated infra files
 * and force-injects the immutable hardcoded foundation. Only the AI's /src code
 * is passed through — the sandbox can never crash on bad dependencies/configs.
 *
 * Also caps the payload size (per-file + total) and skips binary assets so the
 * classic bundler never runs out of memory cloning a massive bulk update.
 */
function toSandpackFiles(files: ProjectFiles): Record<string, string> {
  const result: Record<string, string> = {};
  let totalBytes = 0;
  for (const [path, content] of Object.entries(files)) {
    const bare = path.replace(/^\//, "");
    if (INFRA_FILES.has(bare) || INFRA_FILES.has(path)) continue;
    if (BINARY_EXT.test(bare)) continue; // assets can't be text-bundled
    let value = content == null ? "" : String(content);
    if (value.length > MAX_FILE_BYTES) continue; // skip oversized file
    if (totalBytes + value.length > MAX_TOTAL_BYTES) continue; // protect worker memory
    // Classic bundler has no Vite "@/" alias — rewrite to absolute "/src/".
    value = value.replace(/(['"])@\//g, "$1/src/");
    totalBytes += value.length;
    const key = path.startsWith("/") ? path : `/${path}`;
    result[key] = value;
  }
  // Force-inject the immutable foundation; "main" points the classic bundler at the entry.
  // The entry MUST be a file that actually mounts React AND exists in `result`, otherwise
  // the /index.tsx shim below imports a missing module
  // ("Could not find module './src/main.tsx'").
  let entry = findRenderEntry(result);
  if (entry == null) {
    // Generated project ships no mount file → synthesize /src/main.tsx that renders the
    // discovered root component, so the preview boots instead of crashing on a dead entry.
    const root = findRootComponent(result);
    if (root) {
      result["/src/main.tsx"] = buildMainShim(root);
      entry = "/src/main.tsx";
    } else {
      entry = detectSandpackEntry(result);
    }
  }
  result["/package.json"] = JSON.stringify({ ...JSON.parse(HARDCODED_PACKAGE_JSON), main: entry }, null, 2);
  // The react-ts template always boots from /index.tsx (its default renders "Hello
  // world"). customSetup.entry isn't reliably honored, so we OVERWRITE /index.tsx with
  // a shim that imports the real generated entry — guaranteeing our app mounts.
  // Guard on `entry in result`: never point the shim at a non-existent module.
  if (entry !== "/index.tsx" && entry in result) {
    // Classic bundler needs a RELATIVE specifier WITH extension ("./src/main.tsx"),
    // not an absolute "/src/main" — the latter fails to resolve from /index.tsx.
    result["/index.tsx"] = `import ".${entry}";\n`;
  }
  result["/vite.config.ts"] = HARDCODED_VITE_CONFIG;
  result["/tailwind.config.js"] = HARDCODED_TAILWIND_CONFIG;
  result["/postcss.config.js"] = HARDCODED_POSTCSS_CONFIG;
  result["/index.html"] = HARDCODED_INDEX_HTML;
  return result;
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
      if (isError && /shell|enoent|nodebox|node worker|initializing node|timestamp|vite\.config|dataclone|out of memory/i.test(text)) onCrash();
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

  // Server-side export: returns current session id
  getSessionId?: () => string;
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
  getSessionId,
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
  // Auto-recovery for nodebox shell crashes: remount SandpackProvider with a fresh key
  // (capped) instead of leaving the user stuck on Sandpack's red error overlay.
  const sandpackRetries = useRef(0);
  const [sandpackKey, setSandpackKey] = useState(0);
  const [sandpackError, setSandpackError] = useState(false);

  // ── Bulk-update safety ──────────────────────────────────────────────
  // A massive bulk update (e.g. 58 files recovered via polling) must NOT be
  // HMR-patched into a running Vite/nodebox instance — that triggers the
  // "vite.config.ts.timestamp" ENOENT / "Failed to get shell by ID" race.
  // We debounce the file payload (so writes settle) and force a COLD remount
  // (fresh key) whenever the file SET changes structurally, so Sandpack does a
  // clean boot instead of 50+ hot reloads in a split second. Pure content
  // edits to the same file set keep the same key (normal HMR for live editing).
  const [debouncedFiles, setDebouncedFiles] = useState<Record<string, string>>(() => sandpackFiles);
  const debounceTimer = useRef<ReturnType<typeof setTimeout>>();
  const prevPathsKey = useRef<string>(Object.keys(sandpackFiles).sort().join("|"));
  // Explicit bundler entry — without this the classic bundler falls back to the
  // template's default /index.tsx ("Hello world") instead of the generated app.
  const sandpackEntry = useMemo(() => detectSandpackEntry(debouncedFiles), [debouncedFiles]);

  useEffect(() => {
    if (!reactProject) return;
    if (debounceTimer.current) clearTimeout(debounceTimer.current);
    debounceTimer.current = setTimeout(() => {
      const pathsKey = Object.keys(sandpackFiles).sort().join("|");
      if (pathsKey !== prevPathsKey.current) {
        // Structural change (bulk add/remove) → cold remount, no HMR storm.
        prevPathsKey.current = pathsKey;
        sandpackRetries.current = 0;
        setSandpackError(false);
        setSandpackKey((k) => k + 1);
      }
      setDebouncedFiles(sandpackFiles);
    }, 450);
    return () => {
      if (debounceTimer.current) clearTimeout(debounceTimer.current);
    };
  }, [sandpackFiles, reactProject]);

  const handleSandpackCrash = useCallback(() => {
    if (sandpackRetries.current >= 2) {
      // Exhausted auto-retries → surface a manual recovery UI instead of bricking.
      setSandpackError(true);
      return;
    }
    sandpackRetries.current += 1;
    setSandpackKey((k) => k + 1);
  }, []);

  const restartPreview = useCallback(() => {
    sandpackRetries.current = 0;
    setSandpackError(false);
    setSandpackKey((k) => k + 1);
  }, []);

  // The Nodebox "vite.config.ts.timestamp-*.mjs" ENOENT race surfaces as an
  // UNCAUGHT PROMISE rejection from the runtime worker — it never reaches
  // Sandpack's listen() bus, so SandpackCrashGuard can't see it. Catch it at the
  // window level (narrowly pattern-matched to avoid false remounts) and route it
  // through the same capped auto-remount → Restart Preview recovery.
  useEffect(() => {
    if (!reactProject) return;
    const NODEBOX_RACE = /vite\.config.*timestamp|failed to stat file|nodebox|node worker|initializing node|out of memory|datacloneerror/i;
    const onRejection = (e: PromiseRejectionEvent) => {
      const reason = e.reason;
      const text = `${reason?.message ?? ""} ${String(reason ?? "")}`;
      if (NODEBOX_RACE.test(text)) {
        e.preventDefault();
        handleSandpackCrash();
      }
    };
    const onError = (e: ErrorEvent) => {
      const text = `${e.message ?? ""} ${e.error?.message ?? ""}`;
      if (NODEBOX_RACE.test(text)) handleSandpackCrash();
    };
    window.addEventListener("unhandledrejection", onRejection);
    window.addEventListener("error", onError);
    return () => {
      window.removeEventListener("unhandledrejection", onRejection);
      window.removeEventListener("error", onError);
    };
  }, [reactProject, handleSandpackCrash]);

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
      const sessionId = getSessionId?.() || "";
      if (!sessionId) {
        toast.error("Сессия не найдена — сгенерируйте проект сначала.");
        return;
      }
      const apiBase = import.meta.env.VITE_API_URL || "http://localhost:8080";
      const resp = await fetch(`${apiBase}/api/v1/project/export?session_id=${encodeURIComponent(sessionId)}`);
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
  }, [getSessionId]);

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
                  sandpackError ? (
                    <div className="flex h-full w-full flex-col items-center justify-center gap-4 bg-[hsl(240,6%,7%)] p-6 text-center">
                      <div className="grid h-12 w-12 place-items-center rounded-xl border border-amber-500/30 bg-amber-500/10">
                        <RotateCcw size={20} className="text-amber-400" />
                      </div>
                      <div className="space-y-1">
                        <p className="text-sm font-semibold text-foreground">Предпросмотр перезагружается</p>
                        <p className="max-w-[320px] text-xs text-muted-foreground">
                          Среда Sandpack не справилась с массовым обновлением файлов. Нажмите, чтобы перезапустить предпросмотр — код проекта в безопасности.
                        </p>
                      </div>
                      <button
                        onClick={restartPreview}
                        className="flex items-center gap-2 rounded-lg bg-gradient-primary px-4 py-2 text-xs font-medium text-primary-foreground shadow-glow transition hover:opacity-90"
                      >
                        <RotateCcw size={14} /> Перезапустить предпросмотр
                      </button>
                    </div>
                  ) : (
                    <SandpackProvider
                      key={sandpackKey}
                      template="react-ts"
                      files={debouncedFiles}
                      theme="dark"
                      customSetup={{ entry: sandpackEntry }}
                      options={{ externalResources: ["https://cdn.tailwindcss.com"] }}
                    >
                      <SandpackCrashGuard onCrash={handleSandpackCrash} />
                      <SandpackLivePreview
                        showNavigator={false}
                        style={{ height: "100%", width: "100%" }}
                      />
                    </SandpackProvider>
                  )
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
