"use client";

import { useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { CalendarRange, LayoutDashboard, LogOut, Menu, Ticket, User as UserIcon, X } from "lucide-react";
import { useAuth } from "@/context/AuthContext";
import { Button } from "./ui/Button";
import { Tooltip } from "./ui/Tooltip";
import { ThemeToggle } from "./ThemeToggle";

export function Navbar() {
  const { user, logout, loading } = useAuth();
  const pathname = usePathname();
  const [open, setOpen] = useState(false);

  const links = [{ href: "/", label: "Browse Events" }];

  return (
    <header className="sticky top-0 z-40 border-b border-slate-200 bg-white/90 backdrop-blur dark:border-slate-800 dark:bg-slate-900/90">
      <div className="mx-auto flex max-w-7xl items-center justify-between px-4 py-3 sm:px-6 lg:px-8">
        <Link href="/" className="flex items-center gap-2 font-semibold text-slate-900 dark:text-slate-100">
          <CalendarRange className="h-6 w-6 text-indigo-600 dark:text-indigo-400" />
          <span>Eventura</span>
        </Link>

        <nav className="hidden items-center gap-6 md:flex">
          {links.map((l) => (
            <Link
              key={l.href}
              href={l.href}
              className={`text-sm font-medium ${pathname === l.href ? "text-indigo-600 dark:text-indigo-400" : "text-slate-600 hover:text-slate-900 dark:text-slate-400 dark:hover:text-slate-100"}`}
            >
              {l.label}
            </Link>
          ))}
        </nav>

        <div className="hidden items-center gap-3 md:flex">
          {!loading && !user && (
            <>
              <Link href="/login">
                <Button variant="ghost" size="sm">
                  Log in
                </Button>
              </Link>
              <Link href="/register">
                <Button size="sm">Sign up</Button>
              </Link>
            </>
          )}
          {!loading && user?.role === "customer" && (
            <>
              <Link href="/my-tickets">
                <Button variant="outline" size="sm">
                  <Ticket className="h-4 w-4" /> My tickets
                </Button>
              </Link>
              <Link href="/account" className="flex items-center gap-2 text-sm font-medium text-slate-700 dark:text-slate-300">
                <UserIcon className="h-4 w-4" /> {user.name.split(" ")[0]}
              </Link>
              <Tooltip label="Log out">
                <Button variant="ghost" size="sm" onClick={logout} aria-label="Log out">
                  <LogOut className="h-4 w-4" />
                </Button>
              </Tooltip>
            </>
          )}
          {!loading && user?.role === "organizer" && (
            <>
              <Link href="/organizer/dashboard">
                <Button variant="outline" size="sm">
                  <LayoutDashboard className="h-4 w-4" /> Dashboard
                </Button>
              </Link>
              <Tooltip label="Log out">
                <Button variant="ghost" size="sm" onClick={logout} aria-label="Log out">
                  <LogOut className="h-4 w-4" />
                </Button>
              </Tooltip>
            </>
          )}
          <ThemeToggle />
        </div>

        <div className="flex items-center gap-1 md:hidden">
          <ThemeToggle />
          <button
            className="text-slate-700 dark:text-slate-300"
            onClick={() => setOpen((v) => !v)}
            aria-label="Toggle menu"
          >
            {open ? <X className="h-6 w-6" /> : <Menu className="h-6 w-6" />}
          </button>
        </div>
      </div>

      {open && (
        <div className="border-t border-slate-200 bg-white px-4 py-3 md:hidden dark:border-slate-800 dark:bg-slate-900">
          <div className="flex flex-col gap-2">
            {links.map((l) => (
              <Link
                key={l.href}
                href={l.href}
                className="py-2 text-sm font-medium text-slate-700 dark:text-slate-300"
                onClick={() => setOpen(false)}
              >
                {l.label}
              </Link>
            ))}
            {!loading && !user && (
              <div className="flex gap-2 pt-2">
                <Link href="/login" className="flex-1">
                  <Button variant="outline" size="sm" className="w-full">
                    Log in
                  </Button>
                </Link>
                <Link href="/register" className="flex-1">
                  <Button size="sm" className="w-full">
                    Sign up
                  </Button>
                </Link>
              </div>
            )}
            {!loading && user?.role === "customer" && (
              <div className="flex flex-col gap-2 pt-2">
                <Link href="/my-tickets" onClick={() => setOpen(false)}>
                  <Button variant="outline" size="sm" className="w-full">
                    My tickets
                  </Button>
                </Link>
                <Link href="/account" onClick={() => setOpen(false)}>
                  <Button variant="outline" size="sm" className="w-full">
                    My account
                  </Button>
                </Link>
                <Button variant="ghost" size="sm" onClick={logout}>
                  Log out
                </Button>
              </div>
            )}
            {!loading && user?.role === "organizer" && (
              <div className="flex flex-col gap-2 pt-2">
                <Link href="/organizer/dashboard" onClick={() => setOpen(false)}>
                  <Button variant="outline" size="sm" className="w-full">
                    Dashboard
                  </Button>
                </Link>
                <Button variant="ghost" size="sm" onClick={logout}>
                  Log out
                </Button>
              </div>
            )}
          </div>
        </div>
      )}
    </header>
  );
}
