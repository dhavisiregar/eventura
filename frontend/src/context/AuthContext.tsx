"use client";

import { createContext, useCallback, useContext, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { api, apiErrorMessage } from "@/lib/api";
import { clearTokenCookie, getTokenCookie, setTokenCookie } from "@/lib/cookies";
import type { ApiResponse, User } from "@/lib/types";

interface AuthResponseBody {
  token: string;
  user: User;
}

interface AuthContextValue {
  user: User | null;
  loading: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (payload: {
    name: string;
    email: string;
    password: string;
    role: "customer" | "organizer";
    referral_code?: string;
  }) => Promise<void>;
  logout: () => void;
  refresh: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const router = useRouter();

  const refresh = useCallback(async () => {
    const token = getTokenCookie();
    if (!token) {
      setUser(null);
      setLoading(false);
      return;
    }
    try {
      const res = await api.get<ApiResponse<User>>("/auth/me");
      setUser(res.data.data);
    } catch {
      clearTokenCookie();
      setUser(null);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    refresh();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const login = useCallback(async (email: string, password: string) => {
    try {
      const res = await api.post<ApiResponse<AuthResponseBody>>("/auth/login", { email, password });
      setTokenCookie(res.data.data.token);
      setUser(res.data.data.user);
    } catch (err) {
      throw new Error(apiErrorMessage(err, "Failed to sign in"));
    }
  }, []);

  const register = useCallback(
    async (payload: {
      name: string;
      email: string;
      password: string;
      role: "customer" | "organizer";
      referral_code?: string;
    }) => {
      try {
        const res = await api.post<ApiResponse<AuthResponseBody>>("/auth/register", payload);
        setTokenCookie(res.data.data.token);
        setUser(res.data.data.user);
      } catch (err) {
        throw new Error(apiErrorMessage(err, "Failed to create account"));
      }
    },
    []
  );

  const logout = useCallback(() => {
    clearTokenCookie();
    setUser(null);
    router.push("/login");
  }, [router]);

  return (
    <AuthContext.Provider value={{ user, loading, login, register, logout, refresh }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within an AuthProvider");
  return ctx;
}
