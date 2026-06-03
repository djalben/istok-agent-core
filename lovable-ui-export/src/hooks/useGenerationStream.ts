import { useCallback, useEffect, useRef, useState } from "react";
import { initialAgents, type Agent } from "@/lib/mockData";

export function useGenerationStream() {
  const [agents, setAgents] = useState<Agent[]>(initialAgents);
  const [isStreaming, setIsStreaming] = useState(false);
  const timerRef = useRef<number | null>(null);

  const reset = useCallback(() => {
    setAgents(initialAgents);
  }, []);

  const start = useCallback(() => {
    setIsStreaming(true);
    setAgents((prev) => prev.map((a) => ({ ...a, status: "thinking", task: "Queued…" })));

    const sequence: Array<Partial<Agent> & { id: string }> = [
      { id: "director", status: "working", task: "Routing tasks…" },
      { id: "director", status: "done", task: "Plan dispatched" },
      { id: "researcher", status: "working", task: "Crawling references…" },
      { id: "researcher", status: "done", task: "Indexed 8 sources" },
      { id: "architect", status: "working", task: "Composing routes…" },
      { id: "architect", status: "done", task: "12 files scaffolded" },
      { id: "coder", status: "working", task: "Writing Hero.tsx…" },
      { id: "designer", status: "working", task: "Applying tokens…" },
      { id: "coder", status: "done", task: "Components ready" },
      { id: "designer", status: "done", task: "Theme applied" },
    ];

    let i = 0;
    const tick = () => {
      if (i >= sequence.length) {
        setIsStreaming(false);
        return;
      }
      const step = sequence[i++];
      setAgents((prev) => prev.map((a) => (a.id === step.id ? { ...a, ...step } as Agent : a)));
      timerRef.current = window.setTimeout(tick, 700 + Math.random() * 600);
    };
    tick();
  }, []);

  useEffect(() => {
    return () => {
      if (timerRef.current) window.clearTimeout(timerRef.current);
    };
  }, []);

  return { agents, isStreaming, start, reset };
}
