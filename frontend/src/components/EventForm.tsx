"use client";

import { useRef, useState } from "react";
import { useFieldArray, useForm } from "react-hook-form";
import { Plus, Trash2, Upload } from "lucide-react";
import { api, apiErrorMessage } from "@/lib/api";
import { resolveAssetUrl } from "@/lib/assets";
import { Button } from "./ui/Button";
import { Input, Select, Textarea } from "./ui/Input";
import type { Category, EventItem } from "@/lib/types";

export interface EventFormValues {
  title: string;
  category_id: number;
  description: string;
  location: string;
  city: string;
  is_paid: boolean;
  price: number;
  start_date: string;
  end_date: string;
  total_seats: number;
  banner_url: string;
  status?: string;
  ticket_types: { name: string; price: number; quota: number }[];
}

function toLocalInput(dateStr?: string) {
  if (!dateStr) return "";
  const d = new Date(dateStr);
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

export function EventForm({
  mode,
  categories,
  initial,
  onSubmit,
  submitLabel = "Save",
}: {
  mode: "create" | "edit";
  categories: Category[];
  initial?: EventItem;
  onSubmit: (values: EventFormValues) => Promise<void>;
  submitLabel?: string;
}) {
  const [serverError, setServerError] = useState("");
  const [uploading, setUploading] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const {
    register,
    handleSubmit,
    watch,
    setValue,
    control,
    formState: { errors, isSubmitting },
  } = useForm<EventFormValues>({
    defaultValues: {
      title: initial?.title ?? "",
      category_id: initial?.category_id ?? categories[0]?.id ?? 0,
      description: initial?.description ?? "",
      location: initial?.location ?? "",
      city: initial?.city ?? "",
      is_paid: initial?.is_paid ?? false,
      price: initial?.price ?? 0,
      start_date: toLocalInput(initial?.start_date),
      end_date: toLocalInput(initial?.end_date),
      total_seats: initial?.total_seats ?? 100,
      banner_url: initial?.banner_url ?? "",
      status: initial?.status,
      ticket_types: initial?.ticket_types?.map((t) => ({ name: t.name, price: t.price, quota: t.quota })) ?? [],
    },
  });

  const { fields, append, remove } = useFieldArray({ control, name: "ticket_types" });
  const isPaid = watch("is_paid");
  const bannerUrl = watch("banner_url");

  async function handleUpload(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;
    setUploading(true);
    setServerError("");
    try {
      const formData = new FormData();
      formData.append("file", file);
      const res = await api.post<{ data: { url: string } }>("/organizer/uploads/banners", formData, {
        headers: { "Content-Type": "multipart/form-data" },
      });
      setValue("banner_url", res.data.data.url);
    } catch (err) {
      setServerError(apiErrorMessage(err, "Failed to upload image"));
    } finally {
      setUploading(false);
    }
  }

  async function submit(values: EventFormValues) {
    setServerError("");
    try {
      await onSubmit({
        ...values,
        category_id: Number(values.category_id),
        price: values.is_paid ? Number(values.price) : 0,
        total_seats: Number(values.total_seats),
        start_date: new Date(values.start_date).toISOString(),
        end_date: new Date(values.end_date).toISOString(),
        ticket_types: values.ticket_types.map((t) => ({ ...t, price: Number(t.price), quota: Number(t.quota) })),
      });
    } catch (err) {
      setServerError(apiErrorMessage(err, "Failed to save event"));
    }
  }

  return (
    <form onSubmit={handleSubmit(submit)} className="flex flex-col gap-5">
      <div>
        <p className="mb-1.5 text-sm font-medium text-slate-700 dark:text-slate-300">Banner image</p>
        <div className="flex items-center gap-4">
          <div className="h-24 w-40 shrink-0 overflow-hidden rounded-lg border border-dashed border-slate-300 bg-slate-50 dark:border-slate-700 dark:bg-slate-800">
            {bannerUrl && (
              // eslint-disable-next-line @next/next/no-img-element
              <img src={resolveAssetUrl(bannerUrl) ?? ""} alt="Banner preview" className="h-full w-full object-cover" />
            )}
          </div>
          <div>
            <input ref={fileInputRef} type="file" accept="image/*" hidden onChange={handleUpload} />
            <Button type="button" variant="outline" size="sm" loading={uploading} onClick={() => fileInputRef.current?.click()}>
              <Upload className="h-4 w-4" /> Upload image
            </Button>
            <p className="mt-1 text-xs text-slate-400 dark:text-slate-500">JPG, PNG, or WebP. Max 5MB.</p>
          </div>
        </div>
      </div>

      <Input label="Event title" error={errors.title?.message} {...register("title", { required: "Title is required" })} />

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <Select label="Category" error={errors.category_id?.message} {...register("category_id", { required: true })}>
          {categories.map((c) => (
            <option key={c.id} value={c.id}>
              {c.name}
            </option>
          ))}
        </Select>
        <Input label="City" error={errors.city?.message} {...register("city", { required: "City is required" })} />
      </div>

      <Input label="Venue / location" error={errors.location?.message} {...register("location", { required: "Location is required" })} />

      <Textarea label="Description" error={errors.description?.message} {...register("description", { required: "Description is required" })} />

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <Input
          label="Start date & time"
          type="datetime-local"
          error={errors.start_date?.message}
          {...register("start_date", { required: "Start date is required" })}
        />
        <Input
          label="End date & time"
          type="datetime-local"
          error={errors.end_date?.message}
          {...register("end_date", { required: "End date is required" })}
        />
      </div>

      {mode === "create" && (
        <Input
          label="Total seats"
          type="number"
          min={1}
          error={errors.total_seats?.message}
          {...register("total_seats", { required: true, min: 1 })}
        />
      )}

      {mode === "edit" && (
        <Select label="Status" {...register("status")}>
          <option value="draft">Draft</option>
          <option value="published">Published</option>
          <option value="completed">Completed</option>
          <option value="cancelled">Cancelled</option>
        </Select>
      )}

      <label className="flex items-center gap-2 text-sm font-medium text-slate-700 dark:text-slate-300">
        <input type="checkbox" {...register("is_paid")} /> This is a paid event
      </label>

      {isPaid && (
        <Input
          label="Base price (IDR)"
          type="number"
          min={0}
          hint="Used when no specific ticket types are defined below"
          {...register("price", { min: 0 })}
        />
      )}

      {mode === "create" && (
        <div>
          <div className="mb-2 flex items-center justify-between">
            <p className="text-sm font-medium text-slate-700 dark:text-slate-300">Ticket types (optional)</p>
            <Button type="button" size="sm" variant="outline" onClick={() => append({ name: "", price: 0, quota: 1 })}>
              <Plus className="h-3.5 w-3.5" /> Add type
            </Button>
          </div>
          {fields.length === 0 && (
            <p className="text-xs text-slate-400 dark:text-slate-500">
              No ticket types added — attendees will book at the base price above (or free, if unpaid).
            </p>
          )}
          <div className="flex flex-col gap-3">
            {fields.map((field, index) => (
              <div
                key={field.id}
                className="grid grid-cols-1 gap-2 rounded-lg border border-slate-200 p-3 sm:grid-cols-[1fr_auto_auto_auto] dark:border-slate-800"
              >
                <Input placeholder="Name (e.g. VIP)" {...register(`ticket_types.${index}.name`, { required: true })} />
                <Input type="number" placeholder="Price" min={0} className="sm:w-28" {...register(`ticket_types.${index}.price`, { min: 0 })} />
                <Input type="number" placeholder="Quota" min={1} className="sm:w-24" {...register(`ticket_types.${index}.quota`, { min: 1 })} />
                <Button type="button" variant="ghost" size="sm" onClick={() => remove(index)} aria-label="Remove ticket type">
                  <Trash2 className="h-4 w-4 text-red-500 dark:text-red-400" />
                </Button>
              </div>
            ))}
          </div>
        </div>
      )}

      {serverError && <p className="text-sm text-red-600 dark:text-red-400">{serverError}</p>}

      <Button type="submit" loading={isSubmitting} size="lg" className="mt-2 w-full sm:w-auto">
        {submitLabel}
      </Button>
    </form>
  );
}
