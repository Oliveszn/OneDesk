import { Modal, Form, Input, Button } from "antd";
import { UserIcon, Mail01Icon } from "hugeicons-react";

export default function NewCustomerModal({ open, onClose, onCreate }) {
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
          New customer
        </span>
      }
      width={400}
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
          rules={[{ required: true, message: "Enter a name" }]}
        >
          <Input
            prefix={<UserIcon size={18} className="mr-1 opacity-60" />}
            placeholder="Bright Traders Ltd"
          />
        </Form.Item>
        <Form.Item
          name="email"
          label="Email"
          rules={[{ type: "email", message: "Enter a valid email" }]}
        >
          <Input
            prefix={<Mail01Icon size={18} className="mr-1 opacity-60" />}
            placeholder="ops@brighttraders.com"
          />
        </Form.Item>
        <div className="flex justify-end gap-3 mt-6">
          <Button onClick={onClose}>Cancel</Button>
          <Button type="primary" htmlType="submit">
            Create customer
          </Button>
        </div>
      </Form>
    </Modal>
  );
}
