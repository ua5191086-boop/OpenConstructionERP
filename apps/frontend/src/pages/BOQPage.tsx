import { useState, useEffect } from 'react'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts'

const CBS_COLORS = ['#60a5fa','#fbbf24','#34d399','#f472b6','#a78bfa','#38bdf8','#fb923c','#4ade80','#facc15','#f87171','#818cf8','#fca5a5']

export default function BOQPage() {
  const [data, setData] = useState<any>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [sectionFilter, setSectionFilter] = useState('')
  const [cbsFilter, setCbsFilter] = useState('')

  useEffect(() => {
    async function load() {
      try {
        const [items, sections, complexes, objects, chapters] = await Promise.all([
          fetch('/api/boq/items').then(r => r.ok ? r.json() : []),
          fetch('/api/boq/sections').then(r => r.ok ? r.json() : []),
          fetch('/api/boq/complexes').then(r => r.ok ? r.json() : []),
          fetch('/api/boq/objects').then(r => r.ok ? r.json() : []),
          fetch('/api/boq/cbs-chapters').then(r => r.ok ? r.json() : []),
        ])
        setData({ items, sections, complexes, objects, chapters })
      } catch {
        setError('API unavailable — use demo data or start the Go backend')
      }
      setLoading(false)
    }
    load()
  }, [])

  if (loading) return <div className="p-8 text-center text-[#64748b]">Loading BOQ data...</div>
  if (error) {
    return (
      <div className="p-8">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-8 text-center">
          <div className="text-4xl mb-4">📋</div>
          <h2 className="text-xl font-bold text-white mb-2">BOQ Module</h2>
          <p className="text-[#94a3b8] mb-4">{error}</p>
          <p className="text-sm text-[#64748b]">Start the Go API server on port 8081 to see live data</p>
        </div>
      </div>
    )
  }

  const items = data.items || []
  const sections = data.sections || []
  const complexes = data.complexes || []
  const objects = data.objects || []

  // Compute stats
  const totalCost = items.reduce((s: number, i: any) => s + (i.total_cost || 0), 0)
  const totalItems = items.length

  // CBS chart data
  const cbsTotals: Record<string, number> = {}
  items.forEach((i: any) => {
    const cbs = (i.cbs_code || '00').split('.')[0]
    cbsTotals[cbs] = (cbsTotals[cbs] || 0) + (i.total_cost || 0)
  })
  const cbsChartData = Object.entries(cbsTotals)
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([code, value]) => ({ name: `CBS-${code}`, value }))

  // Section chart data
  const sectionMap: Record<string, string> = {}
  sections.forEach((s: any) => { sectionMap[s.id] = s.name })
  const complexMap: Record<string, string> = {}
  complexes.forEach((c: any) => { complexMap[c.id] = c.section_id })
  const objectMap: Record<string, string> = {}
  objects.forEach((o: any) => { objectMap[o.id] = o.complex_id })
  const sectionTotals: Record<string, number> = {}
  items.forEach((i: any) => {
    const compId = objectMap[i.object_id]
    const secId = compId ? complexMap[compId] : null
    const secName = secId ? sectionMap[secId] || 'Unknown' : 'Unknown'
    sectionTotals[secName] = (sectionTotals[secName] || 0) + (i.total_cost || 0)
  })
  const sectionChartData = Object.entries(sectionTotals).map(([name, value]) => ({ name, value }))

  // Filter items
  let filteredItems = [...items]
  if (sectionFilter) {
    const compIds = new Set(complexes.filter((c: any) => c.section_id === sectionFilter).map((c: any) => c.id))
    const objIds = new Set(objects.filter((o: any) => compIds.has(o.complex_id)).map((o: any) => o.id))
    filteredItems = filteredItems.filter((i: any) => objIds.has(i.object_id))
  }
  if (cbsFilter) {
    filteredItems = filteredItems.filter((i: any) => (i.cbs_code || '').startsWith(cbsFilter))
  }

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-white">📋 BOQ — Bill of Quantities</h1>
        <p className="text-[#94a3b8] mt-1">Railway Infrastructure Project</p>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <div className="text-[#94a3b8] text-xs uppercase tracking-wider">Total BOQ Cost</div>
          <div className="text-2xl font-bold text-[#22c55e] mt-2">${(totalCost / 1e6).toFixed(2)}M</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <div className="text-[#94a3b8] text-xs uppercase tracking-wider">Sections</div>
          <div className="text-2xl font-bold text-[#3b82f6] mt-2">{sections.length}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <div className="text-[#94a3b8] text-xs uppercase tracking-wider">Complexes</div>
          <div className="text-2xl font-bold text-[#a855f7] mt-2">{complexes.length}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <div className="text-[#94a3b8] text-xs uppercase tracking-wider">BOQ Items</div>
          <div className="text-2xl font-bold text-[#f97316] mt-2">{totalItems}</div>
        </div>
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-6">
          <h3 className="text-white font-semibold mb-4">Cost by CBS Chapter</h3>
          <ResponsiveContainer width="100%" height={300}>
            <BarChart data={cbsChartData}>
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
              <XAxis dataKey="name" tick={{ fill: '#94a3b8', fontSize: 11 }} />
              <YAxis tick={{ fill: '#94a3b8', fontSize: 11 }} tickFormatter={(v) => `$${(v / 1e6).toFixed(1)}M`} />
              <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '8px', color: '#e2e8f0' }} />
              <Bar dataKey="value" radius={[4, 4, 0, 0]}>
                {cbsChartData.map((_, i) => <Cell key={i} fill={CBS_COLORS[i % CBS_COLORS.length]} />)}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-6">
          <h3 className="text-white font-semibold mb-4">Cost by Section</h3>
          <ResponsiveContainer width="100%" height={300}>
            <PieChart>
              <Pie data={sectionChartData} cx="50%" cy="50%" outerRadius={100} dataKey="value" label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}>
                {sectionChartData.map((_, i) => <Cell key={i} fill={CBS_COLORS[i % CBS_COLORS.length]} />)}
              </Pie>
              <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '8px', color: '#e2e8f0' }} />
            </PieChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Table */}
      <div className="bg-[#1e293b] border border-[#334155] rounded-xl overflow-hidden">
        <div className="p-4 border-b border-[#334155] flex flex-wrap gap-3 items-center">
          <h3 className="text-white font-semibold">BOQ Items</h3>
          <div className="flex gap-2 ml-auto">
            <select
              className="bg-[#0f172a] text-[#e2e8f0] border border-[#334155] rounded-lg px-3 py-1.5 text-sm"
              value={sectionFilter}
              onChange={(e) => setSectionFilter(e.target.value)}
            >
              <option value="">All Sections</option>
              {sections.map((s: any) => (
                <option key={s.id} value={s.id}>{s.name}</option>
              ))}
            </select>
            <select
              className="bg-[#0f172a] text-[#e2e8f0] border border-[#334155] rounded-lg px-3 py-1.5 text-sm"
              value={cbsFilter}
              onChange={(e) => setCbsFilter(e.target.value)}
            >
              <option value="">All CBS</option>
              {Object.keys(cbsTotals).sort().map((c) => (
                <option key={c} value={c}>CBS-{c}</option>
              ))}
            </select>
          </div>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="text-left text-xs text-[#64748b] uppercase tracking-wider">
                <th className="p-3 pl-4">Code</th>
                <th className="p-3">Item</th>
                <th className="p-3">CBS</th>
                <th className="p-3">Unit</th>
                <th className="p-3 text-right">Qty</th>
                <th className="p-3 text-right">Unit Price</th>
                <th className="p-3 text-right">Total</th>
                <th className="p-3 pr-4">Contractor</th>
              </tr>
            </thead>
            <tbody>
              {filteredItems.map((item: any) => (
                <tr key={item.id} className="border-t border-[#1e293b] hover:bg-[#334155] text-sm">
                  <td className="p-3 pl-4 font-mono text-xs text-[#64748b]">{item.code}</td>
                  <td className="p-3 text-white">{item.name}</td>
                  <td className="p-3">
                    <span className="inline-block px-2 py-0.5 rounded text-xs font-semibold bg-[#1e3a5f] text-[#60a5fa]">
                      {item.cbs_code}
                    </span>
                  </td>
                  <td className="p-3 text-[#94a3b8]">{item.unit}</td>
                  <td className="p-3 text-right text-white">{Number(item.quantity).toLocaleString()}</td>
                  <td className="p-3 text-right text-white">${Number(item.unit_price).toLocaleString()}</td>
                  <td className="p-3 text-right text-[#22c55e]">${Number(item.total_cost).toLocaleString()}</td>
                  <td className="p-3 pr-4 text-[#94a3b8] text-xs">{item.contractor?.name || '—'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
