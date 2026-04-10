import Taro from '@tarojs/taro';
import type { AuthResponse } from '../types';
import { clearAuth, getRefreshToken, getToken, setRefreshToken, setToken, setUser } from './storage';

const DEFAULT_BASE_URL = 'http://127.0.0.1:8080/api/v1';
const BASE_URL = (process.env.TARO_APP_API_BASE_URL || DEFAULT_BASE_URL).replace(/\/+$/, '');
const BASE_ORIGIN = (process.env.TARO_APP_ASSET_BASE_URL || BASE_URL.replace(/\/api\/v1\/?$/, '')).replace(/\/+$/, '');

type RequestMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';

interface RequestConfig {
  url: string;
  method?: RequestMethod;
  data?: unknown;
  headers?: Record<string, string>;
  skipAuthRefresh?: boolean;
}

type AuthChangeListener = (auth: AuthResponse | null) => void;

let authChangeListener: AuthChangeListener | null = null;
let refreshPromise: Promise<AuthResponse | null> | null = null;

export function setAuthChangeListener(listener: AuthChangeListener | null): void {
  authChangeListener = listener;
}

function notifyAuthChange(auth: AuthResponse | null): void {
  authChangeListener?.(auth);
}

function persistAuth(auth: AuthResponse): void {
  setToken(auth.access_token);
  setRefreshToken(auth.refresh_token);
  setUser(auth.user);
  notifyAuthChange(auth);
}

function clearStoredAuth(): void {
  clearAuth();
  notifyAuthChange(null);
}

function createRequestError(statusCode: number, data: unknown): Error & { statusCode: number; data: unknown } {
  const message =
    typeof data === 'object' && data !== null && 'error' in data && typeof data.error === 'string'
      ? data.error
      : `request failed with status ${statusCode}`;
  const error = new Error(message) as Error & { statusCode: number; data: unknown };
  error.statusCode = statusCode;
  error.data = data;
  return error;
}

function isUnauthorizedError(error: unknown): error is Error & { statusCode: number } {
  return typeof error === 'object' && error !== null && 'statusCode' in error && error.statusCode === 401;
}

function parseResponseData<T>(data: T | string): T | unknown {
  if (typeof data !== 'string') {
    return data;
  }

  try {
    return JSON.parse(data);
  } catch {
    return data;
  }
}

async function performRequest<T>(config: RequestConfig, token?: string | null): Promise<T> {
  const response = await Taro.request<T>({
    url: BASE_URL + config.url,
    method: config.method ?? 'GET',
    data: config.data,
    header: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...config.headers,
    },
  });

  if (response.statusCode >= 200 && response.statusCode < 300) {
    return response.data;
  }

  throw createRequestError(response.statusCode, response.data);
}

export async function refreshSession(): Promise<AuthResponse | null> {
  if (refreshPromise) {
    return refreshPromise;
  }

  const refreshToken = getRefreshToken();
  if (!refreshToken) {
    clearStoredAuth();
    return null;
  }

  refreshPromise = (async () => {
    try {
      const auth = await performRequest<AuthResponse>(
        {
          url: '/auth/refresh',
          method: 'POST',
          data: { refresh_token: refreshToken },
          skipAuthRefresh: true,
        },
        null,
      );
      persistAuth(auth);
      return auth;
    } catch (error) {
      console.warn('刷新登录态失败:', error);
      clearStoredAuth();
      return null;
    } finally {
      refreshPromise = null;
    }
  })();

  return refreshPromise;
}

/**
 * 请求封装 - 同时支持微信小程序和 H5
 */
export async function request<T = unknown>(config: RequestConfig): Promise<T> {
  const token = getToken();

  try {
    return await performRequest<T>(config, token);
  } catch (error) {
    if (!token || config.skipAuthRefresh || !isUnauthorizedError(error)) {
      throw error;
    }

    const auth = await refreshSession();
    if (!auth) {
      throw error;
    }

    return performRequest<T>(config, auth.access_token);
  }
}

export async function uploadReviewImage(reviewID: string, filePath: string): Promise<void> {
  const upload = async (token?: string | null) => {
    const response = await Taro.uploadFile({
      url: `${BASE_URL}/reviews/${reviewID}/images`,
      filePath,
      name: 'file',
      header: token ? { Authorization: `Bearer ${token}` } : undefined,
    });

    if (response.statusCode >= 200 && response.statusCode < 300) {
      return;
    }

    throw createRequestError(response.statusCode, parseResponseData(response.data));
  };

  const token = getToken();

  try {
    await upload(token);
  } catch (error) {
    if (!token || !isUnauthorizedError(error)) {
      throw error;
    }

    const auth = await refreshSession();
    if (!auth) {
      throw error;
    }

    await upload(auth.access_token);
  }
}

export function resolveAssetURL(url: string): string {
  if (!url) {
    return url;
  }

  if (/^https?:\/\//i.test(url)) {
    return url;
  }

  return `${BASE_ORIGIN}${url.startsWith('/') ? url : `/${url}`}`;
}

/**
 * 微信登录凭证校验
 */
export async function wxLogin(code: string): Promise<AuthResponse> {
  return request({
    url: '/auth/wechat/login',
    method: 'POST',
    data: { code },
  });
}
