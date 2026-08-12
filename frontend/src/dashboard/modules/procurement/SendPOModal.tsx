import { Modal, Form, Select, Button } from "antd";

export default function SendPOModal({ open, onClose, po, vendors, onSend }) {
  const [form] = Form.useForm();

  const handleFinish = (values) => {
    onSend(po.po_id, values.vendor_id);
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
          Send purchase order
        </span>
      }
      width={400}
      destroyOnClose
    >
      {po && (
        <>
          <p className="text-sm mb-5" style={{ color: "#52585d" }}>
            Assign a vendor to <span className="font-mono">{po.po_id}</span> and
            mark it sent.
          </p>
          <Form
            form={form}
            layout="vertical"
            onFinish={handleFinish}
            requiredMark={false}
          >
            <Form.Item
              name="vendor_id"
              label="Vendor"
              rules={[{ required: true, message: "Choose a vendor" }]}
            >
              <Select
                placeholder="Select a vendor"
                options={vendors.map((v) => ({
                  value: v.vendor_id,
                  label: v.name,
                }))}
              />
            </Form.Item>
            <div className="flex justify-end gap-3 mt-6">
              <Button onClick={onClose}>Cancel</Button>
              <Button type="primary" htmlType="submit">
                Send
              </Button>
            </div>
          </Form>
        </>
      )}
    </Modal>
  );
}
