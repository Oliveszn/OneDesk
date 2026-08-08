import { Table, Tag, Button } from "antd";
import type { ColumnsType } from "antd/es/table";
import { ShoppingCart01Icon } from "hugeicons-react";
import { formatNaira } from "../../shared/format";
import EmptyState from "../../shared/EmptyState";

type Order = {
  order_id: string;
  customer_id: string;
  status: string;
  total: number;
};
type Customer = {
  customer_id: string;
  name: string;
};

type OrdersTableProps = {
  orders: Order[];
  customers: Customer[];
  onView: (order: Order) => void;
};
export default function OrdersTable({
  orders,
  customers,
  onView,
}: OrdersTableProps) {
  if (orders.length === 0) {
    return (
      <EmptyState
        Icon={ShoppingCart01Icon}
        title="No orders yet"
        description="Orders you place will show up here."
      />
    );
  }

  const columns: ColumnsType<Order> = [
    {
      title: "Order",
      dataIndex: "order_id",
      key: "order_id",
      width: 140,
      render: (id) => (
        <span className="font-mono text-[13px]" style={{ color: "#52585d" }}>
          {id}
        </span>
      ),
    },
    {
      title: "Customer",
      key: "customer",
      render: (_, order) => {
        const customer = customers.find(
          (c) => c.customer_id === order.customer_id,
        );
        return (
          <span
            className="font-medium text-[14px]"
            style={{ color: "#16191c" }}
          >
            {customer?.name ?? "Unknown customer"}
          </span>
        );
      },
    },
    {
      title: "Status",
      dataIndex: "status",
      key: "status",
      render: (status) => <StatusTag status={status} />,
    },
    {
      title: "Total",
      key: "total",
      align: "right",
      render: (_, order) => (
        <span
          className="font-mono text-[14px] font-medium"
          style={{ color: "#16191c" }}
        >
          {formatNaira(order.total)}
        </span>
      ),
    },
    {
      title: "",
      key: "actions",
      width: 100,
      render: (_, order) => (
        <Button size="small" onClick={() => onView(order)}>
          View
        </Button>
      ),
    },
  ];

  return (
    <Table
      dataSource={orders}
      columns={columns}
      rowKey="order_id"
      pagination={false}
      className="px-8"
    />
  );
}

function StatusTag({ status }) {
  if (status === "stock_issue") {
    return (
      <Tag color="#b8863a" style={{ margin: 0, fontSize: 11, fontWeight: 600 }}>
        Stock issue
      </Tag>
    );
  }
  return (
    <Tag color="#5b7a63" style={{ margin: 0, fontSize: 11, fontWeight: 600 }}>
      Placed
    </Tag>
  );
}
