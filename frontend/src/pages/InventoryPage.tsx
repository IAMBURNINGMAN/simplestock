import { useEffect, useState } from 'react'
import { api } from '@/api/client'
import type { Inventory, Product, PaginatedResponse } from '@/api/types'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table'
import { Plus, CheckCircle, Eye } from 'lucide-react'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { formatDate } from '@/lib/utils'

export default function InventoryPage() {
  const [inventories, setInventories] = useState<Inventory[]>([])
  const [active, setActive] = useState<Inventory | null>(null)
  const [products, setProducts] = useState<Product[]>([])
  const [addProduct, setAddProduct] = useState('')
  const [addQty, setAddQty] = useState('')
  const [detailOpen, setDetailOpen] = useState(false)
  const [detailInv, setDetailInv] = useState<Inventory | null>(null)

  const load = () => {
    api.get<Inventory[]>('/inventories').then((data) => {
      setInventories(data)
      const act = data.find((i) => i.status === 'active')
      if (act) {
        api.get<Inventory>(`/inventories/${act.id}`).then(setActive)
      } else {
        setActive(null)
      }
    })
  }

  useEffect(() => {
    load()
    api.get<PaginatedResponse<Product>>('/products?page_size=1000').then((r) => setProducts(r.data))
  }, [])

  const startInventory = async () => {
    try {
      await api.post('/inventories')
      load()
    } catch (err: any) {
      alert(err.message)
    }
  }

  const addItem = async () => {
    if (!active || !addProduct || addQty === '') return
    try {
      await api.post(`/inventories/${active.id}/items`, {
        product_id: Number(addProduct),
        actual_quantity: Number(addQty),
      })
      setAddProduct('')
      setAddQty('')
      api.get<Inventory>(`/inventories/${active.id}`).then(setActive)
    } catch (err: any) {
      alert(err.message)
    }
  }

  const complete = async () => {
    if (!active) return
    if (!confirm('Завершить инвентаризацию? Будут созданы корректировки остатков.')) return
    try {
      await api.post(`/inventories/${active.id}/complete`)
      load()
    } catch (err: any) {
      alert(err.message)
    }
  }

  const showDetail = async (id: number) => {
    const inv = await api.get<Inventory>(`/inventories/${id}`)
    setDetailInv(inv)
    setDetailOpen(true)
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Инвентаризация</h1>
        {!active && <Button onClick={startInventory}><Plus className="h-4 w-4 mr-2" />Начать инвентаризацию</Button>}
      </div>

      {active && (
        <Card className="mb-6 border-primary">
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle className="text-lg">
                Активная инвентаризация: {active.inv_number}
              </CardTitle>
              <Button onClick={complete} variant="default">
                <CheckCircle className="h-4 w-4 mr-2" />Завершить
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            <div className="flex items-end gap-4 mb-4">
              <div className="flex-1 space-y-2">
                <label className="text-sm font-medium">Товар</label>
                <select
                  className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                  value={addProduct}
                  onChange={(e) => setAddProduct(e.target.value)}
                >
                  <option value="">Выберите товар...</option>
                  {products.map((p) => (
                    <option key={p.id} value={p.id}>{p.sku} — {p.name}</option>
                  ))}
                </select>
              </div>
              <div className="w-40 space-y-2">
                <label className="text-sm font-medium">Фактическое кол-во</label>
                <Input type="number" min={0} value={addQty} onChange={(e) => setAddQty(e.target.value)} />
              </div>
              <Button onClick={addItem} disabled={!addProduct || addQty === ''}>
                <Plus className="h-4 w-4 mr-1" />Добавить
              </Button>
            </div>

            {active.items && active.items.length > 0 && (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Артикул</TableHead>
                    <TableHead>Товар</TableHead>
                    <TableHead className="text-right">Учётное</TableHead>
                    <TableHead className="text-right">Фактическое</TableHead>
                    <TableHead className="text-right">Расхождение</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {active.items.map((item) => (
                    <TableRow key={item.id}>
                      <TableCell className="font-mono">{item.product_sku}</TableCell>
                      <TableCell>{item.product_name}</TableCell>
                      <TableCell className="text-right">{item.expected_quantity}</TableCell>
                      <TableCell className="text-right">{item.actual_quantity}</TableCell>
                      <TableCell className="text-right">
                        {item.difference !== 0 ? (
                          <Badge variant={item.difference > 0 ? 'success' : 'destructive'}>
                            {item.difference > 0 ? '+' : ''}{item.difference}
                          </Badge>
                        ) : (
                          <Badge variant="secondary">0</Badge>
                        )}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </CardContent>
        </Card>
      )}

      <Card>
        <CardHeader>
          <CardTitle className="text-lg">История инвентаризаций</CardTitle>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Номер</TableHead>
                <TableHead>Начало</TableHead>
                <TableHead>Завершение</TableHead>
                <TableHead>Статус</TableHead>
                <TableHead></TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {inventories.filter(i => i.status === 'completed').map((inv) => (
                <TableRow key={inv.id}>
                  <TableCell className="font-mono">{inv.inv_number}</TableCell>
                  <TableCell>{formatDate(inv.started_at)}</TableCell>
                  <TableCell>{inv.completed_at ? formatDate(inv.completed_at) : '—'}</TableCell>
                  <TableCell><Badge variant="success">Завершена</Badge></TableCell>
                  <TableCell>
                    <Button variant="ghost" size="icon" onClick={() => showDetail(inv.id)}>
                      <Eye className="h-4 w-4" />
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
              {inventories.filter(i => i.status === 'completed').length === 0 && (
                <TableRow><TableCell colSpan={5} className="text-center text-muted-foreground py-8">Нет завершённых инвентаризаций</TableCell></TableRow>
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <Dialog open={detailOpen} onOpenChange={setDetailOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Инвентаризация {detailInv?.inv_number}</DialogTitle>
          </DialogHeader>
          {detailInv?.items && (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Товар</TableHead>
                  <TableHead className="text-right">Учётное</TableHead>
                  <TableHead className="text-right">Фактическое</TableHead>
                  <TableHead className="text-right">Расхождение</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {detailInv.items.map((item) => (
                  <TableRow key={item.id}>
                    <TableCell>{item.product_name}</TableCell>
                    <TableCell className="text-right">{item.expected_quantity}</TableCell>
                    <TableCell className="text-right">{item.actual_quantity}</TableCell>
                    <TableCell className="text-right">
                      <Badge variant={item.difference === 0 ? 'secondary' : item.difference > 0 ? 'success' : 'destructive'}>
                        {item.difference > 0 ? '+' : ''}{item.difference}
                      </Badge>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </DialogContent>
      </Dialog>
    </div>
  )
}
