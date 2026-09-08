"use client";

import { useEffect, useState } from "react";
import {
  Bar,
  BarChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import { CalendarCheck, Receipt, Ticket, Wallet } from "lucide-react";
import { api } from "@/lib/api";
import { formatIDR } from "@/lib/format";
import { Card } from "@/components/ui/Card";
import { Select } from "@/components/ui/Input";
import { EmptyState } from "@/components/EmptyState";
import type { ApiResponse, DashboardStats } from "@/lib/types";

const currentYear = new Date().getFullYear();
const years = Array.from({ length: 6 }, (_, i) => currentYear - i);
const months = Array.from({ length: 12 }, (_, i) => i + 1);

export default function OrganizerDashboardPage() {
  const [range, setRange] = useState<"year" | "month" | "day">("year");
  const [year, setYear] = useState(currentYear);
  const [month, setMonth] = useState(new Date().getMonth() + 1);
  const [day, setDay] = useState(new Date().getDate());
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    api
      .get<ApiResponse<DashboardStats>>("/organizer/dashboard/stats", { params: { range, year, month, day } })
      .then((res) =>
        setStats({ ...res.data.data, series: res.data.data.series ?? [], top_events: res.data.data.top_events ?? [] })
      )
      .finally(() => setLoading(false));
  }, [range, year, month, day]);

  const tiles = stats
    ? [
        { label: "Total events", value: stats.summary.total_events, icon: CalendarCheck },
        { label: "Revenue in range", value: formatIDR(stats.summary.total_revenue), icon: Wallet },
        { label: "Tickets sold", value: stats.summary.total_tickets, icon: Ticket },
        { label: "Orders", value: stats.summary.total_orders, icon: Receipt },
      ]
    : [];

  return (
    <div>
      <div className="mb-6 flex flex-wrap items-center justify-between gap-3">
        <h1 className="text-xl font-bold text-slate-900">Dashboard</h1>
        <div className="flex flex-wrap gap-2">
          <Select value={range} onChange={(e) => setRange(e.target.value as typeof range)} className="w-auto">
            <option value="year">Per year (by month)</option>
            <option value="month">Per month (by day)</option>
            <option value="day">Per day (by hour)</option>
          </Select>
          <Select value={year} onChange={(e) => setYear(Number(e.target.value))} className="w-auto">
            {years.map((y) => (
              <option key={y} value={y}>
                {y}
              </option>
            ))}
          </Select>
          {(range === "month" || range === "day") && (
            <Select value={month} onChange={(e) => setMonth(Number(e.target.value))} className="w-auto">
              {months.map((m) => (
                <option key={m} value={m}>
                  {new Date(2000, m - 1, 1).toLocaleString("en-US", { month: "long" })}
                </option>
              ))}
            </Select>
          )}
          {range === "day" && (
            <Select value={day} onChange={(e) => setDay(Number(e.target.value))} className="w-auto">
              {Array.from({ length: 31 }, (_, i) => i + 1).map((d) => (
                <option key={d} value={d}>
                  {d}
                </option>
              ))}
            </Select>
          )}
        </div>
      </div>

      {loading || !stats ? (
        <p className="text-sm text-slate-500">Loading…</p>
      ) : (
        <>
          <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
            {tiles.map((t) => (
              <Card key={t.label} className="p-4">
                <div className="flex items-center gap-2 text-xs font-medium text-slate-500">
                  <t.icon className="h-4 w-4 text-indigo-500" /> {t.label}
                </div>
                <p className="mt-2 text-xl font-bold text-slate-900">{t.value}</p>
              </Card>
            ))}
          </div>

          <Card className="mt-6 p-5">
            <h2 className="mb-4 text-sm font-semibold text-slate-900">Revenue over time</h2>
            {stats.series.length === 0 ? (
              <EmptyState title="No transactions in this range" description="Try a different year, month, or day." />
            ) : (
              <div className="h-72 w-full">
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={stats.series} margin={{ left: 8, right: 8 }}>
                    <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#e2e8f0" />
                    <XAxis dataKey="label" tick={{ fontSize: 11, fill: "#64748b" }} axisLine={false} tickLine={false} />
                    <YAxis
                      tick={{ fontSize: 11, fill: "#64748b" }}
                      axisLine={false}
                      tickLine={false}
                      tickFormatter={(v) => new Intl.NumberFormat("id-ID", { notation: "compact" }).format(v)}
                    />
                    <Tooltip
                      formatter={(value, name) => [
                        name === "revenue" ? formatIDR(Number(value)) : value,
                        name === "revenue" ? "Revenue" : "Tickets sold",
                      ]}
                      contentStyle={{ borderRadius: 12, borderColor: "#e2e8f0", fontSize: 12 }}
                    />
                    <Bar dataKey="revenue" fill="#6366f1" radius={[6, 6, 0, 0]} maxBarSize={40} />
                  </BarChart>
                </ResponsiveContainer>
              </div>
            )}
          </Card>

          <Card className="mt-6 p-5">
            <h2 className="mb-4 text-sm font-semibold text-slate-900">Top events</h2>
            {stats.top_events.length === 0 ? (
              <p className="text-sm text-slate-500">No sales in this range yet.</p>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-left text-sm">
                  <thead>
                    <tr className="border-b border-slate-100 text-xs uppercase text-slate-400">
                      <th className="py-2 font-medium">Event</th>
                      <th className="py-2 font-medium">Tickets sold</th>
                      <th className="py-2 font-medium">Revenue</th>
                    </tr>
                  </thead>
                  <tbody>
                    {stats.top_events.map((e) => (
                      <tr key={e.title} className="border-b border-slate-50 last:border-0">
                        <td className="py-2.5 text-slate-800">{e.title}</td>
                        <td className="py-2.5 text-slate-600">{e.tickets_sold}</td>
                        <td className="py-2.5 font-medium text-slate-900">{formatIDR(e.revenue)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </Card>
        </>
      )}
    </div>
  );
}
