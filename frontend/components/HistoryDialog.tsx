"use client";

import { useEffect, useState } from "react";
import Modal from "./Modal";
import { api } from "@/lib/api";
import type { Item, Movement } from "@/lib/types";

const LABELS = { in: "Received", out: "Shipped", adjust: "Count set" } as const;

export default function HistoryDialog({ item, onClose }: { item: Item; onClose: () => void }) {
  const [rows, setRows] = useState<Movement[] | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api
      .itemMovements(item.id)
      .then(setRows)
      .catch((e) => setError(e instanceof Error ? e.message : "Could not load history"));
  }, [item.id]);

  return (
    <Modal title={`History: ${item.name}`} onClose={onClose}>
      {error && (
        <p className="form-error" role="alert">
          {error}
        </p>
      )}
      {!rows && !error && <p className="muted">Loading…</p>}
      {rows && rows.length === 0 && (
        <p className="muted">No movements yet. Receive or ship stock to start the history.</p>
      )}
      {rows && rows.length > 0 && (
        <ul className="history">
          {rows.map((m) => (
            <li key={m.id}>
              <div>
                <strong>{LABELS[m.type]}</strong>
                <span className="muted"> {new Date(m.created_at).toLocaleString()}</span>
                {m.note && <div className="note">{m.note}</div>}
              </div>
              <div className="history-nums">
                <span className={m.delta < 0 ? "bad" : "good"}>
                  {m.delta > 0 ? `+${m.delta}` : m.delta}
                </span>
                <span className="muted">{m.balance} left</span>
              </div>
            </li>
          ))}
        </ul>
      )}
    </Modal>
  );
}
