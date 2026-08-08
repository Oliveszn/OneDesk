import { useState } from "react";
import { Modal, Form, Select, InputNumber, Button, message } from "antd";
import { Delete02Icon, PlusSignIcon } from "hugeicons-react";
import { formatNaira } from "../../shared/format";

export default function NewOrderModal({
  open,
  onClose,
  customers,
  products,
  warehouses,
  stockLevels,
  onCreate,
}) {
  const [form] = Form.useForm();
  const [total, setTotal] = useState(0);

  const recomputeTotal = () => {
    const items = form.getFieldValue("items") || [];
    const sum = items.reduce(
      (acc, it) => acc + (it?.quantity || 0) * (it?.unit_price || 0),
      0,
    );
    setTotal(sum);
  };

  const handleFinish = (values) => {
    const insufficient = values.items.some((item) => {
      const levels = stockLevels[item.product_id] ?? [];
      const level = levels.find((l) => l.warehouse_id === item.warehouse_id);
      return (level?.quantity ?? 0) < item.quantity;
    });

    const order = {
      order_id: `o_${Date.now()}`,
      customer_id: values.customer_id,
      status: insufficient ? "stock_issue" : "placed",
      items: values.items,
      total: values.items.reduce(
        (acc, it) => acc + it.quantity * it.unit_price,
        0,
      ),
    };

    onCreate(order);
    form.resetFields();
    setTotal(0);
    onClose();

    if (insufficient) {
      message.warning(
        "Order placed, but at least one item couldn't be fully stocked — marked stock_issue.",
      );
    } else {
      message.success("Order placed.");
    }
  };

  const handleClose = () => {
    form.resetFields();
    setTotal(0);
    onClose();
  };

  return (
    <Modal
      open={open}
      onCancel={handleClose}
      footer={null}
      title={
        <span className="font-semibold text-lg" style={{ color: "#16191c" }}>
          New order
        </span>
      }
      width={560}
      destroyOnHidden
    >
      <Form
        form={form}
        layout="vertical"
        onFinish={handleFinish}
        onValuesChange={recomputeTotal}
        requiredMark={false}
        className="mt-5"
      >
        <Form.Item
          name="customer_id"
          label="Customer"
          rules={[{ required: true, message: "Choose a customer" }]}
        >
          <Select
            placeholder="Select a customer"
            options={customers.map((c) => ({
              value: c.customer_id,
              label: c.name,
            }))}
          />
        </Form.Item>

        <div
          className="text-[13px] font-semibold uppercase tracking-wide mb-3 mt-6"
          style={{ color: "#52585d" }}
        >
          Line items
        </div>

        <Form.List
          name="items"
          rules={[
            {
              validator: async (_, items) => {
                if (!items || items.length < 1)
                  return Promise.reject(new Error("Add at least one item"));
              },
            },
          ]}
        >
          {(fields, { add, remove }, { errors }) => (
            <>
              {fields.map((field) => (
                <div
                  key={field.key}
                  className="flex gap-2 items-start mb-3 p-3"
                  style={{
                    backgroundColor: "#fbfaf8",
                    border: "1px solid #e1ded6",
                  }}
                >
                  <div className="flex-1 grid grid-cols-2 gap-2">
                    <Form.Item
                      name={[field.name, "product_id"]}
                      rules={[{ required: true, message: "Product" }]}
                      className="!mb-0"
                    >
                      <Select
                        placeholder="Product"
                        options={products.map((p) => ({
                          value: p.product_id,
                          label: p.name,
                        }))}
                      />
                    </Form.Item>
                    <Form.Item
                      name={[field.name, "warehouse_id"]}
                      rules={[{ required: true, message: "Warehouse" }]}
                      className="!mb-0"
                    >
                      <Select
                        placeholder="Warehouse"
                        options={warehouses.map((w) => ({
                          value: w.warehouse_id,
                          label: w.name,
                        }))}
                      />
                    </Form.Item>
                    <Form.Item
                      name={[field.name, "quantity"]}
                      rules={[{ required: true, message: "Qty" }]}
                      className="!mb-0"
                    >
                      <InputNumber
                        min={1}
                        placeholder="Quantity"
                        className="!w-full"
                      />
                    </Form.Item>
                    <Form.Item
                      name={[field.name, "unit_price"]}
                      rules={[{ required: true, message: "Price" }]}
                      className="!mb-0"
                    >
                      <InputNumber
                        min={0}
                        placeholder="Unit price (₦)"
                        className="!w-full"
                      />
                    </Form.Item>
                  </div>
                  <Button
                    type="text"
                    icon={<Delete02Icon size={16} color="#9c4a3c" />}
                    onClick={() => {
                      remove(field.name);
                      recomputeTotal();
                    }}
                    className="mt-1"
                  />
                </div>
              ))}
              <Form.Item>
                <Button
                  icon={<PlusSignIcon size={16} />}
                  onClick={() => add()}
                  block
                >
                  Add item
                </Button>
                <Form.ErrorList errors={errors} />
              </Form.Item>
            </>
          )}
        </Form.List>

        <div
          className="flex items-center justify-between py-3 mt-2 border-t"
          style={{ borderColor: "#e1ded6" }}
        >
          <span
            className="text-[14px] font-medium"
            style={{ color: "#52585d" }}
          >
            Estimated total
          </span>
          <span
            className="font-mono text-[16px] font-semibold"
            style={{ color: "#16191c" }}
          >
            {formatNaira(total)}
          </span>
        </div>

        <div className="flex justify-end gap-3 mt-4">
          <Button onClick={handleClose}>Cancel</Button>
          <Button type="primary" htmlType="submit">
            Place order
          </Button>
        </div>
      </Form>
    </Modal>
  );
}
