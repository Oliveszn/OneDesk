import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { ConfigProvider } from "antd";
import "./index.css";
import App from "./App.jsx";

// Ant Design's theme tokens are remapped to our own palette
const theme = {
  token: {
    colorPrimary: "#24344a", // ledger navy
    colorPrimaryHover: "#1a2739", // ledger-dark
    colorLink: "#24344a",
    colorText: "#16191c",
    colorTextSecondary: "#52585d",
    colorBorder: "#e1ded6",
    borderRadius: 2,
    fontFamily: "'Inter', ui-sans-serif, system-ui, sans-serif",
    colorBgContainer: "#fbfaf8",
  },
  components: {
    Button: {
      controlHeight: 44,
      fontWeight: 500,
      primaryShadow: "none",
    },
    Modal: {
      borderRadiusLG: 4,
    },
    Input: {
      controlHeight: 42,
      activeShadow: "0 0 0 2px rgba(36, 52, 74, 0.12)",
    },
  },
};

createRoot(document.getElementById("root")).render(
  <StrictMode>
    <ConfigProvider theme={theme}>
      <App />
    </ConfigProvider>
  </StrictMode>,
);
