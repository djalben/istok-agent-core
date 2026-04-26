import { useNavigate } from "react-router-dom";
import { LogOut, FolderOpen, Settings } from "lucide-react";
import { useAuth } from "@/hooks/useAuth";
import { useLanguage } from "@/hooks/useLanguage";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { toast } from "sonner";

const UserMenu = () => {
  const { user, signOut } = useAuth();
  const navigate = useNavigate();
  const { t } = useLanguage();

  if (!user) {
    return (
      <>
        <button
          onClick={() => navigate("/auth")}
          className="px-4 py-1.5 text-sm text-muted-foreground hover:text-foreground border border-border/50 hover:border-foreground/20 rounded-lg transition-all duration-200"
        >
          {t("login")}
        </button>
        <button
          onClick={() => navigate("/auth")}
          className="px-4 py-1.5 text-sm text-primary-foreground btn-gradient rounded-lg font-medium"
        >
          {t("signup")}
        </button>
      </>
    );
  }

  const displayName = user.display_name || user.email?.split("@")[0] || "User";
  const initials = displayName.slice(0, 2).toUpperCase();

  const handleSignOut = async () => {
    await signOut();
    toast.success(t("authSignedOut"));
    navigate("/");
  };

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <button className="flex items-center gap-2 px-2 py-1 rounded-lg hover:bg-secondary/50 transition-colors">
          <div className="w-7 h-7 rounded-full bg-primary/20 flex items-center justify-center text-xs font-medium text-primary">
            {initials}
          </div>
          <span className="text-sm text-foreground hidden sm:inline max-w-[120px] truncate">
            {displayName}
          </span>
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-48 bg-card border-border/50">
        <div className="px-3 py-2">
          <p className="text-xs text-muted-foreground truncate">{user.email}</p>
        </div>
        <DropdownMenuSeparator />
        <DropdownMenuItem onClick={() => navigate("/projects")} className="text-sm gap-2 cursor-pointer">
          <FolderOpen size={14} />
          {t("myProjects")}
        </DropdownMenuItem>
        <DropdownMenuItem onClick={() => navigate("/settings")} className="text-sm gap-2 cursor-pointer">
          <Settings size={14} />
          {t("settings")}
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem onClick={handleSignOut} className="text-sm gap-2 cursor-pointer text-destructive focus:text-destructive">
          <LogOut size={14} />
          {t("signOut")}
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
};

export default UserMenu;
