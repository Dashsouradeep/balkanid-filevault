import { BrowserRouter as Router, Routes, Route, Navigate } from "react-router-dom";
import LoginSignup from "./pages/LoginSignup";
import Dashboard from "./pages/Dashboard";

function App() {
  const token = localStorage.getItem("token");

  return (
    <Router>
      <Routes>
        <Route path="/" element={<Navigate to={token ? "/dashboard" : "/auth"} />} />
        <Route path="/auth" element={<LoginSignup />} />
        <Route path="/dashboard" element={token ? <Dashboard /> : <Navigate to="/auth" />} />
      </Routes>
    </Router>
  );
}

export default App;
