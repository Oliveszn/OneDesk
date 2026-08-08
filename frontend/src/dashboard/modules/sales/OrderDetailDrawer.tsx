import { Drawer, Tag } from "antd";
import { formatNaira } from "../../shared/format";

export default function OrderDetailDrawer({
  open,
  onClose,
  order,
  customers,
  products,
  warehouses,
}) {
  if (!order) return null;

  const customer = customers.find((c) => c.customer_id === order.customer_id);

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title={
        <span className="font-semibold text-lg" style={{ color: "#16191c" }}>
          Order {order.order_id}
        </span>
      }
      width={480}
    >
      <div className="flex items-center justify-between mb-6">
        <div>
          <div className="text-[13px]" style={{ color: "#52585d" }}>
            Customer
          </div>
          <div className="text-[15px] font-medium" style={{ color: "#16191c" }}>
            {customer?.name ?? "Unknown"}
          </div>
        </div>
        <Tag
          color={order.status === "stock_issue" ? "#b8863a" : "#5b7a63"}
          style={{ margin: 0, fontWeight: 600 }}
        >
          {order.status === "stock_issue" ? "Stock issue" : "Placed"}
        </Tag>
      </div>

      <div
        className="text-[13px] font-semibold uppercase tracking-wide mb-3"
        style={{ color: "#52585d" }}
      >
        Line items
      </div>

      <div className="border" style={{ borderColor: "#e1ded6" }}>
        {order.items.map((item, i) => {
          const product = products.find(
            (p) => p.product_id === item.product_id,
          );
          const warehouse = warehouses.find(
            (w) => w.warehouse_id === item.warehouse_id,
          );
          return (
            <div
              key={i}
              className="px-4 py-3 flex items-center justify-between"
              style={{ borderTop: i > 0 ? "1px solid #e1ded6" : "none" }}
            >
              <div>
                <div
                  className="text-[14px] font-medium"
                  style={{ color: "#16191c" }}
                >
                  {product?.name ?? item.product_id}
                </div>
                <div className="text-[12px]" style={{ color: "#52585d" }}>
                  {item.quantity} × {formatNaira(item.unit_price)} · from{" "}
                  {warehouse?.name ?? item.warehouse_id}
                </div>
              </div>
              <div
                className="font-mono text-[14px]"
                style={{ color: "#16191c" }}
              >
                {formatNaira(item.quantity * item.unit_price)}
              </div>
            </div>
          );
        })}
      </div>

      <div
        className="flex items-center justify-between mt-6 pt-4 border-t"
        style={{ borderColor: "#e1ded6" }}
      >
        <span
          className="text-[15px] font-semibold"
          style={{ color: "#16191c" }}
        >
          Total
        </span>
        <span
          className="font-mono text-[18px] font-semibold"
          style={{ color: "#16191c" }}
        >
          {formatNaira(order.total)}
        </span>
      </div>

      {order.status === "stock_issue" && (
        <div
          className="mt-6 px-4 py-3 text-[13px]"
          style={{
            backgroundColor: "rgba(184,134,58,0.1)",
            border: "1px solid #b8863a",
            color: "#16191c",
          }}
        >
          This order was created, but at least one line item couldn't be fully
          stocked at the time it was placed. Check stock levels in Inventory
          before fulfilling it.
        </div>
      )}
    </Drawer>
  );
}
