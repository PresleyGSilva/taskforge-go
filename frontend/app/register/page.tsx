"use client";
import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { api, persistLogin } from "@/lib/api";

export default function RegisterPage() {
  const router = useRouter();
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [err, setErr] = useState("");

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErr("");
    try {
      const res = await api<{ user: unknown; access_token: string; orgs: { id: string; name: string }[] }>("/auth/register", {
        method: "POST",
        body: JSON.stringify({ name, email, password }),
      });
      persistLogin(res.access_token, res.orgs || []);
      router.push("/");
    } catch (e) {
      setErr(e instanceof Error ? e.message : "Register failed");
    }
  };

  return (
    <div className="max-w-sm mx-auto mt-16 card">
      <h1 className="text-2xl font-bold text-brand mb-4">Create account</h1>
      <form onSubmit={submit} className="flex flex-col gap-3">
        <input placeholder="Name" value={name} onChange={(e) => setName(e.target.value)} />
        <input placeholder="Email" value={email} onChange={(e) => setEmail(e.target.value)} />
        <input placeholder="Password" type="password" value={password} onChange={(e) => setPassword(e.target.value)} />
        {err && <p className="text-red-400 text-sm">{err}</p>}
        <button className="btn" type="submit">Register</button>
      </form>
      <p className="text-xs text-zinc-400 mt-3"><Link href="/login" className="text-brand-light underline">Back to login</Link></p>
    </div>
  );
}
