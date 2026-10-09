"use client";
import { useEffect, useState } from "react";
import { api, getOrgId } from "@/lib/api";
import { useAuth } from "@/components/AuthProvider";

export interface Workspace { id: string; name: string }
export interface Member { id: string; email?: string; role: string }
export interface Project { id: string; name: string }

export default function SettingsPage() {
  const { setOrgId } = useAuth();
  const [workspaces, setWorkspaces] = useState<Workspace[]>([]);
  const [members, setMembers] = useState<Member[]>([]);
  const [projects, setProjects] = useState<Project[]>([]);
  const [wsName, setWsName] = useState("");
  const [inviteEmail, setInviteEmail] = useState("");
  const [msg, setMsg] = useState("");

  const org = typeof window !== "undefined" ? getOrgId() : null;

  const loadWs = async () => {
    const d = await api<{ workspaces?: Workspace[] } | Workspace[]>("/workspaces");
    const list = Array.isArray(d) ? d : d.workspaces || [];
    setWorkspaces(list);
    try {
      const raw = localStorage.getItem("taskforge_orgs");
      if (!raw) localStorage.setItem("taskforge_orgs", JSON.stringify(list));
    } catch {}
  };

  useEffect(() => {
    loadWs().catch((e: Error) => setMsg(e.message));
    if (org) {
      api<{ members?: Member[] } | Member[]>(`/workspaces/${org}/members`).then((d) => setMembers(Array.isArray(d) ? d : d.members || [])).catch(() => {});
      api<{ projects?: Project[] } | Project[]>(`/projects?workspace_id=${org}`).then((d) => setProjects(Array.isArray(d) ? d : d.projects || [])).catch(() => {});
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const createWs = async () => {
    if (!wsName.trim()) return;
    const w = await api<Workspace>("/workspaces", { method: "POST", body: JSON.stringify({ name: wsName }) });
    setWorkspaces((p) => [...p, w]);
    setOrgId(w.id);
    setWsName("");
    loadWs().catch(() => {});
  };

  const invite = async () => {
    if (!org || !inviteEmail.trim()) return;
    const res = await api<{ token?: string; invite?: { token?: string } }>(`/workspaces/${org}/invites`, {
      method: "POST", body: JSON.stringify({ email: inviteEmail }),
    }).catch((e: Error) => { setMsg(e.message); return null; });
    if (res) {
      const tok = res.token || res.invite?.token || "";
      setMsg(tok ? `Invite token: ${tok}` : "Invite sent");
      setInviteEmail("");
    }
  };

  return (
    <div className="flex flex-col gap-4 max-w-2xl">
      <h1 className="text-xl font-bold">Settings</h1>
      {msg && <p className="text-sm text-brand-light">{msg}</p>}
      <div className="card">
        <h2 className="font-semibold mb-2">Workspaces</h2>
        {workspaces.map((w) => (
          <div key={w.id} className="flex justify-between text-sm border-b border-zinc-800 py-1">
            <span>{w.name}</span>
            <button className="btn-ghost" onClick={() => setOrgId(w.id)}>Use</button>
          </div>
        ))}
        <div className="flex gap-2 mt-2">
          <input placeholder="New workspace" value={wsName} onChange={(e) => setWsName(e.target.value)} />
          <button className="btn" onClick={createWs}>Create</button>
        </div>
      </div>
      <div className="card">
        <h2 className="font-semibold mb-2">Members & roles</h2>
        {members.map((m) => <p key={m.id} className="text-sm">{m.email || m.id} · {m.role}</p>)}
        <div className="flex gap-2 mt-2">
          <input placeholder="email to invite" value={inviteEmail} onChange={(e) => setInviteEmail(e.target.value)} />
          <button className="btn" onClick={invite}>Invite</button>
        </div>
      </div>
      <div className="card">
        <h2 className="font-semibold mb-2">Projects</h2>
        {projects.map((p) => <p key={p.id} className="text-sm">{p.name}</p>)}
      </div>
    </div>
  );
}
