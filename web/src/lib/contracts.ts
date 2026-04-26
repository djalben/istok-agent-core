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
  provider: z.enum(["Anthropic Direct", "Replicate", "Local"]).or(z.string()),
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

// ── Runtime helpers ──

/** Safe-parse with a fallback; logs mismatches in dev for contract drift detection. */
export function safeParseContract<T>(
  schema: z.ZodType<T>,
  data: unknown,
  label: string,
): { ok: true; data: T } | { ok: false; error: z.ZodError } {
  const result = schema.safeParse(data);
  if (!result.success) {
    // eslint-disable-next-line no-console
    console.warn(`[contract drift] ${label}:`, result.error.flatten());
    return { ok: false, error: result.error };
  }
  return { ok: true, data: result.data };
}
