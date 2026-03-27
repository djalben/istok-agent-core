import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { motion } from "framer-motion";
import {
  ArrowLeft,
  Mail,
  Calendar,
  FolderOpen,
  Globe,
  Lock,
  Trash2,
  Check,
  Loader2,
  Smile,
  Cat,
  Dog,
  Bird,
  Fish,
  Bug,
  Flower2,
  Star,
  Zap,
  Heart,
  Ghost,
  Flame,
  Rocket,
} from "lucide-react";
import { useAuth } from "@/hooks/useAuth";
import { useLanguage } from "@/hooks/useLanguage";
// import { supabase } from "@/integrations/supabase/client"; // Не используется - переход на Go Auth
import { toast } from "sonner";
import HeaderBar from "@/components/HeaderBar";

const AVATAR_ICONS = [
  { id: "smile", icon: Smile, color: "text-yellow-400" },
  { id: "cat", icon: Cat, color: "text-orange-400" },
  { id: "dog", icon: Dog, color: "text-amber-500" },
  { id: "bird", icon: Bird, color: "text-sky-400" },
  { id: "fish", icon: Fish, color: "text-cyan-400" },
  { id: "bug", icon: Bug, color: "text-green-400" },
  { id: "flower", icon: Flower2, color: "text-pink-400" },
  { id: "star", icon: Star, color: "text-yellow-300" },
  { id: "zap", icon: Zap, color: "text-amber-300" },
  { id: "heart", icon: Heart, color: "text-red-400" },
  { id: "ghost", icon: Ghost, color: "text-purple-400" },
  { id: "flame", icon: Flame, color: "text-orange-500" },
  { id: "rocket", icon: Rocket, color: "text-blue-400" },
];

const Settings = () => {
  const navigate = useNavigate();
  const { user } = useAuth();
  const { t } = useLanguage();

  const [displayName, setDisplayName] = useState("");
  const [selectedAvatar, setSelectedAvatar] = useState("smile");
  const [saving, setSaving] = useState(false);
  const [totalProjects, setTotalProjects] = useState(0);
  const [publishedProjects, setPublishedProjects] = useState(0);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [passwordResetSent, setPasswordResetSent] = useState(false);

  useEffect(() => {
    if (!user) return;
    (async () => {
      const { data } = await supabase
        .from("profiles")
        .select("display_name, avatar_url")
        .eq("id", user.id)
        .single();
      if (data) {
        setDisplayName(data.display_name || "");
        setSelectedAvatar(data.avatar_url || "smile");
      }
      const { count: total } = await supabase
        .from("projects")
        .select("*", { count: "exact", head: true })
        .eq("user_id", user.id);
      setTotalProjects(total || 0);
      const { count: published } = await supabase
        .from("projects")
        .select("*", { count: "exact", head: true })
        .eq("user_id", user.id)
        .eq("is_public", true);
      setPublishedProjects(published || 0);
    })();
  }, [user]);

  const handleSave = async () => {
    if (!user) return;
    setSaving(true);
    const { error } = await supabase
      .from("profiles")
      .update({ display_name: displayName, avatar_url: selectedAvatar, updated_at: new Date().toISOString() })
      .eq("id", user.id);
    setSaving(false);
    if (error) {
      toast.error(t("settingsProfileError"));
    } else {
      toast.success(t("settingsProfileSaved"));
    }
  };

  const handlePasswordReset = async () => {
    if (!user?.email) return;
    // TODO: Реализовать сброс пароля через Go API
    toast.info("Функция сброса пароля будет доступна в следующей версии");
    // setPasswordResetSent(true);
  };

  const handleDeleteAccount = async () => {
    toast.error(t("settingsDeleteContactSupport"), { duration: 5000 });
    setShowDeleteConfirm(false);
  };

  const formatDate = (dateStr?: string) => {
    if (!dateStr) return "—";
    return new Date(dateStr).toLocaleDateString("ru-RU", {
      day: "2-digit",
      month: "long",
      year: "numeric",
    });
  };

  const currentAvatarDef = AVATAR_ICONS.find((a) => a.id === selectedAvatar) || AVATAR_ICONS[0];
  const AvatarIcon = currentAvatarDef.icon;

  return (
    <div className="min-h-screen bg-background">
      <HeaderBar />

      <div className="max-w-2xl mx-auto px-4 md:px-6 py-6 md:py-8">
        <div className="flex items-center gap-4 mb-8 md:mb-10">
          <button
            onClick={() => navigate(-1)}
            className="w-9 h-9 rounded-xl flex items-center justify-center text-muted-foreground hover:text-foreground hover:bg-secondary/50 transition-colors"
          >
            <ArrowLeft size={18} />
          </button>
          <h1 className="text-xl md:text-2xl font-bold text-foreground">{t("settingsTitle")}</h1>
        </div>

        <div className="space-y-6 md:space-y-8">
          {/* Profile */}
          <motion.section initial={{ opacity: 0, y: 12 }} animate={{ opacity: 1, y: 0 }} className="glass rounded-2xl border border-border/20 p-5 md:p-6">
            <h2 className="text-sm font-semibold text-foreground mb-6 uppercase tracking-wider">{t("settingsProfile")}</h2>
            <div className="flex items-center gap-5 mb-6">
              <div className={`w-16 h-16 rounded-2xl bg-secondary/60 flex items-center justify-center ${currentAvatarDef.color}`}>
                <AvatarIcon size={32} />
              </div>
              <div>
                <p className="text-sm font-medium text-foreground">{displayName || t("settingsUser")}</p>
                <p className="text-xs text-muted-foreground">{user?.email}</p>
              </div>
            </div>
            <div className="mb-6">
              <label className="text-xs text-muted-foreground font-medium mb-2 block">{t("settingsChooseAvatar")}</label>
              <div className="flex flex-wrap gap-2">
                {AVATAR_ICONS.map(({ id, icon: Icon, color }) => (
                  <button
                    key={id}
                    onClick={() => setSelectedAvatar(id)}
                    className={`w-10 h-10 rounded-xl flex items-center justify-center transition-all ${
                      selectedAvatar === id
                        ? "bg-primary/20 ring-2 ring-primary/50 scale-110"
                        : "bg-secondary/40 hover:bg-secondary/60"
                    } ${color}`}
                  >
                    <Icon size={18} />
                  </button>
                ))}
              </div>
            </div>
            <div className="mb-6">
              <label className="text-xs text-muted-foreground font-medium mb-2 block">{t("settingsDisplayName")}</label>
              <input
                value={displayName}
                onChange={(e) => setDisplayName(e.target.value)}
                placeholder={t("settingsNamePlaceholder")}
                className="w-full h-11 px-4 rounded-xl bg-secondary/30 border border-border/20 text-sm text-foreground placeholder:text-muted-foreground/40 outline-none focus:border-primary/40 focus:ring-1 focus:ring-primary/20 transition-all"
              />
            </div>
            <button
              onClick={handleSave}
              disabled={saving}
              className="flex items-center gap-2 h-10 px-5 rounded-xl btn-gradient text-primary-foreground text-sm font-medium disabled:opacity-50"
            >
              {saving ? <Loader2 size={14} className="animate-spin" /> : <Check size={14} />}
              {t("settingsSave")}
            </button>
          </motion.section>

          {/* Account */}
          <motion.section initial={{ opacity: 0, y: 12 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.05 }} className="glass rounded-2xl border border-border/20 p-5 md:p-6">
            <h2 className="text-sm font-semibold text-foreground mb-6 uppercase tracking-wider">{t("settingsAccount")}</h2>
            <div className="space-y-4">
              <div className="flex items-center gap-3">
                <div className="w-9 h-9 rounded-xl bg-secondary/40 flex items-center justify-center text-muted-foreground"><Mail size={16} /></div>
                <div>
                  <p className="text-xs text-muted-foreground">{t("settingsEmail")}</p>
                  <p className="text-sm text-foreground">{user?.email}</p>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <div className="w-9 h-9 rounded-xl bg-secondary/40 flex items-center justify-center text-muted-foreground"><Calendar size={16} /></div>
                <div>
                  <p className="text-xs text-muted-foreground">{t("settingsRegDate")}</p>
                  <p className="text-sm text-foreground">{formatDate(user?.created_at)}</p>
                </div>
              </div>
            </div>
          </motion.section>

          {/* Stats */}
          <motion.section initial={{ opacity: 0, y: 12 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.1 }} className="glass rounded-2xl border border-border/20 p-5 md:p-6">
            <h2 className="text-sm font-semibold text-foreground mb-6 uppercase tracking-wider">{t("settingsStats")}</h2>
            <div className="grid grid-cols-2 gap-4">
              <div className="rounded-xl bg-secondary/30 border border-border/10 p-4 text-center">
                <div className="flex items-center justify-center gap-2 mb-2"><FolderOpen size={16} className="text-primary" /></div>
                <p className="text-2xl font-bold text-foreground">{totalProjects}</p>
                <p className="text-xs text-muted-foreground mt-1">{t("settingsTotalProjects")}</p>
              </div>
              <div className="rounded-xl bg-secondary/30 border border-border/10 p-4 text-center">
                <div className="flex items-center justify-center gap-2 mb-2"><Globe size={16} className="text-emerald-400" /></div>
                <p className="text-2xl font-bold text-foreground">{publishedProjects}</p>
                <p className="text-xs text-muted-foreground mt-1">{t("settingsPublished")}</p>
              </div>
            </div>
          </motion.section>

          {/* Security */}
          <motion.section initial={{ opacity: 0, y: 12 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.15 }} className="glass rounded-2xl border border-border/20 p-5 md:p-6">
            <h2 className="text-sm font-semibold text-foreground mb-6 uppercase tracking-wider">{t("settingsSecurity")}</h2>
            <div className="space-y-3">
              <button
                onClick={handlePasswordReset}
                disabled={passwordResetSent}
                className="w-full flex items-center gap-3 h-12 px-4 rounded-xl bg-secondary/30 border border-border/10 hover:border-border/30 transition-all text-left disabled:opacity-50"
              >
                <Lock size={16} className="text-muted-foreground shrink-0" />
                <div className="flex-1">
                  <p className="text-sm text-foreground">{t("settingsChangePassword")}</p>
                  <p className="text-[11px] text-muted-foreground">
                    {passwordResetSent ? t("settingsPasswordResetSent") : t("settingsPasswordResetHint")}
                  </p>
                </div>
              </button>

              {!showDeleteConfirm ? (
                <button
                  onClick={() => setShowDeleteConfirm(true)}
                  className="w-full flex items-center gap-3 h-12 px-4 rounded-xl bg-secondary/30 border border-border/10 hover:border-destructive/30 transition-all text-left group"
                >
                  <Trash2 size={16} className="text-muted-foreground group-hover:text-destructive shrink-0 transition-colors" />
                  <div className="flex-1">
                    <p className="text-sm text-foreground group-hover:text-destructive transition-colors">{t("settingsDeleteAccount")}</p>
                    <p className="text-[11px] text-muted-foreground">{t("settingsDeleteWarning")}</p>
                  </div>
                </button>
              ) : (
                <div className="rounded-xl bg-destructive/10 border border-destructive/20 p-4 space-y-3">
                  <p className="text-sm text-destructive font-medium">{t("settingsDeleteConfirm")}</p>
                  <p className="text-xs text-muted-foreground">{t("settingsDeleteDesc")}</p>
                  <div className="flex items-center gap-2">
                    <button onClick={handleDeleteAccount} className="h-9 px-4 rounded-lg bg-destructive text-destructive-foreground text-xs font-medium">
                      {t("settingsDeleteYes")}
                    </button>
                    <button onClick={() => setShowDeleteConfirm(false)} className="h-9 px-4 rounded-lg bg-secondary/60 text-foreground text-xs">
                      {t("settingsCancel")}
                    </button>
                  </div>
                </div>
              )}
            </div>
          </motion.section>
        </div>
      </div>
    </div>
  );
};

export default Settings;
