import Nav from "../components/Nav";
import Hero from "../components/Hero";
import Modules from "../components/Modules";
import Pricing from "../components/Pricing";
import FinalCTA from "../components/FinalCTA";
import {
  LoginModal,
  RegisterModal,
  useAuthModals,
} from "../components/AuthModals";

function LandingPage() {
  const auth = useAuthModals();

  return (
    <div style={{ backgroundColor: "#f6f5f1" }}>
      <Nav onLogin={auth.openLogin} onRegister={auth.openRegister} />
      <main>
        <Hero onRegister={auth.openRegister} />
        <Modules />
        <Pricing />
        <FinalCTA onRegister={auth.openRegister} />
      </main>

      <LoginModal
        open={auth.mode === "login"}
        onClose={auth.close}
        onSwitchToRegister={auth.openRegister}
      />
      <RegisterModal
        open={auth.mode === "register"}
        onClose={auth.close}
        onSwitchToLogin={auth.openLogin}
      />
    </div>
  );
}

export default LandingPage;
