import type { AuthResponse, User } from '../types';
import {
  api,
  fetchMe,
  login as loginWithClient,
  logout as logoutWithClient,
  refreshTokens,
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
