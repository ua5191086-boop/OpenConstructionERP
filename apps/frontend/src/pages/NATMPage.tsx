import { useState, useEffect } from 'react'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, ScatterChart, Scatter, LineChart, Line } from 'recharts'

export default function NATMPage() {
  const [data, setData] = useState<any>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    async function load() {
      try {
        const [excavation, shotcrete, rockBolts, convergence, faceMapping, mtbmThrust, grouting, settlement, summary] = await Promise.all([
          fetch('/api/v1/natm/excavation').then(r => r.ok ? r.json() : []),
          fetch('/api/v1/natm/shotcrete').then(r => r.ok ? r.json() : []),
          fetch('/api/v1/natm/rock-bolts').then(r => r.ok ? r.json() : []),
          fetch('/api/v1/natm/convergence').then(r => r.ok ? r.json() : []),
          fetch('/api/v1/natm/face-mapping').then(r => r.ok ? r.json() : []),
          fetch('/api/v1/natm/mtbm-thrust').then(r => r.ok ? r.json() : []),
          fetch('/api/v1/natm/grouting').then(r => r.ok ? r.json() : []),
          fetch('/api/v1/natm/settlement').then(r => r.ok ? r.json() : []),
          fetch('/api/v1/natm/summary').then(r => r.ok ? r.json() : {}),
        ])
        setData({ excavation, shotcrete, rockBolts, convergence, faceMapping, mtbmThrust, grouting, settlement, summary })
      } catch {
        setError('API unavailable — start the Go backend on port 8081')
      }
      setLoading(false)
    }
    load()
  }, [])

  if (loading) return <div className="p-8 text-center text-[#64748b]">Loading NATM data...</div>
  if (error) {
    return (
      <div className="p-8">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-8 text-center">
          <div className="text-4xl mb-4">⛰️</div>
          <h2 className="text-xl font-bold text-white mb-2">NATM & Microtunnelling Module</h2>
          <p className="text-[#94a3b8] mb-4">{error}</p>
        </div>
      </div>
    )
  }

  const exc = data.excavation || []
  const shot = data.shotcrete || []
  const bolts = data.rockBolts || []
  const face = data.faceMapping || []
  const thrust = data.mtbmThrust || []
  const settle = data.settlement || []
  const sm = data.summary || {}

  // Excavation progress
  const sortedExc = [...exc].sort((a: any, b: any) => a.round_no - b.round_no)

  // MTBM thrust sorted by pipe
  const sortedThrust = [...thrust].sort((a: any, b: any) => a.pipe_no - b.pipe_no)

  // Shotcrete strength
  const shotStr = shot.filter((s: any) => s.compressive_strength_mpa).slice(0, 50)

  return (
    <div className="p-6 space-y-6">
      <h1 className="text-2xl font-bold text-white">⛰️ NATM & Microtunnelling</h1>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4 text-center">
          <div className="text-xs text-[#94a3b8] uppercase">Excavation Rounds</div><div className="text-2xl font-bold mt-2 text-blue-400">{exc.length}</div></div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4 text-center">
          <div className="text-xs text-[#94a3b8] uppercase">Shotcrete</div><div className="text-2xl font-bold mt-2 text-green-400">{shot.length}</div></div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4 text-center">
          <div className="text-xs text-[#94a3b8] uppercase">Rock Bolts</div><div className="text-2xl font-bold mt-2 text-yellow-400">{bolts.length}</div></div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4 text-center">
          <div className="text-xs text-[#94a3b8] uppercase">Settlement Alarms</div><div className="text-2xl font-bold mt-2 text-red-400">{settle.filter((s: any) => s.alarm_triggered).length}</div></div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <h3 className="text-white font-medium mb-4">📐 Excavation Progress (Chainage)</h3>
          <ResponsiveContainer width="100%" height={250}>
            <LineChart data={sortedExc.map((e: any) => ({ round: `R${e.round_no}`, from: e.chainage_from || 0, to: e.chainage_to || 0 }))}>
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
              <XAxis dataKey="round" stroke="#64748b" tick={false} />
              <YAxis stroke="#64748b" />
              <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155' }} />
              <Line type="monotone" dataKey="from" stroke="#3b82f6" dot={false} name="From" />
              <Line type="monotone" dataKey="to" stroke="#22c55e" dot={false} name="To" />
            </LineChart>
          </ResponsiveContainer>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <h3 className="text-white font-medium mb-4">🔫 MTBM Thrust Force</h3>
          <ResponsiveContainer width="100%" height={250}>
            <BarChart data={sortedThrust.map((t: any) => ({ pipe: `P${t.pipe_no}`, thrust: t.thrust_force_kN || 0 }))}>
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
              <XAxis dataKey="pipe" stroke="#64748b" tick={false} />
              <YAxis stroke="#64748b" />
              <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155' }} />
              <Bar dataKey="thrust" fill="#f97316" name="Thrust (kN)" />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <h3 className="text-white font-medium mb-4">🧱 Shotcrete Strength</h3>
          <ResponsiveContainer width="100%" height={200}>
            <BarChart data={shotStr.map((s: any, i: number) => ({ i, strength: s.compressive_strength_mpa }))}>
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
              <XAxis dataKey="i" stroke="#64748b" tick={false} />
              <YAxis stroke="#64748b" />
              <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155' }} />
              <Bar dataKey="strength" fill="#22c55e" name="MPa" />
            </BarChart>
          </ResponsiveContainer>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <h3 className="text-white font-medium mb-4">🪨 Rock Mass Rating (RMR)</h3>
          <ResponsiveContainer width="100%" height={200}>
            <BarChart data={face.filter((f: any) => f.rmr_score).slice(0, 40).map((f: any, i: number) => ({ i, rmr: f.rmr_score }))}>
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
              <XAxis dataKey="i" stroke="#64748b" tick={false} />
              <YAxis stroke="#64748b" />
              <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155' }} />
              <Bar dataKey="rmr" fill="#a855f7" name="RMR" />
            </BarChart>
          </ResponsiveContainer>
        </div>
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-4">
          <h3 className="text-white font-medium mb-4">📉 Settlement Monitoring</h3>
          <ResponsiveContainer width="100%" height={200}>
            <ScatterChart>
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
              <XAxis dataKey="chainage" stroke="#64748b" name="Chainage" />
              <YAxis dataKey="settlement_mm" stroke="#64748b" name="mm" />
              <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155' }} />
              <Scatter data={settle.slice(0, 100).map((s: any) => ({ chainage: s.chainage || 0, settlement_mm: s.settlement_mm || 0 }))} fill="#3b82f6" />
            </ScatterChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Face Mapping Table */}
      <div className="bg-[#1e293b] border border-[#334155] rounded-xl overflow-hidden">
        <div className="p-4 border-b border-[#334155]"><h3 className="text-white font-medium">🪨 Face Mapping Log</h3></div>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead><tr className="text-xs text-[#64748b] uppercase"><th className="p-3 text-left">Chainage</th><th className="p-3 text-left">Rock Type</th><th className="p-3 text-left">RMR</th><th className="p-3 text-left">Q-Score</th><th className="p-3 text-left">Fault</th><th className="p-3 text-left">Mapped By</th></tr></thead>
            <tbody>
              {face.slice(0, 30).map((f: any) => (
                <tr key={f.id} className="border-t border-[#1e293b] hover:bg-[#334155]">
                  <td className="p-3 text-[#e2e8f0]">{f.chainage?.toFixed(1)}</td>
                  <td className="p-3 text-[#94a3b8]">{f.rock_type}</td>
                  <td className="p-3 text-[#94a3b8]">{f.rmr_score}</td>
                  <td className="p-3 text-[#94a3b8]">{f.q_score}</td>
                  <td className="p-3">{f.fault_zone ? <span className="text-red-400">⚠️ Yes</span> : <span className="text-green-400">No</span>}</td>
                  <td className="p-3 text-[#94a3b8]">{f.mapped_by}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}