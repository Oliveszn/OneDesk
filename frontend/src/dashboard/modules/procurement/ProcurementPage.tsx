import { useState } from "react";
import { Tabs, Button, message } from "antd";
import { PlusSignIcon } from "hugeicons-react";
import PageHeader from "../../shared/PageHeader";
import { mockVendors, mockPurchaseOrders } from "../../mockData/procurement";
import { mockProducts, mockWarehouses } from "../../mockData/inventory";
import PurchaseOrdersTable from "./PurchaseOrdersTable";
import VendorsTable from "./VendorsTable";
import NewVendorModal from "./NewVendorModal";
import SendPOModal from "./SendPOModal";
import PODetailDrawer from "./PODetailDrawer";

export default function ProcurementPage() {
  const [tab, setTab] = useState("purchaseOrders");
  const [vendors, setVendors] = useState(mockVendors);
  const [purchaseOrders, setPurchaseOrders] = useState(mockPurchaseOrders);

  const [vendorModalOpen, setVendorModalOpen] = useState(false);
  const [sendingPO, setSendingPO] = useState(null);
  const [viewingPO, setViewingPO] = useState(null);

  const handleCreateVendor = ({ name }) => {
    setVendors((prev) => [...prev, { vendor_id: `v_${Date.now()}`, name }]);
  };

  const handleSend = (poId, vendorId) => {
    setPurchaseOrders((prev) =>
      prev.map((po) =>
        po.po_id === poId ? { ...po, vendor_id: vendorId, status: "sent" } : po,
      ),
    );
    message.success("Purchase order sent.");
  };

  const handleReceive = (po) => {
    setPurchaseOrders((prev) =>
      prev.map((p) =>
        p.po_id === po.po_id ? { ...p, status: "received" } : p,
      ),
    );
    message.success("Purchase order received — stock updated.");
  };

  return (
    <div>
      <PageHeader
        title="Procurement"
        description="Vendors and purchase orders."
        actions={
          tab === "vendors" ? (
            <Button
              type="primary"
              icon={<PlusSignIcon size={16} />}
              onClick={() => setVendorModalOpen(true)}
            >
              New vendor
            </Button>
          ) : null
        }
      />

      <Tabs
        activeKey={tab}
        onChange={setTab}
        className="px-8"
        items={[
          {
            key: "purchaseOrders",
            label: `Purchase orders (${purchaseOrders.length})`,
          },
          { key: "vendors", label: `Vendors (${vendors.length})` },
        ]}
      />

      {tab === "purchaseOrders" ? (
        <PurchaseOrdersTable
          purchaseOrders={purchaseOrders}
          vendors={vendors}
          onView={setViewingPO}
          onSend={setSendingPO}
          onReceive={handleReceive}
        />
      ) : (
        <VendorsTable
          vendors={vendors}
          onNewVendor={() => setVendorModalOpen(true)}
        />
      )}

      <NewVendorModal
        open={vendorModalOpen}
        onClose={() => setVendorModalOpen(false)}
        onCreate={handleCreateVendor}
      />
      <SendPOModal
        open={!!sendingPO}
        onClose={() => setSendingPO(null)}
        po={sendingPO}
        vendors={vendors}
        onSend={handleSend}
      />
      <PODetailDrawer
        open={!!viewingPO}
        onClose={() => setViewingPO(null)}
        po={viewingPO}
        vendors={vendors}
        products={mockProducts}
        warehouses={mockWarehouses}
      />
    </div>
  );
}
