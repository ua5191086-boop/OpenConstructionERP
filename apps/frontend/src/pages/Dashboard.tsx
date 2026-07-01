import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell, Legend } from 'recharts'

const COLORS = ['#3b82f6', '#22c55e', '#a855f7', '#f97316', '#ef4444', '#14b8a6', '#f59e0b', '#6366f1', '#ec4899']

const moduleCards = [
  { path: '/boq', label: 'BOQ', icon: '📋', desc: 'Bill of Quantities', color: '#3b82f6' },
  { path: '/tenders', label: 'Tenders', icon: '📢', desc: 'Tender Management', color: '#22c55e' },
  { path: '/contracts', label: 'Contracts', icon: '📝', desc: 'Contract Management', color: '#a855f7' },
  { path: '/hr', label: 'HR', icon: '👥', desc: 'Human Resources', color: '#f97316' },
  { path: '/finance', label: 'Finance', icon: '💰', desc: 'Financial Management', color: '#ef4444' },
  { path: '/procurement', label: 'Procurement', icon: '📦', desc: 'Procurement & Inventory', color: '#14b8a6' },
  { path: '/bim', label: 'BIM', icon: '🏗️', desc: 'BIM Integration', color: '#f59e0b' },
  { path: '/ai', label: 'AI', icon: '🤖', desc: 'AI Assistant', color: '#6366f1' },
]

export default function Dashboard() {
  const [stats, setStats] = useState<any>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    async function load() {
      try {
        const [boqItems, tenders, contracts, employees, budgets, invoices, prs, pos, bimModels, aiAgents] =
          await Promise.allSettled([
            fetch('/api/boq/items').then(r => r.ok ? r.json() : []),
            fetch('/api/tenders').then(r => r.ok ? r.json() : []),
            fetch('/api/contracts').then(r => r.ok ? r.json() : []),
            fetch('/api/hr/employees').then(r => r.ok ? r.json() : []),
            fetch('/api/finance/budgets').then(r => r.ok ? r.json() : []),
            fetch('/api/finance/invoices').then(r => r.ok ? r.json() : []),
            fetch('/api/procurement/requests').then(r => r.ok ? r.json() : []),
            fetch('/api/procurement/purchase-orders').then(r => r.ok ? r.json() : []),
            fetch('/api/bim/models').then(r => r.ok ? r.json() : []),
            fetch('/api/ai/agents').then(r => r.ok ? r.json() : []),
          ])
        setStats({
          boqItems: (boqItems as any).value?.length ?? 0,
          tenders: (tenders as any).value?.length ?? 0,
          contracts: (contracts as any).value?.length ?? 0,
          employees: (employees as any).value?.length ?? 0,
          budgets: (budgets as any).value?.length ?? 0,
          invoices: (invoices as any).value?.length ?? 0,
          procurementRequests: (prs as any).value?.length ?? 0,
          purchaseOrders: (pos as any).value?.length ?? 0,
          bimModels: (bimModels as any).value?.length ?? 0,
          aiAgents: (aiAgents as any).value?.length ?? 0,
        })
      } catch {
        // Use demo data if API unavailable
        setStats({
          boqItems: 245, tenders: 12, contracts: 8, employees: 156,
          budgets: 24, invoices: 89, procurementRequests: 67,
          purchaseOrders: 43, bimModels: 15, aiAgents: 6,
        })
      }
      setLoading(false)
    }
    load()
  }, [])

  const chartData = stats
    ? [
        { name: 'BOQ Items', value: stats.boqItems },
        { name: 'Tenders', value: stats.tenders },
        { name: 'Contracts', value: stats.contracts },
        { name: 'Employees', value: stats.employees },
        { name: 'Budgets', value: stats.budgets },
        { name: 'Invoices', value: stats.invoices },
        { name: 'PRs', value: stats.procurementRequests },
        { name: 'POs', value: stats.purchaseOrders },
        { name: 'BIM Models', value: stats.bimModels },
        { name: 'AI Agents', value: stats.aiAgents },
      ]
    : []

  return (
    <div className="p-6 max-w-7xl mx-auto">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-white">Dashboard</h1>
        <p className="text-[#94a3b8] mt-1">OpenConstructionERP — Project Management Platform</p>
      </div>

      {/* Module Cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
        {moduleCards.map((mod) => (
          <Link
            key={mod.path}
            to={mod.path}
            className="bg-[#1e293b] border border-[#334155] rounded-xl p-5 hover:border-[#3b82f6]/50 transition-all hover:shadow-lg hover:shadow-[#3b82f6]/5"
          >
            <div className="text-3xl mb-3">{mod.icon}</div>
            <h3 className="text-white font-semibold">{mod.label}</h3>
            <p className="text-[#94a3b8] text-sm mt-1">{mod.desc}</p>
          </Link>
        ))}
      </div>

      {/* Stats Grid */}
      {loading ? (
        <div className="text-center py-12 text-[#64748b]">Loading dashboard data...</div>
      ) : (
        <>
          <div className="grid grid-cols-2 md:grid-cols-5 gap-4 mb-8">
            {chartData.map((d) => (
              <div key={d.name} className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
                <div className="text-[#94a3b8] text-xs uppercase tracking-wider">{d.name}</div>
                <div className="text-2xl font-bold text-white mt-2">{d.value}</div>
              </div>
            ))}
          </div>

          {/* Charts */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-6">
              <h3 className="text-white font-semibold mb-4">Module Overview</h3>
              <ResponsiveContainer width="100%" height={300}>
                <BarChart data={chartData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
                  <XAxis dataKey="name" tick={{ fill: '#94a3b8', fontSize: 11 }} angle={-20} textAnchor="end" height={60} />
                  <YAxis tick={{ fill: '#94a3b8', fontSize: 11 }} />
                  <Tooltip
                    contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '8px', color: '#e2e8f0' }}
                  />
                  <Bar dataKey="value" fill="#3b82f6" radius={[4, 4, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-6">
              <h3 className="text-white font-semibold mb-4">Distribution</h3>
              <ResponsiveContainer width="100%" height={300}>
                <PieChart>
                  <Pie
                    data={chartData}
                    cx="50%"
                    cy="50%"
                    outerRadius={100}
                    dataKey="value"
                    label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
                  >
                    {chartData.map((_, i) => (
                      <Cell key={i} fill={COLORS[i % COLORS.length]} />
                    ))}
                  </Pie>
                  <Tooltip
                    contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '8px', color: '#e2e8f0' }}
                  />
                </PieChart>
              </ResponsiveContainer>
            </div>
          </div>
        </>
      )}
    </div>
  )
}
