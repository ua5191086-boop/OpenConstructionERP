import { useState, useEffect } from 'react'

export default function PermitsPage() {
  const [apps, setApps] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    fetch('/api/v1/protected/permits/applications').then(r => r.ok ? r.json() : Promise.reject())
      .then(d => setApps(d))
      .catch(() => setError('API unavailable'))
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <div className="p-8 text-center text-[#64748b]">Loading Permits data...</div>

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-white">📋 Permits</h1>
        <p className="text-[#94a3b8] mt-1">V030 — Permit Applications, Inspections, Renewals</p>
      </div>

      {error ? (
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-8 text-center">
          <div className="text-4xl mb-4">📋</div>
          <h2 className="text-xl font-bold text-white mb-2">Permits Module</h2>
          <p className="text-[#94a3b8]">{error}</p>
        </div>
      ) : (
        <>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Applications</div>
              <div className="text-2xl font-bold text-[#3b82f6] mt-2">{apps.length}</div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Approved</div>
              <div className="text-2xl font-bold text-[#22c55e] mt-2">{apps.filter((a: any) => a.status === 'approved').length}</div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Pending</div>
              <div className="text-2xl font-bold text-[#f97316] mt-2">{apps.filter((a: any) => a.status === 'submitted' || a.status === 'under_review').length}</div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Draft</div>
              <div className="text-2xl font-bold text-[#64748b] mt-2">{apps.filter((a: any) => a.status === 'draft').length}</div>
            </div>
          </div>

          <div className="bg-[#1e293b] border border-[#334155] rounded-xl overflow-hidden">
            <div className="p-4 border-b border-[#334155]">
              <h3 className="text-white font-semibold">Permit Applications</h3>
            </div>
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="text-left text-xs text-[#64748b] uppercase tracking-wider">
                    <th className="p-3 pl-4">Permit #</th>
                    <th className="p-3">Type</th>
                    <th className="p-3">Status</th>
                    <th className="p-3">Applied</th>
                    <th className="p-3 pr-4">Decision</th>
                  </tr>
                </thead>
                <tbody>
                  {apps.map((a: any) => (
                    <tr key={a.id} className="border-t border-[#1e293b] hover:bg-[#334155] text-sm">
                      <td className="p-3 pl-4 font-mono text-xs text-white">{a.permit_number || '—'}</td>
                      <td className="p-3 text-[#94a3b8]">{a.permit_type}</td>
                      <td className="p-3">{a.status}</td>
                      <td className="p-3 text-[#94a3b8]">{a.application_date || '—'}</td>
                      <td className="p-3 pr-4 text-[#94a3b8]">{a.decision_date || '—'}</td>
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