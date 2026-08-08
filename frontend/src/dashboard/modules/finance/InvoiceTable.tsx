import { Table, Tag, Button, Popconfirm } from "antd";
import { Invoice03Icon } from "hugeicons-react";
import EmptyState from "../../shared/EmptyState";
import { formatNaira } from "../../shared/format";
import type { ColumnsType } from "antd/es/table";

type Invoice = {
  invoice_id: string;
  order_id: string;
  issued_at: string;
  status: "paid" | "unpaid";
  amount: number;
};

export default function InvoicesTable({ invoices, onPay }) {
  if (invoices.length === 0) {
    return (
      <EmptyState
        Icon={Invoice03Icon}
        title="No invoices yet"
        description="Invoices are generated automatically when an order is placed — nothing to create here directly."
      />
    );
  }

  const columns: ColumnsType<Invoice> = [
    {
      title: "Invoice",
      dataIndex: "invoice_id",
      key: "invoice_id",
      width: 140,
      render: (id) => (
        <span className="font-mono text-[13px]" style={{ color: "#52585d" }}>
          {id}
        </span>
      ),
    },
    {
      title: "Order",
      dataIndex: "order_id",
      key: "order_id",
      render: (id) => (
        <span className="font-mono text-[13px]" style={{ color: "#52585d" }}>
          {id}
        </span>
      ),
    },
    {
      title: "Issued",
      dataIndex: "issued_at",
      key: "issued_at",
      render: (date) => (
        <span className="text-[14px]" style={{ color: "#52585d" }}>
          {new Date(date).toLocaleDateString("en-NG", {
            day: "numeric",
            month: "short",
            year: "numeric",
          })}
        </span>
      ),
    },
    {
      title: "Status",
      dataIndex: "status",
      key: "status",
      render: (status) =>
        status === "paid" ? (
          <Tag
            color="#5b7a63"
            style={{ margin: 0, fontSize: 11, fontWeight: 600 }}
          >
            Paid
          </Tag>
        ) : (
          <Tag
            color="#b8863a"
            style={{ margin: 0, fontSize: 11, fontWeight: 600 }}
          >
            Unpaid
          </Tag>
        ),
    },
    {
      title: "Amount",
      key: "amount",
      align: "right",
      render: (_, inv) => (
        <span
          className="font-mono text-[14px] font-medium"
          style={{ color: "#16191c" }}
        >
          {formatNaira(inv.amount)}
        </span>
      ),
    },
    {
      title: "",
      key: "actions",
      width: 120,
      render: (_, inv) =>
        inv.status === "unpaid" ? (
          <Popconfirm
            title="Mark this invoice as paid?"
            description="This can't be undone in the current design."
            onConfirm={() => onPay(inv.invoice_id)}
            okText="Mark paid"
          >
            <Button size="small">Mark paid</Button>
          </Popconfirm>
        ) : null,
    },
  ];

  return (
    <Table
      dataSource={invoices}
      columns={columns}
      rowKey="invoice_id"
      pagination={false}
      className="px-8"
    />
  );
}
