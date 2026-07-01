import { useState, useEffect } from 'react'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, LineChart, Line, PieChart, Pie, Cell } from 'recharts'

const STATUS_COLORS: Record<string, string> = {
  draft: '#64748b', submitted: '#3b82f6', approved: '#22c55e', paid: '#a855f7', overdue: '#ef4444', pending: '#f97316',
}

export default function FinancePage() {
  const [budgets, setBudgets] = useState<any[]>([])
  const [invoices, setInvoices] = useState<any[]>([])
  const [cashFlow, setCashFlow] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    Promise.all([
      fetch('/api/finance/budgets').then(r => r.ok ? r.json() : Promise.reject()),
      fetch('/api/finance/invoices').then(r => r.ok ? r.json() : Promise.reject()),
      fetch('/api/finance/cash-flow').then(r => r.ok ? r.json() : Promise.reject()),
    ])
      .then(([b, i, c]) => { setBudgets(b); setInvoices(i); setCashFlow(c) })
      .catch(() => setError('API unavailable'))
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <div className="p-8 text-center text-[#64748b]">Loading Finance data...</div>

  const totalBudget = budgets.reduce((s: number, b: any) => s + (b.total_amount || 0), 0)
  const totalInvoiced = invoices.reduce((s: number, i: any) => s + (i.amount || 0), 0)

  const invStatusCounts: Record<string, number> = {}
  invoices.forEach((i: any) => { invStatusCounts[i.status || 'draft'] = (invStatusCounts[i.status || 'draft'] || 0) + 1 })
  const invChartData = Object.entries(invStatusCounts).map(([name, value]) => ({ name, value }))

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-white">💰 Finance</h1>
        <p className="text-[#94a3b8] mt-1">Financial Management</p>
      </div>

      {error ? (
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-8 text-center">
          <div className="text-4xl mb-4">💰</div>
          <h2 className="text-xl font-bold text-white mb-2">Finance Module</h2>
          <p className="text-[#94a3b8]">{error}</p>
        </div>
      ) : (
        <>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Total Budget</div>
              <div className="text-2xl font-bold text-[#22c55e] mt-2">${(totalBudget / 1e6).toFixed(1)}M</div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Invoiced</div>
              <div className="text-2xl font-bold text-[#3b82f6] mt-2">${(totalInvoiced / 1e6).toFixed(1)}M</div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Budgets</div>
              <div className="text-2xl font-bold text-white mt-2">{budgets.length}</div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Invoices</div>
              <div className="text-2xl font-bold text-white mt-2">{invoices.length}</div>
            </div>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-6">
              <h3 className="text-white font-semibold mb-4">Invoices by Status</h3>
              <ResponsiveContainer width="100%" height={300}>
                <BarChart data={invChartData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
                  <XAxis dataKey="name" tick={{ fill: '#94a3b8', fontSize: 11 }} />
                  <YAxis tick={{ fill: '#94a3b8', fontSize: 11 }} />
                  <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '8px', color: '#e2e8f0' }} />
                  <Bar dataKey="value" radius={[4, 4, 0, 0]}>
                    {invChartData.map((d) => <Cell key={d.name} fill={STATUS_COLORS[d.name] || '#64748b'} />)}
                  </Bar>
                </BarChart>
              </ResponsiveContainer>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-6">
              <h3 className="text-white font-semibold mb-4">Cash Flow</h3>
              <ResponsiveContainer width="100%" height={300}>
                <LineChart data={cashFlow.length > 0 ? cashFlow : [{ period: 'No data', amount: 0 }]}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
                  <XAxis dataKey="period" tick={{ fill: '#94a3b8', fontSize: 11 }} />
                  <YAxis tick={{ fill: '#94a3b8', fontSize: 11 }} />
                  <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '8px', color: '#e2e8f0' }} />
                  <Line type="monotone" dataKey="amount" stroke="#22c55e" strokeWidth={2} dot={{ fill: '#22c55e' }} />
                </LineChart>
              </ResponsiveContainer>
            </div>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl overflow-hidden">
              <div className="p-4 border-b border-[#334155]">
                <h3 className="text-white font-semibold">Budgets</h3>
              </div>
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="text-left text-xs text-[#64748b] uppercase tracking-wider">
                      <th className="p-3 pl-4">Name</th>
                      <th className="p-3">Type</th>
                      <th className="p-3">Status</th>
                      <th className="p-3 text-right pr-4">Amount</th>
                    </tr>
                  </thead>
                  <tbody>
                    {budgets.map((b: any) => (
                      <tr key={b.id} className="border-t border-[#1e293b] hover:bg-[#334155] text-sm">
                        <td className="p-3 pl-4 text-white">{b.name}</td>
                        <td className="p-3 text-[#94a3b8]">{b.budget_type}</td>
                        <td className="p-3">
                          <span className="inline-block px-2 py-0.5 rounded text-xs font-semibold"
                            style={{ background: `${STATUS_COLORS[b.status] || '#64748b'}22`, color: STATUS_COLORS[b.status] || '#64748b' }}>
                            {b.status}
                          </span>
                        </td>
                        <td className="p-3 text-right pr-4 text-white">${Number(b.total_amount || 0).toLocaleString()}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl overflow-hidden">
              <div className="p-4 border-b border-[#334155]">
                <h3 className="text-white font-semibold">Invoices</h3>
              </div>
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="text-left text-xs text-[#64748b] uppercase tracking-wider">
                      <th className="p-3 pl-4">Number</th>
                      <th className="p-3">Status</th>
                      <th className="p-3 text-right">Amount</th>
                      <th className="p-3 pr-4">Due Date</th>
                    </tr>
                  </thead>
                  <tbody>
                    {invoices.map((i: any) => (
                      <tr key={i.id} className="border-t border-[#1e293b] hover:bg-[#334155] text-sm">
                        <td className="p-3 pl-4 font-mono text-xs text-[#64748b]">{i.invoice_number}</td>
                        <td className="p-3">
                          <span className="inline-block px-2 py-0.5 rounded text-xs font-semibold"
                            style={{ background: `${STATUS_COLORS[i.status] || '#64748b'}22`, color: STATUS_COLORS[i.status] || '#64748b' }}>
                            {i.status}
                          </span>
                        </td>
                        <td className="p-3 text-right text-white">${Number(i.amount || 0).toLocaleString()}</td>
                        <td className="p-3 pr-4 text-[#94a3b8] text-xs">{i.due_date ? new Date(i.due_date).toLocaleDateString() : '—'}</td>
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
