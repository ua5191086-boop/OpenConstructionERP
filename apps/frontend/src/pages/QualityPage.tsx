import { useState, useEffect } from 'react'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell, LineChart, Line } from 'recharts'

const TYPE_COLORS = ['#3b82f6','#22c55e','#a855f7','#f97316','#ef4444','#14b8a6','#f59e0b','#ec4899','#6366f1','#84cc16']

export default function QualityPage() {
  const [data, setData] = useState<any>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [projectFilter, setProjectFilter] = useState('')

  useEffect(() => {
    async function load() {
      try {
        const [itps, inspections, testResults, ncrs, correctiveActions, calibration, qualityMetrics] = await Promise.all([
          fetch('/api/v1/quality/itps').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/quality/inspections').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/quality/test-results').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/quality/ncrs').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/quality/corrective-actions').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/quality/calibration').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/quality/quality-metrics').then(r => r.ok ? r.json() : { data: [] }),
        ])
        setData({
          itps: itps.data || [],
          inspections: inspections.data || [],
          testResults: testResults.data || [],
          ncrs: ncrs.data || [],
          correctiveActions: correctiveActions.data || [],
          calibration: calibration.data || [],
          qualityMetrics: qualityMetrics.data || [],
        })
      } catch {
        setError('API unavailable — start the Go backend on port 8081')
      }
      setLoading(false)
    }
    load()
  }, [])

  if (loading) return <div className="p-8 text-center text-[#64748b]">Loading Quality data...</div>
  if (error) {
    return (
      <div className="p-8">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-8 text-center">
          <div className="text-4xl mb-4">✅</div>
          <h2 className="text-xl font-bold text-white mb-2">Quality Module</h2>
          <p className="text-[#94a3b8] mb-4">{error}</p>
          <p className="text-sm text-[#64748b]">Start the Go API server on port 8081 to see live data</p>
        </div>
      </div>
    )
  }

  const itps = data.itps || []
  const inspections = data.inspections || []
  const testResults = data.testResults || []
  const ncrs = data.ncrs || []
  const correctiveActions = data.correctiveActions || []
  const calibration = data.calibration || []
  const qualityMetrics = data.qualityMetrics || []

  // Filters
  const projects = [...new Set(inspections.map((i: any) => i.project_name).filter(Boolean))] as string[]
  let filteredInspections = inspections
  let filteredNCRs = ncrs
  let filteredMetrics = qualityMetrics
  if (projectFilter) {
    filteredInspections = filteredInspections.filter((i: any) => i.project_name === projectFilter)
    filteredNCRs = filteredNCRs.filter((n: any) => n.project_name === projectFilter)
    filteredMetrics = filteredMetrics.filter((m: any) => m.project_name === projectFilter)
  }

  // Stats
  const totalInspections = filteredInspections.length
  const passedIns = filteredInspections.filter((i: any) => i.result === 'pass').length
  const failedIns = filteredInspections.filter((i: any) => i.result === 'fail').length
  const openNCRs = filteredNCRs.filter((n: any) => n.status === 'open' || n.status === 'investigating').length
  const activeITPs = itps.filter((i: any) => i.status === 'active').length

  // Inspection results pie
  const insPie = [
    { name: 'Pass', value: passedIns, color: '#22c55e' },
    { name: 'Fail', value: failedIns, color: '#ef4444' },
    { name: 'Conditional', value: filteredInspections.filter((i: any) => i.result === 'conditional_pass').length, color: '#f59e0b' },
    { name: 'Pending', value: filteredInspections.filter((i: any) => i.result === 'pending').length, color: '#64748b' },
  ].filter(d => d.value > 0)

  // NCR severity pie
  const ncrSeverity = [
    { name: 'Critical', value: filteredNCRs.filter((n: any) => n.severity === 'critical').length, color: '#ef4444' },
    { name: 'Major', value: filteredNCRs.filter((n: any) => n.severity === 'major').length, color: '#f97316' },
    { name: 'Minor', value: filteredNCRs.filter((n: any) => n.severity === 'minor').length, color: '#f59e0b' },
  ].filter(d => d.value > 0)

  // Test results pass/fail
  const testPie = [
    { name: 'Pass', value: testResults.filter((t: any) => t.result === 'pass').length, color: '#22c55e' },
    { name: 'Fail', value: testResults.filter((t: any) => t.result === 'fail').length, color: '#ef4444' },
    { name: 'Conditional', value: testResults.filter((t: any) => t.result === 'conditional').length, color: '#f59e0b' },
  ].filter(d => d.value > 0)

  // Quality metrics line chart
  const metricsByMonth = filteredMetrics
    .sort((a: any, b: any) => a.report_month.localeCompare(b.report_month))
    .slice(0, 24)

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">✅ Quality Management</h1>
          <p className="text-sm text-[#94a3b8] mt-1">ITPs, Inspections, Tests, NCRs, Corrective Actions & Calibration</p>
        </div>
        <select
          className="bg-[#1e293b] border border-[#334155] rounded-lg px-3 py-2 text-sm text-white"
          value={projectFilter}
          onChange={e => setProjectFilter(e.target.value)}
        >
          <option value="">All Projects</option>
          {projects.map(p => <option key={p} value={p}>{p}</option>)}
        </select>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-xs text-[#64748b] mb-1">Active ITPs</div>
          <div className="text-2xl font-bold text-[#3b82f6]">{activeITPs}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-xs text-[#64748b] mb-1">Inspections</div>
          <div className="text-2xl font-bold text-white">{totalInspections}</div>
          <div className="text-xs text-[#22c55e]">{passedIns} passed</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-xs text-[#64748b] mb-1">Tests</div>
          <div className="text-2xl font-bold text-white">{testResults.length}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-xs text-[#64748b] mb-1">Open NCRs</div>
          <div className="text-2xl font-bold text-[#ef4444]">{openNCRs}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-xs text-[#64748b] mb-1">Corrective Actions</div>
          <div className="text-2xl font-bold text-[#a855f7]">{correctiveActions.length}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-xs text-[#64748b] mb-1">Calibration Items</div>
          <div className="text-2xl font-bold text-[#14b8a6]">{calibration.length}</div>
        </div>
      </div>

      {/* Charts Row 1 */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-sm font-semibold text-white mb-4">Inspection Results</h3>
          <ResponsiveContainer width="100%" height={240}>
            <PieChart>
              <Pie data={insPie} cx="50%" cy="50%" innerRadius={60} outerRadius={90} dataKey="value" label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}>
                {insPie.map((e, i) => <Cell key={i} fill={e.color} />)}
              </Pie>
              <Tooltip />
            </PieChart>
          </ResponsiveContainer>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-sm font-semibold text-white mb-4">NCR Severity</h3>
          <ResponsiveContainer width="100%" height={240}>
            <PieChart>
              <Pie data={ncrSeverity} cx="50%" cy="50%" innerRadius={60} outerRadius={90} dataKey="value" label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}>
                {ncrSeverity.map((e, i) => <Cell key={i} fill={e.color} />)}
              </Pie>
              <Tooltip />
            </PieChart>
          </ResponsiveContainer>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-sm font-semibold text-white mb-4">Test Results</h3>
          <ResponsiveContainer width="100%" height={240}>
            <PieChart>
              <Pie data={testPie} cx="50%" cy="50%" innerRadius={60} outerRadius={90} dataKey="value" label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}>
                {testPie.map((e, i) => <Cell key={i} fill={e.color} />)}
              </Pie>
              <Tooltip />
            </PieChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* First Pass Yield & NCR Trend */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-sm font-semibold text-white mb-4">First Pass Yield Trend</h3>
          <ResponsiveContainer width="100%" height={280}>
            <LineChart data={metricsByMonth}>
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
              <XAxis dataKey="report_month" tick={{ fill: '#94a3b8', fontSize: 11 }} />
              <YAxis domain={[80, 100]} tick={{ fill: '#94a3b8', fontSize: 11 }} />
              <Tooltip />
              <Line type="monotone" dataKey="first_pass_yield" stroke="#22c55e" strokeWidth={2} dot={{ fill: '#22c55e' }} name="FPY %" />
            </LineChart>
          </ResponsiveContainer>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-sm font-semibold text-white mb-4">NCRs Opened vs Closed (Monthly)</h3>
          <ResponsiveContainer width="100%" height={280}>
            <BarChart data={metricsByMonth}>
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
              <XAxis dataKey="report_month" tick={{ fill: '#94a3b8', fontSize: 11 }} />
              <YAxis tick={{ fill: '#94a3b8', fontSize: 11 }} />
              <Tooltip />
              <Bar dataKey="ncr_opened" fill="#ef4444" name="Opened" />
              <Bar dataKey="ncr_closed" fill="#22c55e" name="Closed" />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Recent NCRs */}
      <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
        <h3 className="text-sm font-semibold text-white mb-4">Recent Non-Conformance Reports</h3>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="text-[#64748b] border-b border-[#334155]">
                <th className="text-left py-2 px-3">Code</th>
                <th className="text-left py-2 px-3">Title</th>
                <th className="text-left py-2 px-3">Project</th>
                <th className="text-left py-2 px-3">Severity</th>
                <th className="text-left py-2 px-3">Status</th>
                <th className="text-left py-2 px-3">Category</th>
              </tr>
            </thead>
            <tbody>
              {filteredNCRs.slice(0, 10).map((ncr: any) => (
                <tr key={ncr.id} className="border-b border-[#1e293b] hover:bg-[#334155]/40">
                  <td className="py-2 px-3 text-white font-mono text-xs">{ncr.ncr_code}</td>
                  <td className="py-2 px-3 text-white">{ncr.title}</td>
                  <td className="py-2 px-3 text-[#94a3b8]">{ncr.project_name}</td>
                  <td className="py-2 px-3">
                    <span className={`px-2 py-0.5 rounded-full text-xs ${
                      ncr.severity === 'critical' ? 'bg-red-500/20 text-red-400' :
                      ncr.severity === 'major' ? 'bg-orange-500/20 text-orange-400' :
                      'bg-yellow-500/20 text-yellow-400'
                    }`}>{ncr.severity}</span>
                  </td>
                  <td className="py-2 px-3 text-[#94a3b8]">{ncr.status}</td>
                  <td className="py-2 px-3 text-[#94a3b8]">{ncr.ncr_category}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* ITPs & Calibration */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-sm font-semibold text-white mb-4">Inspection & Test Plans</h3>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-[#64748b] border-b border-[#334155]">
                  <th className="text-left py-2 px-3">Code</th>
                  <th className="text-left py-2 px-3">Name</th>
                  <th className="text-left py-2 px-3">Type</th>
                  <th className="text-left py-2 px-3">Status</th>
                </tr>
              </thead>
              <tbody>
                {itps.slice(0, 8).map((itp: any) => (
                  <tr key={itp.id} className="border-b border-[#1e293b] hover:bg-[#334155]/40">
                    <td className="py-2 px-3 text-white font-mono text-xs">{itp.itp_code}</td>
                    <td className="py-2 px-3 text-white">{itp.itp_name}</td>
                    <td className="py-2 px-3 text-[#94a3b8]">{itp.itp_type}</td>
                    <td className="py-2 px-3">
                      <span className={`px-2 py-0.5 rounded-full text-xs ${
                        itp.status === 'active' ? 'bg-green-500/20 text-green-400' :
                        itp.status === 'completed' ? 'bg-blue-500/20 text-blue-400' :
                        'bg-yellow-500/20 text-yellow-400'
                      }`}>{itp.status}</span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-sm font-semibold text-white mb-4">Calibration Status</h3>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-[#64748b] border-b border-[#334155]">
                  <th className="text-left py-2 px-3">Equipment</th>
                  <th className="text-left py-2 px-3">Model</th>
                  <th className="text-left py-2 px-3">Last Cal</th>
                  <th className="text-left py-2 px-3">Next Cal</th>
                  <th className="text-left py-2 px-3">Result</th>
                </tr>
              </thead>
              <tbody>
                {calibration.slice(0, 8).map((cal: any) => (
                  <tr key={cal.id} className="border-b border-[#1e293b] hover:bg-[#334155]/40">
                    <td className="py-2 px-3 text-white">{cal.equipment_name}</td>
                    <td className="py-2 px-3 text-[#94a3b8]">{cal.equipment_model}</td>
                    <td className="py-2 px-3 text-[#94a3b8] text-xs">{cal.last_calibration_date}</td>
                    <td className="py-2 px-3 text-[#94a3b8] text-xs">{cal.next_calibration_date}</td>
                    <td className="py-2 px-3">
                      <span className={`px-2 py-0.5 rounded-full text-xs ${
                        cal.calibration_result === 'pass' ? 'bg-green-500/20 text-green-400' :
                        cal.calibration_result === 'conditional' ? 'bg-yellow-500/20 text-yellow-400' :
                        'bg-red-500/20 text-red-400'
                      }`}>{cal.calibration_result}</span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  )
}