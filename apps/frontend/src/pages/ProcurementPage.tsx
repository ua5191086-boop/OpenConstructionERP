import { useState, useEffect } from 'react'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts'

const STATUS_COLORS: Record<string, string> = {
  draft: '#64748b', submitted: '#3b82f6', approved: '#22c55e', rejected: '#ef4444', ordered: '#a855f7', received: '#14b8a6',
}
const PRIORITY_COLORS: Record<string, string> = { low: '#64748b', medium: '#f97316', high: '#ef4444', critical: '#dc2626' }

export default function ProcurementPage() {
  const [requests, setRequests] = useState<any[]>([])
  const [purchaseOrders, setPurchaseOrders] = useState<any[]>([])
  const [inventory, setInventory] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    Promise.all([
      fetch('/api/procurement/requests').then(r => r.ok ? r.json() : Promise.reject()),
      fetch('/api/procurement/purchase-orders').then(r => r.ok ? r.json() : Promise.reject()),
      fetch('/api/procurement/inventory').then(r => r.ok ? r.json() : Promise.reject()),
    ])
      .then(([r, p, i]) => { setRequests(r); setPurchaseOrders(p); setInventory(i) })
      .catch(() => setError('API unavailable'))
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <div className="p-8 text-center text-[#64748b]">Loading Procurement data...</div>

  const statusCounts: Record<string, number> = {}
  requests.forEach((r: any) => { statusCounts[r.status || 'draft'] = (statusCounts[r.status || 'draft'] || 0) + 1 })
  const statusChartData = Object.entries(statusCounts).map(([name, value]) => ({ name, value }))

  const totalEstCost = requests.reduce((s: number, r: any) => s + (r.estimated_cost || 0), 0)
  const totalPO = purchaseOrders.reduce((s: number, p: any) => s + (p.total_amount || 0), 0)

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-white">📦 Procurement</h1>
        <p className="text-[#94a3b8] mt-1">Procurement & Inventory Management</p>
      </div>

      {error ? (
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-8 text-center">
          <div className="text-4xl mb-4">📦</div>
          <h2 className="text-xl font-bold text-white mb-2">Procurement Module</h2>
          <p className="text-[#94a3b8]">{error}</p>
        </div>
      ) : (
        <>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Requests</div>
              <div className="text-2xl font-bold text-white mt-2">{requests.length}</div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Est. Cost</div>
              <div className="text-2xl font-bold text-[#f97316] mt-2">${(totalEstCost / 1e3).toFixed(0)}K</div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Purchase Orders</div>
              <div className="text-2xl font-bold text-[#3b82f6] mt-2">{purchaseOrders.length}</div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Inventory Items</div>
              <div className="text-2xl font-bold text-[#22c55e] mt-2">{inventory.length}</div>
            </div>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-6">
              <h3 className="text-white font-semibold mb-4">Requests by Status</h3>
              <ResponsiveContainer width="100%" height={300}>
                <BarChart data={statusChartData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
                  <XAxis dataKey="name" tick={{ fill: '#94a3b8', fontSize: 11 }} />
                  <YAxis tick={{ fill: '#94a3b8', fontSize: 11 }} />
                  <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '8px', color: '#e2e8f0' }} />
                  <Bar dataKey="value" radius={[4, 4, 0, 0]}>
                    {statusChartData.map((d) => <Cell key={d.name} fill={STATUS_COLORS[d.name] || '#64748b'} />)}
                  </Bar>
                </BarChart>
              </ResponsiveContainer>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-6">
              <h3 className="text-white font-semibold mb-4">Purchase Orders</h3>
              <ResponsiveContainer width="100%" height={300}>
                <PieChart>
                  <Pie data={[
                    { name: 'Total PO Value', value: totalPO },
                    { name: 'Est. Requests', value: totalEstCost },
                  ]} cx="50%" cy="50%" outerRadius={100} dataKey="value" label={({ name, value }) => `${name}: $${(value / 1e3).toFixed(0)}K`}>
                    <Cell fill="#3b82f6" />
                    <Cell fill="#f97316" />
                  </Pie>
                  <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '8px', color: '#e2e8f0' }} />
                </PieChart>
              </ResponsiveContainer>
            </div>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl overflow-hidden">
              <div className="p-4 border-b border-[#334155]">
                <h3 className="text-white font-semibold">Procurement Requests</h3>
              </div>
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="text-left text-xs text-[#64748b] uppercase tracking-wider">
                      <th className="p-3 pl-4">Number</th>
                      <th className="p-3">Status</th>
                      <th className="p-3">Priority</th>
                      <th className="p-3 text-right pr-4">Est. Cost</th>
                    </tr>
                  </thead>
                  <tbody>
                    {requests.map((r: any) => (
                      <tr key={r.id} className="border-t border-[#1e293b] hover:bg-[#334155] text-sm">
                        <td className="p-3 pl-4 font-mono text-xs text-[#64748b]">{r.request_number}</td>
                        <td className="p-3">
                          <span className="inline-block px-2 py-0.5 rounded text-xs font-semibold"
                            style={{ background: `${STATUS_COLORS[r.status] || '#64748b'}22`, color: STATUS_COLORS[r.status] || '#64748b' }}>
                            {r.status}
                          </span>
                        </td>
                        <td className="p-3">
                          <span className="inline-block px-2 py-0.5 rounded text-xs font-semibold"
                            style={{ background: `${PRIORITY_COLORS[r.priority] || '#64748b'}22`, color: PRIORITY_COLORS[r.priority] || '#64748b' }}>
                            {r.priority}
                          </span>
                        </td>
                        <td className="p-3 text-right pr-4 text-white">${Number(r.estimated_cost || 0).toLocaleString()}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl overflow-hidden">
              <div className="p-4 border-b border-[#334155]">
                <h3 className="text-white font-semibold">Inventory</h3>
              </div>
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="text-left text-xs text-[#64748b] uppercase tracking-wider">
                      <th className="p-3 pl-4">Item</th>
                      <th className="p-3">Category</th>
                      <th className="p-3 text-right">Qty</th>
                      <th className="p-3 text-right pr-4">Unit Price</th>
                    </tr>
                  </thead>
                  <tbody>
                    {inventory.map((i: any) => (
                      <tr key={i.id} className="border-t border-[#1e293b] hover:bg-[#334155] text-sm">
                        <td className="p-3 pl-4 text-white">{i.item_name || i.name}</td>
                        <td className="p-3 text-[#94a3b8]">{i.category || '—'}</td>
                        <td className="p-3 text-right text-white">{Number(i.quantity || 0).toLocaleString()}</td>
                        <td className="p-3 text-right pr-4 text-white">${Number(i.unit_price || 0).toLocaleString()}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </>
      )}
    </div>
  )
}
