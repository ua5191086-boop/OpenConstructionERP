"""API authentication (A-02, ADR-009).

Modes (env AUTH_MODE):
  keycloak  — verify RS256 JWT against Keycloak JWKS (production default path)
  static    — verify HS256 JWT with AUTH_STATIC_SECRET (dev/CI)
  disabled  — no auth; must be set EXPLICITLY, logs a loud warning

Every request (except /health, /docs, /openapi.json and static files) requires
Authorization: Bearer <token> when auth is enabled. Realm/client roles are
exposed on request.state.user for future ABAC checks.
"""
import logging
import os
import time
from typing import Optional

import jwt
from fastapi import Request
from fastapi.responses import JSONResponse
from starlette.middleware.base import BaseHTTPMiddleware

log = logging.getLogger("oce.auth")

AUTH_MODE = os.getenv("AUTH_MODE", "static").lower()
STATIC_SECRET = os.getenv("AUTH_STATIC_SECRET", "")
KEYCLOAK_URL = os.getenv("KEYCLOAK_URL", "http://localhost:8084")
KEYCLOAK_REALM = os.getenv("KEYCLOAK_REALM", "oce")
AUDIENCE = os.getenv("AUTH_AUDIENCE")          # optional aud check

PUBLIC_PATHS = ("/health", "/docs", "/openapi.json", "/redoc")

_jwks_cache: dict = {"keys": None, "at": 0.0}


def _jwks():
    if _jwks_cache["keys"] and time.time() - _jwks_cache["at"] < 300:
        return _jwks_cache["keys"]
    import urllib.request, json as _json
    url = f"{KEYCLOAK_URL}/realms/{KEYCLOAK_REALM}/protocol/openid-connect/certs"
    with urllib.request.urlopen(url, timeout=5) as r:
        _jwks_cache["keys"] = _json.loads(r.read())
        _jwks_cache["at"] = time.time()
    return _jwks_cache["keys"]


def _verify(token: str) -> dict:
    if AUTH_MODE == "static":
        if not STATIC_SECRET:
            raise jwt.InvalidTokenError("AUTH_STATIC_SECRET not configured")
        return jwt.decode(token, STATIC_SECRET, algorithms=["HS256"],
                          audience=AUDIENCE, options={"verify_aud": bool(AUDIENCE)})
    # keycloak: RS256 via JWKS
    header = jwt.get_unverified_header(token)
    key = None
    for k in _jwks().get("keys", []):
        if k.get("kid") == header.get("kid"):
            key = jwt.algorithms.RSAAlgorithm.from_jwk(k)
            break
    if key is None:
        _jwks_cache["at"] = 0.0  # force refresh once on unknown kid
        for k in _jwks().get("keys", []):
            if k.get("kid") == header.get("kid"):
                key = jwt.algorithms.RSAAlgorithm.from_jwk(k)
                break
    if key is None:
        raise jwt.InvalidTokenError("Signing key not found in JWKS")
    return jwt.decode(token, key, algorithms=["RS256"], audience=AUDIENCE,
                      options={"verify_aud": bool(AUDIENCE)})


def extract_roles(claims: dict) -> list[str]:
    roles = list(claims.get("realm_access", {}).get("roles", []))
    for client, acc in claims.get("resource_access", {}).items():
        roles += [f"{client}:{r}" for r in acc.get("roles", [])]
    return roles


class AuthMiddleware(BaseHTTPMiddleware):
    async def dispatch(self, request: Request, call_next):
        if AUTH_MODE == "disabled":
            request.state.user = {"sub": "anonymous", "roles": [], "mode": "disabled"}
            return await call_next(request)
        path = request.url.path
        if path in PUBLIC_PATHS or not path.startswith("/api/"):
            return await call_next(request)   # dashboards/static stay open; API guarded
        authz = request.headers.get("authorization", "")
        if not authz.lower().startswith("bearer "):
            return JSONResponse({"detail": "Missing bearer token"}, status_code=401)
        try:
            claims = _verify(authz.split(" ", 1)[1].strip())
        except jwt.ExpiredSignatureError:
            return JSONResponse({"detail": "Token expired"}, status_code=401)
        except jwt.InvalidTokenError as e:
            return JSONResponse({"detail": f"Invalid token: {e}"}, status_code=401)
        request.state.user = {"sub": claims.get("sub"),
                              "username": claims.get("preferred_username"),
                              "roles": extract_roles(claims),
                              "claims": claims}
        return await call_next(request)


def startup_banner():
    if AUTH_MODE == "disabled":
        log.warning("=" * 60)
        log.warning("AUTH IS DISABLED (AUTH_MODE=disabled). "
                    "Every API endpoint is open. NEVER run like this "
                    "outside localhost.")
        log.warning("=" * 60)
    else:
        log.info("Auth mode: %s%s", AUTH_MODE,
                 f" (realm {KEYCLOAK_REALM} @ {KEYCLOAK_URL})"
                 if AUTH_MODE == "keycloak" else "")
