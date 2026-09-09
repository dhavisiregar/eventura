"use client";

import { useCallback, useRef, useState } from "react";
import { Button } from "./Button";

interface ConfirmOptions {
  title: string;
  description?: string;
  confirmLabel?: string;
  cancelLabel?: string;
  danger?: boolean;
}

interface ConfirmState extends ConfirmOptions {
  open: boolean;
}

/**
 * Imperative confirmation dialog: `await confirm({...})` resolves true/false.
 * Render `<dialog />` once near the root of the page that uses it.
 */
export function useConfirmDialog() {
  const [state, setState] = useState<ConfirmState>({ open: false, title: "" });
  const resolver = useRef<(value: boolean) => void>(null);

  const confirm = useCallback((options: ConfirmOptions) => {
    setState({ ...options, open: true });
    return new Promise<boolean>((resolve) => {
      resolver.current = resolve;
    });
  }, []);

  const close = useCallback((result: boolean) => {
    setState((s) => ({ ...s, open: false }));
    resolver.current?.(result);
  }, []);

  const dialog = state.open ? (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 p-4 dark:bg-black/70" role="dialog" aria-modal="true">
      <div className="w-full max-w-sm rounded-2xl bg-white p-6 shadow-xl dark:bg-slate-900">
        <h2 className="text-base font-semibold text-slate-900 dark:text-slate-100">{state.title}</h2>
        {state.description && <p className="mt-2 text-sm text-slate-600 dark:text-slate-400">{state.description}</p>}
        <div className="mt-6 flex justify-end gap-3">
          <Button variant="outline" size="sm" onClick={() => close(false)}>
            {state.cancelLabel ?? "Cancel"}
          </Button>
          <Button variant={state.danger ? "danger" : "primary"} size="sm" onClick={() => close(true)}>
            {state.confirmLabel ?? "Confirm"}
          </Button>
        </div>
      </div>
    </div>
  ) : null;

  return { confirm, dialog };
}
