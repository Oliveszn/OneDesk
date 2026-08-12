export const mockVendors = [
  { vendor_id: "v_acme", name: "Acme Supplies Ltd" },
  { vendor_id: "v_northgate", name: "Northgate Wholesale" },
];

export const mockPurchaseOrders = [
  {
    po_id: "po_2001",
    vendor_id: null,
    status: "suggested",
    items: [
      { product_id: "p_monitor", warehouse_id: "wh_lagos", quantity: 30 },
    ],
  },
  {
    po_id: "po_2002",
    vendor_id: "v_acme",
    status: "sent",
    items: [
      { product_id: "p_keyboard", warehouse_id: "wh_lagos", quantity: 25 },
    ],
  },
  {
    po_id: "po_2003",
    vendor_id: "v_northgate",
    status: "received",
    items: [{ product_id: "p_mouse", warehouse_id: "wh_abuja", quantity: 50 }],
  },
  {
    po_id: "po_2004",
    vendor_id: "v_acme",
    status: "receive_issue",
    items: [{ product_id: "p_hub", warehouse_id: "wh_lagos", quantity: 40 }],
  },
];
