# 📁 BalkanID FileVault

Capstone Internship Task — Secure File Vault System  
Built with **Go (backend)**, **React + Vite (frontend)**, and **PostgreSQL (database)**

---

## ✨ Features

- 🔐 **Authentication & Authorization**
  - User Registration & Login
  - JWT-based authentication

- 📤 **File Management**
  - Upload files with **deduplication**
  - Download files
  - Delete files
  - Enforce user storage **quota**

- 🤝 **File Sharing**
  - Share files with other users (read-only)
  - View files shared with you

- 📊 **Dashboard**
  - View your uploaded files
  - View files shared with you
  - Check quota usage

---

## 🛠 Tech Stack

- **Backend**: [Go](https://go.dev/) (mux, pgx, JWT, bcrypt)
- **Frontend**: [React](https://react.dev/) + [Vite](https://vitejs.dev/) + TypeScript + [MDB React UI Kit](https://mdbootstrap.com/docs/react/)
- **Database**: [PostgreSQL](https://www.postgresql.org/)
- **Authentication**: JSON Web Tokens (JWT)

---

## ⚙️ Setup Instructions

### 1. Clone Repository
```bash
git clone <your-classroom-repo>
cd BalkanID_FileVault
2. Database Setup
Create PostgreSQL database:

sql
Copy code
CREATE DATABASE filevault;
\c filevault
\i schema.sql   -- run the schema file included in repo
Verify tables:

sql
Copy code
\dt
3. Backend Setup
bash
Copy code
cd backend
cp .env.example .env   # copy example env file
Update .env with your database credentials:

env
Copy code
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpassword
DB_NAME=filevault
JWT_SECRET=supersecretkey
Install dependencies & run:

bash
Copy code
go mod tidy
go run .
Backend will run at: http://localhost:8080

4. Frontend Setup
bash
Copy code
cd filevault-frontend
npm install
npm run dev
Frontend will run at: http://localhost:5173

🔑 API Endpoints
Auth
POST /register → Register user

POST /login → Login user (returns JWT)

Files
POST /files → Upload file (multipart form)

GET /files → List user files

GET /files/{id} → Download file

DELETE /files/{id} → Delete file

Sharing
POST /share → Share file with another user

GET /shared → List files shared with logged-in user

Storage
GET /storage → Get quota usage


<img width="548" height="565" alt="image" src="https://github.com/user-attachments/assets/e8c7e718-2734-4490-b447-36eac242dd8d" />
<img width="1658" height="848" alt="image" src="https://github.com/user-attachments/assets/ac933388-80ba-4db0-b26a-0697def97f49" />
<img width="1626" height="860" alt="image" src="https://github.com/user-attachments/assets/93cca8de-292b-42cd-b6a6-a8461d28138b" />

