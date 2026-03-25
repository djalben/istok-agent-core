// API клиент для взаимодействия с бэкендом

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export interface GenerateRequest {
  specification: string;
  language?: string;
  framework?: string;
  analyze_url?: string;
}

export interface GenerateResponse {
  code: string;
  explanation: string;
  tokens_used: number;
  dependencies: string[];
  model: string;
}

export interface AgentStats {
  agent_id: string;
  name: string;
  status: string;
  token_balance: number;
  total_tasks: number;
  success_rate: number;
  knowledge_nodes: number;
  learning_confidence: number;
  average_tokens_per_task: number;
}

export class APIClient {
  private baseURL: string;

  constructor(baseURL: string = API_URL) {
    this.baseURL = baseURL;
  }

  async generateProject(request: GenerateRequest): Promise<GenerateResponse> {
    const response = await fetch(`${this.baseURL}/api/v1/generate`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Unknown error' }));
      throw new Error(error.error || `HTTP error! status: ${response.status}`);
    }

    return response.json();
  }

  async getStats(): Promise<AgentStats> {
    const response = await fetch(`${this.baseURL}/api/v1/stats`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    return response.json();
  }

  async healthCheck(): Promise<{ status: string; uptime: string }> {
    const response = await fetch(`${this.baseURL}/api/v1/health`, {
      method: 'GET',
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    return response.json();
  }
}

export const apiClient = new APIClient();
