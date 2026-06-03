import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import type { Folder, Workspace } from "@/lib/contracts";

/** Folders for the "move to folder" selector (GET /api/v1/folders). */
export function useFolders() {
  return useQuery<Folder[]>({
    queryKey: ["folders"],
    queryFn: () => api.getFolders(),
    staleTime: 60_000,
    retry: false,
  });
}

/** Workspaces for the "transfer" selector (GET /api/v1/workspaces). */
export function useWorkspaces() {
  return useQuery<Workspace[]>({
    queryKey: ["workspaces"],
    queryFn: () => api.getWorkspaces(),
    staleTime: 60_000,
    retry: false,
  });
}
