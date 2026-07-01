import { useState, useEffect } from 'react'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts'

const DISCIPLINE_COLORS: Record<string, string> = {
  architectural: '#3b82f6', structural: '#22c55e', mechanical: '#f97316',
  electrical: '#f59e0b', plumbing: '#14b8a6', civil: '#a855f7',
}

export default function BIMPage() {
  const [models, setModels] = useState<any[]>([])
  const [elements, setElements] = useState<any[]>([])
  const [clashes, setClashes] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    Promise.all([
      fetch('/api/bim/models').then(r => r.ok ? r.json() : Promise.reject()),
      fetch('/api/bim/elements').then(r => r.ok ? r.json() : Promise.reject()),
      fetch('/api/bim/clashes').then(r => r.ok ? r.json() : Promise.reject()),
    ])
      .then(([m, e, c]) => { setModels(m); setElements(e); setClashes(c) })
      .catch(() => setError('API unavailable'))
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <div className="p-8 text-center text-[#64748b]">Loading BIM data...</div>

  const disciplineCounts: Record<string, number> = {}
  models.forEach((m: any) => { disciplineCounts[m.discipline || 'other'] = (disciplineCounts[m.discipline || 'other'] || 0) + 1 })
  const discChartData = Object.entries(disciplineCounts).map(([name, value]) => ({ name, value }))

  const clashOpen = clashes.filter((c: any) => c.status === 'open' || c.status === 'new').length
  const clashResolved = clashes.filter((c: any) => c.status === 'resolved' || c.status === 'closed').length

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-white">🏗️ BIM — Building Information Modeling</h1>
        <p className="text-[#94a3b8] mt-1">BIM Integration</p>
      </div>

      {error ? (
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-8 text-center">
          <div className="text-4xl mb-4">🏗️</div>
          <h2 className="text-xl font-bold text-white mb-2">BIM Module</h2>
          <p className="text-[#94a3b8]">{error}</p>
        </div>
      ) : (
        <>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">BIM Models</div>
              <div className="text-2xl font-bold text-white mt-2">{models.length}</div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Elements</div>
              <div className="text-2xl font-bold text-[#3b82f6] mt-2">{elements.length}</div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Open Clashes</div>
              <div className="text-2xl font-bold text-[#ef4444] mt-2">{clashOpen}</div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Resolved</div>
              <div className="text-2xl font-bold text-[#22c55e] mt-2">{clashResolved}</div>
            </div>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-6">
              <h3 className="text-white font-semibold mb-4">Models by Discipline</h3>
              <ResponsiveContainer width="100%" height={300}>
                <BarChart data={discChartData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
                  <XAxis dataKey="name" tick={{ fill: '#94a3b8', fontSize: 11 }} />
                  <YAxis tick={{ fill: '#94a3b8', fontSize: 11 }} />
                  <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '8px', color: '#e2e8f0' }} />
                  <Bar dataKey="value" radius={[4, 4, 0, 0]}>
                    {discChartData.map((d) => <Cell key={d.name} fill={DISCIPLINE_COLORS[d.name] || '#64748b'} />)}
                  </Bar>
                </BarChart>
              </ResponsiveContainer>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-6">
              <h3 className="text-white font-semibold mb-4">Clash Status</h3>
              <ResponsiveContainer width="100%" height={300}>
                <PieChart>
                  <Pie data={[
                    { name: 'Open', value: clashOpen || 1 },
                    { name: 'Resolved', value: clashResolved || 1 },
                  ]} cx="50%" cy="50%" outerRadius={100} dataKey="value" label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}>
                    <Cell fill="#ef4444" />
                    <Cell fill="#22c55e" />
                  </Pie>
                  <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '8px', color: '#e2e8f0' }} />
                </PieChart>
              </ResponsiveContainer>
            </div>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl overflow-hidden">
              <div className="p-4 border-b border-[#334155]">
                <h3 className="text-white font-semibold">BIM Models</h3>
              </div>
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="text-left text-xs text-[#64748b] uppercase tracking-wider">
                      <th className="p-3 pl-4">Model</th>
                      <th className="p-3">Version</th>
                      <th className="p-3">Discipline</th>
                      <th className="p-3">LOD</th>
                      <th className="p-3 pr-4">Author</th>
                    </tr>
                  </thead>
                  <tbody>
                    {models.map((m: any) => (
                      <tr key={m.id} className="border-t border-[#1e293b] hover:bg-[#334155] text-sm">
                        <td className="p-3 pl-4 text-white">{m.model_name}</td>
                        <td className="p-3 text-[#94a3b8]">{m.model_version}</td>
                        <td className="p-3">
                          <span className="inline-block px-2 py-0.5 rounded text-xs font-semibold"
                            style={{ background: `${DISCIPLINE_COLORS[m.discipline] || '#64748b'}22`, color: DISCIPLINE_COLORS[m.discipline] || '#64748b' }}>
                            {m.discipline}
                          </span>
                        </td>
                        <td className="p-3 text-[#94a3b8]">{m.lod || '—'}</td>
                        <td className="p-3 pr-4 text-[#94a3b8] text-xs">{m.author || '—'}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl overflow-hidden">
              <div className="p-4 border-b border-[#334155]">
                <h3 className="text-white font-semibold">BIM Elements</h3>
              </div>
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="text-left text-xs text-[#64748b] uppercase tracking-wider">
                      <th className="p-3 pl-4">Element</th>
                      <th className="p-3">Type</th>
                      <th className="p-3">IFC Class</th>
                      <th className="p-3 pr-4">Material</th>
                    </tr>
                  </thead>
                  <tbody>
                    {elements.map((e: any) => (
                      <tr key={e.id} className="border-t border-[#1e293b] hover:bg-[#334155] text-sm">
                        <td className="p-3 pl-4 text-white">{e.element_name}</td>
                        <td className="p-3 text-[#94a3b8]">{e.element_type}</td>
                        <td className="p-3 font-mono text-xs text-[#64748b]">{e.ifc_class || '—'}</td>
                        <td className="p-3 pr-4 text-[#94a3b8]">{e.material || '—'}</td>
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
