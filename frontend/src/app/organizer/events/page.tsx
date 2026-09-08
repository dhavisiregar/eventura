"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { CalendarPlus, PenSquare, Ticket, Trash2, Users } from "lucide-react";
import { api, apiErrorMessage } from "@/lib/api";
import { formatDate, formatIDR } from "@/lib/format";
import { Badge, Card } from "@/components/ui/Card";
import { Button } from "@/components/ui/Button";
import { Pagination } from "@/components/Pagination";
import { EmptyState } from "@/components/EmptyState";
import { useConfirmDialog } from "@/components/ui/ConfirmDialog";
import type { ApiListResponse, EventItem, EventStatus } from "@/lib/types";

const statusTone: Record<EventStatus, "slate" | "green" | "amber" | "red" | "indigo"> = {
  draft: "slate",
  published: "green",
  completed: "indigo",
  cancelled: "red",
};

export default function OrganizerEventsPage() {
  const [events, setEvents] = useState<EventItem[]>([]);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const { confirm, dialog } = useConfirmDialog();

  async function load() {
    setLoading(true);
    try {
      const res = await api.get<ApiListResponse<EventItem>>("/organizer/events", { params: { page, limit: 10 } });
      setEvents(res.data.data ?? []);
      setTotalPages(res.data.meta.total_pages);
    } catch (err) {
      setError(apiErrorMessage(err));
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page]);

  async function handleDelete(event: EventItem) {
    const ok = await confirm({
      title: "Delete this event?",
      description: `"${event.title}" will be permanently removed. This only works if the event has no transactions yet.`,
      confirmLabel: "Delete",
      danger: true,
    });
    if (!ok) return;
    try {
      await api.delete(`/organizer/events/${event.id}`);
      load();
    } catch (err) {
      setError(apiErrorMessage(err));
    }
  }

  return (
    <div>
      {dialog}
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-xl font-bold text-slate-900">My events</h1>
        <Link href="/organizer/events/new">
          <Button size="sm">
            <CalendarPlus className="h-4 w-4" /> Create event
          </Button>
        </Link>
      </div>

      {error && <p className="mb-4 text-sm text-red-600">{error}</p>}

      {loading ? (
        <p className="text-sm text-slate-500">Loading…</p>
      ) : events.length === 0 ? (
        <EmptyState
          title="No events yet"
          description="Create your first event to start selling tickets."
          action={
            <Link href="/organizer/events/new">
              <Button size="sm">Create event</Button>
            </Link>
          }
        />
      ) : (
        <>
          <div className="flex flex-col gap-4">
            {events.map((event) => (
              <Card key={event.id} className="p-4">
                <div className="flex flex-wrap items-start justify-between gap-3">
                  <div>
                    <div className="mb-1 flex items-center gap-2">
                      <Badge tone={statusTone[event.status]}>{event.status}</Badge>
                      {event.category && <Badge tone="slate">{event.category.name}</Badge>}
                    </div>
                    <p className="font-semibold text-slate-900">{event.title}</p>
                    <p className="text-xs text-slate-500">
                      {formatDate(event.start_date)} · {event.city} · {event.available_seats}/{event.total_seats} seats left
                    </p>
                  </div>
                  <div className="text-right font-semibold text-indigo-600">
                    {event.is_paid ? formatIDR(event.price) : "Free"}
                  </div>
                </div>
                <div className="mt-3 flex flex-wrap gap-2 border-t border-slate-100 pt-3">
                  <Link href={`/organizer/events/${event.id}/edit`}>
                    <Button size="sm" variant="outline">
                      <PenSquare className="h-3.5 w-3.5" /> Edit
                    </Button>
                  </Link>
                  <Link href={`/organizer/events/${event.id}/vouchers`}>
                    <Button size="sm" variant="outline">
                      <Ticket className="h-3.5 w-3.5" /> Vouchers
                    </Button>
                  </Link>
                  <Link href={`/organizer/events/${event.id}/attendees`}>
                    <Button size="sm" variant="outline">
                      <Users className="h-3.5 w-3.5" /> Attendees
                    </Button>
                  </Link>
                  <Button size="sm" variant="ghost" className="text-red-600 hover:bg-red-50" onClick={() => handleDelete(event)}>
                    <Trash2 className="h-3.5 w-3.5" /> Delete
                  </Button>
                </div>
              </Card>
            ))}
          </div>
          <div className="mt-8">
            <Pagination page={page} totalPages={totalPages} onChange={setPage} />
          </div>
        </>
      )}
    </div>
  );
}
