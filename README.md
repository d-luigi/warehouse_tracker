# Warehouse Tracker

Go API (`backend/`) + PostgreSQL on Render, Next.js frontend (`frontend/`).

## Deploy
1. Push this repo to GitHub (`schema.sql` lives in `backend/internal/store/` and is applied automatically on startup).
2. Render → New → Blueprint → select the repo. It creates `warehouse-db` and `warehouse-tracker-api`.
3. In the API service's Environment tab, set `CORS_ORIGIN` to your frontend URL (e.g. `https://your-app.vercel.app`).
4. Deploy the frontend (e.g. Vercel, root directory `frontend`) with `NEXT_PUBLIC_API_URL` = your Render API URL.

## Local development
    export DATABASE_URL="postgres://user:pass@localhost:5432/warehouse?sslmode=disable"
    cd backend && go mod tidy && go run .
    cd frontend && npm install && npm run dev
