"use client";

import { useEffect, useState } from "react";
import { Check, Copy, Gift, Users } from "lucide-react";
import { api } from "@/lib/api";
import { useAuth } from "@/context/AuthContext";
import { formatDate } from "@/lib/format";
import { Card } from "@/components/ui/Card";
import { Button } from "@/components/ui/Button";
import type { ApiResponse, ReferralInfo } from "@/lib/types";

export default function AccountPage() {
  const { user } = useAuth();
  const [info, setInfo] = useState<ReferralInfo | null>(null);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    api.get<ApiResponse<ReferralInfo>>("/referral/me").then((res) =>
      setInfo({
        ...res.data.data,
        available_coupons: res.data.data.available_coupons ?? [],
        point_ledger: res.data.data.point_ledger ?? [],
      })
    );
  }, []);

  function copyCode() {
    if (!info) return;
    navigator.clipboard.writeText(info.referral_code).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    });
  }

  return (
    <div className="mx-auto max-w-3xl px-4 py-8 sm:px-6 lg:px-8">
      <h1 className="mb-6 text-xl font-bold text-slate-900">My account</h1>

      <Card className="p-5">
        <p className="text-sm text-slate-500">Signed in as</p>
        <p className="font-semibold text-slate-900">{user?.name}</p>
        <p className="text-sm text-slate-500">{user?.email}</p>
      </Card>

      {info && (
        <>
          <div className="mt-6 grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Card className="p-5">
              <div className="flex items-center gap-2 text-sm font-medium text-slate-500">
                <Gift className="h-4 w-4 text-indigo-500" /> Point balance
              </div>
              <p className="mt-2 text-2xl font-bold text-slate-900">{info.point_balance.toLocaleString("id-ID")}</p>
              <p className="mt-1 text-xs text-slate-500">Use points at checkout to reduce your ticket price.</p>
            </Card>
            <Card className="p-5">
              <div className="flex items-center gap-2 text-sm font-medium text-slate-500">
                <Users className="h-4 w-4 text-indigo-500" /> Total referrals
              </div>
              <p className="mt-2 text-2xl font-bold text-slate-900">{info.total_referrals}</p>
              <p className="mt-1 text-xs text-slate-500">People who signed up with your referral code.</p>
            </Card>
          </div>

          <Card className="mt-6 p-5">
            <p className="text-sm font-medium text-slate-700">Your referral code</p>
            <div className="mt-2 flex items-center gap-2">
              <code className="flex-1 rounded-lg bg-slate-100 px-3 py-2 text-sm font-semibold tracking-wide text-slate-900">
                {info.referral_code}
              </code>
              <Button variant="outline" size="sm" onClick={copyCode}>
                {copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
                {copied ? "Copied" : "Copy"}
              </Button>
            </div>
            <p className="mt-2 text-xs text-slate-500">
              Share this code — every friend who signs up with it earns you 10,000 points (valid 3 months).
            </p>
          </Card>

          {info.available_coupons.length > 0 && (
            <Card className="mt-6 p-5">
              <p className="mb-3 text-sm font-medium text-slate-700">Available coupons</p>
              <div className="flex flex-col gap-2">
                {info.available_coupons.map((c) => (
                  <div key={c.id} className="flex items-center justify-between rounded-lg bg-emerald-50 px-3 py-2 text-sm">
                    <span className="font-medium text-emerald-700">{c.code} — {c.discount_value}% off</span>
                    <span className="text-xs text-emerald-600">Expires {formatDate(c.expires_at)}</span>
                  </div>
                ))}
              </div>
            </Card>
          )}

          <Card className="mt-6 p-5">
            <p className="mb-3 text-sm font-medium text-slate-700">Points history</p>
            {info.point_ledger.length === 0 ? (
              <p className="text-sm text-slate-500">No point activity yet.</p>
            ) : (
              <div className="flex flex-col divide-y divide-slate-100">
                {info.point_ledger.map((entry) => (
                  <div key={entry.id} className="flex items-center justify-between py-2 text-sm">
                    <div>
                      <p className="text-slate-700">{entry.description || (entry.type === "earn" ? "Points earned" : "Points redeemed")}</p>
                      <p className="text-xs text-slate-400">{formatDate(entry.created_at)}</p>
                    </div>
                    <span className={entry.type === "earn" ? "font-medium text-emerald-600" : "font-medium text-red-600"}>
                      {entry.type === "earn" ? "+" : "-"}
                      {entry.points.toLocaleString("id-ID")}
                    </span>
                  </div>
                ))}
              </div>
            )}
          </Card>
        </>
      )}
    </div>
  );
}
