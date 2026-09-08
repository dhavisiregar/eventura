export interface CheckoutEstimateInput {
  unitPrice: number;
  quantity: number;
  couponPercent?: number;
  useCoupon?: boolean;
  pointsToUse?: number;
}

export interface CheckoutEstimate {
  subtotal: number;
  couponDiscount: number;
  pointsDiscount: number;
  total: number;
}

/**
 * Client-side estimate of the checkout total, mirroring the backend's discount
 * order (subtotal -> coupon % -> points, IDR 1:1) so the booking panel can show
 * a live preview. The backend remains the source of truth at checkout time.
 */
export function estimateCheckoutTotal({
  unitPrice,
  quantity,
  couponPercent = 0,
  useCoupon = false,
  pointsToUse = 0,
}: CheckoutEstimateInput): CheckoutEstimate {
  const subtotal = Math.max(unitPrice, 0) * Math.max(quantity, 0);
  const couponDiscount = useCoupon && couponPercent > 0 ? (subtotal * couponPercent) / 100 : 0;
  const remainingAfterCoupon = Math.max(subtotal - couponDiscount, 0);
  const pointsDiscount = Math.max(0, Math.min(pointsToUse, remainingAfterCoupon));
  const total = Math.max(remainingAfterCoupon - pointsDiscount, 0);

  return { subtotal, couponDiscount, pointsDiscount, total };
}
