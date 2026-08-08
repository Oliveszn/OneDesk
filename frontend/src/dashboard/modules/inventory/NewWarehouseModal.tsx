import { Modal, Form, Input, Button } from "antd";
import { Building06Icon } from "hugeicons-react";

// Single field, matching CreateWarehouseRequest exactly.
export default function NewWarehouseModal({ open, onClose, onCreate }) {
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
          New warehouse
        </span>
      }
      width={380}
      destroyOnHidden
    >
      <Form
        form={form}
        layout="vertical"
        onFinish={handleFinish}
        requiredMark={false}
        className="mt-5"
      >
        <Form.Item
          name="name"
          label="Name"
          rules={[{ required: true, message: "Enter a warehouse name" }]}
        >
          <Input
            prefix={<Building06Icon size={18} className="mr-1 opacity-60" />}
            placeholder="Port Harcourt Depot"
          />
        </Form.Item>
        <div className="flex justify-end gap-3 mt-6">
          <Button onClick={onClose}>Cancel</Button>
          <Button type="primary" htmlType="submit">
            Create warehouse
          </Button>
        </div>
      </Form>
    </Modal>
  );
}
