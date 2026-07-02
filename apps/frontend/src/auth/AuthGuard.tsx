import { useEffect, useState, type ReactNode } from 'react'
import keycloak from './keycloak'

// ============================================================
// AuthGuard — wraps protected routes, redirects to Keycloak login
// ============================================================

interface AuthGuardProps {
  children: ReactNode
  fallback?: ReactNode
}

export default function AuthGuard({ children, fallback }: AuthGuardProps) {
  const [initialized, setInitialized] = useState(false)
  const [authenticated, setAuthenticated] = useState(false)

  useEffect(() => {
    keycloak
      .init({
        onLoad: 'check-sso',
        silentCheckSsoRedirectUri:
          window.location.origin + '/silent-check-sso.html',
        pkceMethod: 'S256',
        checkLoginIframe: true,
      })
      .then((auth: boolean) => {
        setAuthenticated(auth)
        setInitialized(true)

        // Periodically refresh token (every 60s)
        setInterval(() => {
          keycloak
            .updateToken(30)
            .then((refreshed: boolean) => {
              if (refreshed) {
                console.log('[Auth] Token refreshed')
              }
            })
            .catch(() => {
              console.error('[Auth] Token refresh failed')
              keycloak.login()
            })
        }, 60000)
      })
      .catch((err: Error) => {
        console.error('[Auth] Keycloak init failed', err)
        setInitialized(true)
        setAuthenticated(false)
      })
  }, [])

  if (!initialized) {
    return (
      <div className="flex items-center justify-center h-screen bg-[#0f172a] text-white">
        <div className="text-center">
          <div className="animate-spin h-8 w-8 border-4 border-[#3b82f6] border-t-transparent rounded-full mx-auto mb-4" />
          <p className="text-[#94a3b8]">Checking authentication...</p>
        </div>
      </div>
    )
  }

  if (!authenticated) {
    if (fallback) {
      return <>{fallback}</>
    }
    // Auto-redirect to Keycloak login
    keycloak.login()
    return (
      <div className="flex items-center justify-center h-screen bg-[#0f172a] text-white">
        <div className="text-center">
          <p className="text-[#94a3b8]">Redirecting to login...</p>
        </div>
      </div>
    )
  }

  return <>{children}</>
}

// Helper to get the stored token
export function getToken(): string | undefined {
  return keycloak.token
}

// Helper to get token with refresh
export async function getTokenMinValidity(minValidity = 30): Promise<string | undefined> {
  try {
    await keycloak.updateToken(minValidity)
    return keycloak.token
  } catch {
    keycloak.login()
    return undefined
  }
}

// Helper to extract username
export function getUsername(): string {
  return keycloak.tokenParsed?.preferred_username || 'Unknown'
}

// Helper to extract roles
export function getRoles(): string[] {
  if (!keycloak.realmAccess?.roles) return []
  return keycloak.realmAccess.roles
}

// Helper to check role
export function hasRole(role: string): boolean {
  return keycloak.hasRealmRole(role)
}

// Logout helper
export function logout(): void {
  keycloak.logout({ redirectUri: window.location.origin })
}

export { keycloak }