import { useState } from "react";
import { Modal, Form, Input, Button, message } from "antd";
import { Mail01Icon, LockPasswordIcon, Building06Icon } from "hugeicons-react";

export function LoginModal({ open, onClose, onSwitchToRegister }) {
  const [form] = Form.useForm();

  const handleFinish = (values) => {
    console.log("login form submitted", values);
    message.info("Not wired to the API yet — this is a front-end-only demo.");
  };

  return (
    <Modal
      open={open}
      onCancel={onClose}
      footer={null}
      title={
        <span className="font-semibold text-lg" style={{ color: "#16191c" }}>
          Log in
        </span>
      }
      width={400}
      destroyOnHidden
    >
      <p className="text-sm mb-5" style={{ color: "#52585d" }}>
        Welcome back. Enter your details to access your workspace.
      </p>
      <Form
        form={form}
        layout="vertical"
        onFinish={handleFinish}
        requiredMark={false}
      >
        <Form.Item
          name="email"
          label="Email"
          rules={[
            { required: true, message: "Enter your email" },
            { type: "email", message: "Enter a valid email" },
          ]}
        >
          <Input
            prefix={<Mail01Icon size={18} className="mr-1 opacity-60" />}
            placeholder="you@business.com"
            autoComplete="email"
          />
        </Form.Item>
        <Form.Item
          name="password"
          label="Password"
          rules={[{ required: true, message: "Enter your password" }]}
        >
          <Input.Password
            prefix={<LockPasswordIcon size={18} className="mr-1 opacity-60" />}
            placeholder="••••••••"
            autoComplete="current-password"
          />
        </Form.Item>
        <Button type="primary" htmlType="submit" block className="mt-2">
          Log in
        </Button>
      </Form>
      <p className="text-sm text-center mt-5" style={{ color: "#52585d" }}>
        New here?{" "}
        <button
          onClick={onSwitchToRegister}
          className="font-medium underline underline-offset-2"
          style={{ color: "#24344a" }}
        >
          Create an account
        </button>
      </p>
    </Modal>
  );
}

export function RegisterModal({ open, onClose, onSwitchToLogin }) {
  const [form] = Form.useForm();

  const handleFinish = (values) => {
    console.log("register form submitted", values);
    message.info("Not wired to the API yet — this is a front-end-only demo.");
  };

  return (
    <Modal
      open={open}
      onCancel={onClose}
      footer={null}
      title={
        <span className="font-semibold text-lg" style={{ color: "#16191c" }}>
          Create your workspace
        </span>
      }
      width={420}
      destroyOnHidden
    >
      <p className="text-sm mb-5" style={{ color: "#52585d" }}>
        Free to start. No card required until you outgrow the Free plan.
      </p>
      <Form
        form={form}
        layout="vertical"
        onFinish={handleFinish}
        requiredMark={false}
      >
        <Form.Item
          name="businessName"
          label="Business name"
          rules={[{ required: true, message: "Enter your business name" }]}
        >
          <Input
            prefix={<Building06Icon size={18} className="mr-1 opacity-60" />}
            placeholder="Acme Co"
            autoComplete="organization"
          />
        </Form.Item>
        <Form.Item
          name="email"
          label="Email"
          rules={[
            { required: true, message: "Enter your email" },
            { type: "email", message: "Enter a valid email" },
          ]}
        >
          <Input
            prefix={<Mail01Icon size={18} className="mr-1 opacity-60" />}
            placeholder="you@business.com"
            autoComplete="email"
          />
        </Form.Item>
        <Form.Item
          name="password"
          label="Password"
          rules={[
            { required: true, message: "Enter a password" },
            { min: 8, message: "At least 8 characters" },
          ]}
        >
          <Input.Password
            prefix={<LockPasswordIcon size={18} className="mr-1 opacity-60" />}
            placeholder="At least 8 characters"
            autoComplete="new-password"
          />
        </Form.Item>
        <Button type="primary" htmlType="submit" block className="mt-2">
          Create workspace
        </Button>
      </Form>
      <p className="text-sm text-center mt-5" style={{ color: "#52585d" }}>
        Already have a workspace?{" "}
        <button
          onClick={onSwitchToLogin}
          className="font-medium underline underline-offset-2"
          style={{ color: "#24344a" }}
        >
          Log in
        </button>
      </p>
    </Modal>
  );
}

//Hook to control both modals and the switch btw them
export function useAuthModals() {
  const [mode, setMode] = useState(null); // null | 'login' | 'register'
  return {
    mode,
    openLogin: () => setMode("login"),
    openRegister: () => setMode("register"),
    close: () => setMode(null),
  };
}
