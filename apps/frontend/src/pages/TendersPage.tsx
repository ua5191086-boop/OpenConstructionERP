import { useState, useEffect } from 'react'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts'

const STATUS_COLORS: Record<string, string> = {
  draft: '#64748b', published: '#3b82f6', open: '#22c55e', evaluating: '#f97316',
  awarded: '#a855f7', cancelled: '#ef4444', closed: '#14b8a6',
}

export default function TendersPage() {
  const [tenders, setTenders] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    fetch('/api/tenders')
      .then(r => r.ok ? r.json() : Promise.reject('API unavailable'))
      .then(setTenders)
      .catch(() => setError('API unavailable'))
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <div className="p-8 text-center text-[#64748b]">Loading Tenders...</div>

  const statusCounts: Record<string, number> = {}
  tenders.forEach((t: any) => { statusCounts[t.status || 'draft'] = (statusCounts[t.status || 'draft'] || 0) + 1 })
  const statusChartData = Object.entries(statusCounts).map(([name, value]) => ({ name, value }))

  const totalBudget = tenders.reduce((s: number, t: any) => s + (t.budget_amount || 0), 0)

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-white">📢 Tenders</h1>
        <p className="text-[#94a3b8] mt-1">Tender Management</p>
      </div>

      {error ? (
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-8 text-center">
          <div className="text-4xl mb-4">📢</div>
          <h2 className="text-xl font-bold text-white mb-2">Tenders Module</h2>
          <p className="text-[#94a3b8]">{error} — start the Go API server</p>
        </div>
      ) : (
        <>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Total Tenders</div>
              <div className="text-2xl font-bold text-white mt-2">{tenders.length}</div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Total Budget</div>
              <div className="text-2xl font-bold text-[#22c55e] mt-2">${(totalBudget / 1e6).toFixed(1)}M</div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Open</div>
              <div className="text-2xl font-bold text-[#22c55e] mt-2">{statusCounts['open'] || 0}</div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Awarded</div>
              <div className="text-2xl font-bold text-[#a855f7] mt-2">{statusCounts['awarded'] || 0}</div>
            </div>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-6">
              <h3 className="text-white font-semibold mb-4">Tenders by Status</h3>
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
              <h3 className="text-white font-semibold">All Tenders</h3>
            </div>
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="text-left text-xs text-[#64748b] uppercase tracking-wider">
                    <th className="p-3 pl-4">Code</th>
                    <th className="p-3">Name</th>
                    <th className="p-3">Type</th>
                    <th className="p-3">Status</th>
                    <th className="p-3 text-right">Budget</th>
                    <th className="p-3">Deadline</th>
                    <th className="p-3 pr-4">Bid Open</th>
                  </tr>
                </thead>
                <tbody>
                  {tenders.map((t: any) => (
                    <tr key={t.id} className="border-t border-[#1e293b] hover:bg-[#334155] text-sm">
                      <td className="p-3 pl-4 font-mono text-xs text-[#64748b]">{t.code}</td>
                      <td className="p-3 text-white">{t.name}</td>
                      <td className="p-3 text-[#94a3b8]">{t.tender_type}</td>
                      <td className="p-3">
                        <span className="inline-block px-2 py-0.5 rounded text-xs font-semibold"
                          style={{ background: `${STATUS_COLORS[t.status] || '#64748b'}22`, color: STATUS_COLORS[t.status] || '#64748b' }}>
                          {t.status}
                        </span>
                      </td>
                      <td className="p-3 text-right text-white">${Number(t.budget_amount || 0).toLocaleString()}</td>
                      <td className="p-3 text-[#94a3b8] text-xs">{t.submission_deadline ? new Date(t.submission_deadline).toLocaleDateString() : '—'}</td>
                      <td className="p-3 pr-4 text-[#94a3b8] text-xs">{t.bid_open_date ? new Date(t.bid_open_date).toLocaleDateString() : '—'}</td>
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
