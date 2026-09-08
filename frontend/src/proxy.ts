import { NextRequest, NextResponse } from "next/server";
import { jwtVerify } from "jose";

const JWT_SECRET = process.env.JWT_SECRET || "dev-secret-change-me";

const ORGANIZER_PREFIXES = ["/organizer"];
const CUSTOMER_PREFIXES = ["/checkout", "/my-tickets", "/reviews"];
const AUTH_ONLY_PREFIXES = ["/account"];
const GUEST_ONLY_PATHS = ["/login", "/register"];

function matchesPrefix(pathname: string, prefixes: string[]) {
  return prefixes.some((p) => pathname === p || pathname.startsWith(p + "/"));
}

async function getRole(token: string | undefined): Promise<string | null> {
  if (!token) return null;
  try {
    const { payload } = await jwtVerify(token, new TextEncoder().encode(JWT_SECRET));
    return (payload.role as string) ?? null;
  } catch {
    return null;
  }
}

export async function proxy(req: NextRequest) {
  const { pathname } = req.nextUrl;
  const token = req.cookies.get("token")?.value;
  const role = await getRole(token);

  if (matchesPrefix(pathname, GUEST_ONLY_PATHS) && role) {
    const dest = role === "organizer" ? "/organizer/dashboard" : "/";
    return NextResponse.redirect(new URL(dest, req.url));
  }

  const needsOrganizer = matchesPrefix(pathname, ORGANIZER_PREFIXES);
  const needsCustomer = matchesPrefix(pathname, CUSTOMER_PREFIXES);
  const needsAuth = matchesPrefix(pathname, AUTH_ONLY_PREFIXES);

  if (needsOrganizer || needsCustomer || needsAuth) {
    if (!role) {
      const loginUrl = new URL("/login", req.url);
      loginUrl.searchParams.set("next", pathname);
      return NextResponse.redirect(loginUrl);
    }
    if (needsOrganizer && role !== "organizer") {
      return NextResponse.redirect(new URL("/", req.url));
    }
    if (needsCustomer && role !== "customer") {
      return NextResponse.redirect(new URL("/organizer/dashboard", req.url));
    }
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico|uploads).*)"],
};
