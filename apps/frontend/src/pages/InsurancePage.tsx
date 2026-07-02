import { useState, useEffect } from 'react'

export default function InsurancePage() {
  const [policies, setPolicies] = useState<any[]>([])
  const [claims, setClaims] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    Promise.all([
      fetch('/api/v1/protected/insurance/policies').then(r => r.ok ? r.json() : Promise.reject()),
      fetch('/api/v1/protected/insurance/claims').then(r => r.ok ? r.json() : Promise.reject()),
    ])
      .then(([p, c]) => { setPolicies(p); setClaims(c) })
      .catch(() => setError('API unavailable'))
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <div className="p-8 text-center text-[#64748b]">Loading Insurance data...</div>

  const active = policies.filter((p: any) => p.status === 'active')
  const totalInsured = active.reduce((s: number, p: any) => s + (p.sum_insured || 0), 0)

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-white">🛡️ Insurance</h1>
        <p className="text-[#94a3b8] mt-1">V031 — Policies, Claims, Coverage & Certificates</p>
      </div>

      {error ? (
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-8 text-center">
          <div className="text-4xl mb-4">🛡️</div>
          <h2 className="text-xl font-bold text-white mb-2">Insurance Module</h2>
          <p className="text-[#94a3b8]">{error}</p>
        </div>
      ) : (
        <>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Active Policies</div>
              <div className="text-2xl font-bold text-[#3b82f6] mt-2">{active.length}</div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Total Insured</div>
              <div className="text-2xl font-bold text-[#22c55e] mt-2">${(totalInsured / 1e6).toFixed(1)}M</div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Claims</div>
              <div className="text-2xl font-bold text-[#f97316] mt-2">{claims.length}</div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Settled</div>
              <div className="text-2xl font-bold text-[#22c55e] mt-2">{claims.filter((c: any) => c.status === 'settled' || c.status === 'closed').length}</div>
            </div>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl overflow-hidden">
              <div className="p-4 border-b border-[#334155]">
                <h3 className="text-white font-semibold">Insurance Policies</h3>
              </div>
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="text-left text-xs text-[#64748b] uppercase tracking-wider">
                      <th className="p-3 pl-4">Policy #</th>
                      <th className="p-3">Type</th>
                      <th className="p-3">Insurer</th>
                      <th className="p-3 text-right pr-4">Sum Insured</th>
                    </tr>
                  </thead>
                  <tbody>
                    {active.map((p: any) => (
                      <tr key={p.id} className="border-t border-[#1e293b] hover:bg-[#334155] text-sm">
                        <td className="p-3 pl-4 font-mono text-xs text-white">{p.policy_number}</td>
                        <td className="p-3 text-[#94a3b8]">{p.policy_type}</td>
                        <td className="p-3 text-[#94a3b8]">{p.insurer}</td>
                        <td className="p-3 text-right pr-4 text-white">${Number(p.sum_insured || 0).toLocaleString()}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>

            <div className="bg-[#1e293b] border border-[#334155] rounded-xl overflow-hidden">
              <div className="p-4 border-b border-[#334155]">
                <h3 className="text-white font-semibold">Insurance Claims</h3>
              </div>
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="text-left text-xs text-[#64748b] uppercase tracking-wider">
                      <th className="p-3 pl-4">Claim #</th>
                      <th className="p-3">Status</th>
                      <th className="p-3 text-right">Claimed</th>
                      <th className="p-3 text-right pr-4">Settled</th>
                    </tr>
                  </thead>
                  <tbody>
                    {claims.map((c: any) => (
                      <tr key={c.id} className="border-t border-[#1e293b] hover:bg-[#334155] text-sm">
                        <td className="p-3 pl-4 font-mono text-xs text-[#64748b]">{c.claim_number}</td>
                        <td className="p-3">{c.status}</td>
                        <td className="p-3 text-right text-white">${Number(c.claimed_amount || 0).toLocaleString()}</td>
                        <td className="p-3 text-right pr-4 text-white">${Number(c.settled_amount || 0).toLocaleString()}</td>
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