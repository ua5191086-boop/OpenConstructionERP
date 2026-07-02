#!/bin/bash
# E2E API Tests for OpenConstructionERP
# Tests all major API endpoints

API_BASE="${API_URL:-http://localhost:8085}"
PASS=0
FAIL=0

green() { echo -e "\033[32m✓ $1\033[0m"; ((PASS++)); }
red() { echo -e "\033[31m✗ $1\033[0m"; ((FAIL++)); }

echo "=== OpenConstructionERP E2E Tests ==="
echo "API: $API_BASE"
echo ""

# 1. Health check
echo "--- Health Check ---"
HEALTH=$(curl -sf "$API_BASE/health" 2>/dev/null)
if [ $? -eq 0 ] && echo "$HEALTH" | grep -q '"status":"ok"'; then
  green "Health check returned OK"
else
  red "Health check failed"
fi

# 2. List projects
echo "--- Projects ---"
PROJECTS=$(curl -sf "$API_BASE/api/v1/projects" 2>/dev/null)
if [ $? -eq 0 ]; then
  COUNT=$(echo "$PROJECTS" | python3 -c "import sys,json; data=json.load(sys.stdin); print(len(data) if isinstance(data,list) else 0)" 2>/dev/null)
  if [ "$COUNT" -gt 0 ]; then
    green "Listed $COUNT projects"
  else
    red "No projects returned (expected at least 1 from seed data)"
  fi
else
  red "List projects failed"
fi

# 3. List BOQ items
echo "--- BOQ ---"
BOQ=$(curl -sf "$API_BASE/api/v1/boq-items?project_id=all" 2>/dev/null)
if [ $? -eq 0 ]; then
  green "BOQ items accessible"
else
  red "BOQ items failed"
fi

# 4. List contracts
echo "--- Contracts ---"
CONTRACTS=$(curl -sf "$API_BASE/api/v1/contracts?project_id=all" 2>/dev/null)
if [ $? -eq 0 ]; then
  green "Contracts accessible"
else
  red "Contracts failed"
fi

# 5. List TBM
echo "--- TBM ---"
TBM=$(curl -sf "$API_BASE/api/v1/tbm?project_id=all" 2>/dev/null)
if [ $? -eq 0 ]; then
  green "TBM accessible"
else
  red "TBM failed"
fi

# 6. List HSE incidents
echo "--- HSE ---"
HSE=$(curl -sf "$API_BASE/api/v1/hse-incidents?project_id=all" 2>/dev/null)
if [ $? -eq 0 ]; then
  green "HSE incidents accessible"
else
  red "HSE incidents failed"
fi

# 7. List NCRs
echo "--- NCRs ---"
NCR=$(curl -sf "$API_BASE/api/v1/ncrs?project_id=all" 2>/dev/null)
if [ $? -eq 0 ]; then
  green "NCRs accessible"
else
  red "NCRs failed"
fi

# 8. List risks
echo "--- Risks ---"
RISKS=$(curl -sf "$API_BASE/api/v1/risks?project_id=all" 2>/dev/null)
if [ $? -eq 0 ]; then
  green "Risks accessible"
else
  red "Risks failed"
fi

# 9. List change orders
echo "--- Change Orders ---"
CO=$(curl -sf "$API_BASE/api/v1/change-orders?project_id=all" 2>/dev/null)
if [ $? -eq 0 ]; then
  green "Change orders accessible"
else
  red "Change orders failed"
fi

# 10. List milestones
echo "--- Milestones ---"
MS=$(curl -sf "$API_BASE/api/v1/milestones?project_id=all" 2>/dev/null)
if [ $? -eq 0 ]; then
  green "Milestones accessible"
else
  red "Milestones failed"
fi

# 11. List documents
echo "--- Documents ---"
DOCS=$(curl -sf "$API_BASE/api/v1/documents?project_id=all" 2>/dev/null)
if [ $? -eq 0 ]; then
  green "Documents accessible"
else
  red "Documents failed"
fi

# 12. List equipment
echo "--- Equipment ---"
EQ=$(curl -sf "$API_BASE/api/v1/equipment?project_id=all" 2>/dev/null)
if [ $? -eq 0 ]; then
  green "Equipment accessible"
else
  red "Equipment failed"
fi

# 13. List stakeholders
echo "--- Stakeholders ---"
SH=$(curl -sf "$API_BASE/api/v1/stakeholders?project_id=all" 2>/dev/null)
if [ $? -eq 0 ]; then
  green "Stakeholders accessible"
else
  red "Stakeholders failed"
fi

# 14. List employees
echo "--- Employees ---"
EMP=$(curl -sf "$API_BASE/api/v1/employees" 2>/dev/null)
if [ $? -eq 0 ]; then
  green "Employees accessible"
else
  red "Employees failed"
fi

# 15. List organizations
echo "--- Organizations ---"
ORG=$(curl -sf "$API_BASE/api/v1/organizations" 2>/dev/null)
if [ $? -eq 0 ]; then
  green "Organizations accessible"
else
  red "Organizations failed"
fi

# 16. EVM
echo "--- EVM ---"
EVM=$(curl -sf "$API_BASE/api/v1/evm?project_id=all" 2>/dev/null)
if [ $? -eq 0 ]; then
  green "EVM accessible"
else
  red "EVM failed"
fi

# 17. Budget
echo "--- Budget ---"
BUDGET=$(curl -sf "$API_BASE/api/v1/budget?project_id=all" 2>/dev/null)
if [ $? -eq 0 ]; then
  green "Budget accessible"
else
  red "Budget failed"
fi

# 18. WBS
echo "--- WBS ---"
WBS=$(curl -sf "$API_BASE/api/v1/wbs?project_id=all" 2>/dev/null)
if [ $? -eq 0 ]; then
  green "WBS accessible"
else
  red "WBS failed"
fi

# 19. Lessons Learned
echo "--- Lessons Learned ---"
LL=$(curl -sf "$API_BASE/api/v1/lessons-learned?project_id=all" 2>/dev/null)
if [ $? -eq 0 ]; then
  green "Lessons Learned accessible"
else
  red "Lessons Learned failed"
fi

# 20. Physical Progress
echo "--- Physical Progress ---"
PP=$(curl -sf "$API_BASE/api/v1/physical-progress?project_id=all" 2>/dev/null)
if [ $? -eq 0 ]; then
  green "Physical Progress accessible"
else
  red "Physical Progress failed"
fi

# 21. Cost Transactions
echo "--- Cost Transactions ---"
CT=$(curl -sf "$API_BASE/api/v1/cost-transactions?project_id=all" 2>/dev/null)
if [ $? -eq 0 ]; then
  green "Cost Transactions accessible"
else
  red "Cost Transactions failed"
fi

# 22. Settlement Monitoring
echo "--- Settlement Monitoring ---"
SM=$(curl -sf "$API_BASE/api/v1/settlement-monitoring?project_id=all" 2>/dev/null)
if [ $? -eq 0 ]; then
  green "Settlement Monitoring accessible"
else
  red "Settlement Monitoring failed"
fi

# 23. TBM Telemetry
echo "--- TBM Telemetry ---"
TT=$(curl -sf "$API_BASE/api/v1/tbm-telemetry?tbm_id=all" 2>/dev/null)
if [ $? -eq 0 ]; then
  green "TBM Telemetry accessible"
else
  red "TBM Telemetry failed"
fi

# 24. Tunnel Rings
echo "--- Tunnel Rings ---"
TR=$(curl -sf "$API_BASE/api/v1/tunnel-rings?drive_id=all" 2>/dev/null)
if [ $? -eq 0 ]; then
  green "Tunnel Rings accessible"
else
  red "Tunnel Rings failed"
fi

echo ""
echo "=== Results: $PASS passed, $FAIL failed ==="
exit $FAIL
