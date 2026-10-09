"use client";
import { useState } from "react";
import Link from "next/link";
import { api, getOrgId } from "@/lib/api";

export interface SearchTask { id: string; title: string; priority?: string; assignee?: string }

export default function SearchPage() {
  const [q, setQ] = useState("");
  const [priority, setPriority] = useState("");
  const [assignee, setAssignee] = useState("");
  const [tag, setTag] = useState("");
  const [results, setResults] = useState<SearchTask[]>([]);

  const run = async () => {
    const org = getOrgId();
    const p = new URLSearchParams({ q, workspace_id: org || "" });
    if (priority) p.set("priority", priority);
    if (assignee) p.set("assignee", assignee);
    if (tag) p.set("tag", tag);
    const d = await api<{ tasks?: SearchTask[]; results?: SearchTask[] } | SearchTask[]>(`/search?${p.toString()}`);
    setResults(Array.isArray(d) ? d : d.tasks || d.results || []);
  };

  return (
    <div className="flex flex-col gap-4">
      <h1 className="text-xl font-bold">Search</h1>
      <div className="card flex flex-col gap-2">
        <input placeholder="Query" value={q} onChange={(e) => setQ(e.target.value)} />
        <div className="grid grid-cols-3 gap-2">
          <input placeholder="priority" value={priority} onChange={(e) => setPriority(e.target.value)} />
          <input placeholder="assignee" value={assignee} onChange={(e) => setAssignee(e.target.value)} />
          <input placeholder="tag" value={tag} onChange={(e) => setTag(e.target.value)} />
        </div>
        <button className="btn w-fit" onClick={run}>Search</button>
      </div>
      {results.map((t) => <Link key={t.id} href={`/tasks/${t.id}`} className="card text-sm hover:border-brand">{t.title}</Link>)}
    </div>
  );
}
