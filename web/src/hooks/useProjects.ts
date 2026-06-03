import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { toDisplayProject, type Project } from "@/lib/projectDisplay";

/** Fetch + map the authenticated user's projects (GET /api/v1/projects). */
export function useProjects() {
  return useQuery<Project[]>({
    queryKey: ["projects"],
    queryFn: async () => {
      const list = await api.getProjects();
      return list.map(toDisplayProject);
    },
  });
}

/** Delete a project, then refresh the cached list. */
export function useDeleteProject() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.deleteProject(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["projects"] }),
  });
}
