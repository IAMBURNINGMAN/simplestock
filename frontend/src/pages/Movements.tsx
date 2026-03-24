import { useEffect, useState } from 'react'
import { api } from '@/api/client'
import type { Movement, PaginatedResponse } from '@/api/types'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table'
import { formatDate } from '@/lib/utils'

const typeLabels: Record<string, { label: string; variant: 'default' | 'success' | 'destructive' | 'warning' | 'secondary' }> = {
  incoming: { label: 'Приход', variant: 'success' },
  outgoing: { label: 'Расход', variant: 'destructive' },
  correction_plus: { label: 'Корр. +', variant: 'success' },
  correction_minus: { label: 'Корр. -', variant: 'warning' },
}

export default function Movements() {
  const [movements, setMovements] = useState<Movement[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [from, setFrom] = useState('')
  const [to, setTo] = useState('')
  const [filterType, setFilterType] = useState('')

  const load = () => {
    const params = new URLSearchParams({ page: String(page), page_size: '20' })
    if (from) params.set('from', from)
    if (to) params.set('to', to)
    if (filterType) params.set('type', filterType)

    api.get<PaginatedResponse<Movement>>(`/movements?${params}`).then((res) => {
      setMovements(res.data)
      setTotal(res.total)
    })
  }

  useEffect(() => { load() }, [page, from, to, filterType])

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">История движений</h1>

      <div className="flex items-end gap-4 mb-4">
        <div className="space-y-2">
          <label className="text-sm font-medium">С</label>
          <Input type="date" value={from} onChange={(e) => { setFrom(e.target.value); setPage(1) }} />
        </div>
        <div className="space-y-2">
          <label className="text-sm font-medium">По</label>
          <Input type="date" value={to} onChange={(e) => { setTo(e.target.value); setPage(1) }} />
        </div>
        <div className="space-y-2">
          <label className="text-sm font-medium">Тип</label>
          <select
            className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
            value={filterType}
            onChange={(e) => { setFilterType(e.target.value); setPage(1) }}
          >
            <option value="">Все</option>
            <option value="incoming">Приход</option>
            <option value="outgoing">Расход</option>
            <option value="correction_plus">Корректировка +</option>
            <option value="correction_minus">Корректировка -</option>
          </select>
        </div>
      </div>

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Дата</TableHead>
              <TableHead>Тип</TableHead>
              <TableHead>Артикул</TableHead>
              <TableHead>Товар</TableHead>
              <TableHead className="text-right">Количество</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {movements.map((m) => {
              const t = typeLabels[m.movement_type] || { label: m.movement_type, variant: 'secondary' as const }
              return (
                <TableRow key={m.id}>
                  <TableCell>{formatDate(m.created_at)}</TableCell>
                  <TableCell><Badge variant={t.variant}>{t.label}</Badge></TableCell>
                  <TableCell className="font-mono">{m.product_sku}</TableCell>
                  <TableCell>{m.product_name}</TableCell>
                  <TableCell className="text-right font-medium">{m.quantity}</TableCell>
                </TableRow>
              )
            })}
            {movements.length === 0 && (
              <TableRow><TableCell colSpan={5} className="text-center text-muted-foreground py-8">Нет движений</TableCell></TableRow>
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
    </div>
  )
}
