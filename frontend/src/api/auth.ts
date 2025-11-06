import axios from 'axios';
import type { User } from '../types';

const ACCESS_TOKEN_KEY = 'hdu-food-review-access-token';

const api = axios.create({
    baseURL: '/api/v1',
    withCredentials: true
});

// 添加请求拦截器，确保包含认证token
api.interceptors.request.use(
    (config) => {
        try {
            const token = localStorage.getItem(ACCESS_TOKEN_KEY);
            if (token) {
                config.headers.Authorization = `Bearer ${token}`;
            }
        } catch (err) {
            console.warn('Failed to read access token from storage', err);
        }
        return config;
    },
    (error) => {
        return Promise.reject(error);
    }
);

// 添加响应拦截器，处理错误
api.interceptors.response.use(
    (response) => response,
    (error) => {
        console.error('API Error:', error.response?.data || error.message);
        return Promise.reject(error);
    }
);

export interface AuthResponse {
    access_token: string;
    refresh_token: string;
    user: User;
}

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
    // 发送注册验证码
    sendRegistrationCode: async (email: string): Promise<{ message: string }> => {
        const response = await api.post('/auth/send-code', { email });
        return response.data;
    },

    // 用户注册
    register: async (data: RegisterRequest): Promise<AuthResponse> => {
        const response = await api.post('/auth/register', data);
        return response.data;
    },

    // 用户登录
    login: async (data: LoginRequest): Promise<AuthResponse> => {
        const response = await api.post('/auth/login', data);
        return response.data;
    },

    // 刷新令牌
    refresh: async (data: RefreshRequest): Promise<AuthResponse> => {
        const response = await api.post('/auth/refresh', data);
        return response.data;
    },

    // 用户登出
    logout: async (refreshToken: string): Promise<void> => {
        await api.post('/auth/logout', { refresh_token: refreshToken });
    },

    // 获取当前用户信息
    getCurrentUser: async (): Promise<User> => {
        const response = await api.get('/users/me');
        return response.data;
    },

    // 发送邮箱验证邮件
    sendVerificationEmail: async (): Promise<{ message: string }> => {
        const response = await api.post('/auth/send-verification');
        return response.data;
    },

    // 验证邮箱
    verifyEmail: async (token: string): Promise<{ message: string }> => {
        const response = await api.post('/auth/verify-email', { token });
        return response.data;
    },

    // 获取邮箱验证状态
    getVerificationStatus: async (): Promise<{ email_verified: boolean }> => {
        const response = await api.get('/auth/verification-status');
        return response.data;
    }
};
