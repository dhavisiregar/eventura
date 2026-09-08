"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { api, apiErrorMessage } from "@/lib/api";
import { Card } from "@/components/ui/Card";
import { EventForm, EventFormValues } from "@/components/EventForm";
import type { Category, EventItem } from "@/lib/types";

export default function EditEventPage() {
  const params = useParams<{ id: string }>();
  const router = useRouter();
  const [categories, setCategories] = useState<Category[]>([]);
  const [event, setEvent] = useState<EventItem | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    api.get<{ data: Category[] }>("/categories").then((res) => setCategories(res.data.data));
  }, []);

  useEffect(() => {
    api
      .get<{ data: EventItem }>(`/organizer/events/${params.id}`)
      .then((res) => setEvent(res.data.data))
      .catch((err) => setError(apiErrorMessage(err, "Event not found or you do not have access to it.")));
  }, [params.id]);

  async function handleSubmit(values: EventFormValues) {
    await api.put(`/organizer/events/${params.id}`, values);
    router.push("/organizer/events");
  }

  return (
    <div>
      <h1 className="mb-6 text-xl font-bold text-slate-900">Edit event</h1>
      {error && <p className="text-sm text-red-600">{error}</p>}
      <Card className="max-w-2xl p-6">
        {event && categories.length > 0 && (
          <EventForm mode="edit" categories={categories} initial={event} onSubmit={handleSubmit} submitLabel="Save changes" />
        )}
      </Card>
    </div>
  );
}
