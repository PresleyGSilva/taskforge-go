"use client";
import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { api, getOrgId, getToken } from "@/lib/api";

export interface DashboardData {
  by_column?: Record<string, number>;
  by_priority?: Record<string, number>;
  overdue?: number;
  by_member?: Record<string, number>;
  [k: string]: unknown;
}
export interface Project { id: string; name: string; workspace_id: string }

function recordOf(v: unknown): Record<string, number> {
  if (v && typeof v === "object" && !Array.isArray(v)) {
    const out: Record<string, number> = {};
    for (const [k, val] of Object.entries(v as Record<string, unknown>)) {
      if (typeof val === "number") out[k] = val;
    }
    return out;
  }
  return {};
}

export default function Dashboard() {
  const router = useRouter();
  const [dash, setDash] = useState<DashboardData | null>(null);
  const [projects, setProjects] = useState<Project[]>([]);
  const [newProject, setNewProject] = useState("");
  const [err, setErr] = useState("");

  useEffect(() => {
    if (!getToken()) { router.push("/login"); return; }
    const org = getOrgId();
    if (!org) return;
    api<DashboardData>(`/dashboard?workspace_id=${encodeURIComponent(org)}`).then(setDash).catch((e: Error) => setErr(e.message));
    api<{ projects?: Project[] } | Project[]>(`/projects?workspace_id=${encodeURIComponent(org)}`).then((d) => {
      setProjects(Array.isArray(d) ? d : d.projects || []);
    }).catch(() => {});
  }, [router]);

  const createProject = async () => {
    const org = getOrgId();
    if (!newProject.trim() || !org) return;
    try {
      const p = await api<Project>("/projects", { method: "POST", body: JSON.stringify({ name: newProject, workspace_id: org }) });
      setProjects((prev) => [...prev, p]);
      setNewProject("");
    } catch (e) {
      setErr(e instanceof Error ? e.message : "Create failed");
    }
  };

  const byCol = recordOf(dash?.by_column);
  const byPri = recordOf(dash?.by_priority);
  const byMem = recordOf(dash?.by_member);
  const overdue = typeof dash?.overdue === "number" ? dash.overdue : 0;

  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-xl font-bold">Productivity Dashboard</h1>
      {err && <p className="text-red-400 text-sm">{err}</p>}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
        <div className="card"><p className="text-xs text-zinc-400">Overdue</p><p className="text-2xl font-bold text-red-400">{overdue}</p></div>
        <div className="card"><p className="text-xs text-zinc-400">Projects</p><p className="text-2xl font-bold">{projects.length}</p></div>
        <div className="card"><p className="text-xs text-zinc-400">High priority</p><p className="text-2xl font-bold text-brand-light">{byPri["high"] || 0}</p></div>
        <div className="card"><p className="text-xs text-zinc-400">Total (columns)</p><p className="text-2xl font-bold">{Object.values(byCol).reduce((a, b) => a + b, 0)}</p></div>
      </div>
      <div className="grid md:grid-cols-3 gap-3">
        <div className="card"><h2 className="font-semibold mb-2">By column</h2>{Object.entries(byCol).map(([k, v]) => <p key={k} className="text-sm flex justify-between"><span>{k}</span><span>{v}</span></p>)}</div>
        <div className="card"><h2 className="font-semibold mb-2">By priority</h2>{Object.entries(byPri).map(([k, v]) => <p key={k} className="text-sm flex justify-between"><span>{k}</span><span>{v}</span></p>)}</div>
        <div className="card"><h2 className="font-semibold mb-2">By member</h2>{Object.entries(byMem).map(([k, v]) => <p key={k} className="text-sm flex justify-between"><span>{k}</span><span>{v}</span></p>)}</div>
      </div>
      <div className="card">
        <h2 className="font-semibold mb-2">Projects</h2>
        <div className="flex gap-2 mb-3">
          <input placeholder="New project name" value={newProject} onChange={(e) => setNewProject(e.target.value)} />
          <button className="btn whitespace-nowrap" onClick={createProject}>Create</button>
        </div>
        <div className="flex flex-col gap-2">
          {projects.map((p) => <Link key={p.id} href={`/boards/${p.id}`} className="border border-zinc-800 rounded px-3 py-2 hover:border-brand text-sm">{p.name}</Link>)}
        </div>
      </div>
    </div>
  );
}
