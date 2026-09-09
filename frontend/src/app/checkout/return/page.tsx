"use client";

import { Suspense, useEffect, useState } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { CheckCircle2, Loader2, XCircle } from "lucide-react";
import { api, apiErrorMessage } from "@/lib/api";
import { Button } from "@/components/ui/Button";
import { Card } from "@/components/ui/Card";

type SyncState = "checking" | "success" | "pending" | "error";

function ReturnContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const orderId = searchParams.get("order_id");

  const [state, setState] = useState<SyncState>("checking");
  const [message, setMessage] = useState("");

  useEffect(() => {
    if (!orderId) {
      setState("error");
      setMessage("Missing order reference.");
      return;
    }

    api
      .post<{ data: { status: string } }>(`/transactions/sync/${orderId}`)
      .then((res) => {
        const status = res.data.data.status;
        if (status === "success") {
          setState("success");
        } else if (status === "pending_payment") {
          setState("pending");
        } else {
          setState("error");
          setMessage(`Payment ${status.replace("_", " ")}.`);
        }
      })
      .catch((err) => {
        setState("error");
        setMessage(apiErrorMessage(err, "Could not confirm your payment status."));
      });
  }, [orderId]);

  useEffect(() => {
    if (state === "success" || state === "pending") {
      const timer = setTimeout(() => router.push("/my-tickets"), 2000);
      return () => clearTimeout(timer);
    }
  }, [state, router]);

  return (
    <div className="mx-auto flex min-h-[70vh] max-w-md flex-col items-center justify-center px-4 text-center">
      <Card className="w-full p-8">
        {state === "checking" && (
          <>
            <Loader2 className="mx-auto h-10 w-10 animate-spin text-indigo-500 dark:text-indigo-400" />
            <h1 className="mt-4 text-lg font-semibold text-slate-900 dark:text-slate-100">Confirming your payment…</h1>
            <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">Please wait a moment, this won&apos;t take long.</p>
          </>
        )}
        {state === "success" && (
          <>
            <CheckCircle2 className="mx-auto h-10 w-10 text-emerald-500 dark:text-emerald-400" />
            <h1 className="mt-4 text-lg font-semibold text-slate-900 dark:text-slate-100">Payment confirmed!</h1>
            <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">Redirecting you to your tickets…</p>
          </>
        )}
        {state === "pending" && (
          <>
            <Loader2 className="mx-auto h-10 w-10 animate-spin text-amber-500 dark:text-amber-400" />
            <h1 className="mt-4 text-lg font-semibold text-slate-900 dark:text-slate-100">Payment still processing</h1>
            <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">We&apos;ll keep checking — redirecting you to your tickets…</p>
          </>
        )}
        {state === "error" && (
          <>
            <XCircle className="mx-auto h-10 w-10 text-red-500 dark:text-red-400" />
            <h1 className="mt-4 text-lg font-semibold text-slate-900 dark:text-slate-100">We couldn&apos;t confirm this payment</h1>
            <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">{message}</p>
            <Link href="/my-tickets" className="mt-6 inline-block">
              <Button size="sm">Go to my tickets</Button>
            </Link>
          </>
        )}
      </Card>
    </div>
  );
}

export default function CheckoutReturnPage() {
  return (
    <Suspense>
      <ReturnContent />
    </Suspense>
  );
}
