# AI Services for OpenConstructionERP
# FastAPI-based microservice for AI/ML operations

from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel, Field
from typing import Optional, List, Dict, Any
from datetime import datetime
import uuid
import random
import math

app = FastAPI(
    title="OpenConstructionERP — AI Services",
    description="AI/ML microservice for classification, prediction, extraction, and recommendations",
    version="1.0.0",
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# ─── Models ─────────────────────────────────────────────────────────────────

class ClassifyRequest(BaseModel):
    text: str = Field(..., description="Input text to classify")
    context: Optional[str] = Field(None, description="Optional module/project context")

class ClassifyResponse(BaseModel):
    request_id: str
    category: str
    subcategory: Optional[str] = None
    confidence: float
    labels: List[str] = []
    processing_time_ms: float

class PredictRequest(BaseModel):
    model_type: str = Field(..., description="Prediction model: cost, duration, risk")
    features: Dict[str, Any] = Field(..., description="Input features for prediction")

class PredictResponse(BaseModel):
    request_id: str
    prediction: float
    confidence: float
    lower_bound: Optional[float] = None
    upper_bound: Optional[float] = None
    unit: str
    processing_time_ms: float

class ExtractRequest(BaseModel):
    text: str = Field(..., description="Text to extract data from")
    entity_types: List[str] = Field(default=["all"], description="Entity types to extract")

class ExtractResponse(BaseModel):
    request_id: str
    entities: List[Dict[str, Any]] = []
    raw_text: str
    processing_time_ms: float

class RecommendRequest(BaseModel):
    context: str = Field(..., description="Context for recommendations")
    constraints: Optional[Dict[str, Any]] = Field(None, description="Additional constraints")

class RecommendResponse(BaseModel):
    request_id: str
    recommendations: List[Dict[str, Any]] = []
    reasoning: Optional[str] = None
    processing_time_ms: float

# ─── In-memory "AI" processor (stub — replace with real model calls) ──────

CATEGORIES = [
    "finance", "schedule", "quality", "safety", "contract",
    "procurement", "design", "construction", "risk", "compliance"
]

SUBCATEGORIES = {
    "finance": ["budget", "cashflow", "invoice", "evm", "forecast"],
    "schedule": ["planning", "delay", "milestone", "resource", "critical_path"],
    "quality": ["inspection", "non_conformance", "test", "certification"],
    "safety": ["incident", "hazard", "training", "ppe"],
    "contract": ["variation", "claim", "payment", "termination"],
}

def _classify_stub(text: str) -> tuple:
    """Simple keyword-based classification stub."""
    text_lower = text.lower()
    scores = {}
    for cat in CATEGORIES:
        score = 0
        keywords = {
            "finance": ["cost", "budget", "invoice", "payment", "financial", "evm", "earned"],
            "schedule": ["schedule", "delay", "plan", "timeline", "milestone", "critical"],
            "quality": ["quality", "inspection", "test", "defect", "non-conformance"],
            "safety": ["safety", "incident", "accident", "hazard", "risk"],
            "contract": ["contract", "claim", "variation", "change order"],
            "procurement": ["procurement", "purchase", "vendor", "supplier", "material"],
            "design": ["design", "drawing", "model", "ifc", "bim"],
            "construction": ["construction", "excavation", "concrete", "steel", "installation"],
            "risk": ["risk", "uncertainty", "probability", "impact", "mitigation"],
            "compliance": ["compliance", "regulation", "permit", "license", "audit"],
        }
        for kw in keywords.get(cat, []):
            if kw in text_lower:
                score += 1
        if score > 0:
            scores[cat] = score
    if not scores:
        return "general", None, 0.5
    best = max(scores, key=lambda k: scores[k])
    sub = None
    if best in SUBCATEGORIES:
        for s in SUBCATEGORIES[best]:
            if s in text_lower:
                sub = s
                break
    confidence = min(0.95, 0.5 + scores[best] * 0.1)
    return best, sub, confidence

# ─── Endpoints ──────────────────────────────────────────────────────────────

@app.get("/api/v1/ai/health", tags=["Health"])
async def health_check():
    return {
        "status": "ok",
        "service": "oce-ai-svc",
        "version": "1.0.0",
        "models_loaded": ["classifier-v1", "predictor-v1", "extractor-v1", "recommender-v1"],
        "timestamp": datetime.utcnow().isoformat()
    }

@app.post("/api/v1/ai/classify", response_model=ClassifyResponse, tags=["Classification"])
async def classify(request: ClassifyRequest):
    start = datetime.utcnow()
    category, subcategory, confidence = _classify_stub(request.text)

    # Build label list
    labels = [category]
    if subcategory:
        labels.append(f"{category}:{subcategory}")

    processing_time = (datetime.utcnow() - start).total_seconds() * 1000
    return ClassifyResponse(
        request_id=str(uuid.uuid4()),
        category=category,
        subcategory=subcategory,
        confidence=round(confidence, 4),
        labels=labels,
        processing_time_ms=round(processing_time, 2)
    )

@app.post("/api/v1/ai/predict", response_model=PredictResponse, tags=["Prediction"])
async def predict(request: PredictRequest):
    start = datetime.utcnow()

    if request.model_type == "cost":
        base_cost = request.features.get("base_cost", 1000000)
        complexity = request.features.get("complexity", 0.5)
        inflation = request.features.get("inflation", 0.05)
        prediction = base_cost * (1 + complexity * 0.3) * (1 + inflation)
        unit = "USD"
        lower = prediction * 0.9
        upper = prediction * 1.15

    elif request.model_type == "duration":
        base_days = request.features.get("base_duration_days", 365)
        team_size = request.features.get("team_size", 10)
        scope = request.features.get("scope_factor", 1.0)
        prediction = base_days * scope * (1 + 0.2 * math.exp(-team_size / 20))
        unit = "days"
        lower = prediction * 0.85
        upper = prediction * 1.2

    elif request.model_type == "risk":
        probability = request.features.get("probability", 0.3)
        impact = request.features.get("impact", 50000)
        prediction = probability * impact
        unit = "USD"
        lower = prediction * 0.7
        upper = prediction * 1.5

    else:
        raise HTTPException(status_code=400, detail=f"Unknown model_type: {request.model_type}. Use: cost, duration, risk")

    confidence = round(random.uniform(0.75, 0.95), 4)
    processing_time = (datetime.utcnow() - start).total_seconds() * 1000

    return PredictResponse(
        request_id=str(uuid.uuid4()),
        prediction=round(prediction, 2),
        confidence=confidence,
        lower_bound=round(lower, 2) if lower else None,
        upper_bound=round(upper, 2) if upper else None,
        unit=unit,
        processing_time_ms=round(processing_time, 2)
    )

@app.post("/api/v1/ai/extract", response_model=ExtractResponse, tags=["Extraction"])
async def extract(request: ExtractRequest):
    start = datetime.utcnow()
    text = request.text
    entities = []

    # Simple regex-free stub extraction
    import re

    # Extract monetary amounts
    money_pattern = r'[\$€£]\s*[\d,]+(?:\.\d{2})?'
    for match in re.finditer(money_pattern, text):
        entities.append({
            "type": "monetary_amount",
            "value": match.group(),
            "position": [match.start(), match.end()]
        })

    # Extract dates
    date_pattern = r'\d{1,2}[./-]\d{1,2}[./-]\d{2,4}'
    for match in re.finditer(date_pattern, text):
        entities.append({
            "type": "date",
            "value": match.group(),
            "position": [match.start(), match.end()]
        })

    # Extract percentages
    pct_pattern = r'\d+(?:\.\d+)?%'
    for match in re.finditer(pct_pattern, text):
        entities.append({
            "type": "percentage",
            "value": match.group(),
            "position": [match.start(), match.end()]
        })

    # Extract project codes (uppercase letters + digits, 4-10 chars)
    code_pattern = r'\b[A-Z]{2,5}-\d{3,5}\b'
    for match in re.finditer(code_pattern, text):
        entities.append({
            "type": "project_code",
            "value": match.group(),
            "position": [match.start(), match.end()]
        })

    processing_time = (datetime.utcnow() - start).total_seconds() * 1000
    return ExtractResponse(
        request_id=str(uuid.uuid4()),
        entities=entities,
        raw_text=text,
        processing_time_ms=round(processing_time, 2)
    )

@app.post("/api/v1/ai/recommend", response_model=RecommendResponse, tags=["Recommendations"])
async def recommend(request: RecommendRequest):
    start = datetime.utcnow()

    context = request.context.lower()
    constraints = request.constraints or {}

    recommendations = []

    # Context-based stub recommendations
    if "cost" in context or "budget" in context:
        recommendations.append({
            "type": "action",
            "priority": "high",
            "title": "Провести анализ отклонений по контрольным счетам",
            "description": "Сравнить фактические затраты с планом по каждому CA, выявить превышения >10%",
            "impact": "Снижение перерасхода бюджета до 5%"
        })
        recommendations.append({
            "type": "action",
            "priority": "medium",
            "title": "Пересмотреть EAC с учётом текущего CPI",
            "description": "Использовать метод CPI-based EAC для прогноза завершения",
            "impact": "Точность прогноза ±3%"
        })

    if "schedule" in context or "delay" in context:
        recommendations.append({
            "type": "action",
            "priority": "high",
            "title": "Проанализировать критический путь",
            "description": "Выявить активности с отрицательным плавающим резервом",
            "impact": "Сокращение задержки на 2-4 недели"
        })
        recommendations.append({
            "type": "optimization",
            "priority": "medium",
            "title": "Перераспределить ресурсы на критические задачи",
            "description": "Увеличить количество персонала на активностях с SPI < 0.9",
            "impact": "SPI > 0.95"
        })

    if "risk" in context:
        recommendations.append({
            "type": "mitigation",
            "priority": "high",
            "title": "Разработать план реагирования на ТОП-5 рисков",
            "description": "Для каждого риска с вероятностью >30% и воздействием >$50K",
            "impact": "Снижение вероятности срыва сроков"
        })

    # Default recommendation
    if not recommendations:
        recommendations.append({
            "type": "info",
            "priority": "low",
            "title": "Запросить более конкретный контекст",
            "description": "Укажите область (cost, schedule, risk, quality) для целевых рекомендаций",
            "impact": "—"
        })

    processing_time = (datetime.utcnow() - start).total_seconds() * 1000
    return RecommendResponse(
        request_id=str(uuid.uuid4()),
        recommendations=recommendations,
        reasoning=f"Recommendations generated based on context: {request.context}",
        processing_time_ms=round(processing_time, 2)
    )

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8100)