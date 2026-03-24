import { useEffect, useState, useCallback } from 'react'
import { api } from '@/api/client'
import type { Product, Category, PaginatedResponse } from '@/api/types'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Plus, Search, Pencil, Trash2 } from 'lucide-react'

export default function Products() {
  const [products, setProducts] = useState<Product[]>([])
  const [categories, setCategories] = useState<Category[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editing, setEditing] = useState<Product | null>(null)
  const [form, setForm] = useState({ name: '', sku: '', category_id: '', unit: 'шт', min_stock: '0', purchase_price: '' })

  const load = useCallback(() => {
    api.get<PaginatedResponse<Product>>(`/products?page=${page}&page_size=20&search=${encodeURIComponent(search)}`).then((res) => {
      setProducts(res.data)
      setTotal(res.total)
    })
  }, [page, search])

  useEffect(() => { load() }, [load])
  useEffect(() => { api.get<Category[]>('/categories').then(setCategories) }, [])

  const openCreate = () => {
    setEditing(null)
    setForm({ name: '', sku: '', category_id: '', unit: 'шт', min_stock: '0', purchase_price: '' })
    setDialogOpen(true)
  }

  const openEdit = (p: Product) => {
    setEditing(p)
    setForm({
      name: p.name,
      sku: p.sku,
      category_id: p.category_id?.toString() || '',
      unit: p.unit,
      min_stock: p.min_stock.toString(),
      purchase_price: p.purchase_price?.toString() || '',
    })
    setDialogOpen(true)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    const body = {
      name: form.name,
      sku: form.sku,
      category_id: form.category_id ? Number(form.category_id) : null,
      unit: form.unit,
      min_stock: Number(form.min_stock),
      purchase_price: form.purchase_price ? Number(form.purchase_price) : null,
    }

    if (editing) {
      await api.put(`/products/${editing.id}`, body)
    } else {
      await api.post('/products', body)
    }
    setDialogOpen(false)
    load()
  }

  const handleDelete = async (id: number) => {
    if (!confirm('Удалить товар?')) return
    await api.delete(`/products/${id}`)
    load()
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Номенклатура</h1>
        <Button onClick={openCreate}><Plus className="h-4 w-4 mr-2" />Добавить товар</Button>
      </div>

      <div className="flex items-center gap-2 mb-4">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Поиск по названию или артикулу..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(1) }}
            className="pl-9"
          />
        </div>
      </div>

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Артикул</TableHead>
              <TableHead>Название</TableHead>
              <TableHead>Категория</TableHead>
              <TableHead className="text-right">Остаток</TableHead>
              <TableHead className="text-right">Мин. остаток</TableHead>
              <TableHead className="text-right">Цена</TableHead>
              <TableHead></TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {products.map((p) => (
              <TableRow key={p.id}>
                <TableCell className="font-mono text-sm">{p.sku}</TableCell>
                <TableCell className="font-medium">{p.name}</TableCell>
                <TableCell>{p.category_name || '—'}</TableCell>
                <TableCell className="text-right">
                  {p.min_stock > 0 && p.quantity <= p.min_stock ? (
                    <Badge variant="destructive">{p.quantity} {p.unit}</Badge>
                  ) : (
                    <span>{p.quantity} {p.unit}</span>
                  )}
                </TableCell>
                <TableCell className="text-right">{p.min_stock}</TableCell>
                <TableCell className="text-right">{p.purchase_price ? `${p.purchase_price} ₽` : '—'}</TableCell>
                <TableCell className="text-right">
                  <Button variant="ghost" size="icon" onClick={() => openEdit(p)}>
                    <Pencil className="h-4 w-4" />
                  </Button>
                  <Button variant="ghost" size="icon" onClick={() => handleDelete(p.id)}>
                    <Trash2 className="h-4 w-4 text-destructive" />
                  </Button>
                </TableCell>
              </TableRow>
            ))}
            {products.length === 0 && (
              <TableRow><TableCell colSpan={7} className="text-center text-muted-foreground py-8">Нет товаров</TableCell></TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      {total > 20 && (
        <div className="flex items-center justify-between mt-4">
          <p className="text-sm text-muted-foreground">Всего: {total}</p>
          <div className="flex gap-2">
            <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage(page - 1)}>Назад</Button>
            <Button variant="outline" size="sm" disabled={page * 20 >= total} onClick={() => setPage(page + 1)}>Далее</Button>
          </div>
        </div>
      )}

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editing ? 'Редактировать товар' : 'Новый товар'}</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <label className="text-sm font-medium">Название *</label>
                <Input value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} required />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">Артикул *</label>
                <Input value={form.sku} onChange={(e) => setForm({ ...form, sku: e.target.value })} required />
              </div>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <label className="text-sm font-medium">Категория</label>
                <select
                  className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                  value={form.category_id}
                  onChange={(e) => setForm({ ...form, category_id: e.target.value })}
                >
                  <option value="">Без категории</option>
                  {categories.map((c) => (
                    <option key={c.id} value={c.id}>{c.name}</option>
                  ))}
                </select>
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">Ед. измерения</label>
                <Input value={form.unit} onChange={(e) => setForm({ ...form, unit: e.target.value })} />
              </div>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <label className="text-sm font-medium">Мин. остаток</label>
                <Input type="number" value={form.min_stock} onChange={(e) => setForm({ ...form, min_stock: e.target.value })} />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">Цена закупки</label>
                <Input type="number" step="0.01" value={form.purchase_price} onChange={(e) => setForm({ ...form, purchase_price: e.target.value })} />
              </div>
            </div>
            <div className="flex justify-end gap-2">
              <Button type="button" variant="outline" onClick={() => setDialogOpen(false)}>Отмена</Button>
              <Button type="submit">{editing ? 'Сохранить' : 'Создать'}</Button>
            </div>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  )
}
