"use client";
import { useCallback, useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { api } from "@/lib/api";
import { useWS, WSEvent } from "@/hooks/useWS";

export interface Column { id: string; name: string; position: number }
export interface Task { id: string; title: string; column_id: string; position: number; version: number; priority?: string; assignee?: string; due_date?: string }

export default function BoardPage() {
  const params = useParams();
  const projectId = Array.isArray(params.projectId) ? params.projectId[0] : (params.projectId as string);
  const [columns, setColumns] = useState<Column[]>([]);
  const [tasks, setTasks] = useState<Task[]>([]);
  const [toast, setToast] = useState<string | null>(null);
  const [newTitle, setNewTitle] = useState("");
  const [newCol, setNewCol] = useState("");
  const [dragId, setDragId] = useState<string | null>(null);

  const load = useCallback(async () => {
    const cols = await api<{ columns?: Column[] } | Column[]>(`/projects/${projectId}/columns`);
    setColumns(Array.isArray(cols) ? cols : cols.columns || []);
    const ts = await api<{ tasks?: Task[] } | Task[]>(`/tasks?project_id=${encodeURIComponent(projectId)}`);
    setTasks(Array.isArray(ts) ? ts : ts.tasks || []);
  }, [projectId]);

  useEffect(() => {
    load().catch((e: Error) => setToast(e.message));
  }, [load]);

  useEffect(() => {
    if (!toast) return;
    const t = setTimeout(() => setToast(null), 3500);
    return () => clearTimeout(t);
  }, [toast]);

  const onEvent = useCallback((ev: WSEvent) => {
    if (ev.type.startsWith("task.") || ev.type === "comment.added" || ev.type === "column.updated") load().catch(() => {});
  }, [load]);
  const { connected } = useWS(onEvent);

  const createTask = async (columnId: string) => {
    if (!newTitle.trim()) { setToast("Type a card title first"); return; }
    try {
      const t = await api<Task>("/tasks", { method: "POST", body: JSON.stringify({ title: newTitle, project_id: projectId, column_id: columnId }) });
      setTasks((p) => [...p, t]);
      setNewTitle("");
    } catch (e) { setToast(e instanceof Error ? e.message : "Create failed"); }
  };

  const moveTask = async (task: Task, columnId: string) => {
    try {
      const updated = await api<Task>(`/tasks/${task.id}/move`, { method: "POST", body: JSON.stringify({ column_id: columnId, position: 0, version: task.version }) });
      setTasks((p) => p.map((t) => (t.id === task.id ? updated : t)));
    } catch (e) {
      const err = e as Error & { status?: number };
      setToast(err.status === 409 ? "Conflict (409): card changed — refetching" : err.message);
      load().catch(() => {});
    }
  };

  const createColumn = async () => {
    if (!newCol.trim()) return;
    try {
      const c = await api<Column>(`/projects/${projectId}/columns`, { method: "POST", body: JSON.stringify({ name: newCol }) });
      setColumns((p) => [...p, c]);
      setNewCol("");
    } catch (e) { setToast(e instanceof Error ? e.message : "Column create failed"); }
  };

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center gap-3">
        <h1 className="text-xl font-bold">Board</h1>
        <span className={`text-xs px-2 py-0.5 rounded ${connected ? "bg-green-900 text-green-300" : "bg-zinc-800 text-zinc-400"}`}>{connected ? "● live" : "○ offline"}</span>
      </div>
      <div className="flex gap-2">
        <input placeholder="New card title (then + in a column)" value={newTitle} onChange={(e) => setNewTitle(e.target.value)} />
        <input placeholder="New column" value={newCol} onChange={(e) => setNewCol(e.target.value)} className="max-w-48" />
        <button className="btn whitespace-nowrap" onClick={createColumn}>+ Column</button>
      </div>
      <div className="flex gap-3 overflow-x-auto pb-4">
        {columns.map((col) => (
          <div key={col.id} className="min-w-64 w-64 card flex flex-col gap-2"
            onDragOver={(e) => e.preventDefault()}
            onDrop={() => {
              const t = tasks.find((x) => x.id === dragId);
              if (t && t.column_id !== col.id) moveTask(t, col.id);
              setDragId(null);
            }}>
            <h2 className="font-semibold text-sm">{col.name}</h2>
            {tasks.filter((t) => t.column_id === col.id).map((t) => (
              <div key={t.id} draggable onDragStart={() => setDragId(t.id)}
                className="border border-zinc-700 rounded p-2 text-sm bg-zinc-950 cursor-grab hover:border-brand">
                <Link href={`/tasks/${t.id}`} className="font-medium hover:text-brand-light">{t.title}</Link>
                <p className="text-xs text-zinc-500">v{t.version}{t.priority ? ` · ${t.priority}` : ""}{t.due_date ? ` · ${t.due_date}` : ""}</p>
              </div>
            ))}
            <button className="btn-ghost" onClick={() => createTask(col.id)}>+ card</button>
          </div>
        ))}
      </div>
      {toast && <div className="fixed bottom-4 right-4 bg-zinc-900 border border-brand rounded px-4 py-2 text-sm z-50">{toast}</div>}
    </div>
  );
}
