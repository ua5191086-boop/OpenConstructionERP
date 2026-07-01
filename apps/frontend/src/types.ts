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

// ============================================================
// Schedule Management
// ============================================================
export interface Schedule {
  id: string;
  project_id: number;
  project_name: string;
  schedule_code: string;
  schedule_name: string;
  schedule_type: string;
  calendar: string;
  data_date: string;
  status: string;
  total_float_pct: number;
  created_by: string;
  created_at: string;
}

export interface ScheduleActivity {
  id: string;
  schedule_id: string;
  project_id: number;
  activity_id: string;
  wbs_code: string;
  activity_name: string;
  activity_type: string;
  status: string;
  original_duration: number;
  remaining_duration: number;
  actual_duration: number;
  percent_complete: number;
  early_start: string;
  early_finish: string;
  late_start: string;
  late_finish: string;
  actual_start: string | null;
  actual_finish: string | null;
  start_date: string;
  finish_date: string;
  float_free: number;
  float_total: number;
  is_critical: boolean;
  is_driving: boolean;
  constraint_type: string;
  constraint_date: string | null;
}

export interface ScheduleRelationship {
  id: string;
  schedule_id: string;
  predecessor_activity_id: string;
  successor_activity_id: string;
  relationship_type: string;
  lag_days: number;
}

export interface ScheduleResource {
  id: string;
  schedule_id: string;
  resource_id: string;
  resource_name: string;
  resource_type: string;
  quantity: number;
  unit: string;
  unit_rate: number;
}

export interface ScheduleBaseline {
  id: string;
  schedule_id: string;
  baseline_code: string;
  baseline_name: string;
  baseline_date: string;
  is_current: boolean;
}

export interface ScheduleChange {
  id: string;
  schedule_id: string;
  change_number: string;
  change_type: string;
  description: string;
  old_duration: number;
  new_duration: number;
  old_start: string;
  new_start: string;
  old_finish: string;
  new_finish: string;
  reason: string;
  approved_by: string;
  approved_at: string;
  status: string;
}

// ============================================================
// Equipment Management
// ============================================================
export interface EquipmentCategory {
  id: string;
  category_code: string;
  category_name: string;
  description: string;
  parent_id: string | null;
  equipment_type: string;
  icon: string;
  sort_order: number;
}

export interface EquipmentItem {
  id: string;
  project_id: number;
  project_name: string;
  equipment_code: string;
  equipment_name: string;
  category_id: string;
  equipment_type: string;
  manufacturer: string;
  model: string;
  serial_number: string;
  year_manufactured: number;
  capacity: string;
  capacity_unit: string;
  status: string;
  location: string;
  purchase_date: string;
  purchase_cost: number;
  current_value: number;
  fuel_type: string;
  fuel_capacity: number;
  hourly_rate: number;
  meter_type: string;
  meter_reading: number;
  operator_required: boolean;
  next_service_date: string;
  is_active: boolean;
}

export interface EquipmentMaintenance {
  id: string;
  equipment_id: string;
  maintenance_type: string;
  description: string;
  start_date: string;
  end_date: string;
  cost: number;
  status: string;
  technician: string;
  notes: string;
}

export interface EquipmentMaintenanceSchedule {
  id: string;
  equipment_id: string;
  schedule_code: string;
  maintenance_type: string;
  frequency_days: number;
  frequency_meter: number;
  description: string;
  is_active: boolean;
  next_due_date: string;
}

export interface EquipmentTelemetry {
  id: string;
  equipment_id: string;
  recorded_at: string;
  meter_value: number;
  fuel_level: number;
  engine_temp: number;
  oil_pressure: number;
  gps_lat: number;
  gps_lng: number;
  is_running: boolean;
}

export interface EquipmentFuel {
  id: string;
  equipment_id: string;
  refuel_date: string;
  quantity: number;
  unit: string;
  cost: number;
  fuel_type: string;
  operator: string;
}

export interface EquipmentDowntime {
  id: string;
  equipment_id: string;
  downtime_type: string;
  description: string;
  start_date: string;
  end_date: string;
  duration_hours: number;
  cost: number;
}

export interface EquipmentSparePart {
  id: string;
  equipment_id: string;
  part_code: string;
  part_name: string;
  quantity: number;
  unit: string;
  unit_price: number;
  reorder_level: number;
}

// ============================================================
// HSE Module
// ============================================================
export interface HSEIncident {
  id: string;
  project_id: number;
  project_name: string;
  incident_number: number;
  incident_code: string;
  title: string;
  description: string;
  incident_type: string;
  severity: string;
  incident_date: string;
  incident_time: string;
  location: string;
  area: string;
  reported_by: string;
  reported_at: string;
  affected_person: string | null;
  lost_days: number;
  medical_cost: number;
  property_cost: number;
  total_cost: number;
  root_cause: string | null;
  investigation_status: string;
  investigation_lead: string;
  is_reportable: boolean;
  status: string;
}

export interface HSEPermit {
  id: string;
  project_id: number;
  project_name: string;
  permit_number: string;
  permit_type: string;
  description: string;
  status: string;
  issue_date: string;
  expiry_date: string;
  work_area: string;
  issued_by: string;
  holder: string;
}

export interface HSEAudit {
  id: string;
  project_id: number;
  project_name: string;
  audit_code: string;
  audit_type: string;
  title: string;
  status: string;
  audit_date: string;
  lead_auditor: string;
  score: number;
  findings_count: number;
}

export interface HSEInspection {
  id: string;
  project_id: number;
  project_name: string;
  inspection_code: string;
  inspection_type: string;
  title: string;
  status: string;
  inspection_date: string;
  inspector: string;
  findings_count: number;
  severity: string;
}

export interface HSETraining {
  id: string;
  project_id: number;
  project_name: string;
  training_type: string;
  title: string;
  training_date: string;
  instructor: string;
  attendee_count: number;
  status: string;
}

export interface HSEPPE {
  id: string;
  project_id: number;
  ppe_type: string;
  description: string;
  quantity: number;
  status: string;
  expiry_date: string;
}

export interface HSEDrill {
  id: string;
  project_id: number;
  drill_type: string;
  title: string;
  drill_date: string;
  duration_minutes: number;
  participants: number;
  status: string;
}

export interface HSEStatistics {
  id: string;
  project_id: number;
  period: string;
  total_incidents: number;
  lost_time_injuries: number;
  near_misses: number;
  first_aid_cases: number;
  manhours_worked: number;
  lti_frequency_rate: number;
  lti_severity_rate: number;
}

export interface HSEEmergencyPlan {
  id: string;
  project_id: number;
  plan_code: string;
  plan_name: string;
  emergency_type: string;
  status: string;
  last_review_date: string;
  next_review_date: string;
}

export interface HSEChemical {
  id: string;
  project_id: number;
  chemical_name: string;
  cas_number: string;
  hazard_class: string;
  quantity: number;
  unit: string;
  storage_location: string;
  safety_data_sheet: string;
}

// ============================================================
// Quality Management
// ============================================================
export interface QualityITP {
  id: string;
  project_id: number;
  project_name: string;
  itp_number: number;
  itp_code: string;
  itp_name: string;
  itp_type: string;
  description: string;
  status: string;
}

export interface QualityInspection {
  id: string;
  project_id: number;
  project_name: string;
  record_number: number;
  record_code: string;
  title: string;
  inspection_type: string;
  inspector: string;
  inspection_date: string;
  result: string;
  defects_found: number;
}

export interface QualityTestResult {
  id: string;
  project_id: number;
  project_name: string;
  test_number: number;
  test_code: string;
  test_name: string;
  test_type: string;
  test_date: string;
  measured_value: number;
  min_acceptable: number;
  max_acceptable: number;
  result: string;
  lab_name: string;
}

export interface QualityNCR {
  id: string;
  project_id: number;
  project_name: string;
  ncr_number: number;
  ncr_code: string;
  title: string;
  ncr_category: string;
  severity: string;
  source: string;
  description: string;
  discovered_date: string;
  discovered_by: string;
  root_cause: string | null;
  disposition_type: string | null;
  rework_cost: number;
  schedule_impact: number;
  status: string;
}

export interface QualityCorrectiveAction {
  id: string;
  project_id: number;
  project_name: string;
  ca_number: number;
  ca_code: string;
  title: string;
  action_type: string;
  assigned_to: string;
  priority: string;
  due_date: string;
  effectiveness: string | null;
  status: string;
}

export interface QualityCalibration {
  id: string;
  project_id: number;
  project_name: string;
  equipment_name: string;
  equipment_model: string;
  serial_number: string;
  calibration_frequency_days: number;
  last_calibration_date: string;
  next_calibration_date: string;
  calibration_result: string;
  status: string;
}

export interface QualityMetric {
  id: string;
  project_id: number;
  project_name: string;
  report_month: string;
  total_inspections: number;
  inspections_passed: number;
  inspections_failed: number;
  total_tests: number;
  tests_passed: number;
  tests_failed: number;
  ncr_opened: number;
  ncr_closed: number;
  ncr_critical: number;
  first_pass_yield: number;
  rework_cost: number;
}

// ============================================================
// GIS & Survey
// ============================================================
export interface GISLayer {
  id: string;
  project_id: number;
  project_name: string;
  layer_name: string;
  layer_type: string;
  geometry_type: string;
  source_type: string;
  is_visible: boolean;
  status: string;
}

export interface GISFeature {
  id: string;
  project_id: number;
  project_name: string;
  feature_name: string;
  feature_type: string;
  geometry: any;
  properties: any;
}

export interface GISSurveyPoint {
  id: string;
  project_id: number;
  project_name: string;
  point_number: number;
  point_code: string;
  point_name: string;
  point_type: string;
  latitude: number;
  longitude: number;
  elevation: number;
  northing: number;
  easting: number;
  zone: string;
  accuracy_mm: number;
  method: string;
  survey_date: string;
  status: string;
}

export interface GISSurveyRun {
  id: string;
  project_id: number;
  project_name: string;
  run_number: number;
  run_code: string;
  run_name: string;
  survey_type: string;
  start_date: string;
  end_date: string;
  instrument: string;
  crew_lead: string;
  point_count: number;
  status: string;
}

export interface GISStation {
  id: string;
  project_id: number;
  project_name: string;
  station_number: number;
  station_code: string;
  station_name: string;
  station_type: string;
  northing: number;
  easting: number;
  elevation: number;
}

export interface GISAlignment {
  id: string;
  project_id: number;
  project_name: string;
  alignment_code: string;
  alignment_name: string;
  alignment_type: string;
  start_chainage: number;
  end_chainage: number;
  total_length: number;
  geometry: any;
  status: string;
}

export interface GISCrossSection {
  id: string;
  project_id: number;
  project_name: string;
  section_number: number;
  chainage: number;
  offset_left: number;
  offset_right: number;
  geometry: any;
  points: any[];
  cut_area: number;
  fill_area: number;
}

export interface GISDroneFlight {
  id: string;
  project_id: number;
  project_name: string;
  flight_number: number;
  flight_code: string;
  flight_name: string;
  drone_model: string;
  pilot: string;
  flight_date: string;
  flight_duration_minutes: number;
  altitude_m: number;
  area_covered_ha: number;
  gsd_cm: number;
  overlap_pct: number;
  images_count: number;
  processing_status: string;
  sensor_type: string;
  output_type: string;
  status: string;
}

// ============================================================
// Risk Management
// ============================================================
export interface RiskCategory {
  id: string;
  category_code: string;
  category_name: string;
  category_type: string;
  description: string;
  sort_order: number;
  is_active: boolean;
}

export interface RiskRegister {
  id: string;
  project_id: number;
  project_name: string;
  risk_number: number;
  risk_code: string;
  risk_name: string;
  risk_type: string;
  description: string;
  probability_score: number;
  impact_score: number;
  risk_score: number;
  risk_rating: string;
  cost_impact: number;
  schedule_impact_days: number;
  risk_owner: string;
  risk_response: string;
  status: string;
}

export interface RiskScenario {
  id: string;
  project_id: number;
  project_name: string;
  scenario_number: number;
  scenario_code: string;
  scenario_name: string;
  scenario_type: string;
  description: string;
  cost_impact_min: number;
  cost_impact_max: number;
  cost_impact_ml: number;
  schedule_impact_min: number;
  schedule_impact_max: number;
  schedule_impact_ml: number;
  probability_pct: number;
  severity: string;
  status: string;
}

export interface RiskMitigation {
  id: string;
  project_id: number;
  project_name: string;
  action_number: number;
  action_code: string;
  action_name: string;
  action_type: string;
  assigned_to: string;
  budget: number;
  due_date: string;
  effectiveness: string | null;
  status: string;
}

export interface RiskEscalation {
  id: string;
  project_id: number;
  project_name: string;
  escalation_number: number;
  escalation_code: string;
  title: string;
  reason: string;
  escalated_to: string;
  escalated_by: string;
  decision: string | null;
  status: string;
}

export interface RiskMonteCarlo {
  id: string;
  project_id: number;
  project_name: string;
  run_label: string;
  run_type: string;
  iterations: number;
  p10_value: number;
  p50_value: number;
  p90_value: number;
  mean_value: number;
  confidence_level: number;
  status: string;
}

export interface RiskDashboard {
  id: string;
  project_id: number;
  project_name: string;
  snapshot_date: string;
  total_risks: number;
  open_risks: number;
  extreme_risks: number;
  high_risks: number;
  medium_risks: number;
  low_risks: number;
  threats: number;
  opportunities: number;
  risk_exposure: number;
  mitigation_progress_pct: number;
}

// ============================================================
// Change Management
// ============================================================
export interface ChangeRequest {
  id: string;
  project_id: number;
  project_name: string;
  cr_number: number;
  cr_code: string;
  cr_name: string;
  cr_type: string;
  source: string;
  priority: string;
  description: string;
  reason: string;
  proposed_by: string;
  proposed_date: string;
  required_by_date: string;
  status: string;
}

export interface ChangeOrder {
  id: string;
  project_id: number;
  project_name: string;
  co_number: number;
  co_code: string;
  co_name: string;
  co_type: string;
  scope_change: string;
  cost_change: number;
  schedule_change_days: number;
  justification: string;
  contractor_name: string;
  approved_by: string;
  status: string;
}

export interface ChangeImpactAnalysis {
  id: string;
  project_id: number;
  project_name: string;
  impact_type: string;
  description: string;
  impact_level: string;
  cost_impact: number;
  schedule_impact_days: number;
  analyzed_by: string;
  analysis_date: string;
  status: string;
}

export interface ChangeApprovalWorkflow {
  id: string;
  project_id: number;
  project_name: string;
  step_order: number;
  step_name: string;
  approver_role: string;
  status: string;
}

export interface ChangeLogEntry {
  id: string;
  project_id: number;
  project_name: string;
  log_type: string;
  previous_status: string | null;
  new_status: string | null;
  description: string;
  changed_by: string;
  changed_at: string;
}
