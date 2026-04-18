import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter, Link, Route, Routes, useLocation, useNavigate, useParams } from "react-router-dom";
import { CoreProvider, attachRealtimeSync, getWorkspaceId, useAuthStore, useWorkspaceStore } from "@openzoo/core";
import {
  DashboardLayout, IssuesPage, AgentsPage, InboxPage, ProjectsPage,
  LoginPage, RuntimesPage, ChatPage, SearchPage, SettingsPage,
  NavigationContext, type NavigationAdapter, NotFoundPage,
  IssueDetailPage, MyIssuesPage, SkillsPage, CyclesPage, LabelsPage, ViewsPage,
  CallbackPage, LandingPage,
} from "@openzoo/views";
import { ThemeProvider, ToastProvider, ErrorBoundary, LoadingScreen } from "@openzoo/ui";
import "@openzoo/ui/styles/globals.css";

function useWebNav() {
  const location = useLocation();
  const navigate = useNavigate();
  const params = useParams();

  return React.useMemo<NavigationAdapter>(() => ({
    navigate: (path) => navigate(path),
    goBack: () => navigate(-1),
    getCurrentPath: () => location.pathname,
    createHref: (path) => path,
    usePathname: () => location.pathname,
    useParams: <T extends Record<string, string>>() => params as T,
    Link: ({ to, className, children }) => <Link to={to} className={className}>{children}</Link>,
  }), [location, navigate, params]);
}

function AuthGuard({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  if (!isAuthenticated) return <LoginPage />;
  return <>{children}</>;
}

function DashboardPage() {
  const workspace = useWorkspaceStore((s) => s.currentWorkspace);
  return (
    <div>
      <h2 className="text-xl font-semibold mb-4">Dashboard</h2>
      {workspace ? (
        <p className="text-muted-foreground">Welcome to {workspace.name}. Select a section from the sidebar.</p>
      ) : (
        <p className="text-muted-foreground">Select or create a workspace to get started.</p>
      )}
    </div>
  );
}

function AuthenticatedShell() {
  const navigation = useWebNav();
  const wsId = getWorkspaceId();

  React.useEffect(() => {
    if (wsId) {
      const disconnect = attachRealtimeSync(wsId, {
        onIssueChanged: () => console.log("[realtime] issue changed"),
        onTaskChanged: () => console.log("[realtime] task changed"),
        onInboxChanged: () => console.log("[realtime] inbox changed"),
      });
      return disconnect;
    }
  }, [wsId]);

  return (
    <NavigationContext.Provider value={navigation}>
      <DashboardLayout>
        <React.Suspense fallback={<LoadingScreen />}>
          <Routes>
            <Route path="/" element={<DashboardPage />} />
            <Route path="/my-issues" element={<MyIssuesPage />} />
            <Route path="/issues" element={<IssuesPage />} />
            <Route path="/issues/:issueId" element={<IssueDetailPage />} />
            <Route path="/agents" element={<AgentsPage />} />
            <Route path="/inbox" element={<InboxPage />} />
            <Route path="/projects" element={<ProjectsPage />} />
            <Route path="/runtimes" element={<RuntimesPage />} />
            <Route path="/skills" element={<SkillsPage />} />
            <Route path="/cycles" element={<CyclesPage />} />
            <Route path="/labels" element={<LabelsPage />} />
            <Route path="/views" element={<ViewsPage />} />
            <Route path="/chat" element={<ChatPage />} />
            <Route path="/search" element={<SearchPage />} />
            <Route path="/settings" element={<SettingsPage />} />
            <Route path="*" element={<NotFoundPage />} />
          </Routes>
        </React.Suspense>
      </DashboardLayout>
    </NavigationContext.Provider>
  );
}

function HomeOrDashboard() {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const navigate = useNavigate();

  if (isAuthenticated) {
    return <AuthenticatedShell />;
  }

  return (
    <LandingPage
      onGetStarted={() => navigate("/issues")}
      onLogin={() => navigate("/issues")}
      onGoToDocs={() => window.open("/docs", "_blank")}
    />
  );
}

function AppRoutes() {
  const navigate = useNavigate();

  return (
    <Routes>
      <Route path="/auth/callback" element={
        <CallbackPage onSuccess={() => navigate("/issues")} onError={() => navigate("/")} />
      } />
      <Route path="/" element={<HomeOrDashboard />} />
      <Route path="/*" element={
        <AuthGuard>
          <AuthenticatedShell />
        </AuthGuard>
      } />
    </Routes>
  );
}

function App() {
  return (
    <React.StrictMode>
      <ThemeProvider>
        <CoreProvider onUnauthorized={() => useAuthStore.getState().logout()}>
          <ToastProvider>
            <BrowserRouter>
              <ErrorBoundary>
                <AppRoutes />
              </ErrorBoundary>
            </BrowserRouter>
          </ToastProvider>
        </CoreProvider>
      </ThemeProvider>
    </React.StrictMode>
  );
}

ReactDOM.createRoot(document.getElementById("root")!).render(<App />);
