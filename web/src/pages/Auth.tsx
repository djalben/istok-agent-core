import { useState } from "react";
import { useNavigate, Navigate } from "react-router-dom";
import { motion } from "framer-motion";
import { Mail, Lock, User, ArrowLeft, Loader2, Eye, EyeOff, Sparkles } from "lucide-react";
import { api } from "@/lib/api";
import { useAuth } from "@/hooks/useAuth";
import { toast } from "sonner";

const inputClass =
  "w-full h-11 pl-10 pr-4 rounded-lg bg-elevated border border-border text-sm text-foreground placeholder:text-muted-foreground/50 outline-none transition-all focus:border-primary/60 focus:ring-1 focus:ring-primary/30";

const Auth = () => {
  const navigate = useNavigate();
  const { user, loading: authLoading } = useAuth();
  const [isLogin, setIsLogin] = useState(true);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [loading, setLoading] = useState(false);
  const [showPassword, setShowPassword] = useState(false);

  // Already signed in → bounce to the dashboard.
  if (!authLoading && user) return <Navigate to="/" replace />;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      if (isLogin) {
        await api.login({ email, password });
        toast.success("Добро пожаловать!");
      } else {
        await api.signup({ email, password, display_name: displayName });
        toast.success("Регистрация успешна! Добро пожаловать!");
      }
      // Full reload re-initialises AuthProvider with the fresh token.
      window.location.href = "/";
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Произошла ошибка";
      toast.error(msg);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="relative min-h-screen overflow-hidden bg-background">
      {/* Ambient gradient glow */}
      <div className="pointer-events-none absolute -left-40 -top-40 h-96 w-96 rounded-full bg-primary/20 blur-[120px]" />
      <div className="pointer-events-none absolute -bottom-40 -right-40 h-96 w-96 rounded-full bg-fuchsia-500/15 blur-[120px]" />

      <div className="relative z-10 flex min-h-screen items-center justify-center p-4">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4, ease: [0.22, 1, 0.36, 1] }}
          className="w-full max-w-md"
        >
          <button
            onClick={() => navigate("/")}
            className="mb-8 flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground"
          >
            <ArrowLeft size={16} /> На главную
          </button>

          <div className="rounded-2xl border border-border bg-panel/80 p-8 shadow-glow backdrop-blur-xl">
            {/* Brand */}
            <div className="mb-7 flex flex-col items-center text-center">
              <div className="grid h-12 w-12 place-items-center rounded-xl bg-gradient-primary shadow-glow">
                <Sparkles className="h-5 w-5 text-primary-foreground" />
              </div>
              <h1 className="mt-4 text-2xl font-semibold tracking-tight text-foreground">
                {isLogin ? "С возвращением" : "Создайте аккаунт"}
              </h1>
              <p className="mt-1 text-sm text-muted-foreground">
                {isLogin
                  ? "Войдите, чтобы продолжить работу с Истоком"
                  : "Несколько секунд — и команда агентов в вашем распоряжении"}
              </p>
            </div>

            {/* Segmented tabs */}
            <div className="mb-6 grid grid-cols-2 gap-1 rounded-lg border border-border bg-elevated/60 p-1">
              {[
                { id: true, label: "Вход" },
                { id: false, label: "Регистрация" },
              ].map((tab) => (
                <button
                  key={tab.label}
                  type="button"
                  onClick={() => setIsLogin(tab.id)}
                  className={`relative h-8 rounded-md text-xs font-medium transition-colors ${
                    isLogin === tab.id ? "text-primary-foreground" : "text-muted-foreground hover:text-foreground"
                  }`}
                >
                  {isLogin === tab.id && (
                    <motion.span
                      layoutId="auth-tab"
                      className="absolute inset-0 rounded-md bg-gradient-primary shadow-glow"
                      transition={{ type: "spring", stiffness: 320, damping: 30 }}
                    />
                  )}
                  <span className="relative z-10">{tab.label}</span>
                </button>
              ))}
            </div>

            <form onSubmit={handleSubmit} className="space-y-4">
              {!isLogin && (
                <div className="space-y-1.5">
                  <label className="text-xs font-medium text-muted-foreground">Имя</label>
                  <div className="relative">
                    <User size={15} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground/60" />
                    <input
                      type="text"
                      value={displayName}
                      onChange={(e) => setDisplayName(e.target.value)}
                      placeholder="Ваше имя"
                      className={inputClass}
                    />
                  </div>
                </div>
              )}

              <div className="space-y-1.5">
                <label className="text-xs font-medium text-muted-foreground">Email</label>
                <div className="relative">
                  <Mail size={15} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground/60" />
                  <input
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    placeholder="you@example.com"
                    required
                    className={inputClass}
                  />
                </div>
              </div>

              <div className="space-y-1.5">
                <label className="text-xs font-medium text-muted-foreground">Пароль</label>
                <div className="relative">
                  <Lock size={15} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground/60" />
                  <input
                    type={showPassword ? "text" : "password"}
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    placeholder="Минимум 6 символов"
                    required
                    minLength={6}
                    className={`${inputClass} pr-11`}
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword((v) => !v)}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground/60 transition-colors hover:text-foreground"
                  >
                    {showPassword ? <EyeOff size={15} /> : <Eye size={15} />}
                  </button>
                </div>
              </div>

              <button
                type="submit"
                disabled={loading}
                className="flex h-11 w-full items-center justify-center gap-2 rounded-lg bg-gradient-primary text-sm font-medium text-primary-foreground shadow-glow transition-all hover:opacity-90 disabled:opacity-50"
              >
                {loading ? (
                  <Loader2 size={16} className="animate-spin" />
                ) : isLogin ? (
                  "Войти"
                ) : (
                  "Зарегистрироваться"
                )}
              </button>
            </form>

            <p className="mt-6 text-center text-xs text-muted-foreground">
              {isLogin ? "Нет аккаунта?" : "Уже есть аккаунт?"}{" "}
              <button
                onClick={() => setIsLogin((v) => !v)}
                className="font-medium text-primary transition-colors hover:underline"
              >
                {isLogin ? "Зарегистрируйтесь" : "Войдите"}
              </button>
            </p>
          </div>
        </motion.div>
      </div>
    </div>
  );
};

export default Auth;
