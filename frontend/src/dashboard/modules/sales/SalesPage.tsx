import { useState } from "react";
import { Tabs, Button } from "antd";
import { PlusSignIcon } from "hugeicons-react";
import PageHeader from "../../shared/PageHeader";
import PlanLimitBanner from "../../shared/PlanLimitBanner";
import OrdersTable from "./OrdersTable";
import { mockCustomers, mockOrders } from "../../mockData/sales";
import {
  mockProducts,
  mockWarehouses,
  mockStockLevels,
} from "../../mockData/inventory";
import { mockUsage } from "../../mockData/billing";
import CustomersTable from "./CustomersTable";
import NewCustomerModal from "./NewCustomerModal";
import NewOrderModal from "./NewOrderModal";
import OrderDetailDrawer from "./OrderDetailDrawer";

// Orders (not Customers) are a capped resource on the backend — see
// billing.ResourceOrders — so PlanLimitBanner only appears on that tab,
// same reasoning as InventoryPage only showing it on Products.
export default function SalesPage() {
  const [tab, setTab] = useState("orders");
  const [customers, setCustomers] = useState(mockCustomers);
  const [orders, setOrders] = useState(mockOrders);

  const [customerModalOpen, setCustomerModalOpen] = useState(false);
  const [orderModalOpen, setOrderModalOpen] = useState(false);
  const [viewingOrder, setViewingOrder] = useState(null);

  const atOrderCap =
    mockUsage.orders_cap != null && orders.length >= mockUsage.orders_cap;

  const handleCreateCustomer = (values) => {
    setCustomers((prev) => [
      ...prev,
      { customer_id: `c_${Date.now()}`, ...values },
    ]);
  };

  const handleCreateOrder = (order) => {
    setOrders((prev) => [order, ...prev]);
  };

  return (
    <div>
      <PageHeader
        title="Sales"
        description="Customers and orders."
        actions={
          tab === "orders" ? (
            <Button
              type="primary"
              icon={<PlusSignIcon size={16} />}
              onClick={() => setOrderModalOpen(true)}
              disabled={atOrderCap}
            >
              New order
            </Button>
          ) : (
            <Button
              type="primary"
              icon={<PlusSignIcon size={16} />}
              onClick={() => setCustomerModalOpen(true)}
            >
              New customer
            </Button>
          )
        }
      />

      {tab === "orders" && (
        <PlanLimitBanner
          resourceLabel="order"
          used={orders.length}
          cap={mockUsage.orders_cap}
        />
      )}

      <Tabs
        activeKey={tab}
        onChange={setTab}
        className="px-8"
        items={[
          { key: "orders", label: `Orders (${orders.length})` },
          { key: "customers", label: `Customers (${customers.length})` },
        ]}
      />

      {tab === "orders" ? (
        <OrdersTable
          orders={orders}
          customers={customers}
          onView={setViewingOrder}
        />
      ) : (
        <CustomersTable
          customers={customers}
          onNewCustomer={() => setCustomerModalOpen(true)}
        />
      )}

      <NewCustomerModal
        open={customerModalOpen}
        onClose={() => setCustomerModalOpen(false)}
        onCreate={handleCreateCustomer}
      />
      <NewOrderModal
        open={orderModalOpen}
        onClose={() => setOrderModalOpen(false)}
        customers={customers}
        products={mockProducts}
        warehouses={mockWarehouses}
        stockLevels={mockStockLevels}
        onCreate={handleCreateOrder}
      />
      <OrderDetailDrawer
        open={!!viewingOrder}
        onClose={() => setViewingOrder(null)}
        order={viewingOrder}
        customers={customers}
        products={mockProducts}
        warehouses={mockWarehouses}
      />
    </div>
  );
}
