import { useState, useEffect } from 'react'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell, ScatterChart, Scatter } from 'recharts'

const QC_COLORS: Record<string, string> = { pass: '#22c55e', pending: '#3b82f6', conditional_pass: '#f59e0b', fail: '#ef4444', rework: '#ec4899' }
const STATUS_COLORS: Record<string, string> = { cast: '#f59e0b', curing: '#3b82f6', demolded: '#a855f7', transport: '#f97316', in_stock: '#14b8a6', installed: '#22c55e', rejected: '#ef4444', quarantine: '#ec4899' }

export default function RingBuilderPage() {
  const [data, setData] = useState<any>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    async function load() {
      try {
        const [designs, production, qc, inventory, measurements, summary] = await Promise.all([
          fetch('/api/v1/ringbuilder/designs').then(r => r.ok ? r.json() : []),
          fetch('/api/v1/ringbuilder/production').then(r => r.ok ? r.json() : []),
          fetch('/api/v1/ringbuilder/qc').then(r => r.ok ? r.json() : []),
          fetch('/api/v1/ringbuilder/inventory').then(r => r.ok ? r.json() : []),
          fetch('/api/v1/ringbuilder/measurements').then(r => r.ok ? r.json() : []),
          fetch('/api/v1/ringbuilder/summary').then(r => r.ok ? r.json() : {}),
        ])
        setData({ designs, production, qc, inventory, measurements, summary })
      } catch {
        setError('API unavailable — start the Go backend on port 8081')
      }
      setLoading(false)
    }
    load()
  }, [])

  if (loading) return <div className="p-8 text-center text-[#64748b]">Loading Ring Builder data...</div>
  if (error) {
    return (
      <div className="p-8">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-8 text-center">
          <div className="text-4xl mb-4">🔘</div>
          <h2 className="text-xl font-bold text-white mb-2">Ring Builder & Segment Tracking Module</h2>
          <p className="text-[#94a3b8] mb-4">{error}</p>
        </div>
      </div>
    )
  }

  const prod = data.production || []
  const qc = data.qc || []
  const inv = data.inventory || []
  const meas = data.measurements || []
  const sm = data.summary || {}

  // Status distribution
  const statusCount: Record<string, number> = {}
  prod.forEach((p: any) => { statusCount[p.status] = (statusCount[p.status] || 0) + 1 })

  // QC results
  const qcCount: Record<string, number> = {}
  qc.forEach((q: any) => { qcCount[q.qc_result] = (qcCount[q.qc_result] || 0) + 1 })

  // Inventory
  const invGrouped: Record<string, any> = {}
  inv.forEach((i: any) => { invGrouped[i.segment_type] = { produced: i.quantity_produced || 0, installed: i.quantity_installed || 0, stock: i.quantity_in_stock || 0 } })

  return (
    <div className="p-6 space-y-6">
      <h1 className="text-2xl font-bold text-white">🔘 Ring Builder & Segment Tracking</h1>

      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4 text-center">
          <div className="text-xs text-[#94a3b8] uppercase">Designs</div><div className="text-2xl font-bold mt-2 text-blue-400">{sm.total_designs || 0}</div></div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4 text-center">
          <div className="text-xs text-[#94a3b8] uppercase">Produced</div><div className="text-2xl font-bold mt-2 text-green-400">{sm.total_produced || 0}</div></div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4 text-center">
          <div className="text-xs text-[#94a3b8] uppercase">Passed QC</div><div className="text-2xl font-bold mt-2 text-green-400">{sm.total_passed_qc || 0}</div></div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4 text-center">
          <div className="text-xs text-[#94a3b8] uppercase">Installed</div><div className="text-2xl font-bold mt-2 text-purple-400">{sm.total_installed || 0}</div></div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4 text-center">
          <div className="text-xs text-[#94a3b8] uppercase">Defective</div><div className="text-2xl font-bold mt-2 text-red-400">{sm.total_defective || 0}</div></div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4 text-center">
          <div className="text-xs text-[#94a3b8] uppercase">Measurements</div><div className="text-2xl font-bold mt-2 text-blue-400">{meas.length}</div></div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <h3 className="text-white font-medium mb-4">📦 Production Status</h3>
          <ResponsiveContainer width="100%" height={250}>
            <PieChart><Pie data={Object.entries(statusCount).map(([k, v]) => ({ name: k, value: v }))} cx="50%" cy="50%" outerRadius={80} dataKey="value" label={({ name, value }) => `${name}: ${value}`}>
              {Object.entries(statusCount).map(([k]) => <Cell key={k} fill={STATUS_COLORS[k] || '#64748b'} />)}
            </Pie></PieChart>
          </ResponsiveContainer>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <h3 className="text-white font-medium mb-4">✅ QC Results</h3>
          <ResponsiveContainer width="100%" height={250}>
            <PieChart><Pie data={Object.entries(qcCount).map(([k, v]) => ({ name: k, value: v }))} cx="50%" cy="50%" outerRadius={80} dataKey="value" label={({ name, value }) => `${name}: ${value}`}>
              {Object.entries(qcCount).map(([k]) => <Cell key={k} fill={QC_COLORS[k] || '#64748b'} />)}
            </Pie></PieChart>
          </ResponsiveContainer>
        </div>
      </div>

      <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
        <h3 className="text-white font-medium mb-4">📊 Inventory by Segment Type</h3>
        <ResponsiveContainer width="100%" height={250}>
          <BarChart data={Object.entries(invGrouped).map(([k, v]) => ({ name: k, produced: v.produced, installed: v.installed, stock: v.stock }))}>
            <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
            <XAxis dataKey="name" stroke="#64748b" />
            <YAxis stroke="#64748b" />
            <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155' }} />
            <Bar dataKey="produced" fill="#3b82f6" name="Produced" />
            <Bar dataKey="installed" fill="#22c55e" name="Installed" />
            <Bar dataKey="stock" fill="#a855f7" name="In Stock" />
          </BarChart>
        </ResponsiveContainer>
      </div>

      {/* Convergence measurements */}
      {meas.length > 0 && (
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <h3 className="text-white font-medium mb-4">📐 Ring Convergence — Ovality</h3>
          <ResponsiveContainer width="100%" height={250}>
            <ScatterChart>
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
              <XAxis dataKey="index" stroke="#64748b" name="Measurement" />
              <YAxis dataKey="ovality_pct" stroke="#64748b" name="Ovality %" />
              <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155' }} cursor={{ strokeDasharray: '3 3' }} />
              <Scatter data={meas.slice(0, 50).map((m: any, i: number) => ({ index: i, ovality_pct: m.ovality_pct || 0 }))} fill="#3b82f6" />
            </ScatterChart>
          </ResponsiveContainer>
        </div>
      )}
    </div>
  )
}