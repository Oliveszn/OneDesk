import { useState } from "react";
import { message } from "antd";
import PageHeader from "../../shared/PageHeader";
import { mockInvoices } from "../../mockData/finance";
import InvoicesTable from "./InvoiceTable";

export default function FinancePage() {
  const [invoices, setInvoices] = useState(mockInvoices);

  const handlePay = (invoiceId) => {
    setInvoices((prev) =>
      prev.map((inv) =>
        inv.invoice_id === invoiceId ? { ...inv, status: "paid" } : inv,
      ),
    );
    message.success("Invoice marked paid.");
  };

  return (
    <div>
      <PageHeader
        title="Finance"
        description="Invoices generated automatically from your orders."
      />
      <InvoicesTable invoices={invoices} onPay={handlePay} />
    </div>
  );
}
