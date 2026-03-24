import { useState } from 'react'
import { api } from '@/api/client'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { formatCurrency } from '@/lib/utils'
import { FileDown } from 'lucide-react'

interface StockRow {
  id: number; name: string; sku: string; category: string; unit: string
  min_stock: number; quantity: number; price: number; value: number
}

interface TurnoverRow {
  id: number; name: string; sku: string; unit: string
  total_in: number; total_out: number; corrections_plus: number; corrections_minus: number
}

export default function Reports() {
  const [tab, setTab] = useState<'stock' | 'turnover'>('stock')
  const [stockData, setStockData] = useState<StockRow[]>([])
  const [turnoverData, setTurnoverData] = useState<TurnoverRow[]>([])
  const [from, setFrom] = useState('')
  const [to, setTo] = useState('')
  const [loaded, setLoaded] = useState(false)

  const loadStock = async () => {
    const data = await api.get<StockRow[]>('/reports/stock')
    setStockData(data)
    setLoaded(true)
  }

  const loadTurnover = async () => {
    if (!from || !to) { alert('Укажите период'); return }
    const data = await api.get<TurnoverRow[]>(`/reports/turnover?from=${from}&to=${to}`)
    setTurnoverData(data)
    setLoaded(true)
  }

  const exportExcel = () => {
    // Direct download via API
    window.open(`/api/reports/export/excel?from=${from}&to=${to}`, '_blank')
  }

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Отчёты</h1>

      <div className="flex gap-2 mb-6">
        <Button variant={tab === 'stock' ? 'default' : 'outline'} onClick={() => { setTab('stock'); setLoaded(false) }}>
          Остатки
        </Button>
        <Button variant={tab === 'turnover' ? 'default' : 'outline'} onClick={() => { setTab('turnover'); setLoaded(false) }}>
          Оборотная ведомость
        </Button>
      </div>

      {tab === 'stock' && (
        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle className="text-lg">Текущие остатки</CardTitle>
            <Button onClick={loadStock}>Сформировать</Button>
          </CardHeader>
          <CardContent>
            {loaded && (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Артикул</TableHead>
                    <TableHead>Товар</TableHead>
                    <TableHead>Категория</TableHead>
                    <TableHead className="text-right">Остаток</TableHead>
                    <TableHead className="text-right">Мин.</TableHead>
                    <TableHead className="text-right">Цена</TableHead>
                    <TableHead className="text-right">Стоимость</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {stockData.map((r) => (
                    <TableRow key={r.id}>
                      <TableCell className="font-mono">{r.sku}</TableCell>
                      <TableCell>{r.name}</TableCell>
                      <TableCell>{r.category || '—'}</TableCell>
                      <TableCell className="text-right">
                        {r.min_stock > 0 && r.quantity <= r.min_stock
                          ? <Badge variant="destructive">{r.quantity} {r.unit}</Badge>
                          : <span>{r.quantity} {r.unit}</span>}
                      </TableCell>
                      <TableCell className="text-right">{r.min_stock}</TableCell>
                      <TableCell className="text-right">{r.price ? formatCurrency(r.price) : '—'}</TableCell>
                      <TableCell className="text-right font-medium">{r.value ? formatCurrency(r.value) : '—'}</TableCell>
                    </TableRow>
                  ))}
                  {stockData.length > 0 && (
                    <TableRow className="font-bold">
                      <TableCell colSpan={6} className="text-right">Итого:</TableCell>
                      <TableCell className="text-right">{formatCurrency(stockData.reduce((s, r) => s + r.value, 0))}</TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            )}
          </CardContent>
        </Card>
      )}

      {tab === 'turnover' && (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Оборотная ведомость</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-end gap-4 mb-4">
              <div className="space-y-2">
                <label className="text-sm font-medium">С</label>
                <Input type="date" value={from} onChange={(e) => setFrom(e.target.value)} />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">По</label>
                <Input type="date" value={to} onChange={(e) => setTo(e.target.value)} />
              </div>
              <Button onClick={loadTurnover}>Сформировать</Button>
              {loaded && (
                <Button variant="outline" onClick={exportExcel}>
                  <FileDown className="h-4 w-4 mr-2" />Excel
                </Button>
              )}
            </div>

            {loaded && (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Артикул</TableHead>
                    <TableHead>Товар</TableHead>
                    <TableHead className="text-right">Приход</TableHead>
                    <TableHead className="text-right">Расход</TableHead>
                    <TableHead className="text-right">Корр. +</TableHead>
                    <TableHead className="text-right">Корр. -</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {turnoverData.map((r) => (
                    <TableRow key={r.id}>
                      <TableCell className="font-mono">{r.sku}</TableCell>
                      <TableCell>{r.name}</TableCell>
                      <TableCell className="text-right text-green-600">{r.total_in || '—'}</TableCell>
                      <TableCell className="text-right text-red-600">{r.total_out || '—'}</TableCell>
                      <TableCell className="text-right">{r.corrections_plus || '—'}</TableCell>
                      <TableCell className="text-right">{r.corrections_minus || '—'}</TableCell>
                    </TableRow>
                  ))}
                  {turnoverData.length === 0 && loaded && (
                    <TableRow><TableCell colSpan={6} className="text-center text-muted-foreground py-8">Нет данных за период</TableCell></TableRow>
                  )}
                </TableBody>
              </Table>
            )}
          </CardContent>
        </Card>
      )}
    </div>
  )
}
