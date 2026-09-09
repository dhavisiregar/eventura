"use client";

import { useEffect, useState } from "react";
import { MessageSquareText } from "lucide-react";
import { api, apiErrorMessage } from "@/lib/api";
import { formatDate } from "@/lib/format";
import { Card } from "@/components/ui/Card";
import { Button } from "@/components/ui/Button";
import { Textarea } from "@/components/ui/Input";
import { StarRating } from "@/components/StarRating";
import { EmptyState } from "@/components/EmptyState";
import type { ApiResponse, Transaction } from "@/lib/types";

export default function ReviewsPage() {
  const [reviewable, setReviewable] = useState<Transaction[]>([]);
  const [loading, setLoading] = useState(true);
  const [activeTx, setActiveTx] = useState<number | null>(null);
  const [rating, setRating] = useState(5);
  const [comment, setComment] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");
  const [successIds, setSuccessIds] = useState<number[]>([]);

  async function load() {
    setLoading(true);
    try {
      const res = await api.get<ApiResponse<Transaction[]>>("/reviews/reviewable");
      setReviewable(res.data.data ?? []);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    load();
  }, []);

  async function submitReview(txId: number) {
    setSubmitting(true);
    setError("");
    try {
      await api.post("/reviews", { transaction_id: txId, rating, comment });
      setSuccessIds((ids) => [...ids, txId]);
      setActiveTx(null);
      setComment("");
      setRating(5);
    } catch (err) {
      setError(apiErrorMessage(err));
    } finally {
      setSubmitting(false);
    }
  }

  const pending = reviewable.filter((tx) => !successIds.includes(tx.id));

  return (
    <div className="mx-auto max-w-3xl px-4 py-8 sm:px-6 lg:px-8">
      <h1 className="mb-2 text-xl font-bold text-slate-900 dark:text-slate-100">Rate your events</h1>
      <p className="mb-6 text-sm text-slate-500 dark:text-slate-400">Share feedback for events you&apos;ve attended.</p>

      {loading ? (
        <p className="text-sm text-slate-500 dark:text-slate-400">Loading…</p>
      ) : pending.length === 0 ? (
        <EmptyState icon={MessageSquareText} title="Nothing to review" description="You have no completed events waiting for a review." />
      ) : (
        <div className="flex flex-col gap-4">
          {pending.map((tx) => (
            <Card key={tx.id} className="p-5">
              <div className="flex items-center justify-between">
                <div>
                  <p className="font-semibold text-slate-900 dark:text-slate-100">{tx.event?.title}</p>
                  <p className="text-xs text-slate-500 dark:text-slate-400">
                    Attended {tx.event ? formatDate(tx.event.end_date) : ""} · {tx.quantity} ticket(s)
                  </p>
                </div>
                {activeTx !== tx.id && (
                  <Button size="sm" variant="outline" onClick={() => setActiveTx(tx.id)}>
                    Write a review
                  </Button>
                )}
              </div>

              {activeTx === tx.id && (
                <div className="mt-4 flex flex-col gap-3 border-t border-slate-100 pt-4 dark:border-slate-800">
                  <StarRating value={rating} onChange={setRating} size="lg" />
                  <Textarea
                    placeholder="How was the event? Any suggestions for improvement?"
                    value={comment}
                    onChange={(e) => setComment(e.target.value)}
                  />
                  {error && <p className="text-xs text-red-600 dark:text-red-400">{error}</p>}
                  <div className="flex justify-end gap-2">
                    <Button variant="ghost" size="sm" onClick={() => setActiveTx(null)}>
                      Cancel
                    </Button>
                    <Button size="sm" loading={submitting} onClick={() => submitReview(tx.id)}>
                      Submit review
                    </Button>
                  </div>
                </div>
              )}
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
