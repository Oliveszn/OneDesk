// Matches WarehouseResponse
export const mockWarehouses = [
  { warehouse_id: "wh_lagos", name: "Lagos Main" },
  { warehouse_id: "wh_abuja", name: "Abuja Depot" },
];

// Matches ProductResponse
export const mockProducts = [
  { product_id: "p_mouse", sku: "SKU-001", name: "Wireless Mouse" },
  { product_id: "p_keyboard", sku: "SKU-002", name: "Mechanical Keyboard" },
  { product_id: "p_hub", sku: "SKU-003", name: "USB-C Hub" },
  { product_id: "p_monitor", sku: "SKU-004", name: '27" Monitor' },
];

// Matches StockLevelResponse
export const mockStockLevels = {
  p_mouse: [
    {
      product_id: "p_mouse",
      warehouse_id: "wh_lagos",
      quantity: 42,
      reorder_point: 10,
    },
    {
      product_id: "p_mouse",
      warehouse_id: "wh_abuja",
      quantity: 6,
      reorder_point: 10,
    }, // below reorder point — low-stock state
  ],
  p_keyboard: [
    {
      product_id: "p_keyboard",
      warehouse_id: "wh_lagos",
      quantity: 0,
      reorder_point: 5,
    }, // out of stock
  ],
  p_hub: [
    {
      product_id: "p_hub",
      warehouse_id: "wh_lagos",
      quantity: 120,
      reorder_point: 20,
    },
    {
      product_id: "p_hub",
      warehouse_id: "wh_abuja",
      quantity: 85,
      reorder_point: 20,
    },
  ],
  p_monitor: [
    {
      product_id: "p_monitor",
      warehouse_id: "wh_lagos",
      quantity: 14,
      reorder_point: 15,
    }, // below reorder point
  ],
};
