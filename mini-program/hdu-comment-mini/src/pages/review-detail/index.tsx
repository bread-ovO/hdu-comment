import { useEffect, useRef, useState } from 'react';
import { View, Text, Image } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { useAuthStore } from '../../store/auth';
import { request, resolveAssetURL } from '../../adapters/request';
import NavBar from '../../components/nav-bar';
import type { ReactionResponse, ReactionType, Review, ReviewStats, UserReactionResponse } from '../../types';
import './index.css';

const buildFallbackStats = (review: Review): ReviewStats => ({
  id: '',
  review_id: review.id,
  views: review.view_count ?? 0,
  likes: review.like_count ?? 0,
  dislikes: 0,
  created_at: review.created_at,
  updated_at: review.updated_at,
});

export default function ReviewDetail() {
  const { isLoggedIn, loginByWechat } = useAuthStore();
  const [review, setReview] = useState<Review | null>(null);
  const [stats, setStats] = useState<ReviewStats | null>(null);
  const [userReaction, setUserReaction] = useState<ReactionType | null>(null);
  const [reactionLoading, setReactionLoading] = useState(false);
  const [loading, setLoading] = useState(true);
  const recordedReviewID = useRef<string | null>(null);

  useEffect(() => {
    const params = Taro.getCurrentInstance()?.router?.params;
    if (params?.id) {
      void loadReview(params.id);
    }
  }, []);

  const loadReview = async (id: string) => {
    try {
      setLoading(true);
      const reviewDetail = await request<Review>({
        url: `/reviews/${id}`,
        method: 'GET',
      });
      setReview(reviewDetail);
      setStats(buildFallbackStats(reviewDetail));

      if (reviewDetail.status === 'approved' && recordedReviewID.current !== id) {
        recordedReviewID.current = id;
        try {
          await request({
            url: `/reviews/${id}/view`,
            method: 'POST',
          });
        } catch (viewError) {
          console.error('记录浏览失败:', viewError);
        }
      }

      if (reviewDetail.status === 'approved') {
        await loadStats(id);
        if (isLoggedIn) {
          await loadUserReaction(id);
        } else {
          setUserReaction(null);
        }
      } else {
        setUserReaction(null);
      }
    } catch (error) {
      console.error('加载点评详情失败:', error);
      void Taro.showToast({ title: '加载失败', icon: 'none' });
    } finally {
      setLoading(false);
    }
  };

  const handleImagePreview = (urls: string[], current: string) => {
    void Taro.previewImage({
      urls,
      current,
    });
  };

  const loadStats = async (id: string) => {
    try {
      const latestStats = await request<ReviewStats>({
        url: `/reviews/${id}/stats`,
        method: 'GET',
      });
      setStats(latestStats);
    } catch (statsError) {
      console.error('加载点评统计失败:', statsError);
    }
  };

  const loadUserReaction = async (id: string) => {
    try {
      const response = await request<UserReactionResponse>({
        url: `/reviews/${id}/user-reaction`,
        method: 'GET',
      });
      setUserReaction(response.reaction);
    } catch (reactionError) {
      console.error('加载用户反馈失败:', reactionError);
      setUserReaction(null);
    }
  };

  const handleReaction = async (type: ReactionType) => {
    if (!review) {
      return;
    }

    if (!isLoggedIn) {
      try {
        await loginByWechat();
      } catch (loginError) {
        console.error('登录失败:', loginError);
        void Taro.showToast({ title: '请先登录后再操作', icon: 'none' });
        return;
      }
    }

    setReactionLoading(true);

    try {
      const isCanceling = userReaction === type;
      await request<ReactionResponse>({
        url: `/reviews/${review.id}/react`,
        method: 'POST',
        data: { type },
      });
      await Promise.all([loadStats(review.id), loadUserReaction(review.id)]);
      void Taro.showToast({
        title: isCanceling ? '已取消反馈' : type === 'like' ? '已点赞' : '已点踩',
        icon: 'success',
      });
    } catch (reactionError) {
      console.error('更新反馈失败:', reactionError);
      void Taro.showToast({ title: '操作失败，请重试', icon: 'none' });
    } finally {
      setReactionLoading(false);
    }
  };

  if (loading) {
    return (
      <View className="detail-page">
        <NavBar title="点评详情" showBack />
        <View className="detail-content">
          <View className="loading">加载中...</View>
        </View>
      </View>
    );
  }

  if (!review) {
    return (
      <View className="detail-page">
        <NavBar title="点评详情" showBack />
        <View className="detail-content">
          <View className="empty">点评不存在</View>
        </View>
      </View>
    );
  }

  const imageURLs = review.images?.map((image) => resolveAssetURL(image.url)) ?? [];

  return (
    <View className="detail-page">
      <NavBar title="点评详情" showBack />
      <View className="detail-content">
        <View className="detail-card">
          <View className="detail-header">
            <View className="detail-title-group">
              <Text className="detail-title">{review.title}</Text>
              <Text className="detail-address">{review.address}</Text>
            </View>
            <View className="rating-badge">
              <Text className="rating-score">{review.rating.toFixed(1)}</Text>
              <Text className="rating-star">分</Text>
            </View>
          </View>

          <View className="detail-meta">
            <Text className="meta-item">{review.author?.display_name || '匿名'}</Text>
            <Text className="meta-item">{new Date(review.created_at).toLocaleDateString()}</Text>
          </View>
        </View>

        <View className="detail-panel">
          <Text className="panel-title">用餐体验</Text>
          <Text className="content-text">{review.description}</Text>
        </View>

        {review.status === 'approved' && (
          <View className="detail-panel">
            <Text className="panel-title">你的反馈</Text>
            <Text className="reaction-hint">
              {isLoggedIn ? '觉得这条点评有帮助？点个赞或踩。再次点击同一项会取消。' : '登录后可以点赞或点踩这条点评。'}
            </Text>
            <View className="reaction-row">
              <View
                className={`reaction-button reaction-like ${userReaction === 'like' ? 'reaction-button-active' : ''} ${reactionLoading ? 'reaction-button-disabled' : ''}`}
                onClick={() => void handleReaction('like')}
              >
                <Text className={`reaction-icon ${userReaction === 'like' ? 'reaction-icon-active' : ''}`}>赞</Text>
                <Text className={`reaction-text ${userReaction === 'like' ? 'reaction-text-active' : ''}`}>点赞</Text>
                <Text className={`reaction-count ${userReaction === 'like' ? 'reaction-text-active' : ''}`}>{stats?.likes ?? 0}</Text>
              </View>
              <View
                className={`reaction-button reaction-dislike ${userReaction === 'dislike' ? 'reaction-button-active reaction-button-danger' : ''} ${reactionLoading ? 'reaction-button-disabled' : ''}`}
                onClick={() => void handleReaction('dislike')}
              >
                <Text className={`reaction-icon ${userReaction === 'dislike' ? 'reaction-icon-danger' : ''}`}>踩</Text>
                <Text className={`reaction-text ${userReaction === 'dislike' ? 'reaction-text-danger' : ''}`}>点踩</Text>
                <Text className={`reaction-count ${userReaction === 'dislike' ? 'reaction-text-danger' : ''}`}>{stats?.dislikes ?? 0}</Text>
              </View>
            </View>
          </View>
        )}

        {review.images && review.images.length > 0 && (
          <View className="detail-panel">
            <Text className="panel-title">图片</Text>
            <View className="images-grid">
              {review.images.map((image, index) => (
                <Image
                  key={image.id || index}
                  className="preview-img"
                  src={resolveAssetURL(image.url)}
                  mode="aspectFill"
                  onClick={() => handleImagePreview(imageURLs, resolveAssetURL(image.url))}
                />
              ))}
            </View>
          </View>
        )}

        {review.status === 'approved' && (
          <View className="detail-stats">
            <View className="stat-item">
              <Text className="stat-label">浏览</Text>
              <Text className="stat-value">{stats?.views ?? 0}</Text>
            </View>
            <View className="stat-item">
              <Text className="stat-label">点赞</Text>
              <Text className="stat-value">{stats?.likes ?? 0}</Text>
            </View>
            <View className="stat-item">
              <Text className="stat-label">点踩</Text>
              <Text className="stat-value">{stats?.dislikes ?? 0}</Text>
            </View>
          </View>
        )}
      </View>
    </View>
  );
}
