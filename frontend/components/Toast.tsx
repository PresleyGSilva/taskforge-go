"use client";
import { useEffect, useState } from "react";

export function useToast() {
  const [msg, setMsg] = useState<string | null>(null);
  useEffect(() => {
    if (!msg) return;
    const t = setTimeout(() => setMsg(null), 3500);
    return () => clearTimeout(t);
  }, [msg]);
  return { msg, show: setMsg, Toast: () => (msg ? (
    <div className="fixed bottom-4 right-4 bg-zinc-900 border border-brand rounded px-4 py-2 text-sm shadow-lg z-50">{msg}</div>
  ) : null) };
}
