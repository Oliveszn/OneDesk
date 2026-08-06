import { Modal, Form, Input, Button } from "antd";
import { Tag01Icon, Package01Icon } from "hugeicons-react";

export default function NewProductModal({ open, onClose, onCreate }) {
  const [form] = Form.useForm();

  const handleFinish = (values) => {
    onCreate(values);
    form.resetFields();
    onClose();
  };

  return (
    <Modal
      open={open}
      onCancel={onClose}
      footer={null}
      title={
        <span className="font-semibold text-lg" style={{ color: "#16191c" }}>
          New product
        </span>
      }
      width={400}
      destroyOnClose
    >
      <Form
        form={form}
        layout="vertical"
        onFinish={handleFinish}
        requiredMark={false}
        className="mt-5"
      >
        <Form.Item
          name="sku"
          label="SKU"
          rules={[{ required: true, message: "Enter a SKU" }]}
        >
          <Input
            prefix={<Tag01Icon size={18} className="mr-1 opacity-60" />}
            placeholder="SKU-005"
          />
        </Form.Item>
        <Form.Item
          name="name"
          label="Name"
          rules={[{ required: true, message: "Enter a product name" }]}
        >
          <Input
            prefix={<Package01Icon size={18} className="mr-1 opacity-60" />}
            placeholder="Standing Desk"
          />
        </Form.Item>
        <div className="flex justify-end gap-3 mt-6">
          <Button onClick={onClose}>Cancel</Button>
          <Button type="primary" htmlType="submit">
            Create product
          </Button>
        </div>
      </Form>
    </Modal>
  );
}
