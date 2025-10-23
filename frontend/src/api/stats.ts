import axios from 'axios';
import type { ReviewStats, SiteStats, ReactionResponse, UserReactionResponse } from '../types';

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

export const statsApi = {
    // 获取点评统计信息
    getReviewStats: async (reviewId: string): Promise<ReviewStats> => {
        const response = await api.get(`/reviews/${reviewId}/stats`);
        return response.data;
    },

    // 记录点评浏览
    recordView: async (reviewId: string): Promise<void> => {
        await api.post(`/reviews/${reviewId}/view`);
    },

    // 点赞或踩点评
    toggleReaction: async (reviewId: string, type: 'like' | 'dislike'): Promise<ReactionResponse> => {
        const response = await api.post(`/reviews/${reviewId}/react`, { type });
        return response.data;
    },

    // 获取用户对点评的反应
    getUserReaction: async (reviewId: string): Promise<UserReactionResponse> => {
        const response = await api.get(`/reviews/${reviewId}/user-reaction`);
        return response.data;
    },

    // 获取网站统计信息
    getSiteStats: async (): Promise<SiteStats> => {
        const response = await api.get('/stats/site');
        return response.data;
    },

    // 获取网站总浏览量
    getTotalViews: async (): Promise<{ total_views: number }> => {
        const response = await api.get('/stats/total-views');
        return response.data;
    },
};
