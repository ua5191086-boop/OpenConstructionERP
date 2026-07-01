import { useState, useEffect } from 'react'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts'

const STATUS_COLORS: Record<string, string> = {
  open: '#34d399', answered: '#60a5fa', closed: '#38bdf8', void: '#a78bfa', overdue: '#f87171',
  draft: '#fbbf24', submitted: '#60a5fa', under_review: '#f97316', reviewed: '#a855f7',
  approved: '#34d399', approved_with_comments: '#fde047', rejected: '#f87171', resubmit: '#fb923c',
  sent: '#60a5fa', received: '#34d399', acknowledged: '#a78bfa', archived: '#64748b',
  prepared: '#fbbf24', distributed: '#60a5fa', investigating: '#fbbf24', action_planned: '#f97316',
  action_taken: '#60a5fa', verified: '#a78bfa',
}

const TYPE_COLORS = ['#3b82f6','#22c55e','#a855f7','#f97316','#ef4444','#14b8a6','#f59e0b','#ec4899']

const TABS = [
  { id: 'rfi', label: 'RFI', endpoint: 'rfis', statusField: 'status', typeField: 'discipline' },
  { id: 'ncr', label: 'NCR', endpoint: 'ncrs', statusField: 'status', typeField: 'ncr_type' },
  { id: 'submittals', label: 'Submittals', endpoint: 'submittals', statusField: 'status', typeField: 'submittal_type' },
  { id: 'ms', label: 'Method Stmts', endpoint: 'method-statements', statusField: 'status', typeField: null },
  { id: 'sd', label: 'Shop Drawings', endpoint: 'shop-drawings', statusField: 'status', typeField: 'discipline' },
  { id: 'corr', label: 'Correspondence', endpoint: 'correspondence', statusField: 'status', typeField: 'corr_type' },
  { id: 'mom', label: 'Minutes', endpoint: 'minutes-of-meeting', statusField: 'status', typeField: 'meeting_type' },
  { id: 'dr', label: 'Daily Reports', endpoint: 'daily-reports', statusField: 'status', typeField: null },
  { id: 'dt', label: 'Transmittals', endpoint: 'transmittals', statusField: 'status', typeField: 'purpose' },
]

interface SummaryItem {
  project_id: string
  total_rfi: number; open_rfi: number
  total_ncr: number; open_ncr: number; critical_ncr: number
  total_submittals: number; pending_submittals: number; rejected_submittals: number
  total_method_statements: number; pending_method_statements: number
  total_shop_drawings: number; pending_shop_drawings: number
  total_correspondence: number; active_correspondence: number
  total_minutes_of_meeting: number; recent_daily_reports: number
  total_transmittals: number; pending_transmittals: number
}

export default function DocControlPage() {
  const [summary, setSummary] = useState<SummaryItem[]>([])
  const [data, setData] = useState<Record<string, any[]>>({})
  const [activeTab, setActiveTab] = useState('rfi')
  const [statusFilter, setStatusFilter] = useState('')
  const [typeFilter, setTypeFilter] = useState('')
  const [searchFilter, setSearchFilter] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    async function load() {
      try {
        const [summaryRes, ...tabRes] = await Promise.all([
          fetch('/api/v1/doc-control/summary').then(r => r.ok ? r.json() : { data: [] }),
          ...TABS.map(t => fetch(`/api/v1/doc-control/${t.endpoint}`).then(r => r.ok ? r.json() : { data: [] })),
        ])
        const summaryData: SummaryItem[] = summaryRes.data || []
        const tabData: Record<string, any[]> = {}
        TABS.forEach((t, i) => { tabData[t.id] = tabRes[i].data || [] })
        setSummary(summaryData)
        setData(tabData)
      } catch {
        setError('API unavailable — start the Go backend on port 8081')
      }
      setLoading(false)
    }
    load()
  }, [])

  if (loading) return <div className="p-8 text-center text-[#64748b]">Loading Document Control data...</div>
  if (error) return (
    <div className="p-8">
      <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-8 text-center">
        <div className="text-4xl mb-4">📄</div>
        <h2 className="text-xl font-bold text-white mb-2">Document Control Module</h2>
        <p className="text-[#94a3b8] mb-4">{error}</p>
        <p className="text-sm text-[#64748b]">Start the Go API server on port 8081 to see live data</p>
      </div>
    </div>
  )

  const totalSummary: SummaryItem = summary.length > 0 ? summary.reduce((a, b) => ({
    ...a, project_id: 'all',
    total_rfi: a.total_rfi + b.total_rfi, open_rfi: a.open_rfi + b.open_rfi,
    total_ncr: a.total_ncr + b.total_ncr, open_ncr: a.open_ncr + b.open_ncr, critical_ncr: a.critical_ncr + b.critical_ncr,
    total_submittals: a.total_submittals + b.total_submittals, pending_submittals: a.pending_submittals + b.pending_submittals,
    rejected_submittals: a.rejected_submittals + b.rejected_submittals,
    total_method_statements: a.total_method_statements + b.total_method_statements,
    pending_method_statements: a.pending_method_statements + b.pending_method_statements,
    total_shop_drawings: a.total_shop_drawings + b.total_shop_drawings,
    pending_shop_drawings: a.pending_shop_drawings + b.pending_shop_drawings,
    total_correspondence: a.total_correspondence + b.total_correspondence,
    active_correspondence: a.active_correspondence + b.active_correspondence,
    total_minutes_of_meeting: a.total_minutes_of_meeting + b.total_minutes_of_meeting,
    recent_daily_reports: a.recent_daily_reports + b.recent_daily_reports,
    total_transmittals: a.total_transmittals + b.total_transmittals,
    pending_transmittals: a.pending_transmittals + b.pending_transmittals,
  })) : {} as SummaryItem

  const total = (totalSummary as any).total_rfi || 0

  const cfg = TABS.find(t => t.id === activeTab)!
  let items = (data[activeTab] || []).filter((i: any) => {
    const s = i.status || ''
    if (statusFilter && statusFilter === 'open_only') return ['open', 'overdue'].includes(s)
    if (statusFilter && s !== statusFilter) return false
    if (typeFilter && cfg.typeField) {
      const tv = i[cfg.typeField] || ''
      if (tv !== typeFilter) return false
    }
    if (searchFilter) {
      const search = searchFilter.toLowerCase()
      const code = i.rfi_code || i.ncr_code || i.submittal_code || i.ms_code || i.drawing_code || i.corr_code || i.mom_code || i.transmittal_code || ''
      const subject = i.subject || i.title || ''
      if (!code.toLowerCase().includes(search) && !subject.toLowerCase().includes(search)) return false
    }
    return true
  })

  const statuses = [...new Set((data[activeTab] || []).map((i: any) => i.status).filter(Boolean))] as string[]
  const types = cfg.typeField ? [...new Set((data[activeTab] || []).map((i: any) => i[cfg.typeField!]).filter(Boolean))] as string[] : []

  // Chart data
  const statusCounts: Record<string, number> = {}
  items.forEach((i: any) => { const s = i.status || 'unknown'; statusCounts[s] = (statusCounts[s] || 0) + 1 })
  const pieData = Object.entries(statusCounts).map(([name, value]) => ({ name, value }))

  const typeCounts: Record<string, number> = {}
  if (cfg.typeField) {
    items.forEach((i: any) => { const t = i[cfg.typeField!] || 'other'; typeCounts[t] = (typeCounts[t] || 0) + 1 })
  }
  const typeData = Object.entries(typeCounts).map(([name, value]) => ({ name, value }))

  const monthly: Record<string, number> = {}
  items.forEach((i: any) => {
    const d = i.raised_at || i.submitted_at || i.reported_at || i.sent_at || i.meeting_date || i.report_date || ''
    if (d) { const m = d.substring(0, 7); monthly[m] = (monthly[m] || 0) + 1 }
  })
  const monthData = Object.entries(monthly).sort(([a], [b]) => a.localeCompare(b)).map(([name, value]) => ({ name, value }))

  function badge(value: string, cls = 'bg-[#334155] text-[#94a3b8]') {
    if (!value) return <span className={`px-2 py-0.5 rounded text-xs font-medium ${cls}`}>—</span>
    const color = STATUS_COLORS[value] || '#94a3b8'
    return <span className="px-2 py-0.5 rounded text-xs font-medium" style={{ background: `${color}20`, color }}>{value}</span>
  }

  function renderTable() {
    const rows: any[] = items
    const getCode = (i: any) =>
      i.rfi_code || i.ncr_code || i.submittal_code || i.ms_code || i.drawing_code || i.corr_code || i.mom_code || i.transmittal_code || '—'
    const getSubject = (i: any) => i.subject || i.title || i.meeting_title || '—'
    
    return (
      <div className="bg-[#1e293b] border border-[#334155] rounded-xl overflow-hidden">
        <div className="p-4 border-b border-[#334155] flex justify-between items-center">
          <h3 className="text-white font-semibold">{cfg.label} ({items.length})</h3>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="text-[#64748b] text-xs uppercase tracking-wider">
                <th className="text-left p-3 border-b border-[#334155]">Code</th>
                <th className="text-left p-3 border-b border-[#334155]">Subject</th>
                {cfg.statusField && <th className="text-left p-3 border-b border-[#334155]">Status</th>}
                {cfg.typeField && <th className="text-left p-3 border-b border-[#334155]">{cfg.typeField.replace('_',' ')}</th>}
                <th className="text-left p-3 border-b border-[#334155]">Project</th>
              </tr>
            </thead>
            <tbody>
              {rows.slice(0, 50).map((i: any) => (
                <tr key={i.id} className="hover:bg-[#334155] border-b border-[#1e293b]">
                  <td className="p-3 font-mono text-xs text-[#64748b]">{getCode(i)}</td>
                  <td className="p-3 text-sm text-[#e2e8f0] max-w-[250px] truncate">{getSubject(i)}</td>
                  {cfg.statusField && <td className="p-3">{badge(i.status)}</td>}
                  {cfg.typeField && <td className="p-3 text-sm text-[#94a3b8]">{i[cfg.typeField!] || '—'}</td>}
                  <td className="p-3 text-xs text-[#64748b]">{i.project_name || i.project_id?.substring(0,8) || '—'}</td>
                </tr>
              ))}
              {rows.length === 0 && (
                <tr><td colSpan={5} className="p-8 text-center text-[#64748b]">No documents found</td></tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    )
  }

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-white">📄 Document Control</h1>
        <p className="text-[#94a3b8] mt-1">Document Management — RFI, NCR, Submittals, Method Statements, Shop Drawings, Correspondence, MOM, Daily Reports, Transmittals</p>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 mb-6 bg-[#1e293b] border border-[#334155] rounded-lg p-1 overflow-x-auto">
        {TABS.map(t => (
          <button key={t.id}
            className={`px-3 py-1.5 rounded-md text-sm whitespace-nowrap transition-colors ${
              activeTab === t.id ? 'bg-[#3b82f6] text-white' : 'text-[#94a3b8] hover:text-white hover:bg-[#334155]'
            }`}
            onClick={() => setActiveTab(t.id)}
          >
            {t.label} ({(data[t.id] || []).length})
          </button>
        ))}
      </div>

      {/* Filters */}
      <div className="flex flex-wrap gap-3 mb-6">
        <select className="bg-[#0f172a] text-[#e2e8f0] border border-[#334155] rounded-lg px-3 py-1.5 text-sm"
          value={statusFilter} onChange={e => setStatusFilter(e.target.value)}>
          <option value="">All Statuses</option>
          <option value="open_only">Open / Overdue</option>
          {statuses.map(s => <option key={s} value={s}>{s}</option>)}
        </select>
        {cfg.typeField && (
          <select className="bg-[#0f172a] text-[#e2e8f0] border border-[#334155] rounded-lg px-3 py-1.5 text-sm"
            value={typeFilter} onChange={e => setTypeFilter(e.target.value)}>
            <option value="">All Types</option>
            {types.map(t => <option key={t} value={t}>{t}</option>)}
          </select>
        )}
        <input type="text" placeholder="Search code or title..."
          className="bg-[#0f172a] text-[#e2e8f0] border border-[#334155] rounded-lg px-3 py-1.5 text-sm flex-1 min-w-[200px]"
          value={searchFilter} onChange={e => setSearchFilter(e.target.value)} />
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 md:grid-cols-5 gap-4 mb-6">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-[#94a3b8] text-xs uppercase tracking-wider">Total Docs</div>
          <div className="text-2xl font-bold text-[#3b82f6] mt-1">{items.length}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-[#94a3b8] text-xs uppercase tracking-wider">Open RFIs</div>
          <div className="text-2xl font-bold text-[#f97316] mt-1">{summary.reduce((s, p) => s + p.open_rfi, 0)}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-[#94a3b8] text-xs uppercase tracking-wider">Open NCRs</div>
          <div className="text-2xl font-bold text-[#ef4444] mt-1">{summary.reduce((s, p) => s + p.open_ncr, 0)}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-[#94a3b8] text-xs uppercase tracking-wider">Critical NCRs</div>
          <div className="text-2xl font-bold text-[#ef4444] mt-1">{summary.reduce((s, p) => s + p.critical_ncr, 0)}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-[#94a3b8] text-xs uppercase tracking-wider">Pending Submittals</div>
          <div className="text-2xl font-bold text-[#eab308] mt-1">{summary.reduce((s, p) => s + p.pending_submittals, 0)}</div>
        </div>
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-6">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-white font-semibold mb-4 text-sm">Status Distribution</h3>
          <ResponsiveContainer width="100%" height={260}>
            <PieChart>
              <Pie data={pieData} cx="50%" cy="50%" outerRadius={90} dataKey="value"
                label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}>
                {pieData.map((_, i) => <Cell key={i} fill={TYPE_COLORS[i % TYPE_COLORS.length]} />)}
              </Pie>
              <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '8px', color: '#e2e8f0' }} />
            </PieChart>
          </ResponsiveContainer>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-white font-semibold mb-4 text-sm">Monthly Trend</h3>
          <ResponsiveContainer width="100%" height={260}>
            <BarChart data={monthData}>
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
              <XAxis dataKey="name" tick={{ fill: '#94a3b8', fontSize: 11 }} />
              <YAxis tick={{ fill: '#94a3b8', fontSize: 11 }} />
              <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '8px', color: '#e2e8f0' }} />
              <Bar dataKey="value" radius={[4, 4, 0, 0]} fill="#3b82f6" />
            </BarChart>
          </ResponsiveContainer>
        </div>
        {cfg.typeField && (
          <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
            <h3 className="text-white font-semibold mb-4 text-sm">By {cfg.typeField.replace('_',' ')}</h3>
            <ResponsiveContainer width="100%" height={260}>
              <BarChart data={typeData} layout="vertical">
                <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
                <XAxis type="number" tick={{ fill: '#94a3b8', fontSize: 11 }} />
                <YAxis dataKey="name" type="category" tick={{ fill: '#94a3b8', fontSize: 11 }} width={100} />
                <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '8px', color: '#e2e8f0' }} />
                <Bar dataKey="value" radius={[0, 4, 4, 0]}>
                  {typeData.map((_, i) => <Cell key={i} fill={TYPE_COLORS[i % TYPE_COLORS.length]} />)}
                </Bar>
              </BarChart>
            </ResponsiveContainer>
          </div>
        )}
      </div>

      {/* Table */}
      {renderTable()}
    </div>
  )
}