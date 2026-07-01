// ============================================================
// Shared types for OpenConstructionERP frontend
// ============================================================

export interface ApiResponse<T> {
  data: T;
  error?: string;
}

// --- BOQ ---
export interface BOQItem {
  id: string;
  code: string;
  name: string;
  cbs_code: string;
  unit: string;
  quantity: number;
  unit_price: number;
  total_cost: number;
  contractor?: { name: string };
  object_id: string;
}

export interface BOQSection {
  id: string;
  name: string;
}

export interface BOQComplex {
  id: string;
  name: string;
  section_id: string;
}

export interface BOQObject {
  id: string;
  name: string;
  complex_id: string;
}

export interface BOQSummary {
  total_boq_cost: number;
  total_sections: number;
  total_complexes: number;
  total_boq_items: number;
}

export interface BOQData {
  project: { name: string };
  sections: BOQSection[];
  complexes: BOQComplex[];
  objects: BOQObject[];
  boq_items: BOQItem[];
  summary: BOQSummary;
}

// --- Tenders ---
export interface Tender {
  id: string;
  code: string;
  name: string;
  status: string;
  tender_type: string;
  budget_amount: number;
  currency: string;
  submission_deadline: string;
  bid_open_date: string;
}

// --- Contracts ---
export interface Contract {
  id: string;
  code: string;
  name: string;
  status: string;
  contract_type: string;
  contract_amount: number;
  currency: string;
  start_date: string;
  end_date: string;
  contractor_id: string;
}

// --- HR ---
export interface Employee {
  id: string;
  full_name: string;
  position: string;
  department: string;
  status: string;
  email: string;
  phone: string;
}

// --- Finance ---
export interface Budget {
  id: string;
  name: string;
  budget_type: string;
  total_amount: number;
  currency: string;
  status: string;
}

export interface Invoice {
  id: string;
  invoice_number: string;
  amount: number;
  status: string;
  due_date: string;
}

// --- Procurement ---
export interface ProcurementRequest {
  id: string;
  request_number: string;
  description: string;
  status: string;
  priority: string;
  estimated_cost: number;
  required_date: string;
}

export interface PurchaseOrder {
  id: string;
  po_number: string;
  status: string;
  total_amount: number;
  vendor_name: string;
}

// --- BIM ---
export interface BIMModel {
  id: string;
  model_name: string;
  model_version: string;
  discipline: string;
  status: string;
  lod: string;
  author: string;
}

export interface BIMElement {
  id: string;
  element_name: string;
  element_type: string;
  ifc_class: string;
  material: string;
}

// --- AI ---
export interface AIAgent {
  id: string;
  agent_name: string;
  agent_type: string;
  description: string;
  model_name: string;
  is_active: boolean;
}

export interface AITask {
  id: string;
  task_name: string;
  task_type: string;
  status: string;
  agent_id: string;
}

// --- Project Management ---
export interface PMProject {
  id: number;
  code: string;
  name: string;
  project_type: string;
  status: string;
  phase: string;
  budget_total: number;
  budget_currency: string;
  country: string;
  city: string;
  start_date: string;
  end_date: string;
  duration_days: number;
  risk_class: string;
  complexity: string;
}

export interface PMWBSItem {
  id: number;
  project_id: number;
  parent_id: number | null;
  wbs_code: string;
  name: string;
  wbs_level: number;
  sort_order: number;
  is_leaf: boolean;
  progress_pct: number;
  status: string;
}

export interface PMMilestone {
  id: number;
  project_id: number;
  milestone_code: string;
  name: string;
  milestone_type: string;
  planned_date: string;
  actual_date: string | null;
  status: string;
  weight_pct: number;
  is_gate: boolean;
}

export interface PMPhase {
  id: number;
  project_id: number;
  phase_code: string;
  name: string;
  sort_order: number;
  status: string;
  completion_pct: number;
  budget_amount: number;
}

export interface PMRisk {
  id: number;
  project_id: number;
  risk_code: string;
  name: string;
  risk_category: string;
  probability_score: number;
  impact_score: number;
  risk_score: number;
  probability: string;
  impact: string;
  potential_cost: number;
  mitigation_strategy: string;
  status: string;
}

export interface PMChange {
  id: number;
  project_id: number;
  change_number: string;
  change_type: string;
  source: string;
  description: string;
  cost_impact: number;
  schedule_impact: number;
  status: string;
}

export interface PMLesson {
  id: number;
  project_id: number;
  category: string;
  title: string;
  is_positive: boolean;
  severity: string;
  status: string;
}

export interface PMPortfolio {
  id: number;
  code: string;
  name: string;
  portfolio_type: string;
  parent_id: number | null;
  budget_total: number;
  status: string;
}
