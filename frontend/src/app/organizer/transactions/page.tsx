"use client";

import { useEffect, useState } from "react";
import { Receipt } from "lucide-react";
import { api, apiErrorMessage } from "@/lib/api";
import { formatDateTime, formatIDR } from "@/lib/format";
import { Card, Badge } from "@/components/ui/Card";
import { Select } from "@/components/ui/Input";
import { Pagination } from "@/components/Pagination";
import { EmptyState } from "@/components/EmptyState";
import type { ApiListResponse, Transaction, TxStatus } from "@/lib/types";

const statusTone: Record<TxStatus, "slate" | "green" | "amber" | "red" | "indigo"> = {
  pending_payment: "amber",
  success: "green",
  expired: "slate",
  cancelled: "slate",
  failed: "red",
};

export default function OrganizerTransactionsPage() {
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [status, setStatus] = useState("");
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    setLoading(true);
    api
      .get<ApiListResponse<Transaction>>("/organizer/transactions", {
        params: { page, limit: 15, ...(status ? { status } : {}) },
      })
      .then((res) => {
        setTransactions(res.data.data ?? []);
        setTotalPages(res.data.meta.total_pages);
      })
      .catch((err) => setError(apiErrorMessage(err)))
      .finally(() => setLoading(false));
  }, [page, status]);

  return (
    <div>
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-xl font-bold text-slate-900">Transactions</h1>
        <Select
          value={status}
          onChange={(e) => {
            setStatus(e.target.value);
            setPage(1);
          }}
          className="w-auto"
        >
          <option value="">All statuses</option>
          <option value="pending_payment">Awaiting payment</option>
          <option value="success">Success</option>
          <option value="expired">Expired</option>
          <option value="cancelled">Cancelled</option>
          <option value="failed">Failed</option>
        </Select>
      </div>

      {error && <p className="mb-4 text-sm text-red-600">{error}</p>}

      {loading ? (
        <p className="text-sm text-slate-500">Loading…</p>
      ) : transactions.length === 0 ? (
        <EmptyState icon={Receipt} title="No transactions found" description="Try a different status filter." />
      ) : (
        <>
          <Card className="overflow-x-auto p-0">
            <table className="w-full min-w-[640px] text-left text-sm">
              <thead>
                <tr className="border-b border-slate-100 text-xs uppercase text-slate-400">
                  <th className="px-4 py-3 font-medium">Invoice</th>
                  <th className="px-4 py-3 font-medium">Event</th>
                  <th className="px-4 py-3 font-medium">Buyer</th>
                  <th className="px-4 py-3 font-medium">Total</th>
                  <th className="px-4 py-3 font-medium">Status</th>
                  <th className="px-4 py-3 font-medium">Date</th>
                </tr>
              </thead>
              <tbody>
                {transactions.map((tx) => (
                  <tr key={tx.id} className="border-b border-slate-50 last:border-0">
                    <td className="px-4 py-3 font-mono text-xs text-slate-500">{tx.invoice_no}</td>
                    <td className="px-4 py-3 text-slate-800">{tx.event?.title}</td>
                    <td className="px-4 py-3 text-slate-600">{tx.user?.name}</td>
                    <td className="px-4 py-3 font-medium text-slate-900">{formatIDR(tx.total_price)}</td>
                    <td className="px-4 py-3">
                      <Badge tone={statusTone[tx.status]}>{tx.status.replace("_", " ")}</Badge>
                    </td>
                    <td className="px-4 py-3 text-slate-500">{formatDateTime(tx.created_at)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </Card>
          <div className="mt-8">
            <Pagination page={page} totalPages={totalPages} onChange={setPage} />
          </div>
        </>
      )}
    </div>
  );
}
