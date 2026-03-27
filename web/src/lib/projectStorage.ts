export interface SavedProject {
  id: string;
  prompt: string;
  code: string;
  createdAt: string;
}

const STORAGE_KEY = "istok_projects";

export function loadProjects(): SavedProject[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    return raw ? JSON.parse(raw) : [];
  } catch {
    return [];
  }
}

export function saveProject(prompt: string, code: string): SavedProject {
  const projects = loadProjects();
  const existing = projects.find((p) => p.prompt === prompt);
  if (existing) {
    existing.code = code;
    existing.createdAt = new Date().toISOString();
    localStorage.setItem(STORAGE_KEY, JSON.stringify(projects));
    return existing;
  }
  const project: SavedProject = {
    id: Date.now().toString(),
    prompt,
    code,
    createdAt: new Date().toISOString(),
  };
  const updated = [project, ...projects].slice(0, 50);
  localStorage.setItem(STORAGE_KEY, JSON.stringify(updated));
  return project;
}

export function deleteProject(id: string): void {
  const projects = loadProjects().filter((p) => p.id !== id);
  localStorage.setItem(STORAGE_KEY, JSON.stringify(projects));
}
