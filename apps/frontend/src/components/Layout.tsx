import { NavLink, Outlet } from 'react-router-dom'
import { getUsername, hasRole, logout } from '../auth/AuthGuard'

const navItems = [
  { path: '/', label: 'Dashboard', icon: '📊' },
  { path: '/boq', label: 'BOQ', icon: '📋' },
  { path: '/tenders', label: 'Tenders', icon: '📢' },
  { path: '/contracts', label: 'Contracts', icon: '📝' },
  { path: '/hr', label: 'HR', icon: '👥' },
  { path: '/finance', label: 'Finance', icon: '💰' },
  { path: '/procurement', label: 'Procurement', icon: '📦' },
  { path: '/bim', label: 'BIM', icon: '🏗️' },
  { path: '/ai', label: 'AI', icon: '🤖' },
  { path: '/pm', label: 'PM', icon: '📁' },
  { path: '/doc-control', label: 'Doc Control', icon: '📄' },
  { path: '/schedule', label: 'Schedule', icon: '📅' },
  { path: '/equipment', label: 'Equipment', icon: '🏗️' },
  { path: '/hse', label: 'HSE', icon: '🛡️' },
  { path: '/quality', label: 'Quality', icon: '✅' },
  { path: '/gis', label: 'GIS', icon: '🗺️' },
  { path: '/risk', label: 'Risk', icon: '⚠️' },
  { path: '/change', label: 'Change', icon: '🔄' },
  { path: '/tbm', label: 'TBM', icon: '🛠️' },
  { path: '/ringbuilder', label: 'Ring Builder', icon: '🔘' },
  { path: '/natm', label: 'NATM', icon: '⛰️' },
  { path: '/funding', label: 'Funding', icon: '💰' },
  { path: '/knowledge-graph', label: 'Knowledge Graph', icon: '🕸️' },
  { path: '/lab', label: 'Laboratory', icon: '🔬' },
  { path: '/permits', label: 'Permits', icon: '📋' },
  { path: '/insurance', label: 'Insurance', icon: '🛡️' },
  { path: '/fleet', label: 'Fleet', icon: '🚛' },
]

export default function Layout() {
  const username = getUsername()
  const isAdmin = hasRole('admin')

  return (
    <div className="flex h-screen">
      {/* Sidebar */}
      <aside className="w-64 bg-[#1e293b] border-r border-[#334155] flex flex-col shrink-0">
        <div className="p-5 border-b border-[#334155]">
          <h1 className="text-lg font-bold text-white">🏗️ OCE</h1>
          <p className="text-xs text-[#94a3b8] mt-1">OpenConstructionERP</p>
        </div>
        <nav className="flex-1 overflow-y-auto p-3 space-y-1">
          {navItems.map((item) => (
            <NavLink
              key={item.path}
              to={item.path}
              end={item.path === '/'}
              className={({ isActive }) =>
                `flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm transition-colors ${
                  isActive
                    ? 'bg-[#3b82f6]/20 text-[#3b82f6] font-medium'
                    : 'text-[#94a3b8] hover:bg-[#334155] hover:text-white'
                }`
              }
            >
              <span className="text-lg">{item.icon}</span>
              <span>{item.label}</span>
            </NavLink>
          ))}
        </nav>
        <div className="p-4 border-t border-[#334155] text-xs text-[#64748b]">
          v0.1.0
        </div>
      </aside>

      {/* Main content */}
      <main className="flex-1 overflow-y-auto bg-[#0f172a]">
        {/* Top bar with user info */}
        <header className="flex items-center justify-end gap-4 px-6 py-3 bg-[#1e293b] border-b border-[#334155]">
          <div className="flex items-center gap-3">
            <span className="text-sm text-[#94a3b8]">
              {username}
              {isAdmin && <span className="ml-2 px-2 py-0.5 text-xs bg-[#3b82f6]/20 text-[#3b82f6] rounded-full">Admin</span>}
            </span>
            <button
              onClick={logout}
              className="px-3 py-1.5 text-xs font-medium text-white bg-[#ef4444]/80 hover:bg-[#ef4444] rounded-lg transition-colors"
            >
              Logout
            </button>
          </div>
        </header>
        <div className="p-6">
          <Outlet />
        </div>
      </main>
    </div>
  )
}
