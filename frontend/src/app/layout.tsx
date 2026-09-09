import type { Metadata } from "next";
import Script from "next/script";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import { AuthProvider } from "@/context/AuthContext";
import { ThemeProvider, THEME_INIT_SCRIPT } from "@/context/ThemeContext";
import { Navbar } from "@/components/Navbar";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "Eventura — Discover and host events",
  description: "Browse events, buy tickets, and manage your own events on Eventura.",
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html lang="en" className={`${geistSans.variable} ${geistMono.variable} h-full antialiased`} suppressHydrationWarning>
      <body className="flex min-h-full flex-col bg-slate-50 dark:bg-slate-950" suppressHydrationWarning>
        {/* Sets the theme class before hydration to avoid a flash of the wrong theme. Must use
            next/script (not a raw <script> tag) so Next.js hoists and runs it outside React's
            own render/hydration of this tree. */}
        <Script id="theme-init" strategy="beforeInteractive" dangerouslySetInnerHTML={{ __html: THEME_INIT_SCRIPT }} />
        <ThemeProvider>
          <AuthProvider>
            <Navbar />
            <main className="flex-1">{children}</main>
            <footer className="border-t border-slate-200 bg-white py-8 text-center text-xs text-slate-500 dark:border-slate-800 dark:bg-slate-900 dark:text-slate-400">
              Eventura — Event Management Platform MVP
            </footer>
          </AuthProvider>
        </ThemeProvider>
      </body>
    </html>
  );
}
