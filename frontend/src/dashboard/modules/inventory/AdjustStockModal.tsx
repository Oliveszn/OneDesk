import { useState } from "react";
import { Modal, Form, Select, InputNumber, Button, Alert } from "antd";

export default function AdjustStockModal({
  open,
  onClose,
  product,
  warehouses,
  currentLevels,
  onAdjust,
}) {
  const [form] = Form.useForm();
  const [error, setError] = useState(null);

  const handleValuesChange = () => setError(null);

  const handleFinish = (values) => {
    const level = currentLevels.find(
      (l) => l.warehouse_id === values.warehouse_id,
    );
    const currentQty = level?.quantity ?? 0;
    const resulting = currentQty + values.delta;

    if (resulting < 0) {
      setError(
        `Can't apply — ${product?.name} only has ${currentQty} in stock at this warehouse. This would take it to ${resulting}.`,
      );
      return;
    }

    onAdjust({ warehouse_id: values.warehouse_id, delta: values.delta });
    form.resetFields();
    setError(null);
    onClose();
  };

  const handleClose = () => {
    form.resetFields();
    setError(null);
    onClose();
  };

  return (
    <Modal
      open={open}
      onCancel={handleClose}
      footer={null}
      title={
        <span className="font-semibold text-lg" style={{ color: "#16191c" }}>
          Adjust stock
        </span>
      }
      width={420}
      destroyOnHidden
    >
      {product && (
        <>
          <p className="text-sm mb-4" style={{ color: "#52585d" }}>
            <span className="font-medium" style={{ color: "#16191c" }}>
              {product.name}
            </span>{" "}
            — current stock by warehouse
          </p>

          <div className="mb-6 border" style={{ borderColor: "#e1ded6" }}>
            {currentLevels.length === 0 ? (
              <div
                className="px-4 py-3 text-[13px]"
                style={{ color: "#52585d" }}
              >
                Not stocked at any warehouse yet.
              </div>
            ) : (
              currentLevels.map((l, i) => {
                const wh = warehouses.find(
                  (w) => w.warehouse_id === l.warehouse_id,
                );
                return (
                  <div
                    key={l.warehouse_id}
                    className="flex items-center justify-between px-4 py-2.5 text-[14px]"
                    style={{ borderTop: i > 0 ? "1px solid #e1ded6" : "none" }}
                  >
                    <span style={{ color: "#16191c" }}>
                      {wh?.name ?? l.warehouse_id}
                    </span>
                    <span
                      className="font-mono font-medium"
                      style={{ color: "#16191c" }}
                    >
                      {l.quantity}
                    </span>
                  </div>
                );
              })
            )}
          </div>

          <Form
            form={form}
            layout="vertical"
            onFinish={handleFinish}
            onValuesChange={handleValuesChange}
            requiredMark={false}
          >
            <Form.Item
              name="warehouse_id"
              label="Warehouse"
              rules={[{ required: true, message: "Choose a warehouse" }]}
            >
              <Select
                placeholder="Select a warehouse"
                options={warehouses.map((w) => ({
                  value: w.warehouse_id,
                  label: w.name,
                }))}
              />
            </Form.Item>
            <Form.Item
              name="delta"
              label="Adjustment"
              rules={[{ required: true, message: "Enter an amount" }]}
              extra="Positive to restock, negative to remove."
            >
              <InputNumber className="!w-full" placeholder="e.g. 50 or -10" />
            </Form.Item>

            {error && (
              <Alert
                type="error"
                showIcon
                message={error}
                className="mb-4"
                style={{ borderColor: "#9c4a3c" }}
              />
            )}

            <div className="flex justify-end gap-3 mt-2">
              <Button onClick={handleClose}>Cancel</Button>
              <Button type="primary" htmlType="submit">
                Apply adjustment
              </Button>
            </div>
          </Form>
        </>
      )}
    </Modal>
  );
}
