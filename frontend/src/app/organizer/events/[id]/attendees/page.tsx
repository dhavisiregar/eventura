"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { Users } from "lucide-react";
import { api, apiErrorMessage } from "@/lib/api";
import { formatDate, formatIDR } from "@/lib/format";
import { Card } from "@/components/ui/Card";
import { EmptyState } from "@/components/EmptyState";
import type { EventItem, Transaction } from "@/lib/types";

export default function AttendeesPage() {
  const params = useParams<{ id: string }>();
  const [event, setEvent] = useState<EventItem | null>(null);
  const [attendees, setAttendees] = useState<Transaction[]>([]);
  const [error, setError] = useState("");

  useEffect(() => {
    Promise.all([
      api.get<{ data: EventItem }>(`/organizer/events/${params.id}`),
      api.get<{ data: Transaction[] }>(`/organizer/events/${params.id}/attendees`),
    ])
      .then(([eventRes, attendeesRes]) => {
        setEvent(eventRes.data.data);
        setAttendees(attendeesRes.data.data ?? []);
      })
      .catch((err) => setError(apiErrorMessage(err)));
  }, [params.id]);

  const totalTickets = attendees.reduce((sum, tx) => sum + tx.quantity, 0);

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-xl font-bold text-slate-900">Attendees</h1>
        {event && (
          <p className="text-sm text-slate-500">
            {event.title} · {totalTickets} ticket(s) sold across {attendees.length} order(s)
          </p>
        )}
      </div>

      {error && <p className="mb-4 text-sm text-red-600">{error}</p>}

      {attendees.length === 0 ? (
        <EmptyState icon={Users} title="No attendees yet" description="Confirmed orders for this event will appear here." />
      ) : (
        <Card className="overflow-x-auto p-0">
          <table className="w-full min-w-[520px] text-left text-sm">
            <thead>
              <tr className="border-b border-slate-100 text-xs uppercase text-slate-400">
                <th className="px-4 py-3 font-medium">Attendee</th>
                <th className="px-4 py-3 font-medium">Qty</th>
                <th className="px-4 py-3 font-medium">Paid</th>
                <th className="px-4 py-3 font-medium">Booked on</th>
              </tr>
            </thead>
            <tbody>
              {attendees.map((tx) => (
                <tr key={tx.id} className="border-b border-slate-50 last:border-0">
                  <td className="px-4 py-3">
                    <p className="font-medium text-slate-800">{tx.user?.name}</p>
                    <p className="text-xs text-slate-400">{tx.user?.email}</p>
                  </td>
                  <td className="px-4 py-3 text-slate-600">{tx.quantity}</td>
                  <td className="px-4 py-3 font-medium text-slate-900">{formatIDR(tx.total_price)}</td>
                  <td className="px-4 py-3 text-slate-500">{formatDate(tx.created_at)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </Card>
      )}
    </div>
  );
}
