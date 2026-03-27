import { createContext, useContext, useState, useEffect, useCallback, type ReactNode } from "react";
import { supabase } from "@/integrations/supabase/client";
import { useAuth } from "@/hooks/useAuth";

interface CreditsContextType {
  credits: number | null;
  refreshCredits: () => Promise<void>;
  setCredits: (credits: number) => void;
}

const CreditsContext = createContext<CreditsContextType>({
  credits: null,
  refreshCredits: async () => {},
  setCredits: () => {},
});

export const useCredits = () => useContext(CreditsContext);

export const CreditsProvider = ({ children }: { children: ReactNode }) => {
  const { user } = useAuth();
  const [credits, setCreditsState] = useState<number | null>(null);

  const refreshCredits = useCallback(async () => {
    if (!user) return;
    const { data } = await supabase
      .from("profiles")
      .select("credits")
      .eq("id", user.id)
      .single();
    if (data) setCreditsState(data.credits);
  }, [user]);

  useEffect(() => {
    refreshCredits();
  }, [refreshCredits]);

  const setCredits = (val: number) => setCreditsState(val);

  return (
    <CreditsContext.Provider value={{ credits, refreshCredits, setCredits }}>
      {children}
    </CreditsContext.Provider>
  );
};
