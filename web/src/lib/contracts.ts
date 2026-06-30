import { z } from "zod";

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — API Contracts (Zod schemas)
//  Mirror Go structs 1:1. Any Go-side change MUST
//  be reflected here to keep FE/BE contract aligned.
//
//  Mapping:
//    GenerateProjectRequest   ↔ internal/application/dto/requests.go
//    GenerateProjectResponse  ↔ internal/application/dto/responses.go
//    AgentStatusResponse      ↔ internal/application/dto/responses.go
//    SSE event payloads       ↔ internal/transport/http/generate_handler_sse.go
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// ── Generation modes (in sync with Go application.GenerationMode) ──
export const GenerationModeSchema = z.enum(["agent", "code", "synthesis"]);
export type GenerationMode = z.infer<typeof GenerationModeSchema>;

// ── POST /api/v1/generate  |  /api/v1/generate/stream  (request body) ──
export const GenerateProjectRequestSchema = z.object({
  specification: z.string().min(1, "specification is required"),
  url: z.string().optional().default(""),
  language: z.string().optional().default(""),
  framework: z.string().optional().default(""),
  analyze_url: z.string().optional(),
  mode: GenerationModeSchema.optional(),
  generate_video: z.boolean().optional().default(false),
});
export type GenerateProjectRequest = z.infer<typeof GenerateProjectRequestSchema>;

// ── POST /api/v1/generate  (non-streaming response) ──
export const GenerateProjectResponseSchema = z.object({
  code: z.string(),
  explanation: z.string(),
  tokens_used: z.number().int().nonnegative(),
  dependencies: z.array(z.string()),
  model: z.string(),
});
export type GenerateProjectResponse = z.infer<typeof GenerateProjectResponseSchema>;

// ── GET /api/v1/agents/status ──
export const AgentInfoSchema = z.object({
  role: z.string(),
  model: z.string(),
  provider: z.enum(["Istok Core", "Replicate", "Local"]).or(z.string()),
  description: z.string(),
  thinking: z.boolean(),
  timeout_sec: z.number().int().nonnegative(),
});
export type AgentInfo = z.infer<typeof AgentInfoSchema>;

export const AgentStatusResponseSchema = z.object({
  agents: z.array(AgentInfoSchema),
  fsm_states: z.number().int().nonnegative(),
  event_buffer: z.number().int().nonnegative(),
  pipeline: z.array(z.string()),
});
export type AgentStatusResponse = z.infer<typeof AgentStatusResponseSchema>;

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  SSE event payloads (emitted from generate_handler_sse.go)
//  event kinds: "status" | "fsm" | "file" | "result_meta" | "done" | "error"
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

export const SSEStatusEventSchema = z.object({
  agent: z.string(),
  status: z.enum(["running", "completed", "error", "started"]).or(z.string()),
  state: z.string().optional().default(""),
  message: z.string().optional().default(""),
  progress: z.number().optional().default(0),
  timestamp: z.string().optional(),
});
export type SSEStatusEvent = z.infer<typeof SSEStatusEventSchema>;

export const SSEFSMEventSchema = z.object({
  agent: z.string().optional(),
  from: z.string().optional(),
  to: z.string().optional(),
  state: z.string().optional(),
  reason: z.string().optional(),
  message: z.string().optional(),
  timestamp: z.string().optional(),
});
export type SSEFSMEvent = z.infer<typeof SSEFSMEventSchema>;

export const SSEFileEventSchema = z.object({
  name: z.string(),
  content: z.string(),
});
export type SSEFileEvent = z.infer<typeof SSEFileEventSchema>;

export const SSEResultMetaSchema = z.object({
  file_count: z.number().int().nonnegative(),
  assets: z.record(z.string()).optional(),
  video: z.string().optional(),
  duration: z.string().optional(),
});
export type SSEResultMeta = z.infer<typeof SSEResultMetaSchema>;

export const SSEDoneEventSchema = z.object({
  message: z.string().optional(),
});
export type SSEDoneEvent = z.infer<typeof SSEDoneEventSchema>;

export const SSEErrorEventSchema = z.object({
  message: z.string(),
});
export type SSEErrorEvent = z.infer<typeof SSEErrorEventSchema>;

export const SSEThoughtEventSchema = z.object({
  agent: z.string(),
  tag: z.string(),   // "PLANNING" | "EXECUTION" | "VALIDATION"
  message: z.string(),
  timestamp: z.string().optional(),
});
export type SSEThoughtEvent = z.infer<typeof SSEThoughtEventSchema>;

export const SSEPostMortemEventSchema = z.object({
  report: z.string(),
  timestamp: z.string().optional(),
});
export type SSEPostMortemEvent = z.infer<typeof SSEPostMortemEventSchema>;

export const SSETelemetryEventSchema = z.object({
  agent: z.string(),
  line: z.string(),  // raw engineering log line e.g. "[LLM] model=... | tokens=..."
  timestamp: z.string().optional(),
});
export type SSETelemetryEvent = z.infer<typeof SSETelemetryEventSchema>;

// ── Canonical pipeline (must match backend application.CanonicalPipeline) ──
export const CANONICAL_PIPELINE = [
  "director",
  "researcher",
  "brain",
  "architect",
  "planner",
  "coder",
  "designer",
  "validator",
  "security",
  "tester",
  "ui_reviewer",
  "videographer",
] as const;
export type CanonicalAgentId = (typeof CANONICAL_PIPELINE)[number];

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  Layer 1 — Auth & DB DTOs (Dashboard / Projects / Profile)
//  Mirror Go structs 1:1. See API spec handed to the backend team.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// ── GET /api/v1/projects → { projects: ProjectSummary[] } ──
export const ProjectSummarySchema = z.object({
  id: z.string(),
  name: z.string(),
  description: z.string().optional().default(""),
  framework: z.string().optional().default(""),
  is_public: z.boolean().optional().default(false),
  slug: z.string().nullable().optional().default(null),
  thumbnail_url: z.string().nullable().optional().default(null),
  created_at: z.string(),
  updated_at: z.string(),
});
export type ProjectSummary = z.infer<typeof ProjectSummarySchema>;

export const ProjectListResponseSchema = z.object({
  projects: z.array(ProjectSummarySchema),
});
export type ProjectListResponse = z.infer<typeof ProjectListResponseSchema>;

// ── GET /api/v1/projects/:id → ProjectDetail ──
export const ProjectDetailSchema = ProjectSummarySchema.extend({
  prompt: z.string().optional().default(""),
  files: z.record(z.string()).optional().default({}),
});
export type ProjectDetail = z.infer<typeof ProjectDetailSchema>;

// ── GET /api/v1/user/profile → UserProfile ──
export const ProfileStatsSchema = z.object({
  total_projects: z.number().int().nonnegative().optional().default(0),
  published_projects: z.number().int().nonnegative().optional().default(0),
  total_generations: z.number().int().nonnegative().optional().default(0),
  days_active: z.number().int().nonnegative().optional().default(0),
  current_streak: z.number().int().nonnegative().optional().default(0),
});
export type ProfileStats = z.infer<typeof ProfileStatsSchema>;

export const UserProfileSchema = z.object({
  id: z.string(),
  email: z.string(),
  display_name: z.string().optional().default(""),
  username: z.string().nullable().optional().default(null),
  avatar_url: z.string().nullable().optional().default(null),
  bio: z.string().nullable().optional().default(null),
  location: z.string().nullable().optional().default(null),
  website: z.string().nullable().optional().default(null),
  created_at: z.string(),
  stats: ProfileStatsSchema.optional().default({}),
  // 371 cells (53 weeks × 7 days), each 0..3 contribution intensity. Optional.
  activity: z.array(z.number().int().min(0).max(3)).optional().default([]),
});
export type UserProfile = z.infer<typeof UserProfileSchema>;

// ── POST /api/v1/projects → ProjectDetail (persist a generated project) ──
export const CreateProjectRequestSchema = z.object({
  name: z.string().min(1),
  description: z.string().optional(),
  framework: z.string().optional(),
  prompt: z.string().optional(),
  is_public: z.boolean().optional(),
  files: z.record(z.string()),
});
export type CreateProjectRequest = z.infer<typeof CreateProjectRequestSchema>;

// ── PATCH /api/v1/projects/:id → ProjectSummary (partial update) ──
export const UpdateProjectRequestSchema = z.object({
  name: z.string().min(1).optional(),
  description: z.string().optional(),
  framework: z.string().optional(),
  is_public: z.boolean().optional(),
  folder_id: z.string().nullable().optional(),
  workspace_id: z.string().optional(),
  files: z.record(z.string()).optional(),
});
export type UpdateProjectRequest = z.infer<typeof UpdateProjectRequestSchema>;

// ── POST /api/v1/projects/:id/remix → ProjectSummary (clone) ──
export const RemixProjectRequestSchema = z.object({
  name: z.string().min(1).optional(),
  include_history: z.boolean().optional(),
});
export type RemixProjectRequest = z.infer<typeof RemixProjectRequestSchema>;

// ── GET /api/v1/folders → { folders: Folder[] } ──
export const FolderSchema = z.object({
  id: z.string(),
  name: z.string(),
  project_count: z.number().int().nonnegative().optional().default(0),
});
export type Folder = z.infer<typeof FolderSchema>;

export const FolderListResponseSchema = z.object({ folders: z.array(FolderSchema) });
export type FolderListResponse = z.infer<typeof FolderListResponseSchema>;

// ── GET /api/v1/workspaces → { workspaces: Workspace[] } ──
export const WorkspaceSchema = z.object({
  id: z.string(),
  name: z.string(),
  role: z.enum(["owner", "admin", "member"]).or(z.string()).optional().default("member"),
  is_personal: z.boolean().optional().default(false),
});
export type Workspace = z.infer<typeof WorkspaceSchema>;

export const WorkspaceListResponseSchema = z.object({ workspaces: z.array(WorkspaceSchema) });
export type WorkspaceListResponse = z.infer<typeof WorkspaceListResponseSchema>;

// ── Runtime helpers ──

/** Safe-parse with a fallback; logs mismatches in dev for contract drift detection. */
export function safeParseContract<S extends z.ZodTypeAny>(
  schema: S,
  data: unknown,
  label: string,
): { ok: true; data: z.infer<S> } | { ok: false; error: z.ZodError } {
  const result = schema.safeParse(data);
  if (!result.success) {
    // eslint-disable-next-line no-console
    console.warn(`[contract drift] ${label}:`, result.error.flatten());
    return { ok: false, error: result.error };
  }
  return { ok: true, data: result.data };
}
