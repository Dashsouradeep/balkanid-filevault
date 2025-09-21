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

const API_BASE = "http://localhost:8080";

function LoginSignup() {
  const [isLogin, setIsLogin] = useState(true);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [username, setUsername] = useState(""); // only for register
  const navigate = useNavigate();

  const handleSubmit = async () => {
    try {
      const endpoint = isLogin ? "/login" : "/register";
      const body = isLogin
        ? { email, password }
        : { username, email, password };

      const res = await fetch(`${API_BASE}${endpoint}`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });

      if (!res.ok) {
        const msg = await res.text();
        alert("❌ " + msg);
        return;
      }

      const data = await res.json();

      if (isLogin) {
        localStorage.setItem("token", data.token);
        navigate("/dashboard");
      } else {
        alert("✅ Registered successfully, please login now.");
        setIsLogin(true);
      }
    } catch (err) {
      console.error(err);
      alert("❌ Could not connect to backend");
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
            <p className="lead fw-normal mb-0 me-3">
              {isLogin ? "Login with" : "Sign up with"}
            </p>

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
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              size="lg"
              required
            />
          )}

          <MDBInput
            wrapperClass="mb-4"
            label="Email address"
            type="email"
            size="lg"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
          <MDBInput
            wrapperClass="mb-4"
            label="Password"
            type="password"
            size="lg"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />

          <div className="d-flex justify-content-between mb-4">
            <MDBCheckbox
              name="flexCheck"
              value=""
              id="flexCheckDefault"
              label="Remember me"
            />
            <a href="#!">Forgot password?</a>
          </div>

          <div className="text-center text-md-start mt-4 pt-2">
            <MDBBtn className="mb-0 px-5" size="lg" onClick={handleSubmit}>
              {isLogin ? "Login" : "Sign Up"}
            </MDBBtn>
            <p className="small fw-bold mt-2 pt-1 mb-2">
              {isLogin
                ? "Don't have an account?"
                : "Already have an account?"}{" "}
              <button
                className="link-danger"
                onClick={() => setIsLogin(!isLogin)}
              >
                {isLogin ? "Register" : "Login"}
              </button>
            </p>
          </div>
        </MDBCol>
      </MDBRow>

      <div className="d-flex flex-column flex-md-row text-center text-md-start justify-content-between py-4 px-4 px-xl-5 bg-primary">
        <div className="text-white mb-3 mb-md-0">
          Copyright © 2025. All rights reserved.
        </div>

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

export default LoginSignup;
