import { useState, useEffect, createContext, useContext, useCallback } from "react";
import { api, type User } from "@/lib/api";

interface AuthContextType {
  user: User | null;
  loading: boolean;
  signOut: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType>({
  user: null,
  loading: true,
  signOut: async () => {},
});

export const useAuth = () => useContext(AuthContext);

export const AuthProvider = ({ children }: { children: React.ReactNode }) => {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Проверяем авторизацию при загрузке
    const checkAuth = async () => {
      try {
        if (api.isAuthenticated()) {
          // Сначала пробуем получить из localStorage
          const cachedUser = api.getCurrentUser();
          if (cachedUser) {
            setUser(cachedUser);
          }
          
          // Затем проверяем на сервере
          try {
            const serverUser = await api.getMe();
            setUser(serverUser);
          } catch (error) {
            // Токен невалиден, очищаем
            api.logout();
            setUser(null);
          }
        }
      } catch (error) {
        console.error("Auth check error:", error);
      } finally {
        setLoading(false);
      }
    };

    checkAuth();
  }, []);

  const signOut = useCallback(async () => {
    api.logout();
    setUser(null);
    window.location.href = "/";
  }, []);

  return (
    <AuthContext.Provider value={{ user, loading, signOut }}>
      {children}
    </AuthContext.Provider>
  );
};
