import Link from "next/link";
import { CalendarDays, MapPin, Ticket } from "lucide-react";
import { Card, Badge } from "./ui/Card";
import { formatDate, formatIDR } from "@/lib/format";
import { resolveAssetUrl } from "@/lib/assets";
import type { EventItem } from "@/lib/types";

export function EventCard({ event }: { event: EventItem }) {
  const banner = resolveAssetUrl(event.banner_url);
  const soldOut = event.available_seats <= 0;

  return (
    <Link href={`/events/${event.slug}`} className="group block h-full">
      <Card className="flex h-full flex-col overflow-hidden transition-shadow hover:shadow-md">
        <div className="relative aspect-[16/9] w-full overflow-hidden bg-linear-to-br from-indigo-100 to-slate-100 dark:from-indigo-500/10 dark:to-slate-800">
          {banner ? (
            // eslint-disable-next-line @next/next/no-img-element
            <img
              src={banner}
              alt={event.title}
              className="h-full w-full object-cover transition-transform duration-300 group-hover:scale-105"
            />
          ) : (
            <div className="flex h-full w-full items-center justify-center text-indigo-300 dark:text-indigo-500/40">
              <Ticket className="h-10 w-10" strokeWidth={1.5} />
            </div>
          )}
          <div className="absolute left-3 top-3 flex gap-2">
            {event.category && <Badge tone="indigo">{event.category.name}</Badge>}
            {soldOut && <Badge tone="red">Sold out</Badge>}
          </div>
        </div>

        <div className="flex flex-1 flex-col gap-2 p-4">
          <h3 className="line-clamp-2 text-sm font-semibold text-slate-900 dark:text-slate-100">{event.title}</h3>

          <div className="flex items-center gap-1.5 text-xs text-slate-500 dark:text-slate-400">
            <CalendarDays className="h-3.5 w-3.5 shrink-0" />
            <span>{formatDate(event.start_date)}</span>
          </div>
          <div className="flex items-center gap-1.5 text-xs text-slate-500 dark:text-slate-400">
            <MapPin className="h-3.5 w-3.5 shrink-0" />
            <span className="line-clamp-1">
              {event.location}, {event.city}
            </span>
          </div>

          <div className="mt-auto pt-2 text-sm font-semibold text-indigo-600 dark:text-indigo-400">
            {event.is_paid ? formatIDR(event.price) : "Free"}
          </div>
        </div>
      </Card>
    </Link>
  );
}
