export interface User {
  id: string;
  email: string;
  display_name: string;
  role: 'user' | 'admin';
  email_verified: boolean;
  email_verified_at?: string | null;
  created_at?: string;
}

export interface ReviewImage {
  id: string;
  review_id: string;
  storage_key: string;
  url: string;
  created_at: string;
}

export interface ReviewStats {
  id: string;
  review_id: string;
  views: number;
  likes: number;
  dislikes: number;
  created_at: string;
  updated_at: string;
}

export interface Review {
  id: string;
  title: string;
  address: string;
  description: string;
  rating: number;
  status: 'pending' | 'approved' | 'rejected';
  rejection_reason?: string;
  author_id: string;
  author?: User;
  images?: ReviewImage[];
  stats?: ReviewStats;
  user_reaction?: 'like' | 'dislike' | null;
  created_at: string;
  updated_at: string;
}

export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  user: User;
}

export interface PaginationMeta {
  page: number;
  page_size: number;
  total: number;
  total_pages: number;
}

export interface PaginatedResponse<T> {
  data: T[];
  pagination: PaginationMeta;
}

export interface SiteStats {
  id: string;
  total_views: number;
  created_at: string;
  updated_at: string;
}

export interface ReactionResponse {
  message: string;
  action: string;
}

export interface UserReactionResponse {
  reaction: 'like' | 'dislike' | null;
}
