export const mockUsage = {
  plan_name: "free",
  products_used: 4,
  products_cap: 50,
  orders_used: 61,
  orders_cap: 100,
  users_used: 1,
  users_cap: 3,
};

export const mockPlans = [
  {
    name: "free",
    max_users: 3,
    max_products: 50,
    max_orders_per_month: 100,
    price_amount: 0,
    price_currency: "NGN",
    billing_interval: null,
  },
  {
    name: "paid",
    max_users: null,
    max_products: null,
    max_orders_per_month: null,
    price_amount: 5000,
    price_currency: "NGN",
    billing_interval: "monthly",
  },
];
