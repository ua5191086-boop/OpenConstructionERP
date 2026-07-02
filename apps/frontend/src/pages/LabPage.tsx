import { useState, useEffect } from 'react'

export default function LabPage() {
  const [tests, setTests] = useState<any[]>([])
  const [equipment, setEquipment] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    Promise.all([
      fetch('/api/v1/protected/lab/tests').then(r => r.ok ? r.json() : Promise.reject()),
      fetch('/api/v1/protected/lab/equipment').then(r => r.ok ? r.json() : Promise.reject()),
    ])
      .then(([t, e]) => { setTests(t); setEquipment(e) })
      .catch(() => setError('API unavailable'))
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <div className="p-8 text-center text-[#64748b]">Loading Laboratory data...</div>

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-white">🔬 Laboratory</h1>
        <p className="text-[#94a3b8] mt-1">V029 — Material Testing, Equipment, Sampling</p>
      </div>

      {error ? (
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-8 text-center">
          <div className="text-4xl mb-4">🔬</div>
          <h2 className="text-xl font-bold text-white mb-2">Laboratory Module</h2>
          <p className="text-[#94a3b8]">{error}</p>
        </div>
      ) : (
        <>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Tests</div>
              <div className="text-2xl font-bold text-[#3b82f6] mt-2">{tests.length}</div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Completed</div>
              <div className="text-2xl font-bold text-[#22c55e] mt-2">{tests.filter((t: any) => t.status === 'completed').length}</div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Equipment</div>
              <div className="text-2xl font-bold text-[#a855f7] mt-2">{equipment.length}</div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Pending</div>
              <div className="text-2xl font-bold text-[#f97316] mt-2">{tests.filter((t: any) => t.status === 'pending').length}</div>
            </div>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl overflow-hidden">
              <div className="p-4 border-b border-[#334155]">
                <h3 className="text-white font-semibold">Material Tests</h3>
              </div>
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="text-left text-xs text-[#64748b] uppercase tracking-wider">
                      <th className="p-3 pl-4">Number</th>
                      <th className="p-3">Material</th>
                      <th className="p-3">Type</th>
                      <th className="p-3">Status</th>
                    </tr>
                  </thead>
                  <tbody>
                    {tests.map((t: any) => (
                      <tr key={t.id} className="border-t border-[#1e293b] hover:bg-[#334155] text-sm">
                        <td className="p-3 pl-4 font-mono text-xs text-white">{t.test_number}</td>
                        <td className="p-3 text-[#94a3b8]">{t.material_type}</td>
                        <td className="p-3 text-[#94a3b8]">{t.test_type}</td>
                        <td className="p-3">
                          <span className="inline-block px-2 py-0.5 rounded text-xs font-semibold" style={{
                            background: t.status === 'completed' ? '#22c55e22' : t.status === 'in_progress' ? '#3b82f622' : '#f9731622',
                            color: t.status === 'completed' ? '#22c55e' : t.status === 'in_progress' ? '#3b82f6' : '#f97316'
                          }}>{t.status}</span>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>

            <div className="bg-[#1e293b] border border-[#334155] rounded-xl overflow-hidden">
              <div className="p-4 border-b border-[#334155]">
                <h3 className="text-white font-semibold">Lab Equipment</h3>
              </div>
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="text-left text-xs text-[#64748b] uppercase tracking-wider">
                      <th className="p-3 pl-4">Code</th>
                      <th className="p-3">Name</th>
                      <th className="p-3">Status</th>
                      <th className="p-3">Calibration</th>
                    </tr>
                  </thead>
                  <tbody>
                    {equipment.map((e: any) => (
                      <tr key={e.id} className="border-t border-[#1e293b] hover:bg-[#334155] text-sm">
                        <td className="p-3 pl-4 font-mono text-xs text-[#64748b]">{e.equipment_code}</td>
                        <td className="p-3 text-white">{e.equipment_name}</td>
                        <td className="p-3">{e.status}</td>
                        <td className="p-3 text-[#94a3b8]">{e.calibration_due || '—'}</td>
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