import { BrowserRouter, Routes, Route } from 'react-router-dom'
import Layout from './components/Layout'
import Dashboard from './pages/Dashboard'
import BOQPage from './pages/BOQPage'
import TendersPage from './pages/TendersPage'
import ContractsPage from './pages/ContractsPage'
import HRPage from './pages/HRPage'
import FinancePage from './pages/FinancePage'
import ProcurementPage from './pages/ProcurementPage'
import BIMPage from './pages/BIMPage'
import AIPage from './pages/AIPage'
import PMProjectPage from './pages/PMProjectPage'
import DocControlPage from './pages/DocControlPage'
import SchedulePage from './pages/SchedulePage'
import EquipmentPage from './pages/EquipmentPage'
import HSEPage from './pages/HSEPage'
import QualityPage from './pages/QualityPage'
import GISPage from './pages/GISPage'
import RiskPage from './pages/RiskPage'
import ChangePage from './pages/ChangePage'

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<Layout />}>
          <Route path="/" element={<Dashboard />} />
          <Route path="/boq" element={<BOQPage />} />
          <Route path="/tenders" element={<TendersPage />} />
          <Route path="/contracts" element={<ContractsPage />} />
          <Route path="/hr" element={<HRPage />} />
          <Route path="/finance" element={<FinancePage />} />
          <Route path="/procurement" element={<ProcurementPage />} />
          <Route path="/bim" element={<BIMPage />} />
          <Route path="/ai" element={<AIPage />} />
          <Route path="/pm" element={<PMProjectPage />} />
          <Route path="/doc-control" element={<DocControlPage />} />
          <Route path="/schedule" element={<SchedulePage />} />
          <Route path="/equipment" element={<EquipmentPage />} />
          <Route path="/hse" element={<HSEPage />} />
          <Route path="/quality" element={<QualityPage />} />
          <Route path="/gis" element={<GISPage />} />
          <Route path="/risk" element={<RiskPage />} />
          <Route path="/change" element={<ChangePage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}

export default App
