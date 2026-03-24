import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '@/api/client'
import type { Document, PaginatedResponse } from '@/api/types'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui/table'
import { Plus, Eye, CheckCircle, Trash2 } from 'lucide-react'
import { formatDate } from '@/lib/utils'

interface Props {
  docType: 'incoming' | 'outgoing'
  title: string
  createPath: string
}

const expenseTypeLabels: Record<string, string> = {
  sale: 'Продажа',
  write_off: 'Списание',
  transfer: 'Перемещение',
}

export default function DocumentList({ docType, title, createPath }: Props) {
  const [docs, setDocs] = useState<Document[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)

  const load = () => {
    api.get<PaginatedResponse<Document>>(`/documents?doc_type=${docType}&page=${page}&page_size=20`).then((res) => {
      setDocs(res.data)
      setTotal(res.total)
    })
  }

  useEffect(() => { load() }, [page, docType])

  const postDoc = async (id: number) => {
    try {
      await api.post(`/documents/${id}/post`)
      load()
    } catch (err: any) {
      alert(err.message)
    }
  }

  const deleteDoc = async (id: number) => {
    if (!confirm('Удалить черновик?')) return
    try {
      await api.delete(`/documents/${id}`)
      load()
    } catch (err: any) {
      alert(err.message)
    }
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">{title}</h1>
        <Link to={createPath}>
          <Button><Plus className="h-4 w-4 mr-2" />Создать</Button>
        </Link>
      </div>

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Номер</TableHead>
              <TableHead>Дата</TableHead>
              <TableHead>Контрагент</TableHead>
              {docType === 'outgoing' && <TableHead>Тип</TableHead>}
              <TableHead>Статус</TableHead>
              <TableHead></TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {docs.map((d) => (
              <TableRow key={d.id}>
                <TableCell className="font-mono font-medium">{d.doc_number}</TableCell>
                <TableCell>{formatDate(d.doc_date)}</TableCell>
                <TableCell>{d.counterparty || '—'}</TableCell>
                {docType === 'outgoing' && (
                  <TableCell>{d.expense_type ? expenseTypeLabels[d.expense_type] || d.expense_type : '—'}</TableCell>
                )}
                <TableCell>
                  <Badge variant={d.status === 'posted' ? 'success' : 'secondary'}>
                    {d.status === 'posted' ? 'Проведён' : 'Черновик'}
                  </Badge>
                </TableCell>
                <TableCell className="text-right space-x-1">
                  <Link to={`/${docType === 'incoming' ? 'incoming' : 'outgoing'}/${d.id}`}>
                    <Button variant="ghost" size="icon"><Eye className="h-4 w-4" /></Button>
                  </Link>
                  {d.status === 'draft' && (
                    <>
                      <Button variant="ghost" size="icon" onClick={() => postDoc(d.id)} title="Провести">
                        <CheckCircle className="h-4 w-4 text-green-600" />
                      </Button>
                      <Button variant="ghost" size="icon" onClick={() => deleteDoc(d.id)}>
                        <Trash2 className="h-4 w-4 text-destructive" />
                      </Button>
                    </>
                  )}
                </TableCell>
              </TableRow>
            ))}
            {docs.length === 0 && (
              <TableRow><TableCell colSpan={docType === 'outgoing' ? 6 : 5} className="text-center text-muted-foreground py-8">Нет документов</TableCell></TableRow>
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
