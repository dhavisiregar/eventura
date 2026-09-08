"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { CalendarDays, MapPin, Ticket } from "lucide-react";
import { api, apiErrorMessage } from "@/lib/api";
import { formatDateTime, formatIDR, formatRelativeCountdown } from "@/lib/format";
import { Badge, Card } from "@/components/ui/Card";
import { Button } from "@/components/ui/Button";
import { EmptyState } from "@/components/EmptyState";
import { useConfirmDialog } from "@/components/ui/ConfirmDialog";
import type { ApiListResponse, Transaction, TxStatus } from "@/lib/types";

const statusTone: Record<TxStatus, "slate" | "green" | "amber" | "red" | "indigo"> = {
  pending_payment: "amber",
  success: "green",
  expired: "slate",
  cancelled: "slate",
  failed: "red",
};

const statusLabel: Record<TxStatus, string> = {
  pending_payment: "Awaiting payment",
  success: "Confirmed",
  expired: "Expired",
  cancelled: "Cancelled",
  failed: "Failed",
};

export default function MyTicketsPage() {
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [loading, setLoading] = useState(true);
  const [statusFilter, setStatusFilter] = useState("");
  const [error, setError] = useState("");
  const { confirm, dialog } = useConfirmDialog();

  async function load() {
    setLoading(true);
    try {
      const res = await api.get<ApiListResponse<Transaction>>("/transactions/me", {
        params: statusFilter ? { status: statusFilter } : {},
      });
      setTransactions(res.data.data ?? []);
    } catch (err) {
      setError(apiErrorMessage(err));
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [statusFilter]);

  async function handleCancel(tx: Transaction) {
    const ok = await confirm({
      title: "Cancel this order?",
      description: `This will release your ${tx.quantity} ticket(s) for "${tx.event?.title}". This cannot be undone.`,
      confirmLabel: "Cancel order",
      danger: true,
    });
    if (!ok) return;
    try {
      await api.post(`/transactions/${tx.id}/cancel`);
      load();
    } catch (err) {
      setError(apiErrorMessage(err));
    }
  }

  return (
    <div className="mx-auto max-w-4xl px-4 py-8 sm:px-6 lg:px-8">
      {dialog}
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-xl font-bold text-slate-900">My tickets</h1>
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          className="rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm"
        >
          <option value="">All statuses</option>
          <option value="pending_payment">Awaiting payment</option>
          <option value="success">Confirmed</option>
          <option value="expired">Expired</option>
          <option value="cancelled">Cancelled</option>
        </select>
      </div>

      {error && <p className="mb-4 text-sm text-red-600">{error}</p>}

      {loading ? (
        <p className="text-sm text-slate-500">Loading…</p>
      ) : transactions.length === 0 ? (
        <EmptyState
          icon={Ticket}
          title="No tickets yet"
          description="Once you book an event, your orders will show up here."
          action={
            <Link href="/">
              <Button size="sm">Browse events</Button>
            </Link>
          }
        />
      ) : (
        <div className="flex flex-col gap-4">
          {transactions.map((tx) => (
            <Card key={tx.id} className="p-4">
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <div className="mb-1 flex items-center gap-2">
                    <Badge tone={statusTone[tx.status]}>{statusLabel[tx.status]}</Badge>
                    <span className="text-xs text-slate-400">#{tx.invoice_no}</span>
                  </div>
                  <Link href={`/events/${tx.event?.slug}`} className="font-semibold text-slate-900 hover:text-indigo-600">
                    {tx.event?.title}
                  </Link>
                  <div className="mt-1 flex flex-wrap gap-x-4 gap-y-1 text-xs text-slate-500">
                    {tx.event && (
                      <>
                        <span className="flex items-center gap-1">
                          <CalendarDays className="h-3.5 w-3.5" /> {formatDateTime(tx.event.start_date)}
                        </span>
                        <span className="flex items-center gap-1">
                          <MapPin className="h-3.5 w-3.5" /> {tx.event.city}
                        </span>
                      </>
                    )}
                    <span>{tx.quantity} ticket(s)</span>
                  </div>
                </div>
                <div className="text-right">
                  <div className="font-semibold text-slate-900">{formatIDR(tx.total_price)}</div>
                  {tx.status === "pending_payment" && (
                    <div className="text-xs font-medium text-amber-600">{formatRelativeCountdown(tx.payment_deadline)}</div>
                  )}
                </div>
              </div>

              {tx.status === "pending_payment" && (
                <div className="mt-3 flex gap-2 border-t border-slate-100 pt-3">
                  {tx.midtrans_redirect_url && (
                    <a href={tx.midtrans_redirect_url} target="_blank" rel="noreferrer">
                      <Button size="sm">Continue payment</Button>
                    </a>
                  )}
                  <Button size="sm" variant="outline" onClick={() => handleCancel(tx)}>
                    Cancel order
                  </Button>
                </div>
              )}
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
