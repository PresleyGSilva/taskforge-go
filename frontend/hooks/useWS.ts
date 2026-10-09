"use client";
import { useEffect, useRef, useState } from "react";
import { getOrgId, getToken } from "@/lib/api";

export type WSEvent = { type: string; payload: Record<string, unknown> };

export function useWS(onEvent: (ev: WSEvent) => void) {
  const [connected, setConnected] = useState(false);
  const cb = useRef(onEvent);
  cb.current = onEvent;

  useEffect(() => {
    let ws: WebSocket | null = null;
    let timer: ReturnType<typeof setTimeout> | null = null;
    let closed = false;

    const connect = () => {
      if (closed) return;
      const org = getOrgId();
      const tok = getToken();
      if (!org || !tok) return;
      const base = (process.env.NEXT_PUBLIC_API_URL || "http://localhost:8083").replace(/^http/, "ws");
      try {
        ws = new WebSocket(`${base}/api/v1/ws?workspace=${encodeURIComponent(org)}&token=${encodeURIComponent(tok)}`);
      } catch {
        schedule();
        return;
      }
      ws.onopen = () => setConnected(true);
      ws.onclose = () => { setConnected(false); schedule(); };
      ws.onerror = () => { try { ws?.close(); } catch {} };
      ws.onmessage = (m) => {
        try {
          const ev = JSON.parse(m.data as string) as WSEvent;
          cb.current(ev);
        } catch {}
      };
    };
    const schedule = () => {
      if (closed) return;
      if (timer) return;
      timer = setTimeout(() => { timer = null; connect(); }, 3000);
    };
    connect();
    return () => { closed = true; if (timer) clearTimeout(timer); try { ws?.close(); } catch {} };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return { connected };
}
