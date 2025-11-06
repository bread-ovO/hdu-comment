import axios, { AxiosError, AxiosRequestConfig } from 'axios';
import type { AuthResponse, PaginatedResponse, Review, User } from '../types';

const api = axios.create({
  baseURL: '/api/v1'
});

const rawApi = axios.create({
  baseURL: '/api/v1'
});

let accessToken: string | null = null;
let refreshToken: string | null = null;
let refreshExecutor: (() => Promise<AuthResponse | null>) | null = null;
let refreshPromise: Promise<AuthResponse | null> | null = null;

const setAuthorizationHeader = (config: AxiosRequestConfig, token: string | null) => {
  if (!config.headers) config.headers = {};
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  } else {
    delete config.headers.Authorization;
  }
};

api.interceptors.request.use((config) => {
  if (accessToken) {
    setAuthorizationHeader(config, accessToken);
  }
  return config;
});

api.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const responseStatus = error.response?.status;
    const originalRequest: any = error.config;

    if (responseStatus === 401 && refreshToken && refreshExecutor && !originalRequest?._retry) {
      originalRequest._retry = true;

      if (!refreshPromise) {
        refreshPromise = refreshExecutor().finally(() => {
          refreshPromise = null;
        });
      }

      const result = await refreshPromise;
      if (result && result.access_token) {
        accessToken = result.access_token;
        refreshToken = result.refresh_token;
        setAuthorizationHeader(originalRequest, accessToken);
        return api(originalRequest);
      }
    }

    return Promise.reject(error);
  }
);

export const setAuthTokens = (access: string | null, refresh: string | null) => {
  accessToken = access;
  refreshToken = refresh;
};

export const clearAuthTokens = () => {
  accessToken = null;
  refreshToken = null;
};

export const setRefreshExecutor = (executor: (() => Promise<AuthResponse | null>) | null) => {
  refreshExecutor = executor;
};

export const getRefreshToken = () => refreshToken;

export interface ReviewQueryParams {
  page?: number;
  page_size?: number;
  query?: string;
  sort?: 'created_at' | 'rating';
  order?: 'asc' | 'desc';
}

export const register = async (email: string, password: string, displayName: string, code: string): Promise<AuthResponse> => {
  const { data } = await api.post<AuthResponse>('/auth/register', {
    email,
    password,
    display_name: displayName,
    code
  });
  return data;
};

export const requestRegistrationCode = async (email: string): Promise<void> => {
  await rawApi.post('/auth/send-code', { email });
};

export const login = async (email: string, password: string): Promise<AuthResponse> => {
  const { data } = await api.post<AuthResponse>('/auth/login', { email, password });
  return data;
};

export const refreshTokens = async (token: string): Promise<AuthResponse> => {
  const { data } = await rawApi.post<AuthResponse>('/auth/refresh', { refresh_token: token });
  return data;
};

export const logout = async (token: string): Promise<void> => {
  await api.post('/auth/logout', { refresh_token: token });
};

export const fetchMe = async (): Promise<User> => {
  const { data } = await api.get<User>('/users/me');
  return data;
};

export const fetchReviews = async (params: ReviewQueryParams = {}): Promise<PaginatedResponse<Review>> => {
  const { data } = await api.get<PaginatedResponse<Review>>('/reviews', { params });
  return data;
};

export const fetchMyReviews = async (params: ReviewQueryParams = {}): Promise<PaginatedResponse<Review>> => {
  const { data } = await api.get<PaginatedResponse<Review>>('/reviews/me', { params });
  return data;
};

export const submitReview = async (payload: {
  title: string;
  address: string;
  description: string;
  rating: number;
}): Promise<Review> => {
  const { data } = await api.post<Review>('/reviews', payload);
  return data;
};

export const fetchReviewDetail = async (id: string): Promise<Review> => {
  const { data } = await api.get<Review>(`/reviews/${id}`);
  return data;
};

export const uploadReviewImage = async (id: string, file: File): Promise<void> => {
  const formData = new FormData();
  formData.append('file', file);
  await api.post(`/reviews/${id}/images`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  });
};

export const fetchPendingReviews = async (params: ReviewQueryParams = {}): Promise<PaginatedResponse<Review>> => {
  const { data } = await api.get<PaginatedResponse<Review>>('/admin/reviews/pending', { params });
  return data;
};

export const approveReview = async (id: string): Promise<Review> => {
  const { data } = await api.put<Review>(`/admin/reviews/${id}/approve`);
  return data;
};

export const rejectReview = async (id: string, reason: string): Promise<Review> => {
  const { data } = await api.put<Review>(`/admin/reviews/${id}/reject`, { reason });
  return data;
};

export const deleteReview = async (id: string): Promise<void> => {
  await api.delete(`/admin/reviews/${id}`);
};

export interface AdminUserListResponse {
  data: User[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
  };
}

export const fetchAdminUsers = async (params: { page?: number; page_size?: number }): Promise<AdminUserListResponse> => {
  const { data } = await api.get<AdminUserListResponse>('/admin/users', { params });
  return data;
};

export const deleteAdminUser = async (id: string): Promise<void> => {
  await api.delete(`/admin/users/${id}`);
};
