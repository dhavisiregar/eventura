"use client";

import { Star } from "lucide-react";
import clsx from "clsx";

export function StarRating({
  value,
  onChange,
  size = "md",
  readOnly = false,
}: {
  value: number;
  onChange?: (value: number) => void;
  size?: "sm" | "md" | "lg";
  readOnly?: boolean;
}) {
  const sizes = { sm: "h-3.5 w-3.5", md: "h-5 w-5", lg: "h-7 w-7" };

  return (
    <div className="flex items-center gap-0.5" role={readOnly ? undefined : "radiogroup"} aria-label="Rating">
      {[1, 2, 3, 4, 5].map((star) => (
        <button
          key={star}
          type="button"
          disabled={readOnly}
          onClick={() => onChange?.(star)}
          aria-label={`${star} star${star > 1 ? "s" : ""}`}
          className={clsx(!readOnly && "cursor-pointer", readOnly && "cursor-default")}
        >
          <Star
            className={clsx(sizes[size], star <= value ? "fill-amber-400 text-amber-400" : "fill-transparent text-slate-300")}
          />
        </button>
      ))}
    </div>
  );
}
