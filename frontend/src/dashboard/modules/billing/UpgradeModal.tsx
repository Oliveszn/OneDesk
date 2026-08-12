import { useState } from "react";
import { Modal, Form, Input, Select, Button } from "antd";
import { Mail01Icon } from "hugeicons-react";

const currencyOptions = [
  { value: "NGN", label: "NGN — Nigerian Naira" },
  { value: "USD", label: "USD — US Dollar" },
];

export default function UpgradeModal({ open, onClose, onUpgrade }) {
  const [form] = Form.useForm();
  const [currency, setCurrency] = useState("NGN");

  const handleFinish = (values) => {
    onUpgrade(values);
    form.resetFields();
    setCurrency("NGN");
    onClose();
  };

  const gatewayHint =
    currency === "NGN"
      ? "NGN checkouts route through Paystack first, with Flutterwave as automatic failover."
      : "Non-NGN checkouts route through Flutterwave first, with Paystack as automatic failover.";

  return (
    <Modal
      open={open}
      onCancel={onClose}
      footer={null}
      title={
        <span className="font-semibold text-lg" style={{ color: "#16191c" }}>
          Upgrade to Paid
        </span>
      }
      width={400}
      destroyOnClose
    >
      <p className="text-sm mb-5" style={{ color: "#52585d" }}>
        ₦5,000/month, unlimited users, products, and orders.
      </p>
      <Form
        form={form}
        layout="vertical"
        onFinish={handleFinish}
        requiredMark={false}
        initialValues={{ currency: "NGN" }}
      >
        <Form.Item
          name="email"
          label="Billing email"
          rules={[
            { required: true, message: "Enter a billing email" },
            { type: "email", message: "Enter a valid email" },
          ]}
        >
          <Input
            prefix={<Mail01Icon size={18} className="mr-1 opacity-60" />}
            placeholder="billing@yourbusiness.com"
          />
        </Form.Item>
        <Form.Item
          name="currency"
          label="Currency"
          rules={[{ required: true }]}
        >
          <Select options={currencyOptions} onChange={setCurrency} />
        </Form.Item>

        <p
          className="text-[13px] mb-6 px-3 py-2.5"
          style={{
            backgroundColor: "#fbfaf8",
            border: "1px solid #e1ded6",
            color: "#52585d",
          }}
        >
          {gatewayHint}
        </p>

        <Button type="primary" htmlType="submit" block>
          Continue to checkout
        </Button>
      </Form>
    </Modal>
  );
}
