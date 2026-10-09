"use client";
import Link from "next/link";
import { useAuth } from "./AuthProvider";
import { useEffect, useState } from "react";
import { api } from "@/lib/api";

type Org = { id: string; name: string };

export function Nav() {
  const { orgId, setOrgId, logoutAndGo, token } = useAuth();
  const [orgs, setOrgs] = useState<Org[]>([]);
  const [notifCount, setNotifCount] = useState(0);

  useEffect(() => {
    try {
      const raw = localStorage.getItem("taskforge_orgs");
      if (raw) setOrgs(JSON.parse(raw));
    } catch {}
    if (token) {
      api<{ workspaces?: Org[] } | Org[]>("/workspaces").then((d) => {
        const list = Array.isArray(d) ? d : d.workspaces || [];
        if (list.length) setOrgs(list);
      }).catch(() => {});
      api<{ unread?: number } | Array<unknown>>("/notifications").then((d) => {
        if (!Array.isArray(d)) return;
        setNotifCount(d.filter((n) => !(n as { read?: boolean }).read).length);
      }).catch(() => {});
    }
  }, [token]);

  if (!token) return null;
  return (
    <nav className="border-b border-zinc-800 bg-zinc-950 px-4 py-2 flex items-center gap-4 text-sm">
      <Link href="/" className="font-bold text-brand">⬢ TaskForge</Link>
      <Link href="/" className="text-zinc-300 hover:text-white">Dashboard</Link>
      <Link href="/search" className="text-zinc-300 hover:text-white">Search</Link>
      <Link href="/notifications" className="text-zinc-300 hover:text-white">Notifications{notifCount ? ` (${notifCount})` : ""}</Link>
      <Link href="/settings" className="text-zinc-300 hover:text-white">Settings</Link>
      <select
        className="ml-auto bg-zinc-900 border border-zinc-700 rounded px-2 py-1"
        value={orgId || ""}
        onChange={(e) => setOrgId(e.target.value)}
      >
        {orgs.map((o) => <option key={o.id} value={o.id}>{o.name}</option>)}
      </select>
      <button onClick={logoutAndGo} className="text-zinc-400 hover:text-white">Logout</button>
    </nav>
  );
}
