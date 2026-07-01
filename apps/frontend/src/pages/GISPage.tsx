import { useState, useEffect } from 'react'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell, LineChart, Line } from 'recharts'

const TYPE_COLORS = ['#3b82f6','#22c55e','#a855f7','#f97316','#ef4444','#14b8a6','#f59e0b','#ec4899','#6366f1','#84cc16']

export default function GISPage() {
  const [data, setData] = useState<any>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [projectFilter, setProjectFilter] = useState('')

  useEffect(() => {
    async function load() {
      try {
        const [layers, features, surveyPoints, surveyRuns, stations, alignments, crossSections, droneFlights] = await Promise.all([
          fetch('/api/v1/gis/layers').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/gis/features').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/gis/survey-points').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/gis/survey-runs').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/gis/survey-stations').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/gis/alignments').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/gis/cross-sections').then(r => r.ok ? r.json() : { data: [] }),
          fetch('/api/v1/gis/drone-flights').then(r => r.ok ? r.json() : { data: [] }),
        ])
        setData({
          layers: layers.data || [],
          features: features.data || [],
          surveyPoints: surveyPoints.data || [],
          surveyRuns: surveyRuns.data || [],
          stations: stations.data || [],
          alignments: alignments.data || [],
          crossSections: crossSections.data || [],
          droneFlights: droneFlights.data || [],
        })
      } catch {
        setError('API unavailable — start the Go backend on port 8081')
      }
      setLoading(false)
    }
    load()
  }, [])

  if (loading) return <div className="p-8 text-center text-[#64748b]">Loading GIS data...</div>
  if (error) {
    return (
      <div className="p-8">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-8 text-center">
          <div className="text-4xl mb-4">🗺️</div>
          <h2 className="text-xl font-bold text-white mb-2">GIS & Survey Module</h2>
          <p className="text-[#94a3b8] mb-4">{error}</p>
          <p className="text-sm text-[#64748b]">Start the Go API server on port 8081 to see live data</p>
        </div>
      </div>
    )
  }

  const layers = data.layers || []
  const features = data.features || []
  const surveyPoints = data.surveyPoints || []
  const surveyRuns = data.surveyRuns || []
  const stations = data.stations || []
  const alignments = data.alignments || []
  const crossSections = data.crossSections || []
  const droneFlights = data.droneFlights || []

  // Filters
  const projects = [...new Set(layers.map((l: any) => l.project_name).filter(Boolean))] as string[]
  let filteredLayers = layers
  let filteredSurveyRuns = surveyRuns
  let filteredDroneFlights = droneFlights
  if (projectFilter) {
    filteredLayers = filteredLayers.filter((l: any) => l.project_name === projectFilter)
    filteredSurveyRuns = filteredSurveyRuns.filter((r: any) => r.project_name === projectFilter)
    filteredDroneFlights = filteredDroneFlights.filter((d: any) => d.project_name === projectFilter)
  }

  // Stats
  const vectorLayers = filteredLayers.filter((l: any) => l.layer_type === 'vector').length
  const rasterLayers = filteredLayers.filter((l: any) => l.layer_type === 'raster').length
  const completedRuns = filteredSurveyRuns.filter((r: any) => r.status === 'completed' || r.status === 'reviewed' || r.status === 'approved').length
  const completedFlights = filteredDroneFlights.filter((d: any) => d.status === 'completed').length

  // Layer type breakdown
  const layerTypeData = [
    { name: 'Vector', value: vectorLayers, color: '#3b82f6' },
    { name: 'Raster', value: rasterLayers, color: '#22c55e' },
  ].filter(d => d.value > 0)

  // Survey run status
  const runStatusData = [
    { name: 'Completed/Approved', value: completedRuns, color: '#22c55e' },
    { name: 'Other', value: filteredSurveyRuns.length - completedRuns, color: '#f59e0b' },
  ].filter(d => d.value > 0)

  // Drone flight status
  const flightStatusData = [
    { name: 'Completed', value: completedFlights, color: '#22c55e' },
    { name: 'Other', value: filteredDroneFlights.length - completedFlights, color: '#64748b' },
  ].filter(d => d.value > 0)

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">🗺️ GIS & Survey</h1>
          <p className="text-sm text-[#94a3b8] mt-1">Layers, Features, Survey Points, Runs, Stations, Alignments, Cross Sections & Drone Flights</p>
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
      <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-8 gap-4">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-xs text-[#64748b] mb-1">Layers</div>
          <div className="text-2xl font-bold text-[#3b82f6]">{layers.length}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-xs text-[#64748b] mb-1">Features</div>
          <div className="text-2xl font-bold text-white">{features.length}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-xs text-[#64748b] mb-1">Survey Points</div>
          <div className="text-2xl font-bold text-[#a855f7]">{surveyPoints.length}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-xs text-[#64748b] mb-1">Survey Runs</div>
          <div className="text-2xl font-bold text-[#f97316]">{surveyRuns.length}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-xs text-[#64748b] mb-1">Stations</div>
          <div className="text-2xl font-bold text-[#14b8a6]">{stations.length}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-xs text-[#64748b] mb-1">Alignments</div>
          <div className="text-2xl font-bold text-[#f59e0b]">{alignments.length}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-xs text-[#64748b] mb-1">Cross Sections</div>
          <div className="text-2xl font-bold text-[#ec4899]">{crossSections.length}</div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <div className="text-xs text-[#64748b] mb-1">Drone Flights</div>
          <div className="text-2xl font-bold text-[#6366f1]">{droneFlights.length}</div>
        </div>
      </div>

      {/* Charts Row */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-sm font-semibold text-white mb-4">Layer Type Breakdown</h3>
          <ResponsiveContainer width="100%" height={240}>
            <PieChart>
              <Pie data={layerTypeData} cx="50%" cy="50%" innerRadius={60} outerRadius={90} dataKey="value" label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}>
                {layerTypeData.map((e, i) => <Cell key={i} fill={e.color} />)}
              </Pie>
              <Tooltip />
            </PieChart>
          </ResponsiveContainer>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-sm font-semibold text-white mb-4">Survey Run Status</h3>
          <ResponsiveContainer width="100%" height={240}>
            <PieChart>
              <Pie data={runStatusData} cx="50%" cy="50%" innerRadius={60} outerRadius={90} dataKey="value" label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}>
                {runStatusData.map((e, i) => <Cell key={i} fill={e.color} />)}
              </Pie>
              <Tooltip />
            </PieChart>
          </ResponsiveContainer>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-sm font-semibold text-white mb-4">Drone Flight Status</h3>
          <ResponsiveContainer width="100%" height={240}>
            <PieChart>
              <Pie data={flightStatusData} cx="50%" cy="50%" innerRadius={60} outerRadius={90} dataKey="value" label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}>
                {flightStatusData.map((e, i) => <Cell key={i} fill={e.color} />)}
              </Pie>
              <Tooltip />
            </PieChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Survey Points & Alignments */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-sm font-semibold text-white mb-4">Survey Points by Type</h3>
          <ResponsiveContainer width="100%" height={280}>
            <BarChart data={(() => {
              const types: Record<string, number> = {}
              surveyPoints.forEach((sp: any) => { types[sp.point_type] = (types[sp.point_type] || 0) + 1 })
              return Object.entries(types).map(([k, v]) => ({ name: k, count: v }))
            })()}>
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
              <XAxis dataKey="name" tick={{ fill: '#94a3b8', fontSize: 11 }} />
              <YAxis tick={{ fill: '#94a3b8', fontSize: 11 }} />
              <Tooltip />
              <Bar dataKey="count" fill="#a855f7" radius={[4,4,0,0]} />
            </BarChart>
          </ResponsiveContainer>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-sm font-semibold text-white mb-4">Alignment Lengths</h3>
          <ResponsiveContainer width="100%" height={280}>
            <BarChart data={alignments}>
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
              <XAxis dataKey="alignment_code" tick={{ fill: '#94a3b8', fontSize: 11 }} />
              <YAxis tick={{ fill: '#94a3b8', fontSize: 11 }} />
              <Tooltip />
              <Bar dataKey="total_length" fill="#f59e0b" radius={[4,4,0,0]} name="Length (m)" />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Drone Flights */}
      <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
        <h3 className="text-sm font-semibold text-white mb-4">Recent Drone Flights</h3>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="text-[#64748b] border-b border-[#334155]">
                <th className="text-left py-2 px-3">Code</th>
                <th className="text-left py-2 px-3">Name</th>
                <th className="text-left py-2 px-3">Project</th>
                <th className="text-left py-2 px-3">Drone</th>
                <th className="text-left py-2 px-3">Date</th>
                <th className="text-left py-2 px-3">Duration</th>
                <th className="text-left py-2 px-3">Area (ha)</th>
                <th className="text-left py-2 px-3">Status</th>
              </tr>
            </thead>
            <tbody>
              {droneFlights.slice(0, 10).map((df: any) => (
                <tr key={df.id} className="border-b border-[#1e293b] hover:bg-[#334155]/40">
                  <td className="py-2 px-3 text-white font-mono text-xs">{df.flight_code}</td>
                  <td className="py-2 px-3 text-white">{df.flight_name}</td>
                  <td className="py-2 px-3 text-[#94a3b8]">{df.project_name}</td>
                  <td className="py-2 px-3 text-[#94a3b8]">{df.drone_model}</td>
                  <td className="py-2 px-3 text-[#94a3b8] text-xs">{df.flight_date}</td>
                  <td className="py-2 px-3 text-[#94a3b8]">{df.flight_duration_minutes}m</td>
                  <td className="py-2 px-3 text-[#94a3b8]">{df.area_covered_ha}</td>
                  <td className="py-2 px-3">
                    <span className={`px-2 py-0.5 rounded-full text-xs ${
                      df.status === 'completed' ? 'bg-green-500/20 text-green-400' :
                      df.status === 'processing' ? 'bg-blue-500/20 text-blue-400' :
                      df.status === 'planned' ? 'bg-yellow-500/20 text-yellow-400' :
                      'bg-red-500/20 text-red-400'
                    }`}>{df.status}</span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Layers & Survey Runs */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-sm font-semibold text-white mb-4">Layers</h3>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-[#64748b] border-b border-[#334155]">
                  <th className="text-left py-2 px-3">Layer Name</th>
                  <th className="text-left py-2 px-3">Type</th>
                  <th className="text-left py-2 px-3">Geometry</th>
                  <th className="text-left py-2 px-3">Source</th>
                </tr>
              </thead>
              <tbody>
                {filteredLayers.slice(0, 8).map((lyr: any) => (
                  <tr key={lyr.id} className="border-b border-[#1e293b] hover:bg-[#334155]/40">
                    <td className="py-2 px-3 text-white">{lyr.layer_name}</td>
                    <td className="py-2 px-3 text-[#94a3b8]">{lyr.layer_type}</td>
                    <td className="py-2 px-3 text-[#94a3b8]">{lyr.geometry_type}</td>
                    <td className="py-2 px-3 text-[#94a3b8]">{lyr.source_type}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
          <h3 className="text-sm font-semibold text-white mb-4">Survey Runs</h3>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-[#64748b] border-b border-[#334155]">
                  <th className="text-left py-2 px-3">Code</th>
                  <th className="text-left py-2 px-3">Name</th>
                  <th className="text-left py-2 px-3">Type</th>
                  <th className="text-left py-2 px-3">Points</th>
                  <th className="text-left py-2 px-3">Status</th>
                </tr>
              </thead>
              <tbody>
                {filteredSurveyRuns.slice(0, 8).map((sr: any) => (
                  <tr key={sr.id} className="border-b border-[#1e293b] hover:bg-[#334155]/40">
                    <td className="py-2 px-3 text-white font-mono text-xs">{sr.run_code}</td>
                    <td className="py-2 px-3 text-white">{sr.run_name}</td>
                    <td className="py-2 px-3 text-[#94a3b8]">{sr.survey_type}</td>
                    <td className="py-2 px-3 text-[#94a3b8]">{sr.point_count}</td>
                    <td className="py-2 px-3">
                      <span className={`px-2 py-0.5 rounded-full text-xs ${
                        sr.status === 'approved' || sr.status === 'completed' ? 'bg-green-500/20 text-green-400' :
                        sr.status === 'reviewed' ? 'bg-blue-500/20 text-blue-400' :
                        'bg-yellow-500/20 text-yellow-400'
                      }`}>{sr.status}</span>
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