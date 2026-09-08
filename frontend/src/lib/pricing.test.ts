import { estimateCheckoutTotal } from "./pricing";

describe("estimateCheckoutTotal", () => {
  it("computes a plain subtotal with no discounts", () => {
    const result = estimateCheckoutTotal({ unitPrice: 100000, quantity: 2 });
    expect(result).toEqual({ subtotal: 200000, couponDiscount: 0, pointsDiscount: 0, total: 200000 });
  });

  it("applies a percentage coupon before points", () => {
    const result = estimateCheckoutTotal({
      unitPrice: 100000,
      quantity: 1,
      couponPercent: 10,
      useCoupon: true,
    });
    expect(result.couponDiscount).toBe(10000);
    expect(result.total).toBe(90000);
  });

  it("ignores the coupon percentage when useCoupon is false", () => {
    const result = estimateCheckoutTotal({ unitPrice: 100000, quantity: 1, couponPercent: 10, useCoupon: false });
    expect(result.couponDiscount).toBe(0);
    expect(result.total).toBe(100000);
  });

  it("deducts points 1:1 in IDR after the coupon discount", () => {
    const result = estimateCheckoutTotal({
      unitPrice: 300000,
      quantity: 1,
      couponPercent: 10,
      useCoupon: true,
      pointsToUse: 20000,
    });
    // subtotal 300000 - 10% coupon (30000) = 270000, then -20000 points = 250000
    expect(result.couponDiscount).toBe(30000);
    expect(result.pointsDiscount).toBe(20000);
    expect(result.total).toBe(250000);
  });

  it("clamps points usage to the remaining balance instead of going negative", () => {
    const result = estimateCheckoutTotal({ unitPrice: 5000, quantity: 1, pointsToUse: 20000 });
    expect(result.pointsDiscount).toBe(5000);
    expect(result.total).toBe(0);
  });

  it("never returns a negative total for a free event", () => {
    const result = estimateCheckoutTotal({ unitPrice: 0, quantity: 3, pointsToUse: 500 });
    expect(result.total).toBe(0);
    expect(result.pointsDiscount).toBe(0);
  });
});
