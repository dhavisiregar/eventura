"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Ticket as TicketIcon, Tag } from "lucide-react";
import { useAuth } from "@/context/AuthContext";
import { api, apiErrorMessage } from "@/lib/api";
import { formatIDR } from "@/lib/format";
import { estimateCheckoutTotal } from "@/lib/pricing";
import { Button } from "./ui/Button";
import { Input, Select } from "./ui/Input";
import type { ApiResponse, EventItem, ReferralInfo, Transaction } from "@/lib/types";

export function BookingPanel({ event }: { event: EventItem }) {
  const { user } = useAuth();
  const router = useRouter();

  const [ticketTypeId, setTicketTypeId] = useState<number | "">(event.ticket_types?.[0]?.id ?? "");
  const [quantity, setQuantity] = useState(1);
  const [voucherCode, setVoucherCode] = useState("");
  const [useCoupon, setUseCoupon] = useState(false);
  const [pointsToUse, setPointsToUse] = useState(0);
  const [referral, setReferral] = useState<ReferralInfo | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  const hasTicketTypes = (event.ticket_types?.length ?? 0) > 0;
  const selectedTicketType = event.ticket_types?.find((t) => t.id === ticketTypeId);
  const unitPrice = hasTicketTypes ? selectedTicketType?.price ?? 0 : event.is_paid ? event.price : 0;
  const remainingSeats = hasTicketTypes ? selectedTicketType?.remaining ?? 0 : event.available_seats;

  const availableCoupon = referral?.available_coupons?.[0];
  const { subtotal, couponDiscount, pointsDiscount, total: estimatedTotal } = estimateCheckoutTotal({
    unitPrice,
    quantity,
    couponPercent: availableCoupon?.discount_value ?? 0,
    useCoupon,
    pointsToUse,
  });

  const soldOut = remainingSeats <= 0;
  const eventEnded = new Date(event.end_date) < new Date();

  useEffect(() => {
    if (!user || user.role !== "customer") return;
    api
      .get<ApiResponse<ReferralInfo>>("/referral/me")
      .then((res) => setReferral(res.data.data))
      .catch(() => {
        // non-critical; the booking form still works without points/coupon info
      });
  }, [user]);

  async function handleBook() {
    if (!user) {
      router.push(`/login?next=/events/${event.slug}`);
      return;
    }
    setSubmitting(true);
    setError("");
    try {
      const res = await api.post<ApiResponse<Transaction>>("/transactions", {
        event_id: event.id,
        ticket_type_id: hasTicketTypes ? ticketTypeId : undefined,
        quantity,
        voucher_code: voucherCode || undefined,
        use_coupon: useCoupon,
        points_to_use: pointsToUse,
      });
      const tx = res.data.data;
      if (tx.status === "success") {
        router.push(`/my-tickets?booked=${tx.id}`);
      } else if (tx.midtrans_redirect_url) {
        window.location.href = tx.midtrans_redirect_url;
      } else {
        router.push(`/my-tickets`);
      }
    } catch (err) {
      setError(apiErrorMessage(err, "Failed to start checkout"));
    } finally {
      setSubmitting(false);
    }
  }

  if (eventEnded) {
    return (
      <div className="rounded-2xl border border-slate-200 bg-white p-5 text-sm text-slate-500">
        This event has already ended.
      </div>
    );
  }

  return (
    <div className="rounded-2xl border border-slate-200 bg-white p-5">
      <div className="flex items-center justify-between">
        <span className="text-2xl font-bold text-slate-900">
          {unitPrice > 0 ? formatIDR(unitPrice) : "Free"}
        </span>
        {event.is_paid && <span className="text-xs text-slate-500">per ticket</span>}
      </div>

      <div className="mt-4 flex flex-col gap-3">
        {hasTicketTypes && (
          <Select
            label="Ticket type"
            value={ticketTypeId}
            onChange={(e) => setTicketTypeId(Number(e.target.value))}
          >
            {event.ticket_types!.map((t) => (
              <option key={t.id} value={t.id} disabled={t.remaining <= 0}>
                {t.name} — {t.price > 0 ? formatIDR(t.price) : "Free"} ({t.remaining} left)
              </option>
            ))}
          </Select>
        )}

        <Input
          label="Quantity"
          type="number"
          min={1}
          max={Math.max(Math.min(remainingSeats, 10), 1)}
          value={quantity}
          onChange={(e) => setQuantity(Math.max(1, Number(e.target.value)))}
          disabled={soldOut}
        />

        {event.is_paid && user?.role === "customer" && (
          <>
            <Input
              label="Voucher code"
              placeholder="Optional"
              value={voucherCode}
              onChange={(e) => setVoucherCode(e.target.value.toUpperCase())}
            />

            {availableCoupon && (
              <label className="flex items-center gap-2 text-sm text-slate-700">
                <input type="checkbox" checked={useCoupon} onChange={(e) => setUseCoupon(e.target.checked)} />
                <Tag className="h-3.5 w-3.5 text-indigo-500" />
                Use my {availableCoupon.discount_value}% referral coupon
              </label>
            )}

            {!!referral?.point_balance && (
              <Input
                label={`Use points (balance: ${referral.point_balance.toLocaleString("id-ID")})`}
                type="number"
                min={0}
                max={referral.point_balance}
                value={pointsToUse}
                onChange={(e) =>
                  setPointsToUse(Math.max(0, Math.min(referral.point_balance, Number(e.target.value))))
                }
              />
            )}
          </>
        )}

        <div className="mt-2 space-y-1 border-t border-dashed border-slate-200 pt-3 text-sm">
          <div className="flex justify-between text-slate-500">
            <span>Subtotal</span>
            <span>{formatIDR(subtotal)}</span>
          </div>
          {couponDiscount > 0 && (
            <div className="flex justify-between text-emerald-600">
              <span>Coupon discount</span>
              <span>-{formatIDR(couponDiscount)}</span>
            </div>
          )}
          {pointsDiscount > 0 && (
            <div className="flex justify-between text-emerald-600">
              <span>Points redeemed</span>
              <span>-{formatIDR(pointsDiscount)}</span>
            </div>
          )}
          <div className="flex justify-between pt-1 text-base font-semibold text-slate-900">
            <span>Total</span>
            <span>{formatIDR(estimatedTotal)}</span>
          </div>
        </div>

        {error && <p className="text-xs text-red-600">{error}</p>}

        <Button
          onClick={handleBook}
          loading={submitting}
          disabled={soldOut || user?.role === "organizer"}
          className="mt-1 w-full"
          size="lg"
        >
          <TicketIcon className="h-4 w-4" />
          {soldOut ? "Sold out" : user?.role === "organizer" ? "Organizers can't book tickets" : "Book now"}
        </Button>
        {!user && <p className="text-center text-xs text-slate-500">You&apos;ll be asked to log in first.</p>}
      </div>
    </div>
  );
}
