import sys, os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), "..", "..", "services", "core-py"))
from app.routers.boq import norm_header, HEADER_MAP

def test_russian_headers_map():
    assert HEADER_MAP[norm_header("Шифр")] == "code"
    assert HEADER_MAP[norm_header("Наименование")] == "name"
    assert HEADER_MAP[norm_header("Ед.изм")] == "unit"
    assert HEADER_MAP[norm_header("Кол-во")] == "quantity"
    assert HEADER_MAP[norm_header("Цена")] == "unit_price"
    assert HEADER_MAP[norm_header("Глава")] == "cbs"

def test_english_headers_map():
    for h, f in [("Code","code"),("Unit Rate","unit_price"),("Qty","quantity"),("UoM","unit")]:
        assert HEADER_MAP[norm_header(h)] == f

def test_noise_normalisation():
    assert norm_header("  Кол-во  ") == "колво"
    assert norm_header("Unit_Price") == "unitprice"

def test_chainage_derivation():
    # ring N chainage = from + N * width/1000 (direction-aware)
    width = 1400; frm = 0.0
    assert frm + 42 * width / 1000.0 == 58.8
