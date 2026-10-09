"use client";
import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { api, persistLogin } from "@/lib/api";

export default function LoginPage() {
  const router = useRouter();
  const [email, setEmail] = useState("demo@taskforge.dev");
  const [password, setPassword] = useState("demo1234");
  const [err, setErr] = useState("");

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErr("");
    try {
      const res = await api<{ user: unknown; access_token: string; orgs: { id: string; name: string }[] }>("/auth/login", {
        method: "POST",
        body: JSON.stringify({ email, password }),
      });
      persistLogin(res.access_token, res.orgs || []);
      router.push("/");
    } catch (e) {
      setErr(e instanceof Error ? e.message : "Login failed");
    }
  };

  return (
    <div className="max-w-sm mx-auto mt-16 card">
      <h1 className="text-2xl font-bold text-brand mb-4">⬢ TaskForge</h1>
      <form onSubmit={submit} className="flex flex-col gap-3">
        <input placeholder="Email" value={email} onChange={(e) => setEmail(e.target.value)} />
        <input placeholder="Password" type="password" value={password} onChange={(e) => setPassword(e.target.value)} />
        {err && <p className="text-red-400 text-sm">{err}</p>}
        <button className="btn" type="submit">Login</button>
      </form>
      <p className="text-xs text-zinc-400 mt-3">Demo: demo@taskforge.dev / demo1234 (seeded). No account? <Link href="/register" className="text-brand-light underline">Register</Link></p>
    </div>
  );
}
