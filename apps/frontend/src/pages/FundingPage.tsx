import { useState, useEffect } from 'react'

const STATUS_COLORS: Record<string, string> = {
  active: '#22c55e', planned: '#3b82f6', disbursed: '#a855f7', cancelled: '#ef4444', pending: '#f97316',
}

export default function FundingPage() {
  const [sources, setSources] = useState<any[]>([])
  const [guarantees, setGuarantees] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    Promise.all([
      fetch('/api/v1/protected/funding/sources').then(r => r.ok ? r.json() : Promise.reject()),
      fetch('/api/v1/protected/funding/guarantees').then(r => r.ok ? r.json() : Promise.reject()),
    ])
      .then(([s, g]) => { setSources(s); setGuarantees(g) })
      .catch(() => setError('API unavailable'))
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <div className="p-8 text-center text-[#64748b]">Loading Funding data...</div>

  const totalCommit = sources.reduce((s: number, x: any) => s + (x.commitment_amount || 0), 0)

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-white">💰 Funding & Guarantees</h1>
        <p className="text-[#94a3b8] mt-1">V027 — Funding Sources, Multi-Currency, Guarantees</p>
      </div>

      {error ? (
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-8 text-center">
          <div className="text-4xl mb-4">💰</div>
          <h2 className="text-xl font-bold text-white mb-2">Funding Module</h2>
          <p className="text-[#94a3b8]">{error}</p>
        </div>
      ) : (
        <>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Sources</div>
              <div className="text-2xl font-bold text-[#3b82f6] mt-2">{sources.length}</div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Total Commitment</div>
              <div className="text-2xl font-bold text-[#22c55e] mt-2">${(totalCommit / 1e6).toFixed(1)}M</div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Guarantees</div>
              <div className="text-2xl font-bold text-[#a855f7] mt-2">{guarantees.length}</div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Active Guarantees</div>
              <div className="text-2xl font-bold text-[#22c55e] mt-2">{guarantees.filter((g: any) => g.status === 'active').length}</div>
            </div>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl overflow-hidden">
              <div className="p-4 border-b border-[#334155]">
                <h3 className="text-white font-semibold">Funding Sources</h3>
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
                    {sources.map((s: any) => (
                      <tr key={s.id} className="border-t border-[#1e293b] hover:bg-[#334155] text-sm">
                        <td className="p-3 pl-4 text-white">{s.source_name}</td>
                        <td className="p-3 text-[#94a3b8]">{s.source_type}</td>
                        <td className="p-3">
                          <span className="inline-block px-2 py-0.5 rounded text-xs font-semibold"
                            style={{ background: `${STATUS_COLORS[s.status] || '#64748b'}22`, color: STATUS_COLORS[s.status] || '#64748b' }}>
                            {s.status}
                          </span>
                        </td>
                        <td className="p-3 text-right pr-4 text-white">${Number(s.commitment_amount || 0).toLocaleString()}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>

            <div className="bg-[#1e293b] border border-[#334155] rounded-xl overflow-hidden">
              <div className="p-4 border-b border-[#334155]">
                <h3 className="text-white font-semibold">Guarantees</h3>
              </div>
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="text-left text-xs text-[#64748b] uppercase tracking-wider">
                      <th className="p-3 pl-4">Type</th>
                      <th className="p-3">Number</th>
                      <th className="p-3">Status</th>
                      <th className="p-3 text-right pr-4">Amount</th>
                    </tr>
                  </thead>
                  <tbody>
                    {guarantees.map((g: any) => (
                      <tr key={g.id} className="border-t border-[#1e293b] hover:bg-[#334155] text-sm">
                        <td className="p-3 pl-4 text-white">{g.guarantee_type}</td>
                        <td className="p-3 font-mono text-xs text-[#64748b]">{g.guarantee_number || '—'}</td>
                        <td className="p-3">
                          <span className="inline-block px-2 py-0.5 rounded text-xs font-semibold"
                            style={{ background: `${STATUS_COLORS[g.status] || '#64748b'}22`, color: STATUS_COLORS[g.status] || '#64748b' }}>
                            {g.status}
                          </span>
                        </td>
                        <td className="p-3 text-right pr-4 text-white">${Number(g.amount || 0).toLocaleString()}</td>
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