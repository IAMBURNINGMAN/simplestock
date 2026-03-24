export interface User {
  id: number
  username: string
  full_name: string
  role: string
}

export interface Category {
  id: number
  name: string
}

export interface Product {
  id: number
  name: string
  sku: string
  category_id: number | null
  category_name: string
  unit: string
  min_stock: number
  quantity: number
  purchase_price: number | null
  created_at: string
  updated_at: string
}

export interface DocumentItem {
  id: number
  document_id: number
  product_id: number
  product_name: string
  product_sku: string
  quantity: number
  price: number | null
}

export interface Document {
  id: number
  doc_type: 'incoming' | 'outgoing'
  doc_number: string
  counterparty: string | null
  expense_type: string | null
  status: 'draft' | 'posted'
  user_id: number
  doc_date: string
  created_at: string
  items?: DocumentItem[]
}

export interface Movement {
  id: number
  product_id: number
  product_name: string
  product_sku: string
  document_id: number | null
  inventory_id: number | null
  movement_type: string
  quantity: number
  created_at: string
}

export interface InventoryItem {
  id: number
  inventory_id: number
  product_id: number
  product_name: string
  product_sku: string
  expected_quantity: number
  actual_quantity: number
  difference: number
}

export interface Inventory {
  id: number
  inv_number: string
  status: 'active' | 'completed'
  user_id: number
  started_at: string
  completed_at: string | null
  items?: InventoryItem[]
}

export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
}

export interface DashboardSummary {
  total_products: number
  low_stock_count: number
  today_movements: number
  today_incoming: number
  today_outgoing: number
  total_stock_value: number
}
