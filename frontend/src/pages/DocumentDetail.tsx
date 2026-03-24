import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { api } from '@/api/client'
import type { Document } from '@/api/types'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table'
import { formatDate } from '@/lib/utils'
import { ArrowLeft, CheckCircle } from 'lucide-react'

export default function DocumentDetail() {
  const { id } = useParams()
  const navigate = useNavigate()
  const [doc, setDoc] = useState<Document | null>(null)

  const load = () => {
    api.get<Document>(`/documents/${id}`).then(setDoc)
  }

  useEffect(() => { load() }, [id])

  if (!doc) return <div className="text-muted-foreground">Загрузка...</div>

  const postDoc = async () => {
    try {
      await api.post(`/documents/${doc.id}/post`)
      load()
    } catch (err: any) {
      alert(err.message)
    }
  }

  return (
    <div>
      <Button variant="ghost" className="mb-4" onClick={() => navigate(-1)}>
        <ArrowLeft className="h-4 w-4 mr-2" />Назад
      </Button>

      <Card className="mb-6">
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>{doc.doc_type === 'incoming' ? 'Приходная накладная' : 'Расходная накладная'} {doc.doc_number}</CardTitle>
            <div className="flex items-center gap-2">
              <Badge variant={doc.status === 'posted' ? 'success' : 'secondary'}>
                {doc.status === 'posted' ? 'Проведён' : 'Черновик'}
              </Badge>
              {doc.status === 'draft' && (
                <Button size="sm" onClick={postDoc}>
                  <CheckCircle className="h-4 w-4 mr-1" />Провести
                </Button>
              )}
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-3 gap-4 text-sm">
            <div><span className="text-muted-foreground">Дата:</span> {formatDate(doc.doc_date)}</div>
            <div><span className="text-muted-foreground">Контрагент:</span> {doc.counterparty || '—'}</div>
            {doc.expense_type && <div><span className="text-muted-foreground">Тип:</span> {doc.expense_type}</div>}
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Позиции</CardTitle>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Артикул</TableHead>
                <TableHead>Товар</TableHead>
                <TableHead className="text-right">Количество</TableHead>
                <TableHead className="text-right">Цена</TableHead>
                <TableHead className="text-right">Сумма</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {doc.items?.map((item) => (
                <TableRow key={item.id}>
                  <TableCell className="font-mono">{item.product_sku}</TableCell>
                  <TableCell>{item.product_name}</TableCell>
                  <TableCell className="text-right">{item.quantity}</TableCell>
                  <TableCell className="text-right">{item.price ? `${item.price} ₽` : '—'}</TableCell>
                  <TableCell className="text-right">{item.price ? `${(item.quantity * item.price).toFixed(2)} ₽` : '—'}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  )
}
