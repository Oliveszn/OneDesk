// Matches CustomerResponse
export const mockCustomers = [
  {
    customer_id: "c_bright",
    name: "Bright Traders Ltd",
    email: "ops@brighttraders.com",
  },
  { customer_id: "c_ada", name: "Ada Okafor", email: "ada.okafor@gmail.com" },
];

export const mockOrders = [
  {
    order_id: "o_1001",
    customer_id: "c_bright",
    status: "placed",
    items: [
      {
        product_id: "p_mouse",
        warehouse_id: "wh_lagos",
        quantity: 10,
        unit_price: 8500,
      },
      {
        product_id: "p_hub",
        warehouse_id: "wh_lagos",
        quantity: 4,
        unit_price: 15000,
      },
    ],
    total: 145000,
  },
  {
    order_id: "o_1002",
    customer_id: "c_ada",
    status: "stock_issue",
    items: [
      {
        product_id: "p_keyboard",
        warehouse_id: "wh_lagos",
        quantity: 5,
        unit_price: 32000,
      },
    ],
    total: 160000,
  },
];
