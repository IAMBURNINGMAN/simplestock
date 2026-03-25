import { Routes, Route } from 'react-router-dom'
import { Layout } from '@/components/Layout'
import Login from '@/pages/Login'
import Dashboard from '@/pages/Dashboard'
import Products from '@/pages/Products'
import DocumentList from '@/pages/DocumentList'
import DocumentForm from '@/pages/DocumentForm'
import DocumentDetail from '@/pages/DocumentDetail'
import InventoryPage from '@/pages/InventoryPage'
import Movements from '@/pages/Movements'

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route element={<Layout />}>
        <Route path="/" element={<Dashboard />} />
        <Route path="/products" element={<Products />} />
        <Route path="/incoming" element={<DocumentList docType="incoming" title="Приход товаров" createPath="/incoming/new" />} />
        <Route path="/incoming/new" element={<DocumentForm docType="incoming" title="Новая приходная накладная" />} />
        <Route path="/incoming/:id" element={<DocumentDetail />} />
        <Route path="/outgoing" element={<DocumentList docType="outgoing" title="Расход товаров" createPath="/outgoing/new" />} />
        <Route path="/outgoing/new" element={<DocumentForm docType="outgoing" title="Новая расходная накладная" />} />
        <Route path="/outgoing/:id" element={<DocumentDetail />} />
        <Route path="/inventory" element={<InventoryPage />} />
        <Route path="/movements" element={<Movements />} />
      </Route>
    </Routes>
  )
}
