# Warehouse Tracker – Frontend (Next.js + TypeScript)

Needs Node 18.18+ and the Go backend running (default http://localhost:8080).

    cd warehouse_tracker/frontend
    npm install
    npm run dev          # http://localhost:3000

To point at a different backend, copy `.env.local.example` to `.env.local` and edit it.

## Layout

    frontend/
    ├── package.json, tsconfig.json, next.config.mjs
    ├── app/
    │   ├── layout.tsx            fonts, page metadata, mounts the animated background
    │   ├── page.tsx              the tracker screen (stats, search, item table, dialogs)
    │   └── globals.css           colors, layout, buttons, forms
    ├── components/
    │   ├── Scene.tsx + scene.css animated warehouses, clouds and trucks (pure CSS motion)
    │   ├── Modal.tsx             shared dialog shell (Esc / click outside closes)
    │   ├── ItemDialog.tsx        add / edit item
    │   ├── MovementDialog.tsx    receive, ship or set a count
    │   └── HistoryDialog.tsx     movement history for one item
    └── lib/
        ├── api.ts                calls to the Go backend
        └── types.ts              TypeScript types matching the backend JSON
