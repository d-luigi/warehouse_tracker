"use client";

import { useCallback, useEffect, useState } from "react";
import { API_URL, api } from "@/lib/api";
import type { Item, Summary } from "@/lib/types";
import ItemDialog from "@/components/ItemDialog";
import MovementDialog from "@/components/MovementDialog";
import HistoryDialog from "@/components/HistoryDialog";

type DialogState =
  | { kind: "create" }
  | { kind: "edit"; item: Item }
  | { kind: "move"; item: Item }
  | { kind: "history"; item: Item }
  | null;

export default function Home() {
  const [items, setItems] = useState<Item[]>([]);
  const [summary, setSummary] = useState<Summary | null>(null);
  const [query, setQuery] = useState("");
  const [lowOnly, setLowOnly] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);
  const [dialog, setDialog] = useState<DialogState>(null);

  const load = useCallback(async () => {
    try {
      const [list, sum] = await Promise.all([api.listItems(query, lowOnly), api.summary()]);
      setItems(list);
      setSummary(sum);
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Unknown error");
    } finally {
      setLoading(false);
    }
  }, [query, lowOnly]);

  useEffect(() => {
    const t = setTimeout(load, 200);
    return () => clearTimeout(t);
  }, [load]);

  useEffect(() => {
    if (!notice) return;
    const t = setTimeout(() => setNotice(null), 4000);
    return () => clearTimeout(t);
  }, [notice]);

  const closeDialog = useCallback(() => setDialog(null), []);

  function done(message: string) {
    setDialog(null);
    setNotice(message);
    load();
  }

  async function remove(item: Item) {
    if (!window.confirm(`Delete ${item.name}? Its movement history will be deleted too.`)) return;
    try {
      await api.deleteItem(item.id);
      setNotice(`Deleted ${item.name}`);
      load();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not delete item");
    }
  }

  const filtering = query.trim() !== "" || lowOnly;

  return (
    <main className="shell">
      <header className="top">
        <div className="brand">
          <h1>Warehouse Tracker</h1>
          <p>What&apos;s on the shelves, what came in, what went out.</p>
        </div>
        <button className="btn btn-primary" onClick={() => setDialog({ kind: "create" })}>
          Add item
        </button>
      </header>

      <div className="chips">
        <div className="chip"><strong>{summary?.total_items ?? "–"}</strong><span>items</span></div>
        <div className="chip"><strong>{summary?.total_units ?? "–"}</strong><span>units in stock</span></div>
        <div className={summary && summary.low_stock_items > 0 ? "chip chip-warn" : "chip"}>
          <strong>{summary?.low_stock_items ?? "–"}</strong><span>running low</span>
        </div>
        <div className="chip"><strong>{summary?.total_movements ?? "–"}</strong><span>movements logged</span></div>
      </div>

      <section className="panel">
        <div className="toolbar">
          <input
            className="search"
            type="search"
            placeholder="Search by name, SKU or shelf"
            aria-label="Search items"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
          />
          <label className="toggle">
            <input type="checkbox" checked={lowOnly} onChange={(e) => setLowOnly(e.target.checked)} />
            Running low only
          </label>
        </div>

        {notice && <div className="notice" role="status">{notice}</div>}

        {error && (
          <div className="error-box" role="alert">
            Can&apos;t reach the backend at <code>{API_URL}</code> ({error}). Start it from the
            backend folder with <code>go run .</code>, then reload this page.
          </div>
        )}

        {!error && !loading && items.length === 0 && (
          <div className="empty">
            <h2>{filtering ? "Nothing matches that" : "No items yet"}</h2>
            <p>
              {filtering
                ? "Try a different search or turn off the low stock filter."
                : "Add your first item to start tracking stock."}
            </p>
            {!filtering && (
              <button className="btn btn-primary" onClick={() => setDialog({ kind: "create" })}>
                Add item
              </button>
            )}
          </div>
        )}

        {!error && items.length > 0 && (
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>Item</th>
                  <th>Shelf</th>
                  <th>In stock</th>
                  <th />
                </tr>
              </thead>
              <tbody>
                {items.map((item) => (
                  <tr key={item.id}>
                    <td>
                      <div className="item-name">{item.name}</div>
                      <div className="item-sku">{item.sku}</div>
                    </td>
                    <td>{item.location || "–"}</td>
                    <td>
                      <span className="qty">{item.quantity}</span>
                      {item.low_stock && <span className="badge">Running low</span>}
                    </td>
                    <td>
                      <div className="row-actions">
                        <button className="btn btn-sm btn-primary" onClick={() => setDialog({ kind: "move", item })}>
                          Move stock
                        </button>
                        <button className="btn btn-sm" onClick={() => setDialog({ kind: "history", item })}>
                          History
                        </button>
                        <button className="btn btn-sm" onClick={() => setDialog({ kind: "edit", item })}>
                          Edit
                        </button>
                        <button className="btn btn-sm btn-danger" onClick={() => remove(item)}>
                          Delete
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>

      {dialog?.kind === "create" && <ItemDialog onClose={closeDialog} onDone={done} />}
      {dialog?.kind === "edit" && <ItemDialog item={dialog.item} onClose={closeDialog} onDone={done} />}
      {dialog?.kind === "move" && <MovementDialog item={dialog.item} onClose={closeDialog} onDone={done} />}
      {dialog?.kind === "history" && <HistoryDialog item={dialog.item} onClose={closeDialog} />}
    </main>
  );
}
