"use client";
import { useEffect, useState } from "react";
import { api } from "@/lib/api";

export interface Notif { id: string; message?: string; title?: string; read?: boolean }

export default function NotificationsPage() {
  const [items, setItems] = useState<Notif[]>([]);

  useEffect(() => {
    api<{ notifications?: Notif[] } | Notif[]>("/notifications").then((d) => {
      setItems(Array.isArray(d) ? d : d.notifications || []);
    }).catch(() => {});
  }, []);

  const markRead = async (n: Notif) => {
    await api(`/notifications/${n.id}/read`, { method: "PATCH" }).catch(() => {});
    setItems((p) => p.map((x) => (x.id === n.id ? { ...x, read: true } : x)));
  };

  return (
    <div className="flex flex-col gap-3 max-w-2xl">
      <h1 className="text-xl font-bold">Notifications</h1>
      {items.map((n) => (
        <div key={n.id} className={`card text-sm flex justify-between items-center ${n.read ? "opacity-60" : ""}`}>
          <span>{n.message || n.title || n.id}</span>
          {!n.read && <button className="btn-ghost" onClick={() => markRead(n)}>Mark read</button>}
        </div>
      ))}
      {items.length === 0 && <p className="text-sm text-zinc-500">No notifications.</p>}
    </div>
  );
}
