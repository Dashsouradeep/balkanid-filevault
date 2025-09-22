// src/pages/LoginSignup.tsx
import "./LoginSignup.css";
import React, { useState } from "react";
import {
  MDBContainer,
  MDBCol,
  MDBRow,
  MDBBtn,
  MDBIcon,
  MDBInput,
  MDBCheckbox,
} from "mdb-react-ui-kit";
import { useNavigate } from "react-router-dom";
import "./LoginSignup.css";

const API_BASE = import.meta.env.VITE_API_BASE ?? "http://localhost:8080";

export default function LoginSignup() {
  const navigate = useNavigate();
  const [email, setEmail] = useState("");
  const [username, setUsername] = useState(""); // optional for register
  const [password, setPassword] = useState("");
  const [isLogin, setIsLogin] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const handleSubmit = async () => {
    setError(null);
    setLoading(true);
    try {
      const endpoint = isLogin ? "/login" : "/register";
      const body: any = { email, password };
      if (!isLogin) body.username = username;

      const res = await fetch(`${API_BASE}${endpoint}`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });

      const text = await res.text();
      // backend sometimes returns a JSON token, or error text; try parse
      let data: any;
      try {
        data = JSON.parse(text);
      } catch {
        data = { message: text };
      }

      if (!res.ok) {
        setError(data.message || text || "Request failed");
        setLoading(false);
        return;
      }

      if (isLogin) {
        if (data.token) {
          localStorage.setItem("token", data.token);
          navigate("/dashboard");
        } else {
          setError("Login succeeded but no token received");
        }
      } else {
        // registration - switch to login automatically
        setIsLogin(true);
        setError("Registration successful. Please log in.");
      }
    } catch (err: any) {
      setError(err.message || "Network error");
    } finally {
      setLoading(false);
    }
  };

  return (
    <MDBContainer fluid className="p-3 my-5 h-custom">
      <MDBRow>
        <MDBCol col="10" md="6">
          <img
            src="https://mdbcdn.b-cdn.net/img/Photos/new-templates/bootstrap-login-form/draw2.webp"
            className="img-fluid"
            alt="Sample"
          />
        </MDBCol>

        <MDBCol col="4" md="6">
          <div className="d-flex flex-row align-items-center justify-content-center">
            <p className="lead fw-normal mb-0 me-3">Sign in with</p>

            <MDBBtn floating size="md" tag="a" className="me-2">
              <MDBIcon fab icon="facebook-f" />
            </MDBBtn>

            <MDBBtn floating size="md" tag="a" className="me-2">
              <MDBIcon fab icon="twitter" />
            </MDBBtn>

            <MDBBtn floating size="md" tag="a" className="me-2">
              <MDBIcon fab icon="linkedin-in" />
            </MDBBtn>
          </div>

          <div className="divider d-flex align-items-center my-4">
            <p className="text-center fw-bold mx-3 mb-0">Or</p>
          </div>

          {!isLogin && (
            <MDBInput
              wrapperClass="mb-4"
              label="Username"
              id="usernameInput"
              type="text"
              size="lg"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
            />
          )}

          <MDBInput
            wrapperClass="mb-4"
            label="Email address"
            id="formControlEmail"
            type="email"
            size="lg"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
          />
          <MDBInput
            wrapperClass="mb-4"
            label="Password"
            id="formControlPassword"
            type="password"
            size="lg"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />

          <div className="d-flex justify-content-between mb-4">
            <MDBCheckbox name="flexCheck" id="flexCheckDefault" label="Remember me" />
            <a href="#!">Forgot password?</a>
          </div>

          {error && <p className="text-danger">{error}</p>}

          <div className="text-center text-md-start mt-4 pt-2">
            <MDBBtn
              className="mb-0 px-5"
              size="lg"
              onClick={handleSubmit}
              disabled={loading}
            >
              {loading ? "Please wait..." : isLogin ? "Login" : "Register"}
            </MDBBtn>
            <p className="small fw-bold mt-2 pt-1 mb-2">
              {isLogin ? "Don't have an account?" : "Already have an account?"}{" "}
              <a
                href="#!"
                className="link-danger"
                onClick={() => {
                  setIsLogin(!isLogin);
                  setError(null);
                }}
              >
                {isLogin ? "Register" : "Login"}
              </a>
            </p>
          </div>
        </MDBCol>
      </MDBRow>

      <div className="d-flex flex-column flex-md-row text-center text-md-start justify-content-between py-4 px-4 px-xl-5 bg-primary mt-4">
        <div className="text-white mb-3 mb-md-0">Copyright © 2025. All rights reserved.</div>

        <div>
          <MDBBtn tag="a" color="none" className="mx-3" style={{ color: "white" }}>
            <MDBIcon fab icon="facebook-f" size="md" />
          </MDBBtn>

          <MDBBtn tag="a" color="none" className="mx-3" style={{ color: "white" }}>
            <MDBIcon fab icon="twitter" size="md" />
          </MDBBtn>

          <MDBBtn tag="a" color="none" className="mx-3" style={{ color: "white" }}>
            <MDBIcon fab icon="google" size="md" />
          </MDBBtn>

          <MDBBtn tag="a" color="none" className="mx-3" style={{ color: "white" }}>
            <MDBIcon fab icon="linkedin-in" size="md" />
          </MDBBtn>
        </div>
      </div>
    </MDBContainer>
  );
}
