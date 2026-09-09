"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { useForm } from "react-hook-form";
import { Plus, Ticket, Trash2 } from "lucide-react";
import { api, apiErrorMessage } from "@/lib/api";
import { formatDate, formatIDR } from "@/lib/format";
import { Card, Badge } from "@/components/ui/Card";
import { Button } from "@/components/ui/Button";
import { Input, Select } from "@/components/ui/Input";
import { EmptyState } from "@/components/EmptyState";
import { useConfirmDialog } from "@/components/ui/ConfirmDialog";
import type { EventItem, Voucher } from "@/lib/types";

interface VoucherFormValues {
  code: string;
  discount_type: "amount" | "percentage";
  discount_value: number;
  quota: number;
  start_date: string;
  end_date: string;
}

export default function VouchersPage() {
  const params = useParams<{ id: string }>();
  const [event, setEvent] = useState<EventItem | null>(null);
  const [vouchers, setVouchers] = useState<Voucher[]>([]);
  const [error, setError] = useState("");
  const [creating, setCreating] = useState(false);
  const { confirm, dialog } = useConfirmDialog();

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<VoucherFormValues>({
    defaultValues: { discount_type: "percentage", discount_value: 10, quota: 50 },
  });

  async function load() {
    try {
      const [eventRes, vouchersRes] = await Promise.all([
        api.get<{ data: EventItem }>(`/organizer/events/${params.id}`),
        api.get<{ data: Voucher[] }>(`/organizer/events/${params.id}/vouchers`),
      ]);
      setEvent(eventRes.data.data);
      setVouchers(vouchersRes.data.data ?? []);
    } catch (err) {
      setError(apiErrorMessage(err));
    }
  }

  useEffect(() => {
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [params.id]);

  async function onCreate(values: VoucherFormValues) {
    setError("");
    try {
      await api.post(`/organizer/events/${params.id}/vouchers`, {
        ...values,
        code: values.code.toUpperCase(),
        discount_value: Number(values.discount_value),
        quota: Number(values.quota),
        start_date: new Date(values.start_date).toISOString(),
        end_date: new Date(values.end_date).toISOString(),
      });
      reset({ code: "", discount_type: "percentage", discount_value: 10, quota: 50, start_date: "", end_date: "" });
      setCreating(false);
      load();
    } catch (err) {
      setError(apiErrorMessage(err));
    }
  }

  async function handleDelete(voucher: Voucher) {
    const ok = await confirm({
      title: "Delete this voucher?",
      description: `"${voucher.code}" will no longer be usable at checkout.`,
      confirmLabel: "Delete",
      danger: true,
    });
    if (!ok) return;
    try {
      await api.delete(`/organizer/vouchers/${voucher.id}`);
      load();
    } catch (err) {
      setError(apiErrorMessage(err));
    }
  }

  return (
    <div>
      {dialog}
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-slate-900 dark:text-slate-100">Vouchers</h1>
          {event && <p className="text-sm text-slate-500 dark:text-slate-400">for {event.title}</p>}
        </div>
        <Button size="sm" onClick={() => setCreating((v) => !v)}>
          <Plus className="h-4 w-4" /> {creating ? "Cancel" : "New voucher"}
        </Button>
      </div>

      {error && <p className="mb-4 text-sm text-red-600 dark:text-red-400">{error}</p>}

      {creating && (
        <Card className="mb-6 max-w-xl p-5">
          <form onSubmit={handleSubmit(onCreate)} className="flex flex-col gap-4">
            <Input
              label="Voucher code"
              placeholder="EARLYBIRD"
              error={errors.code?.message}
              {...register("code", { required: "Code is required", minLength: 3 })}
            />
            <div className="grid grid-cols-2 gap-3">
              <Select label="Discount type" {...register("discount_type")}>
                <option value="percentage">Percentage (%)</option>
                <option value="amount">Fixed amount (IDR)</option>
              </Select>
              <Input label="Discount value" type="number" min={1} {...register("discount_value", { required: true, min: 1 })} />
            </div>
            <Input label="Quota (max redemptions)" type="number" min={1} {...register("quota", { required: true, min: 1 })} />
            <div className="grid grid-cols-2 gap-3">
              <Input
                label="Valid from"
                type="datetime-local"
                error={errors.start_date?.message}
                {...register("start_date", { required: "Required" })}
              />
              <Input
                label="Valid until"
                type="datetime-local"
                error={errors.end_date?.message}
                {...register("end_date", { required: "Required" })}
              />
            </div>
            <Button type="submit" loading={isSubmitting} className="w-full sm:w-auto">
              Create voucher
            </Button>
          </form>
        </Card>
      )}

      {vouchers.length === 0 ? (
        <EmptyState icon={Ticket} title="No vouchers yet" description="Create limited-quota discount codes for this event." />
      ) : (
        <div className="flex flex-col gap-3">
          {vouchers.map((v) => (
            <Card key={v.id} className="flex items-center justify-between p-4">
              <div>
                <div className="flex items-center gap-2">
                  <code className="text-sm font-semibold text-slate-900 dark:text-slate-100">{v.code}</code>
                  <Badge tone="indigo">
                    {v.discount_type === "percentage" ? `${v.discount_value}% off` : `${formatIDR(v.discount_value)} off`}
                  </Badge>
                </div>
                <p className="mt-1 text-xs text-slate-500 dark:text-slate-400">
                  {v.used_count}/{v.quota} redeemed · Valid {formatDate(v.start_date)} – {formatDate(v.end_date)}
                </p>
              </div>
              <Button
                variant="ghost"
                size="sm"
                className="text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-500/10"
                onClick={() => handleDelete(v)}
              >
                <Trash2 className="h-4 w-4" />
              </Button>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
