import { Table, Tag, Button } from "antd";
import { PackageOutOfStockIcon } from "hugeicons-react";
import EmptyState from "../../shared/EmptyState";

export default function ProductsTable({
  products,
  stockLevels,
  onAdjustStock,
  onNewProduct,
}) {
  if (products.length === 0) {
    return (
      <EmptyState
        Icon={PackageOutOfStockIcon}
        title="No products yet"
        description="Add your first product to start tracking stock across warehouses."
        action={
          <Button type="primary" onClick={onNewProduct}>
            New product
          </Button>
        }
      />
    );
  }

  const columns = [
    {
      title: "SKU",
      dataIndex: "sku",
      key: "sku",
      width: 140,
      render: (sku) => (
        <span className="font-mono text-[13px]" style={{ color: "#52585d" }}>
          {sku}
        </span>
      ),
    },
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
      title: "Stock",
      key: "stock",
      render: (_, product) => (
        <StockSummary levels={stockLevels[product.product_id] ?? []} />
      ),
    },
    {
      title: "",
      key: "actions",
      width: 140,
      render: (_, product) => (
        <Button size="small" onClick={() => onAdjustStock(product)}>
          Adjust stock
        </Button>
      ),
    },
  ];

  return (
    <Table
      dataSource={products}
      columns={columns}
      rowKey="product_id"
      pagination={false}
      className="px-8"
    />
  );
}

function StockSummary({ levels }) {
  if (levels.length === 0) {
    return (
      <span className="text-[13px]" style={{ color: "#52585d" }}>
        Not stocked yet
      </span>
    );
  }

  const total = levels.reduce((sum, l) => sum + l.quantity, 0);
  const isOut = total === 0;
  const isLow = !isOut && levels.some((l) => l.quantity <= l.reorder_point);

  return (
    <div className="flex items-center gap-2">
      <span
        className="font-mono text-[14px] font-medium"
        style={{ color: "#16191c" }}
      >
        {total}
      </span>
      {isOut && (
        <Tag
          color="#9c4a3c"
          style={{ margin: 0, fontSize: 11, fontWeight: 600 }}
        >
          Out of stock
        </Tag>
      )}
      {isLow && (
        <Tag
          color="#b8863a"
          style={{ margin: 0, fontSize: 11, fontWeight: 600 }}
        >
          Low stock
        </Tag>
      )}
    </div>
  );
}
