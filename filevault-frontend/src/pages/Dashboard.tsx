import { useEffect, useState } from "react";
import {
  MDBCard,
  MDBCardBody,
  MDBTable,
  MDBTableHead,
  MDBTableBody,
  MDBBtn,
  MDBInput,
  MDBIcon,
} from "mdb-react-ui-kit";

const API_BASE = "http://localhost:8080";

// Renamed to avoid conflict with browser File type
interface DbFile {
  id: number;
  filename: string;
  uploaded_at: string;
  user_id?: number;
}

function Dashboard() {
  const [files, setFiles] = useState<DbFile[]>([]);
  const [sharedFiles, setSharedFiles] = useState<DbFile[]>([]);
  const [file, setFile] = useState<File | null>(null); // now using browser File type
  const [userLabel, setUserLabel] = useState<string>("User");

  const token = localStorage.getItem("token");
  if (!token) {
    window.location.href = "/auth";
  }

  // Decode JWT → user id
  useEffect(() => {
    try {
      const payload = JSON.parse(atob(token!.split(".")[1]));
      setUserLabel(`User ID: ${payload.user_id}`);
    } catch {
      setUserLabel("User");
    }
  }, [token]);

  // Load my files
  const loadFiles = async () => {
    try {
      const res = await fetch(`${API_BASE}/files`, {
        headers: { Authorization: "Bearer " + token },
      });
      const data = await res.json();
      setFiles(Array.isArray(data) ? data : []);
    } catch (err) {
      console.error(err);
    }
  };

  // Load shared files
  const loadSharedFiles = async () => {
    try {
      const res = await fetch(`${API_BASE}/shared`, {
        headers: { Authorization: "Bearer " + token },
      });
      const data = await res.json();
      setSharedFiles(Array.isArray(data) ? data : []);
    } catch (err) {
      console.error(err);
    }
  };

  // Upload file
  const handleUpload = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!file) return;

    const formData = new FormData();
    formData.append("file", file); // ✅ proper browser File

    try {
      const res = await fetch(`${API_BASE}/files`, {
        method: "POST",
        headers: { Authorization: "Bearer " + token },
        body: formData,
      });

      if (!res.ok) {
        const errMsg = await res.text();
        throw new Error(errMsg || "Upload failed");
      }

      alert("✅ File uploaded");
      setFile(null);
      loadFiles();
    } catch (err) {
      console.error(err);
      alert("❌ Upload failed");
    }
  };

  // Download file
  const handleDownload = async (id: number, filename: string) => {
    try {
      const res = await fetch(`${API_BASE}/files/${id}`, {
        headers: { Authorization: "Bearer " + token },
      });
      if (!res.ok) throw new Error("Download failed");
      const blob = await res.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = filename;
      a.click();
      window.URL.revokeObjectURL(url);
    } catch {
      alert("❌ Download failed");
    }
  };

  // Delete file
  const handleDelete = async (id: number) => {
    if (!confirm("Are you sure?")) return;
    try {
      const res = await fetch(`${API_BASE}/files/${id}`, {
        method: "DELETE",
        headers: { Authorization: "Bearer " + token },
      });
      if (!res.ok) throw new Error("Delete failed");
      alert("✅ File deleted");
      loadFiles();
    } catch {
      alert("❌ Delete failed");
    }
  };

  // Share file
  const handleShare = async (id: number) => {
    const targetUser = prompt("Enter target user ID:");
    if (!targetUser) return;
    try {
      const res = await fetch(`${API_BASE}/share`, {
        method: "POST",
        headers: {
          Authorization: "Bearer " + token,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ file_id: id, target_user: parseInt(targetUser) }),
      });
      if (!res.ok) throw new Error("Share failed");
      alert("✅ File shared");
      loadSharedFiles();
    } catch {
      alert("❌ Share failed");
    }
  };

  useEffect(() => {
    loadFiles();
    loadSharedFiles();
  }, []);

  return (
    <div className="d-flex" style={{ height: "100vh", width: "100%" }}>
      {/* Sidebar */}
      <div
        className="bg-primary text-white p-3 d-flex flex-column"
        style={{ width: "20%", minWidth: "200px" }}
      >
        <h4 className="fw-bold mb-4">
          <MDBIcon fas icon="vault" className="me-2" />
          File Vault
        </h4>
        <span className="mb-4">{userLabel}</span>
        <hr className="text-light" />

        <MDBBtn
          color="light"
          className="mb-2 text-start"
          block
          onClick={() => loadFiles()}
        >
          <MDBIcon fas icon="folder" className="me-2" /> My Files
        </MDBBtn>

        <MDBBtn
          color="light"
          className="mb-2 text-start"
          block
          onClick={() => loadSharedFiles()}
        >
          <MDBIcon fas icon="share-alt" className="me-2" /> Shared With Me
        </MDBBtn>

        <div className="mt-auto">
          <MDBBtn
            color="danger"
            block
            onClick={() => {
              localStorage.removeItem("token");
              window.location.href = "/auth";
            }}
          >
            <MDBIcon fas icon="sign-out-alt" className="me-2" />
            Logout
          </MDBBtn>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-grow-1 p-4 overflow-auto" style={{ width: "80%" }}>
        <div className="w-100">
          {/* Upload */}
          <MDBCard className="mb-4 shadow-sm w-100">
            <MDBCardBody>
              <h4 className="text-primary mb-3">Upload File</h4>
              <form
                onSubmit={handleUpload}
                className="d-flex flex-column flex-md-row gap-2 w-100"
              >
                <MDBInput
                  type="file"
                  onChange={(e: any) => setFile(e.target.files?.[0] || null)}
                  required
                  className="w-100"
                />
                <MDBBtn color="primary" type="submit">
                  Upload
                </MDBBtn>
              </form>
            </MDBCardBody>
          </MDBCard>

          {/* My Files */}
          <MDBCard className="mb-4 shadow-sm w-100">
            <MDBCardBody>
              <h4 className="text-success mb-3">My Files</h4>
              <MDBTable striped hover responsive className="w-100">
                <MDBTableHead light>
                  <tr>
                    <th>Filename</th>
                    <th>Uploaded At</th>
                    <th className="text-center">Actions</th>
                  </tr>
                </MDBTableHead>
                <MDBTableBody>
                  {files.length === 0 ? (
                    <tr>
                      <td colSpan={3} className="text-center text-muted">
                        No files uploaded yet
                      </td>
                    </tr>
                  ) : (
                    files.map((f) => (
                      <tr key={f.id}>
                        <td>{f.filename}</td>
                        <td>{new Date(f.uploaded_at).toLocaleString()}</td>
                        <td className="d-flex flex-wrap justify-content-center gap-2">
                          <MDBBtn
                            size="sm"
                            color="success"
                            onClick={() => handleDownload(f.id, f.filename)}
                          >
                            Download
                          </MDBBtn>
                          <MDBBtn
                            size="sm"
                            color="danger"
                            onClick={() => handleDelete(f.id)}
                          >
                            Delete
                          </MDBBtn>
                          <MDBBtn
                            size="sm"
                            color="info"
                            onClick={() => handleShare(f.id)}
                          >
                            Share
                          </MDBBtn>
                        </td>
                      </tr>
                    ))
                  )}
                </MDBTableBody>
              </MDBTable>
            </MDBCardBody>
          </MDBCard>

          {/* Shared With Me */}
          <MDBCard className="mb-4 shadow-sm w-100">
            <MDBCardBody>
              <h4 className="text-warning mb-3">Shared With Me</h4>
              <MDBTable striped hover responsive className="w-100">
                <MDBTableHead light>
                  <tr>
                    <th>Filename</th>
                    <th>Shared By</th>
                    <th className="text-center">Actions</th>
                  </tr>
                </MDBTableHead>
                <MDBTableBody>
                  {sharedFiles.length === 0 ? (
                    <tr>
                      <td colSpan={3} className="text-center text-muted">
                        No files shared yet
                      </td>
                    </tr>
                  ) : (
                    sharedFiles.map((f) => (
                      <tr key={f.id}>
                        <td>{f.filename}</td>
                        <td>{f.user_id || "Unknown"}</td>
                        <td className="text-center">
                          <MDBBtn
                            size="sm"
                            color="success"
                            onClick={() => handleDownload(f.id, f.filename)}
                          >
                            Download
                          </MDBBtn>
                        </td>
                      </tr>
                    ))
                  )}
                </MDBTableBody>
              </MDBTable>
            </MDBCardBody>
          </MDBCard>
        </div>
      </div>
    </div>
  );
}

export default Dashboard;
