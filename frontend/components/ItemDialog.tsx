"use client";

import { FormEvent, useState } from "react";
import Modal from "./Modal";
import { api } from "@/lib/api";
import type { Item } from "@/lib/types";

export default function ItemDialog({
  item,
  onClose,
  onDone,
}: {
  item?: Item;
  onClose: () => void;
  onDone: (message: string) => void;
}) {
  const editing = Boolean(item);
  const [sku, setSku] = useState(item?.sku ?? "");
  const [name, setName] = useState(item?.name ?? "");
  const [location, setLocation] = useState(item?.location ?? "");
  const [description, setDescription] = useState(item?.description ?? "");
  const [quantity, setQuantity] = useState("0");
  const [reorder, setReorder] = useState(String(item?.reorder_level ?? 0));
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError(null);
    try {
      const reorder_level = Number(reorder || 0);
      if (item) {
        await api.updateItem(item.id, { sku, name, description, location, reorder_level });
        onDone(`Saved changes to ${name}`);
      } else {
        await api.createItem({
          sku,
          name,
          description,
          location,
          quantity: Number(quantity || 0),
          reorder_level,
        });
        onDone(`Added ${name}`);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Something went wrong");
      setBusy(false);
    }
  }

  return (
    <Modal title={editing ? `Edit ${item?.name}` : "Add item"} onClose={onClose}>
      <form className="form" onSubmit={submit}>
        <div className="field-row">
          <label className="field">
            <span>SKU</span>
            <input value={sku} onChange={(e) => setSku(e.target.value)} required autoFocus />
          </label>
          <label className="field">
            <span>Shelf location</span>
            <input
              value={location}
              onChange={(e) => setLocation(e.target.value)}
              placeholder="A-03-2"
            />
          </label>
        </div>
        <label className="field">
          <span>Name</span>
          <input value={name} onChange={(e) => setName(e.target.value)} required />
        </label>
        <label className="field">
          <span>Description</span>
          <input value={description} onChange={(e) => setDescription(e.target.value)} />
        </label>
        <div className="field-row">
          {!editing && (
            <label className="field">
              <span>Starting stock</span>
              <input
                type="number"
                min={0}
                step={1}
                value={quantity}
                onChange={(e) => setQuantity(e.target.value)}
              />
            </label>
          )}
          <label className="field">
            <span>Warn me when stock drops to</span>
            <input
              type="number"
              min={0}
              step={1}
              value={reorder}
              onChange={(e) => setReorder(e.target.value)}
            />
          </label>
        </div>
        {error && (
          <p className="form-error" role="alert">
            {error}
          </p>
        )}
        <div className="form-actions">
          <button type="button" className="btn btn-quiet" onClick={onClose}>
            Cancel
          </button>
          <button type="submit" className="btn btn-primary" disabled={busy}>
            {editing ? "Save changes" : "Add item"}
          </button>
        </div>
      </form>
    </Modal>
  );
}
