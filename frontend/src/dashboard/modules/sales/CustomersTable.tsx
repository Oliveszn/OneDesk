import { Table, Button } from "antd";
import { UserGroupIcon } from "hugeicons-react";
import EmptyState from "../../shared/EmptyState";

export default function CustomersTable({ customers, onNewCustomer }) {
  if (customers.length === 0) {
    return (
      <EmptyState
        Icon={UserGroupIcon}
        title="No customers yet"
        description="Add a customer before you can place an order for them."
        action={
          <Button type="primary" onClick={onNewCustomer}>
            New customer
          </Button>
        }
      />
    );
  }

  const columns = [
    {
      title: "Name",
      dataIndex: "name",
      key: "name",
      render: (name) => (
        <span className="font-medium text-[14px]" style={{ color: "#16191c" }}>
          {name}
        </span>
      ),
    },
    {
      title: "Email",
      dataIndex: "email",
      key: "email",
      render: (email) => (
        <span className="text-[14px]" style={{ color: "#52585d" }}>
          {email || "—"}
        </span>
      ),
    },
  ];

  return (
    <Table
      dataSource={customers}
      columns={columns}
      rowKey="customer_id"
      pagination={false}
      className="px-8"
    />
  );
}
