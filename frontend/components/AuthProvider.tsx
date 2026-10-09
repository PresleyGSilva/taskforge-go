"use client";
import { createContext, useContext, useEffect, useState, ReactNode } from "react";
import { useRouter } from "next/navigation";
import { getToken, getOrgId, logout } from "@/lib/api";

type Ctx = { token: string | null; orgId: string | null; setOrgId: (id: string) => void; logoutAndGo: () => void };
const AuthCtx = createContext<Ctx>({ token: null, orgId: null, setOrgId: () => {}, logoutAndGo: () => {} });

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(null);
  const [orgId, setOrgIdState] = useState<string | null>(null);
  const router = useRouter();

  useEffect(() => {
    setToken(getToken());
    setOrgIdState(getOrgId());
  }, []);

  const setOrgId = (id: string) => {
    localStorage.setItem("taskforge_org_id", id);
    setOrgIdState(id);
  };
  const logoutAndGo = () => { logout(); setToken(null); router.push("/login"); };

  return <AuthCtx.Provider value={{ token, orgId, setOrgId, logoutAndGo }}>{children}</AuthCtx.Provider>;
}

export const useAuth = () => useContext(AuthCtx);
