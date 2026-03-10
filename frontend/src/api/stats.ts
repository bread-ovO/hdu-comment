import type { ReviewStats, SiteStats, ReactionResponse, UserReactionResponse } from '../types';
import { api } from './client';

export const statsApi = {
  getReviewStats: async (reviewId: string): Promise<ReviewStats> => {
    const response = await api.get(`/reviews/${reviewId}/stats`);
    return response.data;
  },

  recordView: async (reviewId: string): Promise<void> => {
    await api.post(`/reviews/${reviewId}/view`);
  },

  toggleReaction: async (reviewId: string, type: 'like' | 'dislike'): Promise<ReactionResponse> => {
    const response = await api.post(`/reviews/${reviewId}/react`, { type });
    return response.data;
  },

  getUserReaction: async (reviewId: string): Promise<UserReactionResponse> => {
    const response = await api.get(`/reviews/${reviewId}/user-reaction`);
    return response.data;
  },

  getSiteStats: async (): Promise<SiteStats> => {
    const response = await api.get('/stats/site');
    return response.data;
  },

  getTotalViews: async (): Promise<{ total_views: number }> => {
    const response = await api.get('/stats/total-views');
    return response.data;
  }
};
