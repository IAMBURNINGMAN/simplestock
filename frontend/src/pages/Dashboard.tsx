import { useEffect, useState } from 'react'
import { api } from '@/api/client'
import type { DashboardSummary, Product } from '@/api/types'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { formatCurrency, formatNumber } from '@/lib/utils'
import { Package, ArrowDownToLine, ArrowUpFromLine, AlertTriangle, TrendingUp, Boxes } from 'lucide-react'

export default function Dashboard() {
  const [summary, setSummary] = useState<DashboardSummary | null>(null)
  const [lowStock, setLowStock] = useState<Product[]>([])

  useEffect(() => {
    api.get<DashboardSummary>('/dashboard/summary').then(setSummary)
    api.get<Product[]>('/products/low-stock').then(setLowStock)
  }, [])

  if (!summary) return <div className="text-muted-foreground">Загрузка...</div>

  const cards = [
    { title: 'Товаров', value: formatNumber(summary.total_products), icon: Package, color: 'text-blue-600' },
    { title: 'Приход сегодня', value: formatNumber(summary.today_incoming), icon: ArrowDownToLine, color: 'text-green-600' },
    { title: 'Расход сегодня', value: formatNumber(summary.today_outgoing), icon: ArrowUpFromLine, color: 'text-orange-600' },
    { title: 'Ниже минимума', value: formatNumber(summary.low_stock_count), icon: AlertTriangle, color: summary.low_stock_count > 0 ? 'text-red-600' : 'text-green-600' },
    { title: 'Движений сегодня', value: formatNumber(summary.today_movements), icon: TrendingUp, color: 'text-purple-600' },
    { title: 'Стоимость склада', value: formatCurrency(summary.total_stock_value), icon: Boxes, color: 'text-indigo-600' },
  ]

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Дашборд</h1>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mb-8">
        {cards.map((card) => (
          <Card key={card.title}>
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-muted-foreground">{card.title}</p>
                  <p className="text-2xl font-bold mt-1">{card.value}</p>
                </div>
                <card.icon className={`h-8 w-8 ${card.color}`} />
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {lowStock.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg flex items-center gap-2">
              <AlertTriangle className="h-5 w-5 text-red-600" />
              Товары ниже минимального остатка
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {lowStock.map((p) => (
                <div key={p.id} className="flex items-center justify-between py-2 border-b last:border-0">
                  <div>
                    <p className="font-medium">{p.name}</p>
                    <p className="text-sm text-muted-foreground">{p.sku}</p>
                  </div>
                  <div className="text-right">
                    <Badge variant={p.quantity === 0 ? 'destructive' : 'warning'}>
                      {p.quantity} / {p.min_stock} {p.unit}
                    </Badge>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
