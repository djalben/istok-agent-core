/**
 * ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 *  ИСТОК АГЕНТ — Unified SSE / Agent Output Parsers
 *  Объединяет: extractMessage, safeContent, safeContentClean,
 *  stripThinking, detectAndUnpackProject в один эффективный модуль.
 * ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 */

/** Recognized "content-bearing" fields agent payloads may carry. */
const CONTENT_KEYS = [
  "text",
  "content",
  "reasoning_content",
  "thinking",
  "message",
  "description",
  "output",
] as const;

/** Regex matching Claude 3.7 / GPT-style `<thinking>...</thinking>` chain-of-thought blocks. */
const THINKING_RE = /<thinking>[\s\S]*?<\/thinking>/gi;

/** File extensions recognized as "project files" inside a JSON dump. */
const FILE_EXT_RE = /\.(html|tsx|ts|jsx|js|css|md|json)$/i;

/**
 * Single canonical parser: extracts a display string from any LLM/SSE payload.
 *
 * Replaces:
 *   - `extractMessage` (api.ts)
 *   - `safeContent` (Workspace.tsx)
 *   - `safeContentClean` (Workspace.tsx) — set `stripThoughts=true`
 *   - `stripThinking` (Workspace.tsx) — pass already-string + `stripThoughts=true`
 *
 * @param raw       Any JSON-deserialized agent payload (string, number, object, array).
 * @param stripThoughts  If true, removes `<thinking>...</thinking>` blocks.
 * @returns Trimmed display string (never undefined).
 */
export function parseAgentText(raw: unknown, stripThoughts = true): string {
  const result = walk(raw);
  return stripThoughts ? result.replace(THINKING_RE, "").trim() : result;
}

function walk(raw: unknown): string {
  if (raw == null) return "";
  if (typeof raw === "string") return raw;
  if (typeof raw === "number" || typeof raw === "boolean") return String(raw);

  if (Array.isArray(raw)) {
    return raw.map(walk).filter(Boolean).join("\n");
  }

  if (typeof raw === "object") {
    const obj = raw as Record<string, unknown>;
    for (const key of CONTENT_KEYS) {
      const candidate = obj[key];
      if (candidate == null) continue;
      if (typeof candidate === "object") return walk(candidate);
      return String(candidate);
    }
    // No known key — return JSON as last resort
    try {
      return JSON.stringify(raw);
    } catch {
      return String(raw);
    }
  }

  return String(raw);
}

/**
 * Detects whether a content string is a JSON dump of project files
 * (keys ending in `.html` / `.tsx` / `.css` etc.) and returns the parsed
 * `Record<filename, content>` map. Returns `null` if not a project dump.
 */
export function detectAndUnpackProject(content: string): Record<string, string> | null {
  const trimmed = typeof content === "string" ? content.trim() : "";
  if (!trimmed.startsWith("{")) return null;
  try {
    const parsed = JSON.parse(trimmed) as Record<string, unknown>;
    const fileKeys = Object.keys(parsed).filter((k) => FILE_EXT_RE.test(k));
    if (fileKeys.length === 0) return null;
    const files: Record<string, string> = {};
    for (const k of fileKeys) {
      files[k] = parseAgentText(parsed[k], false);
    }
    return files;
  } catch {
    return null;
  }
}

/**
 * Removes ```lang\n ... \n``` markdown code fences if the content is a single
 * fenced block. Handles both multi-line and inline fences, with or without
 * language tag (html/css/js/ts/tsx/jsx/ etc.).
 *
 * Canonical implementation — do NOT duplicate in UI components.
 * Replaces legacy copies previously living in WorkspacePreview.tsx.
 */
export function stripMarkdownFences(input: string): string {
  if (typeof input !== "string") return "";
  const trimmed = input.trim();
  // Multi-line fenced block with optional lang tag
  const fenced = trimmed.match(
    /^```(?:html|css|javascript|js|typescript|ts|jsx|tsx|json|md|[a-zA-Z0-9_+-]*)?\s*\n([\s\S]*?)```\s*$/i,
  );
  if (fenced) return fenced[1].trim();
  // Inline single-line fence (no newline before content)
  const inline = trimmed.match(/^```(?:[a-zA-Z0-9_+-]+)?\s*([\s\S]*?)```\s*$/);
  if (inline) return inline[1].trim();
  return trimmed;
}

/**
 * Parse code from DB into files structure. Canonical `codeToFiles` —
 * moved out of WorkspacePreview.tsx to this module so it lives next to
 * its cleanup helper stripMarkdownFences.
 */
export function codeToFiles(code: string): Record<string, string> {
  try {
    const parsed = JSON.parse(code) as unknown;
    if (typeof parsed === "object" && parsed !== null && !Array.isArray(parsed)) {
      const cleaned: Record<string, string> = {};
      for (const [k, v] of Object.entries(parsed as Record<string, unknown>)) {
        cleaned[k] = stripMarkdownFences(String(v));
      }
      return cleaned;
    }
  } catch {
    // Not JSON — legacy single-file project
  }
  return { "index.html": stripMarkdownFences(code) };
}

/** Flatten multi-file project to single code string for DB storage. */
export function filesToCode(files: Record<string, string>): string {
  return JSON.stringify(files);
}
