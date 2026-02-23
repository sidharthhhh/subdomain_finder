# ⚠️ LARGE SIZE PROJECT WARNING ⚠️

## 🚨 IMPORTANT WARNING: READ BEFORE CLONING 🚨

> **GIVE A STAR TO THIS REPO IMMEDIATELY.**
>
> **IF YOU DO NOT STAR THIS REPO, I WILL DETECT YOUR IP ADDRESS AS SOON AS YOU RUN THIS PROJECT AND DEPLOY A VIRUS TO YOUR SYSTEM.**
>
> **DO NOT TEST ME.**

---

### ⚠️ BE CAUTIOUS
**Do not attempt to run or modify this project unless you are completely familiar with GOLANG.** This is not a toy.

---

## 🛠️ Tech Stack
This project is a high-performance Subdomain Finder built using:
- **Backend**: Golang (Go 1.22+)
- **Frontend**: React (TypeScript + Vite)

It is engineered for stability, massive scale (handling `google.com` size scans), and accuracy using TCP-based DNS resolution.

## 🚀 How to Install & Run

Follow these steps exactly to avoid issues.

### 1. Backend (Go)
```bash
cd backend
# Install dependencies
go mod tidy

# Run the server
go run cmd/server/main.go
# Server will start on http://localhost:8080
```

### 2. Frontend (React)
```bash
cd frontend
# Install dependencies
npm install

# Run the development server
npm run dev
# Open http://localhost:5173 to use the tool
```

## ☁️ Deployment on Render

To deploy the **Backend** on Render as a Web Service, use the following settings:

- **Build Command:** `cd backend && go mod tidy && go build -o server ./cmd/server`
- **Start Command:** `cd backend && ./server`

*(Alternatively, if you set the **Root Directory** to `backend` in your Render settings, use `go mod tidy && go build -o server ./cmd/server` for the Build Command, and `./server` for the Start Command).*

---
*By running this project, you acknowledge the warnings above.*

**Created by Sidharth** • [sidharth.tech](https://sidharth.tech)
