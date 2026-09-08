import { defineConfig, globalIgnores } from "eslint/config";
import nextVitals from "eslint-config-next/core-web-vitals";
import nextTs from "eslint-config-next/typescript";

const eslintConfig = defineConfig([
  ...nextVitals,
  ...nextTs,
  {
    rules: {
      // This app fetches data with plain useEffect + axios (no React Query/SWR).
      // Every such effect here guards against races with a `cancelled` flag, so
      // the new React Compiler "no setState in effect" rule is a false positive
      // for this codebase; keep it visible as a warning rather than a hard error.
      "react-hooks/set-state-in-effect": "warn",
    },
  },
  // Override default ignores of eslint-config-next.
  globalIgnores([
    // Default ignores of eslint-config-next:
    ".next/**",
    "out/**",
    "build/**",
    "next-env.d.ts",
    // Tooling config, not app code.
    "jest.config.js",
  ]),
]);

export default eslintConfig;
