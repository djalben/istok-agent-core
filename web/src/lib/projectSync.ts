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
  const localProjects = loadProjects();
  if (localProjects.length === 0) return 0;

  let synced = 0;
  for (const project of localProjects) {
    const { error } = await supabase.from("projects").insert({
      user_id: userId,
      prompt: project.prompt,
      code: project.code,
    });
    if (!error) synced++;
  }

  // Clear localStorage after sync
  if (synced > 0) {
    localStorage.removeItem("istok_projects");
  }

  return synced;
}

export async function loadCloudProjects(): Promise<CloudProject[]> {
  const { data, error } = await supabase
    .from("projects")
    .select("*")
    .order("updated_at", { ascending: false })
    .limit(50);

  if (error) {
    console.error("Failed to load cloud projects:", error);
    return [];
  }
  return data || [];
}

export async function saveCloudProject(
  userId: string,
  prompt: string,
  code: string
): Promise<CloudProject | null> {
  // Check if project with same prompt exists
  const { data: existing } = await supabase
    .from("projects")
    .select("id")
    .eq("user_id", userId)
    .eq("prompt", prompt)
    .limit(1);

  if (existing && existing.length > 0) {
    const { data, error } = await supabase
      .from("projects")
      .update({ code, updated_at: new Date().toISOString() })
      .eq("id", existing[0].id)
      .select()
      .single();
    if (error) return null;
    return data;
  }

  const { data, error } = await supabase
    .from("projects")
    .insert({ user_id: userId, prompt, code })
    .select()
    .single();
  if (error) return null;
  return data;
}

export async function deleteCloudProject(id: string): Promise<boolean> {
  const { error } = await supabase.from("projects").delete().eq("id", id);
  return !error;
}

export async function publishProject(id: string): Promise<string | null> {
  const { data, error } = await supabase
    .from("projects")
    .update({ is_public: true })
    .eq("id", id)
    .select("slug")
    .single();

  if (error || !data) return null;
  return data.slug;
}

export async function getProjectByPrompt(userId: string, prompt: string): Promise<CloudProject | null> {
  const { data } = await supabase
    .from("projects")
    .select("*")
    .eq("user_id", userId)
    .eq("prompt", prompt)
    .limit(1)
    .maybeSingle();
  return data;
}
