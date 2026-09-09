import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // "standalone" output is needed for the Docker build (see frontend/Dockerfile, which
  // copies .next/standalone) but breaks Vercel's own build/output-tracing pipeline —
  // it expects the default output mode and errors on next-server.js.nft.json otherwise.
  // Vercel sets the VERCEL env var during its builds, so skip standalone there.
  output: process.env.VERCEL ? undefined : "standalone",
};

export default nextConfig;
