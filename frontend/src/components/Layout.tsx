import { Link, useLocation, Outlet, Navigate } from 'react-router-dom'
import { useAuth } from '@/hooks/useAuth'
import { cn } from '@/lib/utils'
import {
  LayoutDashboard,
  Package,
  ArrowDownToLine,
  ArrowUpFromLine,
  ClipboardCheck,
  History,
  BarChart3,
  LogOut,
} from 'lucide-react'

const navItems = [
  { path: '/', label: 'Дашборд', icon: LayoutDashboard },
  { path: '/products', label: 'Номенклатура', icon: Package },
  { path: '/incoming', label: 'Приход', icon: ArrowDownToLine },
  { path: '/outgoing', label: 'Расход', icon: ArrowUpFromLine },
  { path: '/inventory', label: 'Инвентаризация', icon: ClipboardCheck },
  { path: '/movements', label: 'История', icon: History },
  { path: '/reports', label: 'Отчёты', icon: BarChart3 },
]

export function Layout() {
  const { user, loading, logout } = useAuth()
  const location = useLocation()

  if (loading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="text-lg text-muted-foreground">Загрузка...</div>
      </div>
    )
  }

  if (!user) {
    return <Navigate to="/login" replace />
  }

  return (
    <div className="flex h-screen">
      {/* Sidebar */}
      <aside className="w-64 border-r bg-card flex flex-col">
        <div className="p-6 border-b">
          <h1 className="text-xl font-bold text-primary">SimpleStock</h1>
          <p className="text-xs text-muted-foreground mt-1">Складской учёт</p>
        </div>

        <nav className="flex-1 p-4 space-y-1">
          {navItems.map((item) => {
            const isActive = location.pathname === item.path
            return (
              <Link
                key={item.path}
                to={item.path}
                className={cn(
                  'flex items-center gap-3 px-3 py-2.5 rounded-md text-sm font-medium transition-colors',
                  isActive
                    ? 'bg-primary text-primary-foreground'
                    : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
                )}
              >
                <item.icon className="h-4 w-4" />
                {item.label}
              </Link>
            )
          })}
        </nav>

        <div className="p-4 border-t">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium">{user.full_name}</p>
              <p className="text-xs text-muted-foreground">{user.role}</p>
            </div>
            <button
              onClick={logout}
              className="p-2 rounded-md hover:bg-accent text-muted-foreground"
              title="Выйти"
            >
              <LogOut className="h-4 w-4" />
            </button>
          </div>
        </div>
      </aside>

      {/* Main content */}
      <main className="flex-1 overflow-auto">
        <div className="p-8">
          <Outlet />
        </div>
      </main>
    </div>
  )
}
