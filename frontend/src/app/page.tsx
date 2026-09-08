"use client";

import { useEffect, useState } from "react";
import { SlidersHorizontal } from "lucide-react";
import { api } from "@/lib/api";
import { useDebounce } from "@/hooks/useDebounce";
import { SearchBar } from "@/components/SearchBar";
import { EventCard } from "@/components/EventCard";
import { Pagination } from "@/components/Pagination";
import { EmptyState } from "@/components/EmptyState";
import { Select } from "@/components/ui/Input";
import type { ApiListResponse, Category, EventItem, Pagination as PaginationMeta } from "@/lib/types";

const LIMIT = 12;

export default function HomePage() {
  const [search, setSearch] = useState("");
  const [city, setCity] = useState("");
  const [category, setCategory] = useState("");
  const [priceFilter, setPriceFilter] = useState<"" | "true" | "false">("");
  const [sort, setSort] = useState("");
  const [page, setPage] = useState(1);

  const debouncedSearch = useDebounce(search, 400);
  const debouncedCity = useDebounce(city, 400);

  const [categories, setCategories] = useState<Category[]>([]);
  const [events, setEvents] = useState<EventItem[]>([]);
  const [meta, setMeta] = useState<PaginationMeta | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    api.get<{ data: Category[] }>("/categories").then((res) => setCategories(res.data.data));
  }, []);

  // Reset to page 1 whenever a filter changes.
  useEffect(() => {
    setPage(1);
  }, [debouncedSearch, debouncedCity, category, priceFilter, sort]);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError("");

    const params: Record<string, string | number> = { page, limit: LIMIT };
    if (debouncedSearch) params.q = debouncedSearch;
    if (debouncedCity) params.city = debouncedCity;
    if (category) params.category = category;
    if (priceFilter) params.is_paid = priceFilter;
    if (sort) params.sort = sort;

    api
      .get<ApiListResponse<EventItem>>("/events", { params })
      .then((res) => {
        if (cancelled) return;
        setEvents(res.data.data ?? []);
        setMeta(res.data.meta);
      })
      .catch(() => {
        if (!cancelled) setError("Failed to load events. Please try again.");
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [page, debouncedSearch, debouncedCity, category, priceFilter, sort]);

  const hasActiveFilters = Boolean(search || city || category || priceFilter);

  return (
    <div>
      <section className="border-b border-slate-200 bg-linear-to-b from-indigo-50 to-slate-50">
        <div className="mx-auto max-w-7xl px-4 py-14 sm:px-6 lg:px-8">
          <h1 className="text-3xl font-bold tracking-tight text-slate-900 sm:text-4xl">
            Find your next unforgettable event
          </h1>
          <p className="mt-3 max-w-2xl text-slate-600">
            Discover concerts, workshops, and conferences near you — and book tickets in minutes.
          </p>
          <div className="mt-6 max-w-xl">
            <SearchBar value={search} onChange={setSearch} />
          </div>
        </div>
      </section>

      <section className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
        <div className="mb-6 flex flex-col gap-3 rounded-2xl border border-slate-200 bg-white p-4 sm:flex-row sm:items-center">
          <div className="flex items-center gap-2 text-sm font-medium text-slate-500">
            <SlidersHorizontal className="h-4 w-4" /> Filters
          </div>
          <div className="grid flex-1 grid-cols-2 gap-3 sm:grid-cols-4">
            <Select aria-label="Category" value={category} onChange={(e) => setCategory(e.target.value)}>
              <option value="">All categories</option>
              {categories.map((c) => (
                <option key={c.id} value={c.slug}>
                  {c.name}
                </option>
              ))}
            </Select>
            <input
              aria-label="City"
              placeholder="City"
              value={city}
              onChange={(e) => setCity(e.target.value)}
              className="w-full rounded-lg border border-slate-300 bg-white px-3.5 py-2.5 text-sm text-slate-900 placeholder:text-slate-400 focus:border-indigo-500 focus:outline-none focus:ring-2 focus:ring-indigo-100"
            />
            <Select
              aria-label="Price"
              value={priceFilter}
              onChange={(e) => setPriceFilter(e.target.value as "" | "true" | "false")}
            >
              <option value="">Free & paid</option>
              <option value="false">Free only</option>
              <option value="true">Paid only</option>
            </Select>
            <Select aria-label="Sort" value={sort} onChange={(e) => setSort(e.target.value)}>
              <option value="">Soonest first</option>
              <option value="newest">Newest</option>
              <option value="price_asc">Price: low to high</option>
              <option value="price_desc">Price: high to low</option>
            </Select>
          </div>
        </div>

        {loading ? (
          <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
            {Array.from({ length: 8 }).map((_, i) => (
              <div key={i} className="h-72 animate-pulse rounded-2xl bg-slate-100" />
            ))}
          </div>
        ) : error ? (
          <EmptyState title="Couldn't load events" description={error} />
        ) : events.length === 0 ? (
          <EmptyState
            title="No events found"
            description={
              hasActiveFilters
                ? "Try adjusting your search or filters."
                : "There are no upcoming events published yet — check back soon."
            }
          />
        ) : (
          <>
            <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
              {events.map((event) => (
                <EventCard key={event.id} event={event} />
              ))}
            </div>
            {meta && (
              <div className="mt-10">
                <Pagination page={meta.page} totalPages={meta.total_pages} onChange={setPage} />
              </div>
            )}
          </>
        )}
      </section>
    </div>
  );
}
