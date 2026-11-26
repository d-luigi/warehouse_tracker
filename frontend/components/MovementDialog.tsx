"use client";

import { FormEvent, useState } from "react";
import Modal from "./Modal";
import { api, units } from "@/lib/api";
import type { Item, MovementType } from "@/lib/types";

const MODES: { type: MovementType; label: string; qtyLabel: string }[] = [
  { type: "in", label: "Receive", qtyLabel: "Units received" },
  { type: "out", label: "Ship", qtyLabel: "Units shipped" },
  { type: "adjust", label: "Set count", qtyLabel: "Counted units" },
];

export default function MovementDialog({
  item,
  onClose,
  onDone,
}: {
  item: Item;
  onClose: () => void;
  onDone: (message: string) => void;
}) {
  const [type, setType] = useState<MovementType>("in");
  const [qty, setQty] = useState("");
  const [note, setNote] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const mode = MODES.find((m) => m.type === type)!;
  const n = Number(qty);
  const validNumber = qty !== "" && Number.isInteger(n) && (type === "adjust" ? n >= 0 : n > 0);
  const after = !validNumber
    ? null
    : type === "in"
      ? item.quantity + n
      : type === "out"
        ? item.quantity - n
        : n;
  const tooMany = after !== null && after < 0;

  async function submit(e: FormEvent) {
    e.preventDefault();
    if (!validNumber || tooMany) return;
    setBusy(true);
    setError(null);
    try {
      await api.recordMovement(item.id, { type, quantity: n, note });
      const verb =
        type === "in"
          ? `Received ${units(n)} of ${item.name}`
          : type === "out"
            ? `Shipped ${units(n)} of ${item.name}`
            : `Set ${item.name} to ${units(n)}`;
      onDone(verb);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Something went wrong");
      setBusy(false);
    }
  }

  return (
    <Modal title={`Move stock: ${item.name}`} onClose={onClose}>
      <form className="form" onSubmit={submit}>
        <div className="segmented" role="radiogroup" aria-label="Type of movement">
          {MODES.map((m) => (
            <button
              key={m.type}
              type="button"
              role="radio"
              aria-checked={type === m.type}
              className={type === m.type ? "seg seg-on" : "seg"}
              onClick={() => setType(m.type)}
            >
              {m.label}
            </button>
          ))}
        </div>

        <label className="field">
          <span>{mode.qtyLabel}</span>
          <input
            type="number"
            min={type === "adjust" ? 0 : 1}
            step={1}
            value={qty}
            onChange={(e) => setQty(e.target.value)}
            autoFocus
            required
          />
        </label>
        <label className="field">
          <span>Note</span>
          <input
            value={note}
            onChange={(e) => setNote(e.target.value)}
            placeholder="Order number, supplier, reason"
          />
        </label>

        <p className="preview">
          Now in stock: <strong>{item.quantity}</strong>
          {after !== null && (
            <>
              {" "}
              After this: <strong className={tooMany ? "bad" : ""}>{after}</strong>
            </>
          )}
        </p>
        {tooMany && (
          <p className="form-error" role="alert">
            Only {units(item.quantity)} in stock, so you can&apos;t ship {units(n)}.
          </p>
        )}
        {error && (
          <p className="form-error" role="alert">
            {error}
          </p>
        )}

        <div className="form-actions">
          <button type="button" className="btn btn-quiet" onClick={onClose}>
            Cancel
          </button>
          <button type="submit" className="btn btn-primary" disabled={busy || !validNumber || tooMany}>
            Record movement
          </button>
        </div>
      </form>
    </Modal>
  );
}
