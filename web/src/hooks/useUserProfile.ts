import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import type { UserProfile } from "@/lib/contracts";

/** Fetch the authenticated user's profile + aggregated stats (GET /api/v1/user/profile). */
export function useUserProfile() {
  return useQuery<UserProfile>({
    queryKey: ["user-profile"],
    queryFn: () => api.getUserProfile(),
    retry: 1,
  });
}
