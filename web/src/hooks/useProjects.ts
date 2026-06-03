import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { toDisplayProject, type Project } from "@/lib/projectDisplay";
import type { UpdateProjectRequest, RemixProjectRequest } from "@/lib/contracts";

const PROJECTS_KEY = ["projects"] as const;

/** Fetch + map the authenticated user's projects (GET /api/v1/projects). */
export function useProjects() {
  return useQuery<Project[]>({
    queryKey: PROJECTS_KEY,
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
    onSuccess: () => qc.invalidateQueries({ queryKey: PROJECTS_KEY }),
  });
}

/**
 * PATCH a project (rename / move / transfer) with optimistic cache update.
 * Visible fields (name/description) update instantly; rolls back on error.
 */
export function useUpdateProject() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, patch }: { id: string; patch: UpdateProjectRequest }) =>
      api.updateProject(id, patch),
    onMutate: async ({ id, patch }) => {
      await qc.cancelQueries({ queryKey: PROJECTS_KEY });
      const previous = qc.getQueryData<Project[]>(PROJECTS_KEY);
      if (previous) {
        qc.setQueryData<Project[]>(
          PROJECTS_KEY,
          previous.map((p) =>
            p.id === id
              ? {
                  ...p,
                  ...(patch.name !== undefined ? { name: patch.name } : {}),
                  ...(patch.description !== undefined ? { description: patch.description } : {}),
                  ...(patch.framework !== undefined ? { framework: patch.framework || "—" } : {}),
                }
              : p,
          ),
        );
      }
      return { previous };
    },
    onError: (_err, _vars, ctx) => {
      if (ctx?.previous) qc.setQueryData(PROJECTS_KEY, ctx.previous);
    },
    onSettled: () => qc.invalidateQueries({ queryKey: PROJECTS_KEY }),
  });
}

/** Clone a project (POST /projects/:id/remix), then refresh the list. */
export function useRemixProject() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: RemixProjectRequest }) =>
      api.remixProject(id, payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: PROJECTS_KEY }),
  });
}
