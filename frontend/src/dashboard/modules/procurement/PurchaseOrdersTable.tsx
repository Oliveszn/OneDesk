import { Table, Tag, Button, Popconfirm } from "antd";
import { Package01Icon } from "hugeicons-react";
import EmptyState from "../../shared/EmptyState";

export default function PurchaseOrdersTable({
  purchaseOrders,
  vendors,
  onView,
  onSend,
  onReceive,
}) {
  if (purchaseOrders.length === 0) {
    return (
      <EmptyState
        Icon={Package01Icon}
        title="No purchase orders yet"
        description="A purchase order is suggested automatically when a product's stock falls to or below its reorder point."
      />
    );
  }

  const columns = [
    {
      title: "PO",
      dataIndex: "po_id",
      key: "po_id",
      width: 120,
      render: (id) => (
        <span className="font-mono text-[13px]" style={{ color: "#52585d" }}>
          {id}
        </span>
      ),
    },
    {
      title: "Vendor",
      key: "vendor",
      render: (_, po) => {
        if (!po.vendor_id) {
          return (
            <span className="text-[13px] italic" style={{ color: "#52585d" }}>
              Unassigned
            </span>
          );
        }
        const vendor = vendors.find((v) => v.vendor_id === po.vendor_id);
        return (
          <span
            className="font-medium text-[14px]"
            style={{ color: "#16191c" }}
          >
            {vendor?.name ?? po.vendor_id}
          </span>
        );
      },
    },
    {
      title: "Items",
      key: "items",
      render: (_, po) => (
        <span className="text-[14px]" style={{ color: "#52585d" }}>
          {po.items.length} line{po.items.length !== 1 ? "s" : ""}
        </span>
      ),
    },
    {
      title: "Status",
      dataIndex: "status",
      key: "status",
      render: (status) => <StatusTag status={status} />,
    },
    {
      title: "",
      key: "actions",
      width: 200,
      render: (_, po) => (
        <div className="flex gap-2 justify-end">
          <Button size="small" onClick={() => onView(po)}>
            View
          </Button>
          {po.status === "suggested" && (
            <Button size="small" type="primary" onClick={() => onSend(po)}>
              Send
            </Button>
          )}
          {po.status === "sent" && (
            <Popconfirm
              title="Mark this purchase order received?"
              description="This restocks the ordered items."
              onConfirm={() => onReceive(po)}
              okText="Receive"
            >
              <Button size="small" type="primary">
                Receive
              </Button>
            </Popconfirm>
          )}
        </div>
      ),
    },
  ];

  return (
    <Table
      dataSource={purchaseOrders}
      columns={columns}
      rowKey="po_id"
      pagination={false}
      className="px-8"
    />
  );
}

const statusConfig = {
  suggested: { color: "#b8863a", label: "Suggested" },
  sent: { color: "#24344a", label: "Sent" },
  received: { color: "#5b7a63", label: "Received" },
  receive_issue: { color: "#9c4a3c", label: "Receive issue" },
};

function StatusTag({ status }) {
  const config = statusConfig[status] ?? { color: "#52585d", label: status };
  return (
    <Tag
      color={config.color}
      style={{ margin: 0, fontSize: 11, fontWeight: 600 }}
    >
      {config.label}
    </Tag>
  );
}
