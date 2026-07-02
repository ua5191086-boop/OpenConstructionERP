import Keycloak from 'keycloak-js'

// ============================================================
// Keycloak initialisation for OpenConstructionERP Frontend
// ============================================================

const keycloakConfig = {
  url: import.meta.env.VITE_KEYCLOAK_URL || 'http://localhost:8084',
  realm: import.meta.env.VITE_KEYCLOAK_REALM || 'OpenConstructionERP',
  clientId: import.meta.env.VITE_KEYCLOAK_CLIENT_ID || 'oce-frontend',
}

const keycloak = new Keycloak(keycloakConfig)

export default keycloak