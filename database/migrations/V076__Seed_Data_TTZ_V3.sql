-- ============================================================
-- V076: Seed Data — TTZ-V3 Demo Project
-- Uses CTE with RETURNING for cross-table references
-- ============================================================
BEGIN;

-- ============================================================
-- 1. ORGANIZATIONS
-- ============================================================
WITH org AS (
  INSERT INTO organizations (code, name, org_type, country, is_active) VALUES
    ('ALMATY-METRO', 'АО Метрополитен Алматы', 'client', 'KZ', true),
    ('TUNNEL-JV', 'ТОО ТоннельСтрой Сервис', 'contractor', 'KZ', true),
    ('DESIGN-INST', 'АО КазНИИПИ Транспорт', 'consultant', 'KZ', true),
    ('SUPERVISION', 'ТОО ТехНадзор', 'consultant', 'KZ', true)
  RETURNING id, code
),
-- ============================================================
-- 2. USERS (for created_by references)
-- ============================================================
u AS (
  INSERT INTO users (email, full_name, organization_id, principal_type, is_active)
  SELECT 'pm@openconstructionerp.com', 'Иванов Иван Иванович', org.id, 'human', true
  FROM org WHERE org.code='TUNNEL-JV'
  RETURNING id, email
),
-- ============================================================
-- 3. EMPLOYEES
-- ============================================================
emp AS (
  INSERT INTO employees (employee_code, full_name, first_name, last_name, position, department, position_type, status, hire_date, salary_currency) VALUES
    ('EMP-001', 'Иванов Иван Иванович', 'Иван', 'Иванов', 'Project Manager', 'Управление проектами', 'full_time', 'active', '2024-01-15', 'KZT'),
    ('EMP-002', 'Петров Петр Петрович', 'Петр', 'Петров', 'Site Engineer', 'Тоннельное производство', 'full_time', 'active', '2024-03-01', 'KZT'),
    ('EMP-003', 'Сидоров Алексей Викторович', 'Алексей', 'Сидоров', 'Supervisor', 'Технадзор', 'full_time', 'active', '2024-02-01', 'KZT'),
    ('EMP-004', 'Козлова Мария Сергеевна', 'Мария', 'Козлова', 'HSE Engineer', 'Охрана труда', 'full_time', 'active', '2024-04-01', 'KZT'),
    ('EMP-005', 'Нурланов Бекжан Аскарович', 'Бекжан', 'Нурланов', 'TBM Operator', 'Тоннельное производство', 'full_time', 'active', '2024-06-01', 'KZT')
  RETURNING id, employee_code
),
-- ============================================================
-- 3. PROJECTS
-- ============================================================
proj AS (
  INSERT INTO projects (code, name, name_ru, organization_id, client_id, project_manager_id, project_type, status, country, currency, start_date, finish_date, contract_value, priority)
  SELECT 'TTZ-V3', 'Tunnel West-East — Section V3', 'Тоннель Запад-Восток — Участок V3', o1.id, o2.id, u.id, 'metro', 'execution', 'KZ', 'KZT', '2025-01-15', '2028-06-30', 285000000.00, 'high'
  FROM org o1, org o2, u WHERE o1.code='TUNNEL-JV' AND o2.code='ALMATY-METRO'
  RETURNING id, code
),
-- ============================================================
-- 4. BOQ SECTIONS, COMPLEXES, OBJECTS, CBS
-- ============================================================
bs AS (
  INSERT INTO boq_sections (project_id, code, name, section_type, sort_order)
  SELECT id, 'SEC-01', 'Основной тоннель TBM', 'Track', 1 FROM proj
  UNION ALL SELECT id, 'SEC-02', 'НАТМ участки', 'Track', 2 FROM proj
  UNION ALL SELECT id, 'SEC-03', 'Станционные сооружения', 'Station', 3 FROM proj
  RETURNING id, code
),
bc AS (
  INSERT INTO boq_complexes (project_id, section_id, code, name, sort_order)
  SELECT p.id, bs.id, 'CMP-01', 'TBM проходка D=6.3м', 1
  FROM proj p, bs WHERE bs.code='SEC-01'
  UNION ALL
  SELECT p.id, bs.id, 'CMP-02', 'Обделка тоннеля', 2
  FROM proj p, bs WHERE bs.code='SEC-01'
  UNION ALL
  SELECT p.id, bs.id, 'CMP-03', 'НАТМ разработка', 3
  FROM proj p, bs WHERE bs.code='SEC-02'
  RETURNING id, code
),
bo AS (
  INSERT INTO boq_objects (project_id, complex_id, code, name, sort_order)
  SELECT p.id, bc.id, 'OBJ-01', 'Разработка грунта TBM', 1
  FROM proj p, bc WHERE bc.code='CMP-01'
  UNION ALL
  SELECT p.id, bc.id, 'OBJ-02', 'Монтаж колец обделки', 2
  FROM proj p, bc WHERE bc.code='CMP-02'
  UNION ALL
  SELECT p.id, bc.id, 'OBJ-03', 'Тампонаж заобделочного пространства', 3
  FROM proj p, bc WHERE bc.code='CMP-02'
  UNION ALL
  SELECT p.id, bc.id, 'OBJ-04', 'Разработка NATM', 4
  FROM proj p, bc WHERE bc.code='CMP-03'
  RETURNING id, code
),
cbs AS (
  INSERT INTO cbs_chapters (project_id, code, name, name_ru, level, sort_order)
  SELECT id, '01', 'Подготовительные работы', 'Подготовительные работы', 1, 1 FROM proj
  UNION ALL SELECT id, '02', 'Тоннельные работы', 'Тоннельные работы', 1, 2 FROM proj
  UNION ALL SELECT id, '03', 'НАТМ работы', 'НАТМ работы', 1, 3 FROM proj
  UNION ALL SELECT id, '04', 'Вентиляция и сантехника', 'Вентиляция и сантехника', 1, 4 FROM proj
  UNION ALL SELECT id, '05', 'Электрика и автоматика', 'Электрика и автоматика', 1, 5 FROM proj
  UNION ALL SELECT id, '06', 'Пусконаладочные работы', 'Пусконаладочные работы', 1, 6 FROM proj
  UNION ALL SELECT id, '07', 'Резерв', 'Непредвиденные работы', 1, 7 FROM proj
  RETURNING id, code
),
-- ============================================================
-- 5. CONTRACTS
-- ============================================================
con AS (
  INSERT INTO contracts (project_id, code, name, contract_type, contract_form, client_id, contractor_id, currency, contract_value, signed_at, status, start_date, end_date, retention_pct, penalty_rate_daily, payment_terms_type)
  SELECT p.id, 'TTZ-V3-CON-001', 'Основной EPC контракт — Тоннель V3', 'main', 'epc', o1.id, o2.id, 'KZT', 285000000.00, '2025-01-15', 'active', '2025-01-15', '2028-06-30', 5.0, 0.05, 'monthly'
  FROM proj p, org o1, org o2 WHERE o1.code='ALMATY-METRO' AND o2.code='TUNNEL-JV'
  RETURNING id, code
),
-- ============================================================
-- 6. BOQ ITEMS
-- ============================================================
boq AS (
  INSERT INTO boq_items (project_id, object_id, cbs_chapter_id, code, name, description, unit, quantity, unit_price, currency, contract_id, status)
  SELECT p.id, bo.id, cbs.id, '02.01', 'Разработка грунта TBM (EPB)', 'Разработка грунта EPB щитом D=6.3м', 'm3', 72000, 850.00, 'KZT', con.id, 'active'
  FROM proj p, bo, cbs, con WHERE bo.code='OBJ-01' AND cbs.code='02'
  UNION ALL
  SELECT p.id, bo.id, cbs.id, '02.02', 'Монтаж ж/б колец обделки', 'Монтаж ж/б колец обделки, внутр. D=5.6м', 'ring', 1534, 45000.00, 'KZT', con.id, 'active'
  FROM proj p, bo, cbs, con WHERE bo.code='OBJ-02' AND cbs.code='02'
  UNION ALL
  SELECT p.id, bo.id, cbs.id, '02.03', 'Тампонаж заобделочного пространства', 'Цементация заобделочного пространства', 'm3', 11500, 3200.00, 'KZT', con.id, 'active'
  FROM proj p, bo, cbs, con WHERE bo.code='OBJ-03' AND cbs.code='02'
  UNION ALL
  SELECT p.id, bo.id, cbs.id, '03.01', 'Разработка грунта NATM', 'Разработка NATM калотта+штросса', 'm3', 18000, 1200.00, 'KZT', con.id, 'active'
  FROM proj p, bo, cbs, con WHERE bo.code='OBJ-04' AND cbs.code='03'
  UNION ALL
  SELECT p.id, bo.id, cbs.id, '03.02', 'Набрызг-бетон (SFRS)', 'Набрызг-бетон SFRS t=250mm', 'm3', 4500, 4200.00, 'KZT', con.id, 'active'
  FROM proj p, bo, cbs, con WHERE bo.code='OBJ-04' AND cbs.code='03'
  UNION ALL
  SELECT p.id, bo.id, cbs.id, '01.01', 'Подготовительные работы', 'Строительная площадка, бытовки, ограждение', 'lump', 1, 8500000.00, 'KZT', con.id, 'active'
  FROM proj p, bo, cbs, con WHERE bo.code='OBJ-01' AND cbs.code='01'
  RETURNING id, code
),
-- ============================================================
-- 7. TBM
-- ============================================================
tbm AS (
  INSERT INTO tbm (project_id, code, manufacturer, model, tbm_type, diameter_mm, commissioning_date, status)
  SELECT id, 'TBM-01', 'Herrenknecht AG', 'S-1200', 'EPB', 6300, '2025-03-01', 'boring' FROM proj
  RETURNING id, code
),
-- ============================================================
-- 8. TUNNEL DRIVES
-- ============================================================
drv AS (
  INSERT INTO tunnel_drives (project_id, tbm_id, code, name, method, chainage_from, chainage_to, ring_width_mm, design_rings, status, started_at)
  SELECT p.id, tbm.id, 'DRV-01', 'Основная проходка TBM — левый путь', 'TBM', 0, 2300, 1500, 1534, 'boring', '2025-03-20'
  FROM proj p, tbm
  RETURNING id, code
),
-- ============================================================
-- 9. TUNNEL RINGS (100)
-- ============================================================
rings AS (
  INSERT INTO tunnel_rings (drive_id, ring_no, chainage, built_at, shift, ring_type, key_position, advance_mm, grout_volume_m3, grout_pressure_bar)
  SELECT
    drv.id, n, (n-1) * 1.5,
    CASE WHEN n <= 80 THEN '2025-06-01'::date + (n || ' days')::interval ELSE '2025-09-01'::date + ((n-80) || ' days')::interval END,
    CASE WHEN n % 2 = 0 THEN 'day' ELSE 'night' END,
    CASE WHEN n % 3 = 0 THEN 'universal' WHEN n % 3 = 1 THEN 'left' ELSE 'right' END,
    CASE WHEN n % 4 = 0 THEN 0 WHEN n % 4 = 1 THEN 90 WHEN n % 4 = 2 THEN 180 ELSE 270 END,
    1500 + (random()*50)::int,
    6.5 + random()*3.0,
    1.5 + random()*1.0
  FROM drv, generate_series(1, 100) AS n
  RETURNING id, ring_no
),
-- ============================================================
-- 10. EQUIPMENT
-- ============================================================
equip AS (
  INSERT INTO equipment (project_id, equipment_code, equipment_name, equipment_type, manufacturer, model, status, location, purchase_cost, hourly_rate)
  SELECT p.id, 'LOCO-01', 'Locomotive PL-60 #1', 'heavy_machine', 'Paus', 'PL-60', 'in_use', 'Забой TBM', 45000000.00, 8500.00 FROM proj p
  UNION ALL SELECT p.id, 'LOCO-02', 'Locomotive PL-60 #2', 'heavy_machine', 'Paus', 'PL-60', 'available', 'Депо', 45000000.00, 8500.00 FROM proj p
  UNION ALL SELECT p.id, 'CRANE-01', 'Gantry crane 20t', 'crane', 'DEMAG', 'DC-20', 'in_use', 'Завод ЖБИ', 28000000.00, 12000.00 FROM proj p
  UNION ALL SELECT p.id, 'PUMP-01', 'Grout pump P-200', 'pump', 'Putzmeister', 'P-200', 'in_use', 'TBM хвост', 8500000.00, 3500.00 FROM proj p
  UNION ALL SELECT p.id, 'GENER-01', 'Diesel generator 500kVA', 'generator', 'Caterpillar', 'C18', 'available', 'Площадка', 12000000.00, 5000.00 FROM proj p
  RETURNING id, equipment_code
),
-- ============================================================
-- 11. HSE INCIDENTS
-- ============================================================
hse AS (
  INSERT INTO hse_incidents (project_id, incident_number, incident_code, title, incident_type, severity, incident_date, location, description, reported_by, status, affected_employees, lost_days, medical_cost, property_cost)
  SELECT p.id, 'HSE-2025-001', 'HSE-001', 'Падение породы в забое', 'near_miss', 'low', '2025-03-15 14:30:00+06'::timestamptz, 'PK 12+450', 'Падение породы с кровли выработки в забое TBM. Персонал эвакуирован, пострадавших нет.', emp.id, 'closed', 0, 0, 0, 0
  FROM proj p, emp WHERE emp.employee_code='EMP-002'
  UNION ALL
  SELECT p.id, 'HSE-2025-002', 'HSE-002', 'Травма руки при монтаже кольца', 'injury', 'medium', '2025-05-20 09:15:00+06'::timestamptz, 'PK 12+800', 'Травма левой кисти при установке сегмента кольца. 3 дня больничного.', emp.id, 'closed', 1, 3, 45000.00, 0
  FROM proj p, emp WHERE emp.employee_code='EMP-002'
  UNION ALL
  SELECT p.id, 'HSE-2025-003', 'HSE-003', 'Повреждение сегмента при транспортировке', 'property_damage', 'low', '2025-07-01 11:00:00+06'::timestamptz, 'Завод ЖБИ', 'Повреждение угла сегмента при разгрузке с трейлера. Сегмент забракован.', emp.id, 'open', 0, 0, 0, 120000.00
  FROM proj p, emp WHERE emp.employee_code='EMP-002'
  RETURNING id, incident_code
),
-- ============================================================
-- 12. NCRs
-- ============================================================
ncr AS (
  INSERT INTO ncrs (project_id, number, code, title, description, severity, location, boq_item_id, raised_by, status)
  SELECT p.id, 1, 'NCR-001', 'Отклонение геометрии кольца #42', 'Зазор между сегментами 8мм при допуске 5мм.', 'minor', 'PK 12+563', boq.id, emp.employee_code, 'open'
  FROM proj p, boq, emp WHERE boq.code='02.02' AND emp.employee_code='EMP-003'
  UNION ALL
  SELECT p.id, 2, 'NCR-002', 'Прочность бетона сегментов ниже проектной', 'Прочность бетона сегментов партии B-023: B35 вместо B40.', 'major', 'Завод ЖБИ', boq.id, emp.employee_code, 'open'
  FROM proj p, boq, emp WHERE boq.code='02.02' AND emp.employee_code='EMP-003'
  UNION ALL
  SELECT p.id, 3, 'NCR-003', 'Несоответствие толщины набрызг-бетона', 'Толщина набрызг-бетона на NATM участке: 200мм вместо 250мм.', 'minor', 'PK 13+200', boq.id, emp.employee_code, 'closed'
  FROM proj p, boq, emp WHERE boq.code='03.02' AND emp.employee_code='EMP-003'
  RETURNING id, code
),
-- ============================================================
-- 13. PROJECT MILESTONES
-- ============================================================
ms AS (
  INSERT INTO project_milestones (project_id, milestone_code, name, description, milestone_type, category, planned_date, forecast_date, actual_date, status, weight_pct, is_gate)
  SELECT p.id, 'MS-01', 'Мобилизация TBM', 'Доставка, сборка и наладка TBM', 'mobilization', 'technical', '2025-03-01'::date, '2025-03-01'::date, '2025-03-05'::date, 'completed', 5.0, true FROM proj p
  UNION ALL SELECT p.id, 'MS-02', 'Начало проходки TBM', 'Первый метр проходки TBM', 'construction', 'technical', '2025-03-15'::date, '2025-03-20'::date, '2025-03-20'::date, 'completed', 5.0, true FROM proj p
  UNION ALL SELECT p.id, 'MS-03', 'Кольцо #500', 'Монтаж 500-го кольца обделки', 'construction', 'technical', '2025-10-15'::date, '2025-11-01'::date, NULL, 'in_progress', 15.0, false FROM proj p
  UNION ALL SELECT p.id, 'MS-04', 'Кольцо #1000', 'Монтаж 1000-го кольца обделки', 'construction', 'technical', '2026-05-01'::date, '2026-06-15'::date, NULL, 'planned', 20.0, false FROM proj p
  UNION ALL SELECT p.id, 'MS-05', 'Прорыв TBM', 'Выход TBM на приемную камеру', 'construction', 'technical', '2027-09-01'::date, '2027-12-01'::date, NULL, 'planned', 25.0, true FROM proj p
  UNION ALL SELECT p.id, 'MS-06', 'Завершение отделочных работ', 'Полное завершение отделочных работ', 'finishing', 'technical', '2028-04-01'::date, '2028-05-01'::date, NULL, 'planned', 15.0, false FROM proj p
  UNION ALL SELECT p.id, 'MS-07', 'Ввод в эксплуатацию', 'Подписание акта ввода в эксплуатацию', 'commissioning', 'contractual', '2028-06-30'::date, '2028-08-01'::date, NULL, 'planned', 15.0, true FROM proj p
  RETURNING id, milestone_code
),
-- ============================================================
-- 14. DOCUMENTS
-- ============================================================
doc AS (
  INSERT INTO documents (project_id, doc_number, title, doc_type, discipline, originator, revision, state)
  SELECT p.id, 'TTZ-V3-DWG-001', 'План трассы V3 (M 1:500)', 'DWG', 'GE', 'DESIGN-INST', 'C', 'Published' FROM proj p
  UNION ALL SELECT p.id, 'TTZ-V3-SPC-001', 'Спецификация колец обделки', 'SPC', 'ST', 'DESIGN-INST', 'B', 'Published' FROM proj p
  UNION ALL SELECT p.id, 'TTZ-V3-RPT-001', 'Отчёт о геологических изысканиях', 'RPT', 'GE', 'DESIGN-INST', 'D', 'Published' FROM proj p
  UNION ALL SELECT p.id, 'TTZ-V3-PLN-001', 'План производства работ (PPR)', 'PLN', 'PM', 'TUNNEL-JV', 'A', 'Published' FROM proj p
  RETURNING id, doc_number
),
-- ============================================================
-- 15. PROJECT RISKS
-- ============================================================
risk AS (
  INSERT INTO project_risks (project_id, risk_code, name, description, risk_category, risk_type, probability, impact, probability_score, impact_score, potential_cost, mitigation_strategy, mitigation_plan, status)
  SELECT p.id, 'R-001', 'Встреча плывуна в зоне TBM', 'Водонасыщенные грунты на PK 14+200 — PK 14+500', 'geological', 'technical', 'possible', 'major', 4, 5, 45000000.00, 'mitigate', 'Дополнительное геологическое бурение, пенный режим TBM', 'open' FROM proj p
  UNION ALL SELECT p.id, 'R-002', 'Задержка поставки колец', 'Завод ЖБИ может не обеспечить своевременную поставку', 'supply_chain', 'schedule', 'likely', 'moderate', 4, 3, 28000000.00, 'mitigate', 'Страховой запас 50 колец, альтернативный поставщик', 'open' FROM proj p
  UNION ALL SELECT p.id, 'R-003', 'Износ режущего инструмента TBM', 'Абразивный износ резцов, замена в забое', 'technical', 'technical', 'possible', 'moderate', 3, 3, 15000000.00, 'mitigate', 'Плановая замена резцов на PK 13+500', 'open' FROM proj p
  UNION ALL SELECT p.id, 'R-004', 'Превышение бюджета по статье бетон', 'Рост цен на цемент и заполнители', 'financial', 'cost', 'likely', 'minor', 4, 2, 12000000.00, 'accept', 'Мониторинг цен, пересмотр контракта', 'open' FROM proj p
  RETURNING id, risk_code
),
-- ============================================================
-- 16. CHANGE ORDERS
-- ============================================================
co AS (
  INSERT INTO change_orders (project_id, co_number, co_code, co_name, co_type, scope_change, cost_change, cost_currency, schedule_change_days, status)
  SELECT p.id, 1, 'CO-001', 'Дополнительное укрепление NATM участка', 'variation', 'Увеличение толщины набрызг-бетона на 30мм на PK 13+800 — PK 14+200', 3200000.00, 'KZT', 15, 'approved' FROM proj p
  UNION ALL SELECT p.id, 2, 'CO-002', 'Дополнительная вентиляционная шахта', 'variation', 'Устройство дополнительной вентшахты на PK 13+000', 5800000.00, 'KZT', 30, 'submitted' FROM proj p
  RETURNING id, co_code
),
-- ============================================================
-- 17. PHYSICAL PROGRESS (IPC)
-- ============================================================
pp AS (
  INSERT INTO physical_progress (project_id, contract_id, boq_item_id, measurement_date, item_code, description, unit, contract_quantity, prev_cumulative_qty, current_qty, total_cumulative_qty, completion_pct, unit_price, ipc_amount)
  SELECT p.id, con.id, boq.id, '2025-06-30'::date, '02.01', 'Разработка грунта TBM', 'm3', 72000, 0, 10800, 10800, 15.0, 850.00, 9180000.00
  FROM proj p, con, boq WHERE boq.code='02.01'
  UNION ALL
  SELECT p.id, con.id, boq.id, '2025-06-30'::date, '02.02', 'Монтаж колец обделки', 'ring', 1534, 0, 80, 80, 5.2, 45000.00, 3600000.00
  FROM proj p, con, boq WHERE boq.code='02.02'
  UNION ALL
  SELECT p.id, con.id, boq.id, '2025-06-30'::date, '02.03', 'Тампонаж заобделочного пространства', 'm3', 11500, 0, 600, 600, 5.2, 3200.00, 1920000.00
  FROM proj p, con, boq WHERE boq.code='02.03'
  RETURNING id, item_code
),
-- ============================================================
-- 18. EVM
-- ============================================================
evm_ca AS (
  INSERT INTO evm_control_accounts (project_id, ca_code, ca_name, wbs_code, responsible)
  SELECT p.id, 'CA-TBM-01', 'TBM Проходка', '02.01', 'Петров П.П.' FROM proj p
  UNION ALL SELECT p.id, 'CA-RING-01', 'Монтаж обделки', '02.02', 'Петров П.П.' FROM proj p
  RETURNING id, ca_code
),
evm_ac AS (
  INSERT INTO evm_actuals (project_id, control_account_id, period_date, actual_cost, earned_value, progress_pct)
  SELECT p.id, evm_ca.id, '2025-03-31'::date, 8100000.00, 8200000.00, 11.4
  FROM proj p, evm_ca WHERE evm_ca.ca_code='CA-TBM-01'
  UNION ALL
  SELECT p.id, evm_ca.id, '2025-06-30'::date, 27500000.00, 26200000.00, 36.4
  FROM proj p, evm_ca WHERE evm_ca.ca_code='CA-TBM-01'
  UNION ALL
  SELECT p.id, evm_ca.id, '2025-06-30'::date, 3200000.00, 3600000.00, 5.2
  FROM proj p, evm_ca WHERE evm_ca.ca_code='CA-RING-01'
  RETURNING id
),
-- ============================================================
-- 19. STAKEHOLDERS
-- ============================================================
sh AS (
  INSERT INTO stakeholders (project_id, stakeholder_type, name, organization, contact_person, interest_level, influence_level, status)
  SELECT p.id, 'sponsor', 'Акимат г. Алматы', 'Акимат г. Алматы', 'Заместитель акима по строительству', 'high', 'high', 'active' FROM proj p
  UNION ALL SELECT p.id, 'regulator', 'Управление строительства', 'Управление строительства г. Алматы', 'Начальник управления', 'medium', 'high', 'active' FROM proj p
  UNION ALL SELECT p.id, 'community', 'Жители мкр. Шугыла', 'КСК Шугыла', 'Председатель КСК', 'high', 'low', 'active' FROM proj p
  RETURNING id, name
),
-- ============================================================
-- 20. SETTLEMENT MONITORING
-- ============================================================
smp AS (
  INSERT INTO settlement_monitoring_points (project_id, point_code, point_name, point_type, chainage_m, offset_m, initial_level_m, trigger_alert_mm, trigger_urgent_mm, status)
  SELECT p.id, 'SP-' || LPAD(n::text, 3, '0'), 'Точка осадки PK ' || (n * 25),
    CASE WHEN n % 5 = 0 THEN 'building' WHEN n % 3 = 0 THEN 'subsurface' ELSE 'surface' END,
    n * 25.0, CASE WHEN n % 2 = 0 THEN 5.0 ELSE -5.0 END,
    1000.0 + (random()*5 - 2.5)::numeric(8,3),
    10.0, 25.0, 'active'
  FROM proj p, generate_series(1, 20) AS n
  RETURNING id, point_code, initial_level_m
),
sr AS (
  INSERT INTO settlement_readings (point_id, reading_time, level_m, settlement_mm, rate_mm_per_day, is_alert, is_urgent)
  SELECT smp.id, '2025-04-01'::date + (n || ' days')::interval,
    smp.initial_level_m - (random()*12)::numeric(8,3),
    (random()*12)::numeric(8,2),
    (random()*1.5)::numeric(6,2),
    random() < 0.05, random() < 0.01
  FROM smp, generate_series(0, 90, 7) AS n
  WHERE smp.point_code LIKE 'SP-%'
  LIMIT 200
  RETURNING id
),
-- ============================================================
-- 21. BUDGET VERSIONS
-- ============================================================
bv AS (
  INSERT INTO budget_versions (project_id, version_number, version_name, status, total_amount, notes)
  SELECT p.id, 1, 'Initial Budget V1', 'approved', 285000000.00, 'Утверждённый бюджет на 2025-2028' FROM proj p
  UNION ALL SELECT p.id, 2, 'Revised Budget V2', 'approved', 288200000.00, 'Бюджет с учётом CO-001' FROM proj p
  RETURNING id, version_number
),
-- ============================================================
-- 22. WBS ITEMS
-- ============================================================
wbs AS (
  INSERT INTO wbs_items (project_id, wbs_code, name, wbs_level, is_leaf, planned_start, planned_end, planned_cost, status)
  SELECT p.id, '01', 'Подготовительные работы', 1, false, '2025-01-15'::date, '2025-03-01'::date, 8500000.00, 'completed' FROM proj p
  UNION ALL SELECT p.id, '02', 'Тоннельные работы (TBM)', 1, false, '2025-03-01'::date, '2027-09-01'::date, 185000000.00, 'in_progress' FROM proj p
  UNION ALL SELECT p.id, '02.01', 'Разработка грунта TBM', 2, true, '2025-03-15'::date, '2027-06-01'::date, 61200000.00, 'in_progress' FROM proj p
  UNION ALL SELECT p.id, '02.02', 'Монтаж колец обделки', 2, true, '2025-03-20'::date, '2027-08-01'::date, 69030000.00, 'in_progress' FROM proj p
  UNION ALL SELECT p.id, '03', 'НАТМ работы', 1, false, '2026-01-01'::date, '2027-06-01'::date, 45000000.00, 'planned' FROM proj p
  UNION ALL SELECT p.id, '04', 'Вентиляция и сантехника', 1, false, '2027-06-01'::date, '2028-01-01'::date, 12500000.00, 'planned' FROM proj p
  UNION ALL SELECT p.id, '05', 'Электрика и автоматика', 1, false, '2027-08-01'::date, '2028-03-01'::date, 18000000.00, 'planned' FROM proj p
  UNION ALL SELECT p.id, '06', 'Пусконаладочные работы', 1, false, '2028-03-01'::date, '2028-06-01'::date, 8000000.00, 'planned' FROM proj p
  RETURNING id, wbs_code
),
-- ============================================================
-- 23. TBM TELEMETRY
-- ============================================================
tbm_tel AS (
  INSERT INTO tbm_telemetry (tbm_id, recorded_at, thrust_force_kn, torque_knm, advance_rate_mmmin, cutterhead_rpm, face_pressure_bar, epb_face_pressure_bar, epb_screw_speed_rpm, thrust_speed_mmmin, total_power_kw)
  SELECT tbm.id, '2025-06-01'::date + (n || ' days')::interval + time '08:00:00',
    18000 + (random()*4000)::int, 2500 + (random()*800)::int,
    35 + (random()*20)::int, 2.5 + random()*1.0,
    0.8 + random()*0.3, 0.8 + random()*0.3,
    8 + (random()*4)::int, 40 + (random()*20)::int,
    200 + (random()*100)::int
  FROM tbm, generate_series(1, 50) AS n
  RETURNING id
),
-- ============================================================
-- 24. COST TRANSACTIONS
-- ============================================================
ct AS (
  INSERT INTO cost_transactions (project_id, boq_item_id, cbs_chapter_id, contract_id, transaction_type, amount, currency, period, description, created_by)
  SELECT p.id, boq.id, cbs.id, con.id, 'Actual', 28500000.00, 'KZT', '2025-02-15'::date, 'Авансовый платёж 10% по контракту TTZ-V3-CON-001', u.id
  FROM proj p, boq, cbs, con, u WHERE boq.code='02.01' AND cbs.code='02'
  UNION ALL
  SELECT p.id, boq.id, cbs.id, con.id, 'Actual', 9180000.00, 'KZT', '2025-04-30'::date, 'Промежуточный платёж №1 — TBM проходка', u.id
  FROM proj p, boq, cbs, con, u WHERE boq.code='02.01' AND cbs.code='02'
  UNION ALL
  SELECT p.id, boq.id, cbs.id, con.id, 'Actual', 3600000.00, 'KZT', '2025-07-31'::date, 'Промежуточный платёж №2 — кольца 1-80', u.id
  FROM proj p, boq, cbs, con, u WHERE boq.code='02.02' AND cbs.code='02'
  RETURNING id
),
-- ============================================================
-- 25. LESSONS LEARNED
-- ============================================================
ll AS (
  INSERT INTO project_lessons (project_id, title, description, category, severity, status, root_cause, impact, recommendation)
  SELECT p.id, 'Усилить геологическое бурение перед TBM', 'На PK 12+400 — PK 12+600 неожиданно встретили водоносный слой, не выявленный при изысканиях.', 'geotechnical', 'high', 'active', 'Недостаточная плотность геологических скважин', 'Задержка проходки на 5 дней, дополнительные расходы 2.5M KZT', 'Бурение опережающих скважин каждые 50м в зонах потенциального риска' FROM proj p
  UNION ALL SELECT p.id, 'Страховой запас колец обязателен', 'Задержка поставки колец на 3 дня из-за поломки формы на заводе ЖБИ.', 'logistics', 'medium', 'active', 'Поломка формы на заводе-поставщике', 'Простой TBM на 1 смену', 'Поддерживать страховой запас не менее 20 колец на площадке' FROM proj p
  RETURNING id, title
)
SELECT 'Seed data loaded successfully' AS result;

COMMIT;
