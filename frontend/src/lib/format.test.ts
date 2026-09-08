import { formatIDR, formatRelativeCountdown } from "./format";

describe("formatIDR", () => {
  it("formats a number as an IDR currency string with no decimals", () => {
    const result = formatIDR(150000);
    expect(result).toContain("150.000");
    expect(result).toMatch(/Rp/);
  });

  it("formats zero as free-priced currency", () => {
    expect(formatIDR(0)).toContain("0");
  });
});

describe("formatRelativeCountdown", () => {
  it("reports Expired for a deadline in the past", () => {
    const past = new Date(Date.now() - 60_000).toISOString();
    expect(formatRelativeCountdown(past)).toBe("Expired");
  });

  it("reports minutes left for a near-future deadline", () => {
    const soon = new Date(Date.now() + 5 * 60_000).toISOString();
    expect(formatRelativeCountdown(soon)).toMatch(/\dm left/);
  });

  it("reports hours left for a same-day deadline", () => {
    const later = new Date(Date.now() + 3 * 60 * 60_000).toISOString();
    expect(formatRelativeCountdown(later)).toMatch(/\dh \dm left/);
  });
});
