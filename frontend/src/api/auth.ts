import type { AuthResponse, User } from '../types';
import {
  api,
  fetchMe,
  getQQLoginURL,
  login as loginWithClient,
  loginByQQ,
  loginBySMS,
  logout as logoutWithClient,
  refreshTokens,
  sendSMSLoginCode,
  register as registerWithClient
} from './client';

export interface RegisterRequest {
  email: string;
  password: string;
  display_name: string;
  code: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RefreshRequest {
  refresh_token: string;
}

export interface VerifyEmailRequest {
  token: string;
}

export interface QQLoginRequest {
  code: string;
  state: string;
}

export interface SMSLoginRequest {
  phone: string;
  code: string;
}

export const authApi = {
  sendRegistrationCode: async (email: string): Promise<{ message: string }> => {
    const response = await api.post<{ message: string }>('/auth/send-code', { email });
    return response.data;
  },

  register: async (data: RegisterRequest): Promise<AuthResponse> => {
    return registerWithClient(data.email, data.password, data.display_name, data.code);
  },

  login: async (data: LoginRequest): Promise<AuthResponse> => {
    return loginWithClient(data.email, data.password);
  },

  getQQLoginURL: async (): Promise<{ url: string; state: string }> => {
    return getQQLoginURL();
  },

  qqLogin: async (data: QQLoginRequest): Promise<AuthResponse> => {
    return loginByQQ(data.code, data.state);
  },

  sendSMSCode: async (phone: string): Promise<{ message: string; debug_code?: string }> => {
    return sendSMSLoginCode(phone);
  },

  smsLogin: async (data: SMSLoginRequest): Promise<AuthResponse> => {
    return loginBySMS(data.phone, data.code);
  },

  refresh: async (data: RefreshRequest): Promise<AuthResponse> => {
    return refreshTokens(data.refresh_token);
  },

  logout: async (refreshToken: string): Promise<void> => {
    await logoutWithClient(refreshToken);
  },

  getCurrentUser: async (): Promise<User> => {
    return fetchMe();
  },

  sendVerificationEmail: async (): Promise<{ message: string }> => {
    const response = await api.post<{ message: string }>('/auth/send-verification');
    return response.data;
  },

  verifyEmail: async (token: string): Promise<{ message: string }> => {
    const response = await api.post<{ message: string }>('/auth/verify-email', { token });
    return response.data;
  },

  getVerificationStatus: async (): Promise<{ email_verified: boolean }> => {
    const response = await api.get<{ email_verified: boolean }>('/auth/verification-status');
    return response.data;
  }
};
