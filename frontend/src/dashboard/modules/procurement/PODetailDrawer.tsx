import { Drawer, Tag } from "antd";

export default function PODetailDrawer({
  open,
  onClose,
  po,
  vendors,
  products,
  warehouses,
}) {
  if (!po) return null;

  const vendor = po.vendor_id
    ? vendors.find((v) => v.vendor_id === po.vendor_id)
    : null;

  const statusConfig = {
    suggested: { color: "#b8863a", label: "Suggested" },
    sent: { color: "#24344a", label: "Sent" },
    received: { color: "#5b7a63", label: "Received" },
    receive_issue: { color: "#9c4a3c", label: "Receive issue" },
  };
  const config = statusConfig[po.status] ?? {
    color: "#52585d",
    label: po.status,
  };

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title={
        <span className="font-semibold text-lg" style={{ color: "#16191c" }}>
          {po.po_id}
        </span>
      }
      width={480}
    >
      <div className="flex items-center justify-between mb-6">
        <div>
          <div className="text-[13px]" style={{ color: "#52585d" }}>
            Vendor
          </div>
          <div className="text-[15px] font-medium" style={{ color: "#16191c" }}>
            {vendor?.name ?? (po.vendor_id ? po.vendor_id : "Not yet assigned")}
          </div>
        </div>
        <Tag color={config.color} style={{ margin: 0, fontWeight: 600 }}>
          {config.label}
        </Tag>
      </div>

      <div
        className="text-[13px] font-semibold uppercase tracking-wide mb-3"
        style={{ color: "#52585d" }}
      >
        Line items
      </div>

      <div className="border" style={{ borderColor: "#e1ded6" }}>
        {po.items.map((item, i) => {
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
              <div
                className="text-[14px] font-medium"
                style={{ color: "#16191c" }}
              >
                {product?.name ?? item.product_id}
              </div>
              <div className="text-[13px]" style={{ color: "#52585d" }}>
                {item.quantity} units → {warehouse?.name ?? item.warehouse_id}
              </div>
            </div>
          );
        })}
      </div>

      {po.status === "receive_issue" && (
        <div
          className="mt-6 px-4 py-3 text-[13px]"
          style={{
            backgroundColor: "rgba(156,74,60,0.08)",
            border: "1px solid #9c4a3c",
            color: "#16191c",
          }}
        >
          This PO was marked received, but restocking failed on at least one
          line item. Stock levels may not reflect this delivery — check
          Inventory and reconcile manually.
        </div>
      )}

      {po.status === "suggested" && (
        <div
          className="mt-6 px-4 py-3 text-[13px]"
          style={{
            backgroundColor: "rgba(184,134,58,0.1)",
            border: "1px solid #b8863a",
            color: "#16191c",
          }}
        >
          Auto-suggested from a low-stock alert. Assign a vendor to send it.
        </div>
      )}
    </Drawer>
  );
}
