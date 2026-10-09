const API_BASE = (process.env.NEXT_PUBLIC_API_URL || "http://localhost:8083") + "/api/v1";

export const TOKEN_KEY = "taskforge_token";
export const ORG_KEY = "taskforge_org_id";

export function getToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(TOKEN_KEY);
}
export function getOrgId(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(ORG_KEY);
}

async function refreshToken(): Promise<boolean> {
  try {
    const res = await fetch(API_BASE + "/auth/refresh", { method: "POST", credentials: "include" });
    if (!res.ok) return false;
    const body = await res.json();
    const tok = body?.data?.access_token;
    if (tok && typeof window !== "undefined") localStorage.setItem(TOKEN_KEY, tok);
    return !!tok;
  } catch {
    return false;
  }
}

export async function api<T = unknown>(path: string, opts: RequestInit = {}): Promise<T> {
  const doFetch = async (): Promise<Response> => {
    const headers: Record<string, string> = { "Content-Type": "application/json" };
    const tok = getToken();
    if (tok) headers["Authorization"] = `Bearer ${tok}`;
    const org = getOrgId();
    if (org) headers["X-Org-ID"] = org;
    const extra = (opts.headers as Record<string, string>) || {};
    return fetch(API_BASE + path, { ...opts, headers: { ...headers, ...extra }, credentials: "include" });
  };
  let res = await doFetch();
  if (res.status === 401 && getToken()) {
    const ok = await refreshToken();
    if (ok) res = await doFetch();
  }
  if (!res.ok) {
    let msg = `HTTP ${res.status}`;
    try {
      const b = await res.json();
      msg = b?.error?.message || msg;
      const e = new Error(msg) as Error & { code?: string; status?: number };
      e.code = b?.error?.code;
      e.status = res.status;
      throw e;
    } catch (err) {
      if (err instanceof Error && (err as { status?: number }).status) throw err;
      throw new Error(msg);
    }
  }
  if (res.status === 204) return undefined as T;
  const body = await res.json();
  return (body?.data !== undefined ? body.data : body) as T;
}

export type Org = { id: string; name: string };
export type User = { id: string; name: string; email: string };

export function persistLogin(accessToken: string, orgs: Org[]) {
  localStorage.setItem(TOKEN_KEY, accessToken);
  localStorage.setItem("taskforge_orgs", JSON.stringify(orgs));
  if (orgs.length > 0) localStorage.setItem(ORG_KEY, orgs[0].id);
}

export function logout() {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(ORG_KEY);
  localStorage.removeItem("taskforge_orgs");
  fetch(API_BASE + "/auth/logout", { method: "POST", credentials: "include" }).catch(() => {});
}
