import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { motion } from "framer-motion";
import { Mail, Lock, User, ArrowLeft, Loader2, Eye, EyeOff } from "lucide-react";
import { api } from "@/lib/api";
import { toast } from "sonner";

const Auth = () => {
  const navigate = useNavigate();
  const [isLogin, setIsLogin] = useState(true);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [loading, setLoading] = useState(false);
  const [showPassword, setShowPassword] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);

    try {
      if (isLogin) {
        await api.login({ email, password });
        toast.success("Добро пожаловать!");
        // Перезагружаем страницу для обновления состояния авторизации
        window.location.href = "/";
      } else {
        await api.signup({ 
          email, 
          password, 
          display_name: displayName 
        });
        toast.success("Регистрация успешна! Добро пожаловать!");
        // Перезагружаем страницу для обновления состояния авторизации
        window.location.href = "/";
      }
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Произошла ошибка";
      toast.error(msg);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-background flex items-center justify-center p-4">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="w-full max-w-md"
      >
        <button
          onClick={() => navigate("/")}
          className="flex items-center gap-1.5 text-muted-foreground hover:text-foreground transition-colors text-sm mb-8"
        >
          <ArrowLeft size={16} />
          На главную
        </button>

        <div className="glass rounded-2xl p-8 border border-border/30">
          <div className="text-center mb-8">
            <h1 className="text-2xl font-bold text-foreground mb-2">
              {isLogin ? "Вход в Исток" : "Регистрация"}
            </h1>
            <p className="text-sm text-muted-foreground">
              {isLogin
                ? "Войдите, чтобы сохранять проекты в облаке"
                : "Создайте аккаунт для начала работы"}
            </p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-4">
            {!isLogin && (
              <div className="space-y-1.5">
                <label className="text-xs text-muted-foreground font-medium">Имя</label>
                <div className="relative">
                  <User size={15} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground/50" />
                  <input
                    type="text"
                    value={displayName}
                    onChange={(e) => setDisplayName(e.target.value)}
                    placeholder="Ваше имя"
                    className="w-full h-11 pl-10 pr-4 rounded-xl bg-secondary/50 border border-border/30 text-sm text-foreground placeholder:text-muted-foreground/40 outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/20 transition-all"
                  />
                </div>
              </div>
            )}

            <div className="space-y-1.5">
              <label className="text-xs text-muted-foreground font-medium">Email</label>
              <div className="relative">
                <Mail size={15} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground/50" />
                <input
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="you@example.com"
                  required
                  className="w-full h-11 pl-10 pr-4 rounded-xl bg-secondary/50 border border-border/30 text-sm text-foreground placeholder:text-muted-foreground/40 outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/20 transition-all"
                />
              </div>
            </div>

            <div className="space-y-1.5">
              <label className="text-xs text-muted-foreground font-medium">Пароль</label>
              <div className="relative">
                <Lock size={15} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground/50" />
                <input
                  type={showPassword ? "text" : "password"}
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="Минимум 6 символов"
                  required
                  minLength={6}
                  className="w-full h-11 pl-10 pr-11 rounded-xl bg-secondary/50 border border-border/30 text-sm text-foreground placeholder:text-muted-foreground/40 outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/20 transition-all"
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground/50 hover:text-muted-foreground transition-colors"
                >
                  {showPassword ? <EyeOff size={15} /> : <Eye size={15} />}
                </button>
              </div>
            </div>

            <button
              type="submit"
              disabled={loading}
              className="w-full h-11 rounded-xl btn-gradient text-primary-foreground font-medium text-sm flex items-center justify-center gap-2 disabled:opacity-50 transition-all"
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

          <div className="mt-6 text-center">
            <button
              onClick={() => setIsLogin(!isLogin)}
              className="text-sm text-muted-foreground hover:text-primary transition-colors"
            >
              {isLogin ? "Нет аккаунта? Зарегистрируйтесь" : "Уже есть аккаунт? Войдите"}
            </button>
          </div>
        </div>
      </motion.div>
    </div>
  );
};

export default Auth;
