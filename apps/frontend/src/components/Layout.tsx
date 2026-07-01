import { NavLink, Outlet } from 'react-router-dom'

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
]

export default function Layout() {
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
        <Outlet />
      </main>
    </div>
  )
}
