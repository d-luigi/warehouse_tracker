export type MovementType = "in" | "out" | "adjust";

export interface Item {
  id: number;
  sku: string;
  name: string;
  description?: string;
  location: string;
  quantity: number;
  reorder_level: number;
  low_stock: boolean;
  created_at: string;
  updated_at: string;
}

export interface ItemInput {
  sku: string;
  name: string;
  description: string;
  location: string;
  quantity: number;
  reorder_level: number;
}

export type ItemUpdate = Omit<ItemInput, "quantity">;

export interface Movement {
  id: number;
  item_id: number;
  type: MovementType;
  delta: number;
  balance: number;
  note?: string;
  created_at: string;
}

export interface MovementInput {
  type: MovementType;
  quantity: number;
  note: string;
}

export interface Summary {
  total_items: number;
  total_units: number;
  low_stock_items: number;
  total_movements: number;
}
