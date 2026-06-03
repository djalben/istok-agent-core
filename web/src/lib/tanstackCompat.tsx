// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  @tanstack/react-router → react-router-dom compat shim
//  Aliased in vite.config.ts + tsconfig so the grafted
//  Lovable pages run unmodified on our SPA router.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
import * as React from "react";
import {
  Link as RRLink,
  Navigate,
  Outlet as RROutlet,
  useLocation,
  useNavigate as useRRNavigate,
  useParams as useRRParams,
} from "react-router-dom";

function resolvePath(to: string, params?: Record<string, string | number>): string {
  if (!to) return "/";
  let path = to;
  if (params) {
    for (const key of Object.keys(params)) {
      path = path.replace(`$${key}`, String(params[key]));
    }
  }
  return path;
}

// ── Route factory no-ops: just return the options object. ──
export function createFileRoute(_path?: string) {
  return (options: any): any => options;
}
export const createRootRoute = (options?: any): any => options;
export function createRootRouteWithContext<_T = unknown>() {
  return (options?: any): any => options;
}

export function redirect(opts: { to: string }): never {
  const err = new Error("redirect") as Error & { redirectTo?: string };
  err.redirectTo = opts.to;
  throw err;
}
export function notFound(): Error {
  return new Error("not-found");
}

export const Outlet = RROutlet;
export const HeadContent: React.FC = () => null;
export const Scripts: React.FC = () => null;

type NavOpts = {
  to: string;
  params?: Record<string, string | number>;
  search?: unknown;
  replace?: boolean;
  state?: unknown;
};

export function useNavigate() {
  const navigate = useRRNavigate();
  return (opts: NavOpts | string) => {
    if (typeof opts === "string") return navigate(opts);
    return navigate(resolvePath(opts.to, opts.params), { replace: opts.replace, state: opts.state });
  };
}

export function useRouter() {
  const navigate = useRRNavigate();
  return {
    navigate: (opts: NavOpts) => navigate(resolvePath(opts.to, opts.params), { replace: opts?.replace, state: opts?.state }),
    history: { back: () => navigate(-1) },
    invalidate: () => {},
  };
}

export function useRouterState<T = unknown>(opts?: { select?: (s: any) => T }): T {
  const location = useLocation();
  const state = {
    location: {
      pathname: location.pathname,
      search: location.search,
      hash: location.hash,
    },
  };
  return opts?.select ? opts.select(state) : (state as unknown as T);
}

export function useParams<T = Record<string, string>>(_opts?: unknown): T {
  return useRRParams() as unknown as T;
}

export const Link = React.forwardRef<HTMLAnchorElement, any>(function Link(
  { to, params, search: _search, activeProps: _activeProps, activeOptions: _activeOptions, ...rest },
  ref,
) {
  return <RRLink ref={ref} to={resolvePath(to, params)} {...rest} />;
});

export { Navigate };
