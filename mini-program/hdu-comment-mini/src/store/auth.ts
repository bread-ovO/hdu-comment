import { create } from 'zustand';
import Taro from '@tarojs/taro';
import type { User, AuthResponse } from '../types';
import { getToken, getRefreshToken, getUser, setToken, setRefreshToken, setUser, clearAuth } from '../adapters/storage';
import { refreshSession, setAuthChangeListener, wxLogin } from '../adapters/request';

interface AuthState {
  user: User | null;
  token: string | null;
  refreshToken: string | null;
  loading: boolean;
  isLoggedIn: boolean;

  // Actions
  init: () => Promise<void>;
  loginByWechat: () => Promise<void>;
  logout: () => void;
}

const initialUser = getUser();
const initialToken = getToken();
const initialRefreshToken = getRefreshToken();

const buildLoggedOutState = () => ({
  user: null,
  token: null,
  refreshToken: null,
  loading: false,
  isLoggedIn: false,
});

const buildAuthenticatedState = (auth: AuthResponse) => ({
  user: auth.user,
  token: auth.access_token,
  refreshToken: auth.refresh_token,
  loading: false,
  isLoggedIn: true,
});

export const useAuthStore = create<AuthState>((set) => ({
  user: initialUser,
  token: initialToken,
  refreshToken: initialRefreshToken,
  loading: Boolean(initialRefreshToken && !initialUser),
  isLoggedIn: Boolean(initialUser && initialToken),

  init: async () => {
    const token = getToken();
    const refreshToken = getRefreshToken();
    const cachedUser = getUser();

    if (cachedUser && token) {
      set({
        user: cachedUser,
        token,
        refreshToken,
        loading: false,
        isLoggedIn: true,
      });
      return;
    }

    if (refreshToken) {
      const auth = await refreshSession();
      if (auth) {
        set(buildAuthenticatedState(auth));
        return;
      }
    }

    set(buildLoggedOutState());
  },

  loginByWechat: async () => {
    try {
      // 微信小程序登录
      const loginRes = await Taro.login();
      console.log('wx login result', loginRes);


      if (!loginRes.code) {
        throw new Error('微信登录失败');
      }

      // 调用后端接口进行登录
      const auth = await wxLogin(loginRes.code);

      // 保存 token 和用户信息
      setToken(auth.access_token);
      setRefreshToken(auth.refresh_token);
      setUser(auth.user);

      set(buildAuthenticatedState(auth));
    } catch (error) {
      console.error('微信登录失败:', error);
      throw error;
    }
  },

  logout: () => {
    clearAuth();
    set(buildLoggedOutState());
  },
}));

setAuthChangeListener((auth) => {
  if (auth) {
    useAuthStore.setState(buildAuthenticatedState(auth));
    return;
  }

  useAuthStore.setState(buildLoggedOutState());
});
