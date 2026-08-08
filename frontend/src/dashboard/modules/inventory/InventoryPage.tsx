import { useState } from "react";
import { Tabs, Button } from "antd";
import { PlusSignIcon } from "hugeicons-react";
import PageHeader from "../../shared/PageHeader";
import PlanLimitBanner from "../../shared/PlanLimitBanner";
import {
  mockProducts,
  mockStockLevels,
  mockWarehouses,
} from "../../mockData/inventory";
import ProductsTable from "./ProductsTable";
import WarehousesTable from "./WarehousesTable";
import NewProductModal from "./NewProductModal";
import NewWarehouseModal from "./NewWarehouseModal";
import AdjustStockModal from "./AdjustStockModal";
import { mockUsage } from "../../mockData/billing";

export default function InventoryPage() {
  const [tab, setTab] = useState("products");
  const [products, setProducts] = useState(mockProducts);
  const [warehouses, setWarehouses] = useState(mockWarehouses);
  const [stockLevels, setStockLevels] = useState(mockStockLevels);

  const [productModalOpen, setProductModalOpen] = useState(false);
  const [warehouseModalOpen, setWarehouseModalOpen] = useState(false);
  const [adjustingProduct, setAdjustingProduct] = useState(null);

  const atProductCap =
    mockUsage.products_cap != null && products.length >= mockUsage.products_cap;

  const handleCreateProduct = ({ sku, name }) => {
    setProducts((prev) => [
      ...prev,
      { product_id: `p_${Date.now()}`, sku, name },
    ]);
  };

  const handleCreateWarehouse = ({ name }) => {
    setWarehouses((prev) => [
      ...prev,
      { warehouse_id: `wh_${Date.now()}`, name },
    ]);
  };

  const handleAdjustStock = ({ warehouse_id, delta }) => {
    setStockLevels((prev) => {
      const productId = adjustingProduct.product_id;
      const existing = prev[productId] ?? [];
      const idx = existing.findIndex((l) => l.warehouse_id === warehouse_id);

      let updated;
      if (idx === -1) {
        updated = [
          ...existing,
          {
            product_id: productId,
            warehouse_id,
            quantity: delta,
            reorder_point: 0,
          },
        ];
      } else {
        updated = existing.map((l, i) =>
          i === idx ? { ...l, quantity: l.quantity + delta } : l,
        );
      }
      return { ...prev, [productId]: updated };
    });
  };

  return (
    <div>
      <PageHeader
        title="Inventory"
        description="Products and warehouses across your business."
        actions={
          tab === "products" ? (
            <Button
              type="primary"
              icon={<PlusSignIcon size={16} />}
              onClick={() => setProductModalOpen(true)}
              disabled={atProductCap}
            >
              New product
            </Button>
          ) : (
            <Button
              type="primary"
              icon={<PlusSignIcon size={16} />}
              onClick={() => setWarehouseModalOpen(true)}
            >
              New warehouse
            </Button>
          )
        }
      />

      {tab === "products" && (
        <PlanLimitBanner
          resourceLabel="product"
          used={products.length}
          cap={mockUsage.products_cap}
        />
      )}

      <Tabs
        activeKey={tab}
        onChange={setTab}
        className="px-8"
        items={[
          { key: "products", label: `Products (${products.length})` },
          { key: "warehouses", label: `Warehouses (${warehouses.length})` },
        ]}
      />

      {tab === "products" ? (
        <ProductsTable
          products={products}
          stockLevels={stockLevels}
          onAdjustStock={(product) => setAdjustingProduct(product)}
          onNewProduct={() => setProductModalOpen(true)}
        />
      ) : (
        <WarehousesTable
          warehouses={warehouses}
          onNewWarehouse={() => setWarehouseModalOpen(true)}
        />
      )}

      <NewProductModal
        open={productModalOpen}
        onClose={() => setProductModalOpen(false)}
        onCreate={handleCreateProduct}
      />
      <NewWarehouseModal
        open={warehouseModalOpen}
        onClose={() => setWarehouseModalOpen(false)}
        onCreate={handleCreateWarehouse}
      />
      <AdjustStockModal
        open={!!adjustingProduct}
        onClose={() => setAdjustingProduct(null)}
        product={adjustingProduct}
        warehouses={warehouses}
        currentLevels={
          adjustingProduct
            ? (stockLevels[adjustingProduct.product_id] ?? [])
            : []
        }
        onAdjust={handleAdjustStock}
      />
    </div>
  );
}
