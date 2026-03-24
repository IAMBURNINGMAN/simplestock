import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '@/api/client'
import type { Product, PaginatedResponse } from '@/api/types'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table'
import { Plus, Trash2 } from 'lucide-react'

interface Props {
  docType: 'incoming' | 'outgoing'
  title: string
}

interface LineItem {
  product_id: number
  product_name: string
  quantity: number
  price: number | null
  max_stock?: number
}

export default function DocumentForm({ docType, title }: Props) {
  const navigate = useNavigate()
  const [products, setProducts] = useState<Product[]>([])
  const [counterparty, setCounterparty] = useState('')
  const [expenseType, setExpenseType] = useState('sale')
  const [docDate, setDocDate] = useState(new Date().toISOString().split('T')[0])
  const [items, setItems] = useState<LineItem[]>([])
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    api.get<PaginatedResponse<Product>>('/products?page_size=1000').then((res) => setProducts(res.data))
  }, [])

  const addItem = () => {
    setItems([...items, { product_id: 0, product_name: '', quantity: 1, price: null }])
  }

  const updateItem = (idx: number, field: string, value: any) => {
    const updated = [...items]
    if (field === 'product_id') {
      const p = products.find((p) => p.id === Number(value))
      if (p) {
        updated[idx] = { ...updated[idx], product_id: p.id, product_name: p.name, price: p.purchase_price, max_stock: p.quantity }
      }
    } else {
      (updated[idx] as any)[field] = value
    }
    setItems(updated)
  }

  const removeItem = (idx: number) => {
    setItems(items.filter((_, i) => i !== idx))
  }

  const handleSubmit = async (postImmediately: boolean) => {
    setError('')
    if (items.length === 0) {
      setError('Добавьте хотя бы одну позицию')
      return
    }

    setSubmitting(true)
    try {
      const body = {
        doc_type: docType,
        counterparty: counterparty || null,
        expense_type: docType === 'outgoing' ? expenseType : null,
        doc_date: docDate,
        items: items.map((it) => ({
          product_id: it.product_id,
          quantity: Number(it.quantity),
          price: it.price ? Number(it.price) : null,
        })),
      }

      const doc: any = await api.post('/documents', body)

      if (postImmediately) {
        await api.post(`/documents/${doc.id}/post`)
      }

      navigate(docType === 'incoming' ? '/incoming' : '/outgoing')
    } catch (err: any) {
      setError(err.message)
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">{title}</h1>

      {error && <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive mb-4">{error}</div>}

      <Card className="mb-6">
        <CardHeader>
          <CardTitle className="text-lg">Данные документа</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Дата</label>
              <Input type="date" value={docDate} onChange={(e) => setDocDate(e.target.value)} />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">{docType === 'incoming' ? 'Поставщик' : 'Получатель'}</label>
              <Input value={counterparty} onChange={(e) => setCounterparty(e.target.value)} placeholder="Контрагент" />
            </div>
            {docType === 'outgoing' && (
              <div className="space-y-2">
                <label className="text-sm font-medium">Тип расхода</label>
                <select
                  className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                  value={expenseType}
                  onChange={(e) => setExpenseType(e.target.value)}
                >
                  <option value="sale">Продажа</option>
                  <option value="write_off">Списание</option>
                  <option value="transfer">Перемещение</option>
                </select>
              </div>
            )}
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="text-lg">Позиции</CardTitle>
          <Button size="sm" onClick={addItem}><Plus className="h-4 w-4 mr-1" />Добавить</Button>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Товар</TableHead>
                <TableHead className="w-32">Количество</TableHead>
                <TableHead className="w-32">Цена</TableHead>
                {docType === 'outgoing' && <TableHead className="w-28">На складе</TableHead>}
                <TableHead className="w-12"></TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((item, idx) => (
                <TableRow key={idx}>
                  <TableCell>
                    <select
                      className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                      value={item.product_id}
                      onChange={(e) => updateItem(idx, 'product_id', e.target.value)}
                    >
                      <option value={0}>Выберите товар...</option>
                      {products.map((p) => (
                        <option key={p.id} value={p.id}>{p.sku} — {p.name} ({p.quantity} {p.unit})</option>
                      ))}
                    </select>
                  </TableCell>
                  <TableCell>
                    <Input
                      type="number"
                      min={1}
                      value={item.quantity}
                      onChange={(e) => updateItem(idx, 'quantity', Number(e.target.value))}
                    />
                  </TableCell>
                  <TableCell>
                    <Input
                      type="number"
                      step="0.01"
                      value={item.price || ''}
                      onChange={(e) => updateItem(idx, 'price', e.target.value ? Number(e.target.value) : null)}
                      placeholder="—"
                    />
                  </TableCell>
                  {docType === 'outgoing' && (
                    <TableCell className="text-center text-sm text-muted-foreground">{item.max_stock ?? '—'}</TableCell>
                  )}
                  <TableCell>
                    <Button variant="ghost" size="icon" onClick={() => removeItem(idx)}>
                      <Trash2 className="h-4 w-4 text-destructive" />
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
              {items.length === 0 && (
                <TableRow>
                  <TableCell colSpan={docType === 'outgoing' ? 5 : 4} className="text-center text-muted-foreground py-8">
                    Нажмите "Добавить" для добавления позиций
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>

          <div className="flex justify-end gap-2 mt-6">
            <Button variant="outline" onClick={() => navigate(-1)} disabled={submitting}>Отмена</Button>
            <Button variant="secondary" onClick={() => handleSubmit(false)} disabled={submitting}>Сохранить черновик</Button>
            <Button onClick={() => handleSubmit(true)} disabled={submitting}>Провести</Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
