// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  Builder IDE types + adapters (Lovable graft)
//  Bridges our real SSE data (useGeneration) to the
//  Lovable visual components (Agent[], FileNode tree).
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
import { AGENT_PIPELINE, type AgentMilestone, type AgentPipelineId } from "@/hooks/useGeneration";

export type AgentStatus = "idle" | "thinking" | "working" | "done" | "error";

export interface Agent {
  id: string;
  name: string;
  role: string;
  status: AgentStatus;
  task: string;
}

export interface FileNode {
  name: string;
  path: string;
  type: "file" | "folder";
  language?: string;
  content?: string;
  children?: FileNode[];
}

/** Russian display metadata for each backend agent. */
const AGENT_META: Record<AgentPipelineId, { name: string; role: string; idle: string }> = {
  director: { name: "Директор", role: "Управляет сборкой", idle: "Ожидает запрос" },
  researcher: { name: "Исследователь", role: "Собирает контекст и источники", idle: "Готов к поиску" },
  brain: { name: "Аналитик", role: "Анализирует требования", idle: "Ожидает план" },
  architect: { name: "Архитектор", role: "Проектирует структуру файлов", idle: "Ожидает анализ" },
  planner: { name: "Планировщик", role: "Разбивает задачи на шаги", idle: "Ожидает архитектуру" },
  coder: { name: "Кодер", role: "Пишет компоненты и логику", idle: "Ожидает план" },
  designer: { name: "Дизайнер", role: "Применяет дизайн-токены", idle: "Ожидает макет" },
  validator: { name: "Валидатор", role: "Проверяет целостность кода", idle: "Ожидает код" },
  security: { name: "Безопасность", role: "Аудит уязвимостей", idle: "Ожидает код" },
  tester: { name: "Тестировщик", role: "Прогоняет проверки", idle: "Ожидает сборку" },
  ui_reviewer: { name: "UI-ревьюер", role: "Оценивает интерфейс", idle: "Ожидает превью" },
  videographer: { name: "Видеограф", role: "Готовит медиа-ассеты", idle: "Ожидает дизайн" },
};

/** Backend milestone status → Lovable agent status. */
function mapStatus(status: AgentMilestone["status"]): AgentStatus {
  if (status === "running") return "working";
  if (status === "completed") return "done";
  if (status === "error") return "error";
  return "idle";
}

/**
 * Converts our live milestones + activeAgent into the Lovable Agent[] shape
 * for the "Активность агентов" pulse. Only shows agents that have activity
 * (a milestone) plus the currently active one, preserving pipeline order.
 */
export function milestonesToAgents(
  milestones: AgentMilestone[],
  activeAgent: AgentPipelineId | null,
): Agent[] {
  const byAgent = new Map<string, AgentMilestone>();
  for (const m of milestones) {
    const key = m.agent.toLowerCase().replace(/\s+/g, "_");
    byAgent.set(key, m);
  }

  const agents: Agent[] = [];
  for (const id of AGENT_PIPELINE) {
    const m = byAgent.get(id);
    const meta = AGENT_META[id];
    const isActive = activeAgent === id;
    // Skip agents with no activity and not active — keeps the list focused.
    if (!m && !isActive) continue;
    agents.push({
      id,
      name: meta.name,
      role: meta.role,
      status: isActive ? "working" : m ? mapStatus(m.status) : "idle",
      task: m?.message || meta.idle,
    });
  }
  return agents;
}

const EXT_LANG: Record<string, string> = {
  tsx: "tsx", ts: "ts", jsx: "jsx", js: "js",
  css: "css", json: "json", md: "md", html: "html",
};

/**
 * Converts a flat { path: content } map into a nested FileNode tree
 * for the explorer. Folders are inferred from path segments.
 */
export function filesToTree(files: Record<string, string>): FileNode[] {
  const root: FileNode[] = [];
  const folderMap = new Map<string, FileNode>();

  const sortedPaths = Object.keys(files).sort();
  for (const rawPath of sortedPaths) {
    const path = rawPath.replace(/^\//, "");
    const parts = path.split("/");
    let currentLevel = root;
    let accum = "";

    for (let i = 0; i < parts.length; i++) {
      const part = parts[i];
      accum = accum ? `${accum}/${part}` : part;
      const isLeaf = i === parts.length - 1;

      if (isLeaf) {
        const ext = part.split(".").pop() ?? "";
        currentLevel.push({
          name: part,
          path: rawPath,
          type: "file",
          language: EXT_LANG[ext] ?? ext,
          content: files[rawPath],
        });
      } else {
        let folder = folderMap.get(accum);
        if (!folder) {
          folder = { name: part, path: accum, type: "folder", children: [] };
          folderMap.set(accum, folder);
          currentLevel.push(folder);
        }
        currentLevel = folder.children!;
      }
    }
  }
  return root;
}

export function flattenFiles(nodes: FileNode[]): FileNode[] {
  const out: FileNode[] = [];
  const walk = (n: FileNode) => {
    if (n.type === "file") out.push(n);
    n.children?.forEach(walk);
  };
  nodes.forEach(walk);
  return out;
}
