"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { api } from "@/lib/api";
import { Card } from "@/components/ui/Card";
import { EventForm, EventFormValues } from "@/components/EventForm";
import type { ApiResponse, Category, EventItem } from "@/lib/types";

export default function CreateEventPage() {
  const router = useRouter();
  const [categories, setCategories] = useState<Category[]>([]);

  useEffect(() => {
    api.get<{ data: Category[] }>("/categories").then((res) => setCategories(res.data.data));
  }, []);

  async function handleSubmit(values: EventFormValues) {
    const res = await api.post<ApiResponse<EventItem>>("/organizer/events", values);
    router.push(`/organizer/events/${res.data.data.id}/edit?created=1`);
  }

  return (
    <div>
      <h1 className="mb-6 text-xl font-bold text-slate-900 dark:text-slate-100">Create event</h1>
      <Card className="max-w-2xl p-6">
        {categories.length > 0 && (
          <EventForm mode="create" categories={categories} onSubmit={handleSubmit} submitLabel="Publish event" />
        )}
      </Card>
    </div>
  );
}
