import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import { Toaster as Sonner } from "@/components/ui/sonner";
import { Toaster } from "@/components/ui/toaster";
import { TooltipProvider } from "@/components/ui/tooltip";
import { AuthProvider } from "@/hooks/useAuth";
import { ThemeProvider } from "@/hooks/useTheme";
import { LanguageProvider } from "@/hooks/useLanguage";
import { CreditsProvider } from "@/hooks/useCredits";
import { Navigate } from "react-router-dom";
import ProtectedRoute from "@/components/ProtectedRoute";
import Workspace from "./pages/Workspace.tsx";
import Auth from "./pages/Auth.tsx";
import ViewProject from "./pages/ViewProject.tsx";
import Admin from "./pages/Admin.tsx";
import NotFound from "./pages/NotFound.tsx";

// ── Grafted Lovable UI pages (run via @tanstack/react-router compat shim) ──
import { Route as DashboardRoute } from "./lovable-routes/index.tsx";
import { Route as ResourcesRoute } from "./lovable-routes/resources.tsx";
import { Route as ProfileRoute } from "./lovable-routes/profile.tsx";
import { Route as PromptsRoute } from "./lovable-routes/prompts.tsx";
import { Route as DocsRoute } from "./lovable-routes/docs.tsx";
import { Route as HelpRoute } from "./lovable-routes/help.tsx";
import { Route as StatusRoute } from "./lovable-routes/status.tsx";
import { Route as TermsRoute } from "./lovable-routes/terms.tsx";
import { Route as SettingsRoute } from "./lovable-routes/settings.tsx";
import { Route as SettingsAccountRoute } from "./lovable-routes/settings.account.tsx";
import { Route as SettingsAppsRoute } from "./lovable-routes/settings.apps.tsx";
import { Route as SettingsWorkspaceRoute } from "./lovable-routes/settings.workspace.tsx";
import { Route as SettingsBillingRoute } from "./lovable-routes/settings.billing.tsx";
import { Route as SettingsCloudRoute } from "./lovable-routes/settings.cloud.tsx";
import { Route as SettingsPeopleRoute } from "./lovable-routes/settings.people.tsx";
import { Route as SettingsKnowledgeRoute } from "./lovable-routes/settings.knowledge.tsx";
import { Route as SettingsSkillsRoute } from "./lovable-routes/settings.skills.tsx";
import { Route as SettingsTemplatesRoute } from "./lovable-routes/settings.templates.tsx";
import { Route as SettingsDesignRoute } from "./lovable-routes/settings.design-systems.tsx";
import { Route as SettingsGitRoute } from "./lovable-routes/settings.git.tsx";
import { Route as SettingsDomainsRoute } from "./lovable-routes/settings.domains.tsx";
import { Route as SettingsPrivacyRoute } from "./lovable-routes/settings.privacy.tsx";
import { Route as SettingsSecurityRoute } from "./lovable-routes/settings.security-center.tsx";
import { Route as SettingsAuditRoute } from "./lovable-routes/settings.audit-logs.tsx";
import { Route as SettingsProjectRoute } from "./lovable-routes/settings.project.tsx";

const queryClient = new QueryClient();

const App = () => (
  <QueryClientProvider client={queryClient}>
    <AuthProvider>
      <ThemeProvider>
        <LanguageProvider>
          <CreditsProvider>
          <TooltipProvider>
            <Toaster />
            <Sonner />
            <BrowserRouter>
              <Routes>
                {/* ── New Lovable UI — authenticated app (dashboard is the home) ── */}
                <Route
                  path="/"
                  element={
                    <ProtectedRoute>
                      <DashboardRoute.component />
                    </ProtectedRoute>
                  }
                />
                <Route
                  path="/resources"
                  element={
                    <ProtectedRoute>
                      <ResourcesRoute.component />
                    </ProtectedRoute>
                  }
                />
                <Route
                  path="/profile"
                  element={
                    <ProtectedRoute>
                      <ProfileRoute.component />
                    </ProtectedRoute>
                  }
                />
                <Route
                  path="/prompts"
                  element={
                    <ProtectedRoute>
                      <PromptsRoute.component />
                    </ProtectedRoute>
                  }
                />

                {/* ── Public informational pages ── */}
                <Route path="/docs" element={<DocsRoute.component />} />
                <Route path="/help" element={<HelpRoute.component />} />
                <Route path="/status" element={<StatusRoute.component />} />
                <Route path="/terms" element={<TermsRoute.component />} />

                {/* ── Settings (nested layout + sidebar, protected) ── */}
                <Route
                  path="/settings"
                  element={
                    <ProtectedRoute>
                      <SettingsRoute.component />
                    </ProtectedRoute>
                  }
                >
                  <Route index element={<Navigate to="/settings/account" replace />} />
                  <Route path="account" element={<SettingsAccountRoute.component />} />
                  <Route path="apps" element={<SettingsAppsRoute.component />} />
                  <Route path="workspace" element={<SettingsWorkspaceRoute.component />} />
                  <Route path="billing" element={<SettingsBillingRoute.component />} />
                  <Route path="cloud" element={<SettingsCloudRoute.component />} />
                  <Route path="people" element={<SettingsPeopleRoute.component />} />
                  <Route path="knowledge" element={<SettingsKnowledgeRoute.component />} />
                  <Route path="skills" element={<SettingsSkillsRoute.component />} />
                  <Route path="templates" element={<SettingsTemplatesRoute.component />} />
                  <Route path="design-systems" element={<SettingsDesignRoute.component />} />
                  <Route path="git" element={<SettingsGitRoute.component />} />
                  <Route path="domains" element={<SettingsDomainsRoute.component />} />
                  <Route path="privacy" element={<SettingsPrivacyRoute.component />} />
                  <Route path="security-center" element={<SettingsSecurityRoute.component />} />
                  <Route path="audit-logs" element={<SettingsAuditRoute.component />} />
                  <Route path="project" element={<SettingsProjectRoute.component />} />
                </Route>

                {/* ── Auth ── */}
                <Route path="/auth" element={<Auth />} />

                {/* ── Real Builder workspace (our SSE + Sandpack) ── */}
                <Route
                  path="/project/new"
                  element={
                    <ProtectedRoute>
                      <Workspace />
                    </ProtectedRoute>
                  }
                />
                <Route
                  path="/builder"
                  element={
                    <ProtectedRoute>
                      <Workspace />
                    </ProtectedRoute>
                  }
                />
                <Route
                  path="/builder/:id"
                  element={
                    <ProtectedRoute>
                      <Workspace />
                    </ProtectedRoute>
                  }
                />
                <Route path="/view/:projectId" element={<ViewProject />} />
                <Route
                  path="/admin"
                  element={
                    <ProtectedRoute>
                      <Admin />
                    </ProtectedRoute>
                  }
                />
                {/* ADD ALL CUSTOM ROUTES ABOVE THE CATCH-ALL "*" ROUTE */}
                <Route path="*" element={<NotFound />} />
              </Routes>
            </BrowserRouter>
          </TooltipProvider>
          </CreditsProvider>
        </LanguageProvider>
      </ThemeProvider>
    </AuthProvider>
  </QueryClientProvider>
);

export default App;
