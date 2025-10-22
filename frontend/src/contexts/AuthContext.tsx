import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode
} from 'react';
import {
  login as apiLogin,
  register as apiRegister,
  refreshTokens,
  logout as apiLogout,
  setAuthTokens,
  clearAuthTokens,
  setRefreshExecutor,
  getRefreshToken
} from '../api/client';
import type { AuthResponse, User } from '../types';

interface AuthContextValue {
  user: User | null;
  token: string | null;
  refreshToken: string | null;
  loading: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string, displayName: string) => Promise<void>;
  logout: () => Promise<void>;
  refreshProfile: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined);

const TOKEN_KEY = 'hdu-food-review-access-token';
const REFRESH_KEY = 'hdu-food-review-refresh-token';
const USER_KEY = 'hdu-food-review-user';

const readLocalToken = (key: string) => {
  try {
    return localStorage.getItem(key);
  } catch (err) {
    console.warn('unable to read from localStorage', err);
    return null;
  }
};

const writeLocalToken = (key: string, value: string | null) => {
  try {
    if (value) {
      localStorage.setItem(key, value);
    } else {
      localStorage.removeItem(key);
    }
  } catch (err) {
    console.warn('unable to write to localStorage', err);
  }
};

const readLocalUser = (): User | null => {
  try {
    const raw = localStorage.getItem(USER_KEY);
    if (!raw) return null;
    return JSON.parse(raw) as User;
  } catch (err) {
    console.warn('unable to read cached user', err);
    return null;
  }
};

const writeLocalUser = (value: User | null) => {
  try {
    if (value) {
      localStorage.setItem(USER_KEY, JSON.stringify(value));
    } else {
      localStorage.removeItem(USER_KEY);
    }
  } catch (err) {
    console.warn('unable to cache user', err);
  }
};

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [user, setUser] = useState<User | null>(() => readLocalUser());
  const [token, setToken] = useState<string | null>(() => readLocalToken(TOKEN_KEY));
  const [refreshToken, setRefreshToken] = useState<string | null>(() => readLocalToken(REFRESH_KEY));
  const [loading, setLoading] = useState<boolean>(true);

  const refreshRef = useRef<string | null>(refreshToken);
  useEffect(() => {
    refreshRef.current = refreshToken;
  }, [refreshToken]);

  const clearState = useCallback(() => {
    setToken(null);
    setRefreshToken(null);
    setUser(null);
    clearAuthTokens();
    writeLocalToken(TOKEN_KEY, null);
    writeLocalToken(REFRESH_KEY, null);
    writeLocalUser(null);
  }, []);

  const persist = useCallback((auth: AuthResponse) => {
    setToken(auth.access_token);
    setRefreshToken(auth.refresh_token);
    setUser(auth.user);
    setAuthTokens(auth.access_token, auth.refresh_token);
    writeLocalToken(TOKEN_KEY, auth.access_token);
    writeLocalToken(REFRESH_KEY, auth.refresh_token);
    writeLocalUser(auth.user);
  }, []);

  const refreshAccessToken = useCallback(async (): Promise<AuthResponse | null> => {
    const currentRefresh = refreshRef.current;
    if (!currentRefresh) {
      return null;
    }
    try {
      const updated = await refreshTokens(currentRefresh);
      persist(updated);
      return updated;
    } catch (err) {
      console.warn('refresh token failed', err);
      clearState();
      return null;
    }
  }, [persist, clearState]);

  useEffect(() => {
    setAuthTokens(token, refreshToken);
    setRefreshExecutor(refreshAccessToken);
  }, [token, refreshToken, refreshAccessToken]);

  useEffect(() => {
    const bootstrap = async () => {
      if (!token) {
        setLoading(false);
        return;
      }

      if (user) {
        setLoading(false);
        return;
      }

      const refreshed = await refreshAccessToken();
      if (!refreshed) {
        clearState();
      }
      setLoading(false);
    };

    bootstrap();
  }, [token, user, refreshAccessToken, clearState]);

  const handleLogin = useCallback(async (email: string, password: string) => {
    const auth = await apiLogin(email, password);
    persist(auth);
  }, [persist]);

  const handleRegister = useCallback(
    async (email: string, password: string, displayName: string) => {
      const auth = await apiRegister(email, password, displayName);
      persist(auth);
    },
    [persist]
  );

  const handleLogout = useCallback(async () => {
    try {
      const rt = getRefreshToken();
      if (rt) {
        await apiLogout(rt);
      }
    } catch (err) {
      console.warn('logout request failed', err);
    } finally {
      clearState();
    }
  }, [clearState]);

  const refreshProfile = useCallback(async () => {
    const updated = await refreshAccessToken();
    if (updated) {
      setUser(updated.user);
      writeLocalUser(updated.user);
    }
  }, [refreshAccessToken]);

  const value = useMemo<AuthContextValue>(() => ({
    user,
    token,
    refreshToken,
    loading,
    login: handleLogin,
    register: handleRegister,
    logout: handleLogout,
    refreshProfile
  }), [user, token, refreshToken, loading, handleLogin, handleRegister, handleLogout, refreshProfile]);

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export const useAuthContext = () => {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error('useAuthContext must be used inside AuthProvider');
  }
  return ctx;
};
