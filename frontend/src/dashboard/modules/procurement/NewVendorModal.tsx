import { Modal, Form, Input, Button } from "antd";
import { TruckDeliveryIcon } from "hugeicons-react";

export default function NewVendorModal({ open, onClose, onCreate }) {
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
          New vendor
        </span>
      }
      width={380}
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
          name="name"
          label="Name"
          rules={[{ required: true, message: "Enter a vendor name" }]}
        >
          <Input
            prefix={<TruckDeliveryIcon size={18} className="mr-1 opacity-60" />}
            placeholder="Acme Supplies Ltd"
          />
        </Form.Item>
        <div className="flex justify-end gap-3 mt-6">
          <Button onClick={onClose}>Cancel</Button>
          <Button type="primary" htmlType="submit">
            Create vendor
          </Button>
        </div>
      </Form>
    </Modal>
  );
}
