export type Role = "customer" | "organizer";

export interface User {
  id: number;
  name: string;
  email: string;
  role: Role;
  referral_code: string;
  referred_by: number | null;
  avatar_url: string;
  created_at: string;
  updated_at: string;
}

export interface Category {
  id: number;
  name: string;
  slug: string;
}

export type EventStatus = "draft" | "published" | "completed" | "cancelled";

export interface TicketType {
  id: number;
  event_id: number;
  name: string;
  price: number;
  quota: number;
  remaining: number;
}

export interface Voucher {
  id: number;
  event_id: number;
  code: string;
  discount_type: "amount" | "percentage";
  discount_value: number;
  quota: number;
  used_count: number;
  start_date: string;
  end_date: string;
  created_at: string;
}

export interface EventItem {
  id: number;
  organizer_id: number;
  organizer?: User;
  category_id: number;
  category?: Category;
  title: string;
  slug: string;
  description: string;
  location: string;
  city: string;
  is_paid: boolean;
  price: number;
  start_date: string;
  end_date: string;
  total_seats: number;
  available_seats: number;
  banner_url: string;
  status: EventStatus;
  created_at: string;
  updated_at: string;
  ticket_types?: TicketType[];
  vouchers?: Voucher[];
}

export type TxStatus = "pending_payment" | "success" | "expired" | "cancelled" | "failed";

export interface Transaction {
  id: number;
  invoice_no: string;
  user_id: number;
  user?: User;
  event_id: number;
  event?: EventItem;
  ticket_type_id: number | null;
  quantity: number;
  unit_price: number;
  subtotal: number;
  voucher_id: number | null;
  voucher_discount: number;
  coupon_id: number | null;
  coupon_discount: number;
  points_used: number;
  points_discount: number;
  total_price: number;
  status: TxStatus;
  midtrans_redirect_url: string;
  payment_deadline: string;
  created_at: string;
  updated_at: string;
}

export interface Review {
  id: number;
  transaction_id: number;
  event_id: number;
  user_id: number;
  user?: User;
  rating: number;
  comment: string;
  created_at: string;
}

export interface Coupon {
  id: number;
  code: string;
  discount_type: "amount" | "percentage";
  discount_value: number;
  is_used: boolean;
  expires_at: string;
}

export interface PointLedgerEntry {
  id: number;
  points: number;
  type: "earn" | "redeem";
  description: string;
  expires_at: string | null;
  created_at: string;
}

export interface ReferralInfo {
  referral_code: string;
  total_referrals: number;
  point_balance: number;
  available_coupons: Coupon[];
  point_ledger: PointLedgerEntry[];
}

export interface Pagination {
  page: number;
  limit: number;
  total: number;
  total_pages: number;
}

export interface ApiListResponse<T> {
  data: T[];
  meta: Pagination;
}

export interface ApiResponse<T> {
  data: T;
}

export interface ApiErrorBody {
  error: { message: string };
}

export interface DashboardSeriesPoint {
  label: string;
  revenue: number;
  tickets_sold: number;
}

export interface DashboardTopEvent {
  title: string;
  tickets_sold: number;
  revenue: number;
}

export interface DashboardStats {
  range: "year" | "month" | "day";
  from: string;
  to: string;
  summary: {
    total_events: number;
    total_revenue: number;
    total_tickets: number;
    total_orders: number;
  };
  series: DashboardSeriesPoint[];
  top_events: DashboardTopEvent[];
}
