import { useState, useEffect } from 'react'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts'

const STATUS_COLORS: Record<string, string> = {
  draft: '#64748b', active: '#22c55e', completed: '#3b82f6', terminated: '#ef4444', on_hold: '#f97316',
}

export default function ContractsPage() {
  const [contracts, setContracts] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    fetch('/api/contracts')
      .then(r => r.ok ? r.json() : Promise.reject('API unavailable'))
      .then(setContracts)
      .catch(() => setError('API unavailable'))
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <div className="p-8 text-center text-[#64748b]">Loading Contracts...</div>

  const statusCounts: Record<string, number> = {}
  contracts.forEach((c: any) => { statusCounts[c.status || 'draft'] = (statusCounts[c.status || 'draft'] || 0) + 1 })
  const statusChartData = Object.entries(statusCounts).map(([name, value]) => ({ name, value }))

  const totalAmount = contracts.reduce((s: number, c: any) => s + (c.contract_amount || 0), 0)

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-white">📝 Contracts</h1>
        <p className="text-[#94a3b8] mt-1">Contract Management</p>
      </div>

      {error ? (
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-8 text-center">
          <div className="text-4xl mb-4">📝</div>
          <h2 className="text-xl font-bold text-white mb-2">Contracts Module</h2>
          <p className="text-[#94a3b8]">{error}</p>
        </div>
      ) : (
        <>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Total Contracts</div>
              <div className="text-2xl font-bold text-white mt-2">{contracts.length}</div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Total Value</div>
              <div className="text-2xl font-bold text-[#22c55e] mt-2">${(totalAmount / 1e6).toFixed(1)}M</div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Active</div>
              <div className="text-2xl font-bold text-[#22c55e] mt-2">{statusCounts['active'] || 0}</div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Completed</div>
              <div className="text-2xl font-bold text-[#3b82f6] mt-2">{statusCounts['completed'] || 0}</div>
            </div>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-6">
              <h3 className="text-white font-semibold mb-4">Contracts by Status</h3>
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
              <h3 className="text-white font-semibold mb-4">Status Distribution</h3>
              <ResponsiveContainer width="100%" height={300}>
                <PieChart>
                  <Pie data={statusChartData} cx="50%" cy="50%" outerRadius={100} dataKey="value" label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}>
                    {statusChartData.map((d) => <Cell key={d.name} fill={STATUS_COLORS[d.name] || '#64748b'} />)}
                  </Pie>
                  <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '8px', color: '#e2e8f0' }} />
                </PieChart>
              </ResponsiveContainer>
            </div>
          </div>

          <div className="bg-[#1e293b] border border-[#334155] rounded-xl overflow-hidden">
            <div className="p-4 border-b border-[#334155]">
              <h3 className="text-white font-semibold">All Contracts</h3>
            </div>
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="text-left text-xs text-[#64748b] uppercase tracking-wider">
                    <th className="p-3 pl-4">Code</th>
                    <th className="p-3">Name</th>
                    <th className="p-3">Type</th>
                    <th className="p-3">Status</th>
                    <th className="p-3 text-right">Amount</th>
                    <th className="p-3">Start</th>
                    <th className="p-3 pr-4">End</th>
                  </tr>
                </thead>
                <tbody>
                  {contracts.map((c: any) => (
                    <tr key={c.id} className="border-t border-[#1e293b] hover:bg-[#334155] text-sm">
                      <td className="p-3 pl-4 font-mono text-xs text-[#64748b]">{c.code}</td>
                      <td className="p-3 text-white">{c.name}</td>
                      <td className="p-3 text-[#94a3b8]">{c.contract_type}</td>
                      <td className="p-3">
                        <span className="inline-block px-2 py-0.5 rounded text-xs font-semibold"
                          style={{ background: `${STATUS_COLORS[c.status] || '#64748b'}22`, color: STATUS_COLORS[c.status] || '#64748b' }}>
                          {c.status}
                        </span>
                      </td>
                      <td className="p-3 text-right text-white">${Number(c.contract_amount || 0).toLocaleString()}</td>
                      <td className="p-3 text-[#94a3b8] text-xs">{c.start_date ? new Date(c.start_date).toLocaleDateString() : '—'}</td>
                      <td className="p-3 pr-4 text-[#94a3b8] text-xs">{c.end_date ? new Date(c.end_date).toLocaleDateString() : '—'}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </>
      )}
    </div>
  )
}
