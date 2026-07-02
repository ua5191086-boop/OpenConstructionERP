package models

import (
	"database/sql"
	"time"
)

// ============================================================================
// BOQ Module
// ============================================================================

type CBSChapter struct {
	ID        string         `json:"id"`
	ProjectID *string        `json:"project_id,omitempty"`
	Code      string         `json:"code"`
	Name      string         `json:"name"`
	NameRU    *string        `json:"name_ru,omitempty"`
	ParentID  *string        `json:"parent_id,omitempty"`
	Level     int            `json:"level"`
	SortOrder int            `json:"sort_order"`
	Path      *string        `json:"path,omitempty"`
	IsActive  bool           `json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type BOQSection struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	SectionType string    `json:"section_type"`
	StartKM     *float64  `json:"start_km,omitempty"`
	EndKM       *float64  `json:"end_km,omitempty"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type BOQComplex struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	SectionID string    `json:"section_id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BOQObject struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	ComplexID string    `json:"complex_id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BOQItem struct {
	ID             string         `json:"id"`
	ProjectID      string         `json:"project_id"`
	ObjectID       string         `json:"object_id"`
	CBSChapterID   string         `json:"cbs_chapter_id"`
	Code           string         `json:"code"`
	Name           string         `json:"name"`
	Description    *string        `json:"description,omitempty"`
	Unit           string         `json:"unit"`
	Quantity       float64        `json:"quantity"`
	UnitPrice      float64        `json:"unit_price"`
	TotalCost      float64        `json:"total_cost"`
	Currency       string         `json:"currency"`
	ContractorID   *string        `json:"contractor_id,omitempty"`
	ContractID     *string        `json:"contract_id,omitempty"`
	FundingSource  *string        `json:"funding_source,omitempty"`
	Phase          *string        `json:"phase,omitempty"`
	Status         string         `json:"status"`
	Notes          *string        `json:"notes,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

type CostTransaction struct {
	ID              string    `json:"id"`
	ProjectID       string    `json:"project_id"`
	BOQItemID       string    `json:"boq_item_id"`
	CBSChapterID    string    `json:"cbs_chapter_id"`
	ContractorID    *string   `json:"contractor_id,omitempty"`
	ContractID      *string   `json:"contract_id,omitempty"`
	TransactionType string    `json:"transaction_type"`
	Amount          float64   `json:"amount"`
	Currency        string    `json:"currency"`
	ExchangeRate    float64   `json:"exchange_rate"`
	Period          string    `json:"period"`
	FundingSource   *string   `json:"funding_source,omitempty"`
	Description     *string   `json:"description,omitempty"`
	ReferenceType   *string   `json:"reference_type,omitempty"`
	ReferenceID     *string   `json:"reference_id,omitempty"`
	CreatedBy       *string   `json:"created_by,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

type BudgetVersion struct {
	ID            string     `json:"id"`
	ProjectID     string     `json:"project_id"`
	VersionNumber int        `json:"version_number"`
	VersionName   *string    `json:"version_name,omitempty"`
	Status        string     `json:"status"`
	TotalAmount   *float64   `json:"total_amount,omitempty"`
	ApprovedBy    *string    `json:"approved_by,omitempty"`
	ApprovedAt    *time.Time `json:"approved_at,omitempty"`
	Notes         *string    `json:"notes,omitempty"`
	CreatedBy     *string    `json:"created_by,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// ============================================================================
// Tenders Module
// ============================================================================

type Tender struct {
	ID                  string     `json:"id"`
	Code                string     `json:"code"`
	Name                string     `json:"name"`
	Description         *string    `json:"description,omitempty"`
	TenderType          string     `json:"tender_type"`
	Status              string     `json:"status"`
	ClientID            *string    `json:"client_id,omitempty"`
	ProjectID           *string    `json:"project_id,omitempty"`
	BudgetAmount        *float64   `json:"budget_amount,omitempty"`
	Currency            string     `json:"currency"`
	PublishedAt         *time.Time `json:"published_at,omitempty"`
	SubmissionDeadline  *time.Time `json:"submission_deadline,omitempty"`
	BidOpenDate         *time.Time `json:"bid_open_date,omitempty"`
	AwardDate           *time.Time `json:"award_date,omitempty"`
	ContractStart       *string    `json:"contract_start,omitempty"`
	ContractEnd         *string    `json:"contract_end,omitempty"`
	BidBondPct          *float64   `json:"bid_bond_pct,omitempty"`
	PerformanceBondPct  *float64   `json:"performance_bond_pct,omitempty"`
	AdvancePaymentPct   *float64   `json:"advance_payment_pct,omitempty"`
	RetentionPct        *float64   `json:"retention_pct,omitempty"`
	RetentionReleaseDays *int     `json:"retention_release_days,omitempty"`
	ProcurementMethod   *string    `json:"procurement_method,omitempty"`
	FundingSource       *string    `json:"funding_source,omitempty"`
	Notes               *string    `json:"notes,omitempty"`
	CreatedBy           *string    `json:"created_by,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type TenderLot struct {
	ID              string    `json:"id"`
	TenderID        string    `json:"tender_id"`
	LotNumber       int       `json:"lot_number"`
	Name            string    `json:"name"`
	Description     *string   `json:"description,omitempty"`
	EstimatedAmount *float64  `json:"estimated_amount,omitempty"`
	Currency        string    `json:"currency"`
	SectionID       *string   `json:"section_id,omitempty"`
	Status          string    `json:"status"`
	AwardDecision   *string   `json:"award_decision,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

type TenderBidder struct {
	ID             string     `json:"id"`
	TenderID       string     `json:"tender_id"`
	LotID          *string    `json:"lot_id,omitempty"`
	ContractorID   string     `json:"contractor_id"`
	BidNumber      *string    `json:"bid_number,omitempty"`
	Status         string     `json:"status"`
	BidAmount      *float64   `json:"bid_amount,omitempty"`
	Currency       string     `json:"currency"`
	BidBondAmount  *float64   `json:"bid_bond_amount,omitempty"`
	ValidityDays   *int       `json:"validity_days,omitempty"`
	SubmissionDate *time.Time `json:"submission_date,omitempty"`
	IsWinner       bool       `json:"is_winner"`
	AwardAmount    *float64   `json:"award_amount,omitempty"`
	AwardReason    *string    `json:"award_reason,omitempty"`
	Notes          *string    `json:"notes,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// ============================================================================
// Contracts Module
// ============================================================================

type Contract struct {
	ID                   string     `json:"id"`
	Code                 string     `json:"code"`
	Name                 string     `json:"name"`
	Description          *string    `json:"description,omitempty"`
	ContractType         string     `json:"contract_type"`
	Status               string     `json:"status"`
	TenderID             *string    `json:"tender_id,omitempty"`
	LotID                *string    `json:"lot_id,omitempty"`
	ClientID             string     `json:"client_id"`
	ContractorID         string     `json:"contractor_id"`
	ProjectID            *string    `json:"project_id,omitempty"`
	ContractAmount       float64    `json:"contract_amount"`
	Currency             string     `json:"currency"`
	AdvanceAmount        *float64   `json:"advance_amount,omitempty"`
	AdvancePct           *float64   `json:"advance_pct,omitempty"`
	SignedAt             *string    `json:"signed_at,omitempty"`
	StartDate            *string    `json:"start_date,omitempty"`
	EndDate              *string    `json:"end_date,omitempty"`
	DurationDays         *int       `json:"duration_days,omitempty"`
	PerformanceBondAmount *float64  `json:"performance_bond_amount,omitempty"`
	PerformanceBondPct   *float64   `json:"performance_bond_pct,omitempty"`
	WarrantyPeriodDays   *int       `json:"warranty_period_days,omitempty"`
	RetentionPct         *float64   `json:"retention_pct,omitempty"`
	RetentionReleaseDays *int       `json:"retention_release_days,omitempty"`
	PenaltyRateDaily     *float64   `json:"penalty_rate_daily,omitempty"`
	PenaltyMaxPct        *float64   `json:"penalty_max_pct,omitempty"`
	LiquidatedDamages    *float64   `json:"liquidated_damages,omitempty"`
	FundingSource        *string    `json:"funding_source,omitempty"`
	PaymentTerms         *string    `json:"payment_terms,omitempty"`
	PaymentTermsType     *string    `json:"payment_terms_type,omitempty"`
	DocumentPath         *string    `json:"document_path,omitempty"`
	Notes                *string    `json:"notes,omitempty"`
	CreatedBy            *string    `json:"created_by,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type ContractMilestone struct {
	ID              string     `json:"id"`
	ContractID      string     `json:"contract_id"`
	MilestoneNumber int        `json:"milestone_number"`
	Name            string     `json:"name"`
	Description     *string    `json:"description,omitempty"`
	MilestoneType   string     `json:"milestone_type"`
	PlannedDate     *string    `json:"planned_date,omitempty"`
	ActualDate      *string    `json:"actual_date,omitempty"`
	Amount          *float64   `json:"amount,omitempty"`
	AmountPct       *float64   `json:"amount_pct,omitempty"`
	Status          string     `json:"status"`
	CompletionPct   *float64   `json:"completion_pct,omitempty"`
	Notes           *string    `json:"notes,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

type ContractPayment struct {
	ID            string     `json:"id"`
	ContractID    string     `json:"contract_id"`
	AcceptanceID  *string    `json:"acceptance_id,omitempty"`
	MilestoneID   *string    `json:"milestone_id,omitempty"`
	PaymentNumber string     `json:"payment_number"`
	PaymentDate   string     `json:"payment_date"`
	Amount        float64    `json:"amount"`
	Currency      string     `json:"currency"`
	PaymentType   string     `json:"payment_type"`
	PaymentMethod *string    `json:"payment_method,omitempty"`
	Status        string     `json:"status"`
	BankRef       *string    `json:"bank_ref,omitempty"`
	Notes         *string    `json:"notes,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// ============================================================================
// HR Module
// ============================================================================

type Employee struct {
	ID                   string     `json:"id"`
	EmployeeCode         string     `json:"employee_code"`
	FullName             string     `json:"full_name"`
	FirstName            *string    `json:"first_name,omitempty"`
	LastName             *string    `json:"last_name,omitempty"`
	Patronymic           *string    `json:"patronymic,omitempty"`
	BirthDate            *string    `json:"birth_date,omitempty"`
	Gender               *string    `json:"gender,omitempty"`
	Nationality          *string    `json:"nationality,omitempty"`
	Email                *string    `json:"email,omitempty"`
	Phone                *string    `json:"phone,omitempty"`
	PhoneEmergency       *string    `json:"phone_emergency,omitempty"`
	Address              *string    `json:"address,omitempty"`
	Position             string     `json:"position"`
	Department           *string    `json:"department,omitempty"`
	PositionType         string     `json:"position_type"`
	PositionCategory     *string    `json:"position_category,omitempty"`
	Grade                *string    `json:"grade,omitempty"`
	Status               string     `json:"status"`
	HireDate             string     `json:"hire_date"`
	ContractEnd          *string    `json:"contract_end,omitempty"`
	TerminationDate      *string    `json:"termination_date,omitempty"`
	TerminationReason    *string    `json:"termination_reason,omitempty"`
	SalaryBase           *float64   `json:"salary_base,omitempty"`
	SalaryCurrency       string     `json:"salary_currency"`
	HourlyRate           *float64   `json:"hourly_rate,omitempty"`
	BankName             *string    `json:"bank_name,omitempty"`
	BankAccount          *string    `json:"bank_account,omitempty"`
	TaxID                *string    `json:"tax_id,omitempty"`
	SocialSecurityID     *string    `json:"social_security_id,omitempty"`
	Education            *string    `json:"education,omitempty"`
	Certifications       *string    `json:"certifications,omitempty"`
	Skills               *string    `json:"skills,omitempty"`
	ExperienceYears      *int       `json:"experience_years,omitempty"`
	PassportNumber       *string    `json:"passport_number,omitempty"`
	PassportExpiry       *string    `json:"passport_expiry,omitempty"`
	WorkPermit           *string    `json:"work_permit,omitempty"`
	WorkPermitExpiry     *string    `json:"work_permit_expiry,omitempty"`
	MedicalCheckupDate   *string    `json:"medical_checkup_date,omitempty"`
	MedicalCheckupValidUntil *string `json:"medical_checkup_valid_until,omitempty"`
	PhotoPath            *string    `json:"photo_path,omitempty"`
	ResumePath           *string    `json:"resume_path,omitempty"`
	Notes                *string    `json:"notes,omitempty"`
	CreatedBy            *string    `json:"created_by,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type Department struct {
	ID            string    `json:"id"`
	Code          string    `json:"code"`
	Name          string    `json:"name"`
	Description   *string   `json:"description,omitempty"`
	ParentID      *string   `json:"parent_id,omitempty"`
	HeadEmployeeID *string  `json:"head_employee_id,omitempty"`
	CostCenter    *string   `json:"cost_center,omitempty"`
	Location      *string   `json:"location,omitempty"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
}

type TimeAttendance struct {
	ID           string     `json:"id"`
	EmployeeID   string     `json:"employee_id"`
	WorkDate     string     `json:"work_date"`
	DayType      string     `json:"day_type"`
	HoursWorked  float64    `json:"hours_worked"`
	HoursOvertime float64   `json:"hours_overtime"`
	Status       string     `json:"status"`
	Reason       *string    `json:"reason,omitempty"`
	ApprovedBy   *string    `json:"approved_by,omitempty"`
	ApprovedAt   *time.Time `json:"approved_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

type EmployeeLeave struct {
	ID         string     `json:"id"`
	EmployeeID string     `json:"employee_id"`
	LeaveType  string     `json:"leave_type"`
	StartDate  string     `json:"start_date"`
	EndDate    string     `json:"end_date"`
	DaysCount  int        `json:"days_count"`
	Status     string     `json:"status"`
	Reason     *string    `json:"reason,omitempty"`
	ApprovedBy *string    `json:"approved_by,omitempty"`
	ApprovedAt *time.Time `json:"approved_at,omitempty"`
	Notes      *string    `json:"notes,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// ============================================================================
// Finance Module
// ============================================================================

type ProjectBudget struct {
	ID                string     `json:"id"`
	ProjectID         string     `json:"project_id"`
	Version           string     `json:"version"`
	Name              string     `json:"name"`
	Description       *string    `json:"description,omitempty"`
	BudgetType        string     `json:"budget_type"`
	TotalAmount       float64    `json:"total_amount"`
	Currency          string     `json:"currency"`
	ContingencyPct    *float64   `json:"contingency_pct,omitempty"`
	ContingencyAmount *float64   `json:"contingency_amount,omitempty"`
	Status            string     `json:"status"`
	ApprovedBy        *string    `json:"approved_by,omitempty"`
	ApprovedAt        *time.Time `json:"approved_at,omitempty"`
	IsActive          bool       `json:"is_active"`
	Notes             *string    `json:"notes,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

type BudgetItem struct {
	ID               string    `json:"id"`
	BudgetID         string    `json:"budget_id"`
	ParentID         *string   `json:"parent_id,omitempty"`
	ItemCode         string    `json:"item_code"`
	Name             string    `json:"name"`
	Description      *string   `json:"description,omitempty"`
	ItemType         string    `json:"item_type"`
	CBSCode          *string   `json:"cbs_code,omitempty"`
	PlannedAmount    float64   `json:"planned_amount"`
	ActualAmount     float64   `json:"actual_amount"`
	CommittedAmount  float64   `json:"committed_amount"`
	RemainingAmount  *float64  `json:"remaining_amount,omitempty"`
	Currency         string    `json:"currency"`
	SortOrder        int       `json:"sort_order"`
	IsLeaf           bool      `json:"is_leaf"`
	Notes            *string   `json:"notes,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

type CashFlow struct {
	ID            string     `json:"id"`
	ProjectID     *string    `json:"project_id,omitempty"`
	ContractID    *string    `json:"contract_id,omitempty"`
	EntryDate     string     `json:"entry_date"`
	EntryType     string     `json:"entry_type"`
	Category      string     `json:"category"`
	Amount        float64    `json:"amount"`
	Currency      string     `json:"currency"`
	IsPlanned     bool       `json:"is_planned"`
	Description   *string    `json:"description,omitempty"`
	ReferenceType *string    `json:"reference_type,omitempty"`
	ReferenceID   *string    `json:"reference_id,omitempty"`
	Status        string     `json:"status"`
	ReconciledAt  *time.Time `json:"reconciled_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

type Invoice struct {
	ID            string     `json:"id"`
	InvoiceNumber string     `json:"invoice_number"`
	InvoiceType   string     `json:"invoice_type"`
	ContractID    *string    `json:"contract_id,omitempty"`
	AcceptanceID  *string    `json:"acceptance_id,omitempty"`
	IssuerID      *string    `json:"issuer_id,omitempty"`
	RecipientID   *string    `json:"recipient_id,omitempty"`
	InvoiceDate   string     `json:"invoice_date"`
	DueDate       *string    `json:"due_date,omitempty"`
	Amount        float64    `json:"amount"`
	TaxAmount     float64    `json:"tax_amount"`
	TaxRate       float64    `json:"tax_rate"`
	TotalAmount   float64    `json:"total_amount"`
	Currency      string     `json:"currency"`
	Status        string     `json:"status"`
	PaidAt        *time.Time `json:"paid_at,omitempty"`
	PaymentRef    *string    `json:"payment_ref,omitempty"`
	Notes         *string    `json:"notes,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// ============================================================================
// Procurement Module
// ============================================================================

type ProcurementRequest struct {
	ID             string     `json:"id"`
	RequestNumber  string     `json:"request_number"`
	ProjectID      *string    `json:"project_id,omitempty"`
	SectionID      *string    `json:"section_id,omitempty"`
	RequestedBy    *string    `json:"requested_by,omitempty"`
	RequestDate    string     `json:"request_date"`
	RequiredDate   *string    `json:"required_date,omitempty"`
	Priority       string     `json:"priority"`
	Status         string     `json:"status"`
	Description    *string    `json:"description,omitempty"`
	Justification  *string    `json:"justification,omitempty"`
	EstimatedCost  *float64   `json:"estimated_cost,omitempty"`
	Currency       string     `json:"currency"`
	BudgetItemID   *string    `json:"budget_item_id,omitempty"`
	ApprovedBy     *string    `json:"approved_by,omitempty"`
	ApprovedAt     *time.Time `json:"approved_at,omitempty"`
	Notes          *string    `json:"notes,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type PurchaseOrder struct {
	ID              string     `json:"id"`
	PONumber        string     `json:"po_number"`
	RequestID       *string    `json:"request_id,omitempty"`
	ProjectID       *string    `json:"project_id,omitempty"`
	VendorID        string     `json:"vendor_id"`
	OrderDate       string     `json:"order_date"`
	DeliveryDate    *string    `json:"delivery_date,omitempty"`
	DeliveryAddress *string    `json:"delivery_address,omitempty"`
	PaymentTerms    *string    `json:"payment_terms,omitempty"`
	ShippingTerms   *string    `json:"shipping_terms,omitempty"`
	Subtotal        *float64   `json:"subtotal,omitempty"`
	TaxAmount       float64    `json:"tax_amount"`
	TaxRate         float64    `json:"tax_rate"`
	ShippingCost    float64    `json:"shipping_cost"`
	TotalAmount     float64    `json:"total_amount"`
	Currency        string     `json:"currency"`
	Status          string     `json:"status"`
	ApprovedBy      *string    `json:"approved_by,omitempty"`
	ApprovedAt      *time.Time `json:"approved_at,omitempty"`
	Notes           *string    `json:"notes,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type InventoryItem struct {
	ID                string    `json:"id"`
	ItemCode          string    `json:"item_code"`
	Name              string    `json:"name"`
	Description       *string   `json:"description,omitempty"`
	Category          *string   `json:"category,omitempty"`
	Unit              string    `json:"unit"`
	UnitPrice         *float64  `json:"unit_price,omitempty"`
	Currency          string    `json:"currency"`
	MinQuantity       float64   `json:"min_quantity"`
	MaxQuantity       *float64  `json:"max_quantity,omitempty"`
	CurrentQuantity   float64   `json:"current_quantity"`
	ReservedQuantity  float64   `json:"reserved_quantity"`
	AvailableQuantity float64   `json:"available_quantity"`
	StorageLocation   *string   `json:"storage_location,omitempty"`
	Warehouse         *string   `json:"warehouse,omitempty"`
	MaterialType      *string   `json:"material_type,omitempty"`
	IsActive          bool      `json:"is_active"`
	Notes             *string   `json:"notes,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// ============================================================================
// BIM Module
// ============================================================================

type BIMModel struct {
	ID          string     `json:"id"`
	ProjectID   string     `json:"project_id"`
	ModelName   string     `json:"model_name"`
	ModelVersion string    `json:"model_version"`
	Description *string    `json:"description,omitempty"`
	Discipline  string     `json:"discipline"`
	Author      *string    `json:"author,omitempty"`
	Software    *string    `json:"software,omitempty"`
	FileFormat  *string    `json:"file_format,omitempty"`
	FilePath    *string    `json:"file_path,omitempty"`
	FileSize    *int64     `json:"file_size,omitempty"`
	IFCSchema   *string    `json:"ifc_schema,omitempty"`
	LOD         *string    `json:"lod,omitempty"`
	Status      string     `json:"status"`
	Checksum    *string    `json:"checksum,omitempty"`
	IsLatest    bool       `json:"is_latest"`
	Notes       *string    `json:"notes,omitempty"`
	UploadedBy  *string    `json:"uploaded_by,omitempty"`
	UploadedAt  time.Time  `json:"uploaded_at"`
}

type BIMElement struct {
	ID          string    `json:"id"`
	ModelID     string    `json:"model_id"`
	IFCGlobalID *string   `json:"ifc_global_id,omitempty"`
	IFCType     string    `json:"ifc_type"`
	IFCClass    *string   `json:"ifc_class,omitempty"`
	Name        *string   `json:"name,omitempty"`
	Description *string   `json:"description,omitempty"`
	Level       *string   `json:"level,omitempty"`
	Material    *string   `json:"material,omitempty"`
	Volume      *float64  `json:"volume,omitempty"`
	Area        *float64  `json:"area,omitempty"`
	Length      *float64  `json:"length,omitempty"`
	Weight      *float64  `json:"weight,omitempty"`
	Elevation   *float64  `json:"elevation,omitempty"`
	XPosition   *float64  `json:"x_position,omitempty"`
	YPosition   *float64  `json:"y_position,omitempty"`
	ZPosition   *float64  `json:"z_position,omitempty"`
	Properties  *string   `json:"properties,omitempty"`
	Status      string    `json:"status"`
	BOQItemID   *string   `json:"boq_item_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type BIMClash struct {
	ID            string     `json:"id"`
	ModelID       string     `json:"model_id"`
	ClashGroup    *string    `json:"clash_group,omitempty"`
	ClashType     string     `json:"clash_type"`
	Severity      string     `json:"severity"`
	Status        string     `json:"status"`
	ElementAID    *string    `json:"element_a_id,omitempty"`
	ElementBID    *string    `json:"element_b_id,omitempty"`
	ElementAName  *string    `json:"element_a_name,omitempty"`
	ElementBName  *string    `json:"element_b_name,omitempty"`
	Distance      *float64   `json:"distance,omitempty"`
	Tolerance     *float64   `json:"tolerance,omitempty"`
	LocationX     *float64   `json:"location_x,omitempty"`
	LocationY     *float64   `json:"location_y,omitempty"`
	LocationZ     *float64   `json:"location_z,omitempty"`
	ScreenshotPath *string   `json:"screenshot_path,omitempty"`
	AssignedTo    *string    `json:"assigned_to,omitempty"`
	Resolution    *string    `json:"resolution,omitempty"`
	ResolvedBy    *string    `json:"resolved_by,omitempty"`
	ResolvedAt    *time.Time `json:"resolved_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// ============================================================================
// AI Module
// ============================================================================

type AIAgent struct {
	ID            string    `json:"id"`
	AgentName     string    `json:"agent_name"`
	AgentType     string    `json:"agent_type"`
	Description   *string   `json:"description,omitempty"`
	ModelName     *string   `json:"model_name,omitempty"`
	ModelProvider *string   `json:"model_provider,omitempty"`
	SystemPrompt  *string   `json:"system_prompt,omitempty"`
	Temperature   *float64  `json:"temperature,omitempty"`
	MaxTokens     *int      `json:"max_tokens,omitempty"`
	IsActive      bool      `json:"is_active"`
	Version       string    `json:"version"`
	Config        *string   `json:"config,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type AITask struct {
	ID                string     `json:"id"`
	AgentID           *string    `json:"agent_id,omitempty"`
	TaskType          string     `json:"task_type"`
	InputData         string     `json:"input_data"`
	InputFormat       string     `json:"input_format"`
	OutputData        *string    `json:"output_data,omitempty"`
	OutputFormat      string     `json:"output_format"`
	Confidence        *float64   `json:"confidence,omitempty"`
	TokensUsed        *int       `json:"tokens_used,omitempty"`
	Cost              *float64   `json:"cost,omitempty"`
	ProcessingTimeMs  *int       `json:"processing_time_ms,omitempty"`
	Status            string     `json:"status"`
	ErrorMessage      *string    `json:"error_message,omitempty"`
	SourceType        *string    `json:"source_type,omitempty"`
	SourceID          *string    `json:"source_id,omitempty"`
	CreatedBy         *string    `json:"created_by,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
}

type AIConversation struct {
	ID                string    `json:"id"`
	SessionID         string    `json:"session_id"`
	UserMessage       string    `json:"user_message"`
	AssistantMessage  string    `json:"assistant_message"`
	Intent            *string   `json:"intent,omitempty"`
	Entities          *string   `json:"entities,omitempty"`
	ModuleUsed        *string   `json:"module_used,omitempty"`
	ActionTaken       *string   `json:"action_taken,omitempty"`
	TokensUsed        *int      `json:"tokens_used,omitempty"`
	ProcessingTimeMs  *int      `json:"processing_time_ms,omitempty"`
	FeedbackScore     *int      `json:"feedback_score,omitempty"`
	FeedbackText      *string   `json:"feedback_text,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

// ============================================================================
// EVM Module (V025)
// ============================================================================

type EVMControlAccount struct {
	ID            string    `json:"id"`
	ProjectID     string    `json:"project_id"`
	CACode        string    `json:"ca_code"`
	CAName        string    `json:"ca_name"`
	Description   *string   `json:"description,omitempty"`
	WBSCode       *string   `json:"wbs_code,omitempty"`
	Responsible   *string   `json:"responsible,omitempty"`
	SortOrder     int       `json:"sort_order"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type EVMBaseline struct {
	ID            string     `json:"id"`
	ProjectID     string     `json:"project_id"`
	BaselineName  string     `json:"baseline_name"`
	BaselineType  string     `json:"baseline_type"`
	Version       string     `json:"version"`
	Description   *string    `json:"description,omitempty"`
	IsApproved    bool       `json:"is_approved"`
	ApprovedBy    *string    `json:"approved_by,omitempty"`
	ApprovedAt    *time.Time `json:"approved_at,omitempty"`
	IsActive      bool       `json:"is_active"`
	CreatedAt     time.Time  `json:"created_at"`
}

type EVMPeriod struct {
	ID               string    `json:"id"`
	ProjectID        string    `json:"project_id"`
	ControlAccountID *string   `json:"control_account_id,omitempty"`
	BaselineID       *string   `json:"baseline_id,omitempty"`
	PeriodDate       string    `json:"period_date"`
	PeriodType       string    `json:"period_type"`
	PlannedValue     float64   `json:"planned_value"`
	PlannedHours     float64   `json:"planned_hours"`
	PlannedProgress  float64   `json:"planned_progress"`
	IsCumulative     bool      `json:"is_cumulative"`
	Notes            *string   `json:"notes,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

type EVMActual struct {
	ID               string    `json:"id"`
	ProjectID        string    `json:"project_id"`
	ControlAccountID *string   `json:"control_account_id,omitempty"`
	PeriodDate       string    `json:"period_date"`
	ActualCost       float64   `json:"actual_cost"`
	ActualHours      float64   `json:"actual_hours"`
	EarnedValue      float64   `json:"earned_value"`
	ProgressPct      float64   `json:"progress_pct"`
	PhysicalPct      *float64  `json:"physical_pct,omitempty"`
	DataSource       string    `json:"data_source"`
	SourceID         *string   `json:"source_id,omitempty"`
	RecordedAt       time.Time `json:"recorded_at"`
	CreatedAt        time.Time `json:"created_at"`
}

type EVMMetric struct {
	ID               string    `json:"id"`
	ProjectID        string    `json:"project_id"`
	ControlAccountID *string   `json:"control_account_id,omitempty"`
	PeriodDate       string    `json:"period_date"`
	PV               float64   `json:"pv"`
	EV               float64   `json:"ev"`
	AC               float64   `json:"ac"`
	BAC              float64   `json:"bac"`
	SV               float64   `json:"sv"`
	CV               float64   `json:"cv"`
	SVPct            float64   `json:"sv_pct"`
	CVPct            float64   `json:"cv_pct"`
	SPI              float64   `json:"spi"`
	CPI              float64   `json:"cpi"`
	EAC              float64   `json:"eac"`
	ETC              float64   `json:"etc"`
	VAC              float64   `json:"vac"`
	TCPI             float64   `json:"tcpi"`
	MetricScope      string    `json:"metric_scope"`
	IsCumulative     bool      `json:"is_cumulative"`
	CalculatedAt     time.Time `json:"calculated_at"`
	CreatedAt        time.Time `json:"created_at"`
}

type EVMForecast struct {
	ID               string     `json:"id"`
	ProjectID        string     `json:"project_id"`
	ControlAccountID *string    `json:"control_account_id,omitempty"`
	ForecastDate     string     `json:"forecast_date"`
	ForecastType     string     `json:"forecast_type"`
	Method           string     `json:"method"`
	EACValue         *float64   `json:"eac_value,omitempty"`
	ETCValue         *float64   `json:"etc_value,omitempty"`
	VACValue         *float64   `json:"vac_value,omitempty"`
	CompletionDate   *string    `json:"completion_date,omitempty"`
	ConfidencePct    *float64   `json:"confidence_pct,omitempty"`
	Notes            *string    `json:"notes,omitempty"`
	CreatedBy        *string    `json:"created_by,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

type EVMRule struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	RuleName  string    `json:"rule_name"`
	RuleType  string    `json:"rule_type"`
	Description *string `json:"description,omitempty"`
	WeightPct float64   `json:"weight_pct"`
	Config    *string   `json:"config,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type EVMProject struct {
	ID               string    `json:"id"`
	ProjectID        string    `json:"project_id"`
	EVMEnabled       bool      `json:"evm_enabled"`
	DefaultBaselineID *string  `json:"default_baseline_id,omitempty"`
	ReportingFreq    string    `json:"reporting_freq"`
	Currency         string    `json:"currency"`
	ThresholdSPI     float64   `json:"threshold_spi"`
	ThresholdCPI     float64   `json:"threshold_cpi"`
	ThresholdSVPct   float64   `json:"threshold_sv_pct"`
	ThresholdCVPct   float64   `json:"threshold_cv_pct"`
	Config           *string   `json:"config,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// ============================================================================
// P6 Connector Module (V026)
// ============================================================================

type P6Project struct {
	ID             string     `json:"id"`
	ProjectID      string     `json:"project_id"`
	P6ProjectID    string     `json:"p6_project_id"`
	P6UID          *string    `json:"p6_uid,omitempty"`
	P6ProjectCode  *string    `json:"p6_project_code,omitempty"`
	P6ProjectName  *string    `json:"p6_project_name,omitempty"`
	LastSyncAt     *time.Time `json:"last_sync_at,omitempty"`
	SyncStatus     string     `json:"sync_status"`
	SyncError      *string    `json:"sync_error,omitempty"`
	Config         *string    `json:"config,omitempty"`
	IsActive       bool       `json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type P6WBS struct {
	ID                string    `json:"id"`
	P6ProjectID       string    `json:"p6_project_id"`
	P6WBSID           string    `json:"p6_wbs_id"`
	P6WBSCode         *string   `json:"p6_wbs_code,omitempty"`
	P6WBSName         *string   `json:"p6_wbs_name,omitempty"`
	P6ParentWBSID     *string   `json:"p6_parent_wbs_id,omitempty"`
	Level             int       `json:"level"`
	WBSPath           *string   `json:"wbs_path,omitempty"`
	MappedElementType *string   `json:"mapped_element_type,omitempty"`
	MappedElementID   *string   `json:"mapped_element_id,omitempty"`
	IsActive          bool      `json:"is_active"`
	CreatedAt         time.Time `json:"created_at"`
}

type P6Activity struct {
	ID                  string     `json:"id"`
	P6ProjectID         string     `json:"p6_project_id"`
	P6WBSID             *string    `json:"p6_wbs_id,omitempty"`
	P6ActivityID        string     `json:"p6_activity_id"`
	P6ActivityCode      *string    `json:"p6_activity_code,omitempty"`
	P6ActivityName      *string    `json:"p6_activity_name,omitempty"`
	ActivityType        *string    `json:"activity_type,omitempty"`
	Status              *string    `json:"status,omitempty"`
	PlannedStart        *time.Time `json:"planned_start,omitempty"`
	PlannedFinish       *time.Time `json:"planned_finish,omitempty"`
	ActualStart         *time.Time `json:"actual_start,omitempty"`
	ActualFinish        *time.Time `json:"actual_finish,omitempty"`
	RemainingDuration   *int       `json:"remaining_duration,omitempty"`
	AtCompletionDuration *int      `json:"at_completion_duration,omitempty"`
	PercentComplete     *float64   `json:"percent_complete,omitempty"`
	PhysicalComplete    *float64   `json:"physical_complete,omitempty"`
	DurationType        *string    `json:"duration_type,omitempty"`
	MappedToType        *string    `json:"mapped_to_type,omitempty"`
	MappedElementID     *string    `json:"mapped_element_id,omitempty"`
	IsActive            bool       `json:"is_active"`
	LastSyncAt          *time.Time `json:"last_sync_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
}

type P6Relationship struct {
	ID               string    `json:"id"`
	P6ProjectID      string    `json:"p6_project_id"`
	PredecessorID    string    `json:"predecessor_id"`
	SuccessorID      string    `json:"successor_id"`
	RelationshipType string    `json:"relationship_type"`
	LagDays          int       `json:"lag_days"`
	LagType          string    `json:"lag_type"`
	IsActive         bool      `json:"is_active"`
	CreatedAt        time.Time `json:"created_at"`
}

type P6Resource struct {
	ID              string    `json:"id"`
	P6ProjectID     string    `json:"p6_project_id"`
	P6ResourceID    string    `json:"p6_resource_id"`
	P6ResourceName  *string   `json:"p6_resource_name,omitempty"`
	ResourceType    *string   `json:"resource_type,omitempty"`
	UnitOfMeasure   *string   `json:"unit_of_measure,omitempty"`
	UnitPrice       *float64  `json:"unit_price,omitempty"`
	Currency        string    `json:"currency"`
	MappedToType    *string   `json:"mapped_to_type,omitempty"`
	MappedElementID *string   `json:"mapped_element_id,omitempty"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
}

type P6SyncLog struct {
	ID               string     `json:"id"`
	ProjectID        *string    `json:"project_id,omitempty"`
	P6ProjectID      *string    `json:"p6_project_id,omitempty"`
	SyncType         string     `json:"sync_type"`
	Status           string     `json:"status"`
	StartedAt        time.Time  `json:"started_at"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
	DurationSec      *int       `json:"duration_sec,omitempty"`
	RecordsProcessed int        `json:"records_processed"`
	RecordsCreated   int        `json:"records_created"`
	RecordsUpdated   int        `json:"records_updated"`
	RecordsDeleted   int        `json:"records_deleted"`
	SyncFile         *string    `json:"sync_file,omitempty"`
	ErrorMessage     *string    `json:"error_message,omitempty"`
	Details          *string    `json:"details,omitempty"`
}

// ============================================================================
// Common types
// ============================================================================

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Total   *int        `json:"total,omitempty"`
}

type PaginationParams struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// Nullable helpers for scanning
func NewNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}

func NewNullInt64(i *int64) sql.NullInt64 {
	if i == nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: *i, Valid: true}
}

func NewNullFloat64(f *float64) sql.NullFloat64 {
	if f == nil {
		return sql.NullFloat64{Valid: false}
	}
	return sql.NullFloat64{Float64: *f, Valid: true}
}

func NewNullBool(b *bool) sql.NullBool {
	if b == nil {
		return sql.NullBool{Valid: false}
	}
	return sql.NullBool{Bool: *b, Valid: true}
}
