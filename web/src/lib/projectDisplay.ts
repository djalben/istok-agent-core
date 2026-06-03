// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  Project display model + presentation helpers
//  Maps backend ProjectSummary → the shape the dashboard cards render.
//  (gradient & relative time are derived client-side, not stored.)
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
import type { ProjectSummary } from "@/lib/contracts";

/** Display shape consumed by ProjectsGrid / ProjectsView / ProjectCardMenu. */
export interface Project {
  id: string;
  name: string;
  description: string;
  framework: string;
  updatedAt: string;
  updatedAtMs: number;
  gradient: string;
}

const GRADIENTS = [
  "bg-gradient-to-br from-violet-500 to-fuchsia-600",
  "bg-gradient-to-br from-cyan-400 to-blue-600",
  "bg-gradient-to-br from-amber-400 to-orange-600",
  "bg-gradient-to-br from-emerald-400 to-teal-600",
  "bg-gradient-to-br from-rose-400 to-red-600",
  "bg-gradient-to-br from-indigo-400 to-violet-600",
  "bg-gradient-to-br from-pink-500 to-rose-600",
  "bg-gradient-to-br from-sky-400 to-indigo-600",
];

/** Deterministic gradient from a project id (stable across reloads). */
export function gradientForId(id: string): string {
  let hash = 0;
  for (let i = 0; i < id.length; i++) hash = (hash * 31 + id.charCodeAt(i)) >>> 0;
  return GRADIENTS[hash % GRADIENTS.length];
}

/** Russian relative time from an ISO timestamp. */
export function relativeTime(iso: string): string {
  const then = new Date(iso).getTime();
  if (Number.isNaN(then)) return "";
  const diff = Date.now() - then;
  const min = Math.floor(diff / 60_000);
  const hr = Math.floor(diff / 3_600_000);
  const day = Math.floor(diff / 86_400_000);
  if (min < 1) return "только что";
  if (min < 60) return `${min} мин назад`;
  if (hr < 24) return `${hr} ч назад`;
  if (day === 1) return "вчера";
  if (day < 7) return `${day} дн назад`;
  if (day < 30) return `${Math.floor(day / 7)} нед назад`;
  if (day < 365) return `${Math.floor(day / 30)} мес назад`;
  return `${Math.floor(day / 365)} г назад`;
}

export function toDisplayProject(p: ProjectSummary): Project {
  return {
    id: p.id,
    name: p.name,
    description: p.description,
    framework: p.framework || "—",
    updatedAt: relativeTime(p.updated_at),
    updatedAtMs: new Date(p.updated_at).getTime() || 0,
    gradient: gradientForId(p.id),
  };
}
