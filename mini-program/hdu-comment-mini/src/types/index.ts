export interface User {
  id: string;
  email: string;
  display_name: string;
  avatar_url?: string;
  email_verified: boolean;
  role: 'user' | 'admin';
  created_at: string;
}

export interface ReviewImage {
  id: string;
  review_id: string;
  storage_key: string;
  url: string;
  created_at: string;
}

export interface Review {
  id: string;
  title: string;
  address: string;
  description: string;
  rating: number;
  author_id: string;
  author?: User;
  images?: ReviewImage[];
  view_count: number;
  like_count: number;
  status: 'pending' | 'approved' | 'rejected';
  created_at: string;
  updated_at: string;
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

export type ReactionType = 'like' | 'dislike';

export interface ReactionResponse {
  message: string;
}

export interface UserReactionResponse {
  reaction: ReactionType | null;
}

export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  user: User;
}

export interface PaginatedResponse<T> {
  data: T[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
  };
}
