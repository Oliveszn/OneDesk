import { Table, Button } from "antd";
import { Building06Icon } from "hugeicons-react";
import EmptyState from "../../shared/EmptyState";

export default function WarehousesTable({ warehouses, onNewWarehouse }) {
  if (warehouses.length === 0) {
    return (
      <EmptyState
        Icon={Building06Icon}
        title="No warehouses yet"
        description="Add a warehouse before you can stock any products."
        action={
          <Button type="primary" onClick={onNewWarehouse}>
            New warehouse
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
  ];

  return (
    <Table
      dataSource={warehouses}
      columns={columns}
      rowKey="warehouse_id"
      pagination={false}
      className="px-8"
    />
  );
}
