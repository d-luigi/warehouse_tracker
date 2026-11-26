import type {
  Item,
  ItemInput,
  ItemUpdate,
  Movement,
  MovementInput,
  Summary,
} from "./types";

export const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export class ApiError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.status = status;
  }
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers: Record<string, string> = {};
  if (init.body) headers["Content-Type"] = "application/json";

  const res = await fetch(`${API_URL}${path}`, {
    ...init,
    headers,
    cache: "no-store",
  });

  if (res.status === 204) return undefined as T;

  const body = await res.json().catch(() => null);
  if (!res.ok) {
    throw new ApiError(body?.error ?? res.statusText, res.status);
  }
  return body as T;
}

export const api = {
  summary: () => request<Summary>("/api/summary"),

  listItems: (q: string, lowStockOnly: boolean) => {
    const params = new URLSearchParams();
    if (q.trim()) params.set("q", q.trim());
    if (lowStockOnly) params.set("low_stock", "true");
    const qs = params.toString();
    return request<Item[]>(`/api/items${qs ? `?${qs}` : ""}`);
  },

  createItem: (input: ItemInput) =>
    request<Item>("/api/items", { method: "POST", body: JSON.stringify(input) }),

  updateItem: (id: number, input: ItemUpdate) =>
    request<Item>(`/api/items/${id}`, { method: "PUT", body: JSON.stringify(input) }),

  deleteItem: (id: number) => request<void>(`/api/items/${id}`, { method: "DELETE" }),

  itemMovements: (id: number) => request<Movement[]>(`/api/items/${id}/movements`),

  recordMovement: (id: number, input: MovementInput) =>
    request<{ movement: Movement; item: Item }>(`/api/items/${id}/movements`, {
      method: "POST",
      body: JSON.stringify(input),
    }),
};

export function units(n: number): string {
  return `${n} ${Math.abs(n) === 1 ? "unit" : "units"}`;
}
