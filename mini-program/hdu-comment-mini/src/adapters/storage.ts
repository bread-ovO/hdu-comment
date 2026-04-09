/**
 * 存储适配器 - 适配微信小程序和 H5 环境
 */

import Taro from '@tarojs/taro';
import type { User } from '../types';

const isWeapp = process.env.TARO_ENV === 'weapp';

export const storage = {
  get(key: string): string | null {
    if (isWeapp) {
      return Taro.getStorageSync<string>(key) || null;
    }
    return localStorage.getItem(key);
  },

  set(key: string, value: string): void {
    if (isWeapp) {
      Taro.setStorageSync(key, value);
    } else {
      localStorage.setItem(key, value);
    }
  },

  remove(key: string): void {
    if (isWeapp) {
      Taro.removeStorageSync(key);
    } else {
      localStorage.removeItem(key);
    }
  },

  clear(): void {
    if (isWeapp) {
      Taro.clearStorageSync();
    } else {
      localStorage.clear();
    }
  },
};

export const TOKEN_KEY = 'hdu-comment-access-token';
export const REFRESH_KEY = 'hdu-comment-refresh-token';
export const USER_KEY = 'hdu-comment-user';

export const getToken = () => storage.get(TOKEN_KEY);
export const setToken = (token: string) => storage.set(TOKEN_KEY, token);
export const getRefreshToken = () => storage.get(REFRESH_KEY);
export const setRefreshToken = (token: string) => storage.set(REFRESH_KEY, token);
export const getUser = (): User | null => {
  const raw = storage.get(USER_KEY);
  if (!raw) {
    return null;
  }

  try {
    return JSON.parse(raw) as User;
  } catch (error) {
    console.warn('读取用户缓存失败:', error);
    storage.remove(USER_KEY);
    return null;
  }
};
export const setUser = (user: User) => storage.set(USER_KEY, JSON.stringify(user));
export const clearAuth = () => {
  storage.remove(TOKEN_KEY);
  storage.remove(REFRESH_KEY);
  storage.remove(USER_KEY);
};
