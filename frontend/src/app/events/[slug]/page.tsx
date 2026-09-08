"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { CalendarDays, MapPin, Star, Users } from "lucide-react";
import { api } from "@/lib/api";
import { formatDateTime } from "@/lib/format";
import { resolveAssetUrl } from "@/lib/assets";
import { Badge, Card } from "@/components/ui/Card";
import { BookingPanel } from "@/components/BookingPanel";
import { EmptyState } from "@/components/EmptyState";
import { StarRating } from "@/components/StarRating";
import type { EventItem, Review } from "@/lib/types";

interface EventDetailResponse {
  data: { event: EventItem; average_rating: number; review_count: number };
}

export default function EventDetailPage() {
  const params = useParams<{ slug: string }>();
  const [event, setEvent] = useState<EventItem | null>(null);
  const [avgRating, setAvgRating] = useState(0);
  const [reviewCount, setReviewCount] = useState(0);
  const [reviews, setReviews] = useState<Review[]>([]);
  const [loading, setLoading] = useState(true);
  const [notFound, setNotFound] = useState(false);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    api
      .get<EventDetailResponse>(`/events/${params.slug}`)
      .then((res) => {
        if (cancelled) return;
        setEvent(res.data.data.event);
        setAvgRating(res.data.data.average_rating);
        setReviewCount(res.data.data.review_count);
      })
      .catch(() => {
        if (!cancelled) setNotFound(true);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    api
      .get<{ data: Review[] }>(`/events/${params.slug}/reviews`)
      .then((res) => !cancelled && setReviews(res.data.data ?? []))
      .catch(() => {});

    return () => {
      cancelled = true;
    };
  }, [params.slug]);

  if (loading) {
    return <div className="mx-auto max-w-5xl px-4 py-16 text-center text-slate-500">Loading event…</div>;
  }
  if (notFound || !event) {
    return (
      <div className="mx-auto max-w-5xl px-4 py-16">
        <EmptyState title="Event not found" description="This event may have been removed or the link is incorrect." />
      </div>
    );
  }

  const banner = resolveAssetUrl(event.banner_url);

  return (
    <div className="mx-auto max-w-6xl px-4 py-8 sm:px-6 lg:px-8">
      <div className="aspect-[21/9] w-full overflow-hidden rounded-2xl bg-linear-to-br from-indigo-100 to-slate-100">
        {banner && (
          // eslint-disable-next-line @next/next/no-img-element
          <img src={banner} alt={event.title} className="h-full w-full object-cover" />
        )}
      </div>

      <div className="mt-8 grid grid-cols-1 gap-8 lg:grid-cols-3">
        <div className="lg:col-span-2">
          {event.category && <Badge tone="indigo">{event.category.name}</Badge>}
          <h1 className="mt-3 text-2xl font-bold text-slate-900 sm:text-3xl">{event.title}</h1>

          <div className="mt-4 flex flex-wrap gap-x-6 gap-y-2 text-sm text-slate-600">
            <span className="flex items-center gap-1.5">
              <CalendarDays className="h-4 w-4 text-indigo-500" />
              {formatDateTime(event.start_date)}
            </span>
            <span className="flex items-center gap-1.5">
              <MapPin className="h-4 w-4 text-indigo-500" />
              {event.location}, {event.city}
            </span>
            <span className="flex items-center gap-1.5">
              <Users className="h-4 w-4 text-indigo-500" />
              {event.available_seats} / {event.total_seats} seats left
            </span>
            {reviewCount > 0 && (
              <span className="flex items-center gap-1.5">
                <Star className="h-4 w-4 fill-amber-400 text-amber-400" />
                {avgRating.toFixed(1)} ({reviewCount} review{reviewCount === 1 ? "" : "s"})
              </span>
            )}
          </div>

          <Card className="mt-6 p-5">
            <h2 className="mb-2 text-sm font-semibold text-slate-900">About this event</h2>
            <p className="whitespace-pre-line text-sm leading-relaxed text-slate-600">{event.description}</p>
          </Card>

          <div className="mt-8">
            <h2 className="mb-4 text-sm font-semibold text-slate-900">Reviews ({reviewCount})</h2>
            {reviews.length === 0 ? (
              <p className="text-sm text-slate-500">No reviews yet.</p>
            ) : (
              <div className="flex flex-col gap-4">
                {reviews.map((r) => (
                  <Card key={r.id} className="p-4">
                    <div className="flex items-center justify-between">
                      <span className="text-sm font-medium text-slate-900">{r.user?.name ?? "Anonymous"}</span>
                      <StarRating value={r.rating} readOnly size="sm" />
                    </div>
                    {r.comment && <p className="mt-2 text-sm text-slate-600">{r.comment}</p>}
                  </Card>
                ))}
              </div>
            )}
          </div>
        </div>

        <div className="lg:sticky lg:top-20 lg:self-start">
          <BookingPanel event={event} />
        </div>
      </div>
    </div>
  );
}
