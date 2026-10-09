"use client";
import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { api } from "@/lib/api";

export interface TaskDetail { id: string; title: string; description?: string; version: number; priority?: string; assignee?: string; tags?: string[]; due_date?: string; column_id?: string }
export interface Comment { id: string; body: string; author?: string }
export interface Checklist { id: string; title?: string; items?: { id: string; text: string; done: boolean }[] }
export interface Activity { id: string; action: string; created_at?: string }

export default function TaskPage() {
  const params = useParams();
  const id = Array.isArray(params.id) ? params.id[0] : (params.id as string);
  const [task, setTask] = useState<TaskDetail | null>(null);
  const [comments, setComments] = useState<Comment[]>([]);
  const [checklists, setChecklists] = useState<Checklist[]>([]);
  const [activities, setActivities] = useState<Activity[]>([]);
  const [files, setFiles] = useState<{ id: string; filename: string; size: number; mime: string }[]>([]);
  const [toast, setToast] = useState("");
  const [comment, setComment] = useState("");
  const [edit, setEdit] = useState({ title: "", priority: "", assignee: "", due_date: "", tags: "" });

  useEffect(() => {
    api<TaskDetail>(`/tasks/${id}`).then((t) => {
      setTask(t);
      setEdit({ title: t.title || "", priority: t.priority || "", assignee: t.assignee || "", due_date: t.due_date || "", tags: (t.tags || []).join(",") });
    }).catch((e: Error) => setToast(e.message));
    api<{ comments?: Comment[] } | Comment[]>(`/tasks/${id}/comments`).then((d) => setComments(Array.isArray(d) ? d : d.comments || [])).catch(() => {});
    api<{ checklists?: Checklist[] } | Checklist[]>(`/tasks/${id}/checklists`).then((d) => setChecklists(Array.isArray(d) ? d : d.checklists || [])).catch(() => {});
    api<{ activities?: Activity[] } | Activity[]>(`/activities?task_id=${id}`).then((d) => setActivities(Array.isArray(d) ? d : d.activities || [])).catch(() => {});
    api<{ attachments?: { id: string; filename: string; size: number; mime: string }[] } | { id: string; filename: string; size: number; mime: string }[]>(`/tasks/${id}/attachments`).then((d) => {
      const list = Array.isArray(d) ? d : d.attachments || [];
      setFiles(list);
    }).catch(() => {});
  }, [id]);

  useEffect(() => {
    if (!toast) return;
    const t = setTimeout(() => setToast(""), 3500);
    return () => clearTimeout(t);
  }, [toast]);

  const save = async () => {
    if (!task) return;
    try {
      const updated = await api<TaskDetail>(`/tasks/${id}`, {
        method: "PATCH",
        body: JSON.stringify({ title: edit.title, priority: edit.priority || undefined, assignee: edit.assignee || undefined, due_date: edit.due_date || undefined, tags: edit.tags ? edit.tags.split(",").map((s) => s.trim()) : undefined, version: task.version }),
      });
      setTask(updated);
    } catch (e) {
      const err = e as Error & { status?: number };
      setToast(err.status === 409 ? "Conflict (409): version mismatch — reloaded" : err.message);
      api<TaskDetail>(`/tasks/${id}`).then(setTask).catch(() => {});
    }
  };

  const addComment = async () => {
    if (!comment.trim()) return;
    const c = await api<Comment>(`/tasks/${id}/comments`, { method: "POST", body: JSON.stringify({ body: comment }) });
    setComments((p) => [...p, c]);
    setComment("");
  };

  const addChecklist = async () => {
    const c = await api<Checklist>(`/tasks/${id}/checklists`, { method: "POST", body: JSON.stringify({ title: "Checklist" }) });
    setChecklists((p) => [...p, c]);
  };

  const addMeta = async () => {
    const m = await api<{ id: string; filename: string; size: number; mime: string }>(`/tasks/${id}/attachments`, {
      method: "POST", body: JSON.stringify({ filename: "spec.pdf", size: 1234, mime: "application/pdf" }),
    });
    setFiles((p) => [...p, m]);
  };

  if (!task) return <p className="text-sm text-zinc-400">Loading… {toast}</p>;

  return (
    <div className="flex flex-col gap-4 max-w-3xl">
      <h1 className="text-xl font-bold">{task.title}</h1>
      <div className="card flex flex-col gap-2">
        <input value={edit.title} onChange={(e) => setEdit({ ...edit, title: e.target.value })} placeholder="Title" />
        <div className="grid grid-cols-2 gap-2">
          <input value={edit.priority} onChange={(e) => setEdit({ ...edit, priority: e.target.value })} placeholder="priority (low|medium|high)" />
          <input value={edit.assignee} onChange={(e) => setEdit({ ...edit, assignee: e.target.value })} placeholder="assignee" />
          <input value={edit.due_date} onChange={(e) => setEdit({ ...edit, due_date: e.target.value })} placeholder="due date YYYY-MM-DD" />
          <input value={edit.tags} onChange={(e) => setEdit({ ...edit, tags: e.target.value })} placeholder="tags comma separated" />
        </div>
        <button className="btn w-fit" onClick={save}>Save (v{task.version})</button>
      </div>
      <div className="card">
        <h2 className="font-semibold mb-2">Checklists</h2>
        <button className="btn-ghost mb-2" onClick={addChecklist}>+ checklist</button>
        {checklists.map((c) => (
          <div key={c.id} className="border border-zinc-800 rounded p-2 mb-2 text-sm">
            <p className="font-medium">{c.title || c.id}</p>
            {(c.items || []).map((it) => <p key={it.id} className="text-zinc-300">{it.done ? "☑" : "☐"} {it.text}</p>)}
          </div>
        ))}
      </div>
      <div className="card">
        <h2 className="font-semibold mb-2">Comments</h2>
        {comments.map((c) => <p key={c.id} className="text-sm border-b border-zinc-800 py-1"><span className="text-zinc-500">{c.author || ""} </span>{c.body}</p>)}
        <div className="flex gap-2 mt-2">
          <input value={comment} onChange={(e) => setComment(e.target.value)} placeholder="Write a comment" />
          <button className="btn" onClick={addComment}>Send</button>
        </div>
      </div>
      <div className="card">
        <h2 className="font-semibold mb-2">Attachments (metadata)</h2>
        <button className="btn-ghost mb-2" onClick={addMeta}>+ add metadata</button>
        {files.map((f) => <p key={f.id} className="text-sm text-zinc-300">{f.filename} · {f.size}b · {f.mime}</p>)}
      </div>
      <div className="card">
        <h2 className="font-semibold mb-2">Activity</h2>
        {activities.map((a) => <p key={a.id} className="text-sm text-zinc-400">{a.action} {a.created_at || ""}</p>)}
      </div>
      {toast && <div className="fixed bottom-4 right-4 bg-zinc-900 border border-brand rounded px-4 py-2 text-sm z-50">{toast}</div>}
    </div>
  );
}
