import { Table, Button } from "antd";
import { TruckDeliveryIcon } from "hugeicons-react";
import EmptyState from "../../shared/EmptyState";

export default function VendorsTable({ vendors, onNewVendor }) {
  if (vendors.length === 0) {
    return (
      <EmptyState
        Icon={TruckDeliveryIcon}
        title="No vendors yet"
        description="Add a vendor before you can send a purchase order to one."
        action={
          <Button type="primary" onClick={onNewVendor}>
            New vendor
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
      dataSource={vendors}
      columns={columns}
      rowKey="vendor_id"
      pagination={false}
      className="px-8"
    />
  );
}
