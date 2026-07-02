import { useState, useEffect } from 'react'

export default function FleetPage() {
  const [vehicles, setVehicles] = useState<any[]>([])
  const [drivers, setDrivers] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    Promise.all([
      fetch('/api/v1/protected/fleet/vehicles').then(r => r.ok ? r.json() : Promise.reject()),
      fetch('/api/v1/protected/fleet/drivers').then(r => r.ok ? r.json() : Promise.reject()),
    ])
      .then(([v, d]) => { setVehicles(v); setDrivers(d) })
      .catch(() => setError('API unavailable'))
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <div className="p-8 text-center text-[#64748b]">Loading Fleet data...</div>

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-white">🚛 Fleet</h1>
        <p className="text-[#94a3b8] mt-1">V032 — Vehicles, Drivers, Fuel, Maintenance & Tracking</p>
      </div>

      {error ? (
        <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-8 text-center">
          <div className="text-4xl mb-4">🚛</div>
          <h2 className="text-xl font-bold text-white mb-2">Fleet Module</h2>
          <p className="text-[#94a3b8]">{error}</p>
        </div>
      ) : (
        <>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Vehicles</div>
              <div className="text-2xl font-bold text-[#3b82f6] mt-2">{vehicles.length}</div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Operational</div>
              <div className="text-2xl font-bold text-[#22c55e] mt-2">{vehicles.filter((v: any) => v.status === 'operational').length}</div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">Drivers</div>
              <div className="text-2xl font-bold text-[#a855f7] mt-2">{drivers.length}</div>
            </div>
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl p-5">
              <div className="text-[#94a3b8] text-xs uppercase">In Maintenance</div>
              <div className="text-2xl font-bold text-[#f97316] mt-2">{vehicles.filter((v: any) => v.status === 'under_maintenance').length}</div>
            </div>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
            <div className="bg-[#1e293b] border border-[#334155] rounded-xl overflow-hidden">
              <div className="p-4 border-b border-[#334155]">
                <h3 className="text-white font-semibold">Fleet Vehicles</h3>
              </div>
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="text-left text-xs text-[#64748b] uppercase tracking-wider">
                      <th className="p-3 pl-4">Type</th>
                      <th className="p-3">Make/Model</th>
                      <th className="p-3">Plate</th>
                      <th className="p-3">Status</th>
                    </tr>
                  </thead>
                  <tbody>
                    {vehicles.map((v: any) => (
                      <tr key={v.id} className="border-t border-[#1e293b] hover:bg-[#334155] text-sm">
                        <td className="p-3 pl-4 text-white">{v.vehicle_type}</td>
                        <td className="p-3 text-[#94a3b8]">{v.make || ''} {v.model || ''}</td>
                        <td className="p-3 font-mono text-xs text-[#64748b]">{v.license_plate || '—'}</td>
                        <td className="p-3">{v.status}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>

            <div className="bg-[#1e293b] border border-[#334155] rounded-xl overflow-hidden">
              <div className="p-4 border-b border-[#334155]">
                <h3 className="text-white font-semibold">Drivers</h3>
              </div>
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="text-left text-xs text-[#64748b] uppercase tracking-wider">
                      <th className="p-3 pl-4">Name</th>
                      <th className="p-3">License</th>
                      <th className="p-3">Type</th>
                      <th className="p-3 pr-4">Status</th>
                    </tr>
                  </thead>
                  <tbody>
                    {drivers.map((d: any) => (
                      <tr key={d.id} className="border-t border-[#1e293b] hover:bg-[#334155] text-sm">
                        <td className="p-3 pl-4 text-white">{d.driver_name}</td>
                        <td className="p-3 font-mono text-xs text-[#64748b]">{d.license_number || '—'}</td>
                        <td className="p-3 text-[#94a3b8]">{d.license_type || '—'}</td>
                        <td className="p-3 pr-4">{d.status}</td>
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