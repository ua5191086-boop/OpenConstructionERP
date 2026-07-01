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
