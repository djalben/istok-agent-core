// import { supabase } from "@/integrations/supabase/client"; // Не используется - переход на Go Auth
import { loadProjects, type SavedProject } from "./projectStorage";

export interface CloudProject {
  id: string;
  user_id: string;
  prompt: string;
  code: string;
  is_public: boolean;
  slug: string | null;
  created_at: string;
  updated_at: string;
}

export async function syncLocalToCloud(userId: string): Promise<number> {
  return 0;
}

export async function loadCloudProjects(): Promise<CloudProject[]> {
  return [];
}

export async function saveCloudProject(
  userId: string,
  prompt: string,
  code: string
): Promise<CloudProject | null> {
  return null;
}

export async function deleteCloudProject(id: string): Promise<boolean> {
  return false;
}

export async function publishProject(id: string): Promise<string | null> {
  return null;
}

export async function getProjectByPrompt(userId: string, prompt: string): Promise<CloudProject | null> {
  return null;
}
