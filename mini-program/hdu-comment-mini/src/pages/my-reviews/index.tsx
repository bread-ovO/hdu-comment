import { useEffect, useState } from 'react';
import { View, Text } from '@tarojs/components';
import Taro, { useDidShow } from '@tarojs/taro';
import { useAuthStore } from '../../store/auth';
import { request } from '../../adapters/request';
import NavBar from '../../components/nav-bar';
import type { Review } from '../../types';
import './index.css';

export default function MyReviews() {
  const { isLoggedIn, user, logout, loginByWechat } = useAuthStore();
  const [reviews, setReviews] = useState<Review[]>([]);
  const [loading, setLoading] = useState(false);

  useDidShow(() => {
    if (isLoggedIn) {
      void loadMyReviews();
      return;
    }

    setReviews([]);
  });

  useEffect(() => {
    if (!isLoggedIn) {
      setReviews([]);
    }
  }, [isLoggedIn]);

  const approvedCount = reviews.filter((review) => review.status === 'approved').length;
  const pendingCount = reviews.filter((review) => review.status === 'pending').length;

  const loadMyReviews = async () => {
    try {
      setLoading(true);
      const res = await request<{ data: Review[] }>({
        url: '/reviews/me',
        method: 'GET',
      });
      setReviews(res.data || []);
    } catch (error) {
      console.error('加载我的点评失败:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleLogout = async () => {
    const res = await Taro.showModal({
      title: '提示',
      content: '确定要退出登录吗？',
    });

    if (res.confirm) {
      logout();
    }
  };

  const handleReviewClick = (review: Review) => {
    void Taro.navigateTo({ url: `/pages/review-detail/index?id=${review.id}` });
  };

  const getStatusText = (status: string) => {
    switch (status) {
      case 'pending':
        return '审核中';
      case 'approved':
        return '已发布';
      case 'rejected':
        return '未通过';
      default:
        return status;
    }
  };

  const getStatusClass = (status: string) => {
    switch (status) {
      case 'pending':
        return 'status-pending';
      case 'approved':
        return 'status-approved';
      case 'rejected':
        return 'status-rejected';
      default:
        return '';
    }
  };

  if (!isLoggedIn) {
    return (
      <View className="my-reviews-page">
        <NavBar title="我的" />
        <View className="my-reviews-content">
          <View className="login-card">
            <Text className="login-title">登录后查看我的点评</Text>
            <Text className="login-desc">这里会集中显示你写过的美食记录、审核状态和发布时间。</Text>
            <View className="login-btn" onClick={loginByWechat}>
              <Text className="login-btn-text">微信一键登录</Text>
            </View>
          </View>
        </View>
      </View>
    );
  }

  return (
    <View className="my-reviews-page">
      <NavBar title="我的" />
      <View className="my-reviews-content">
        <View className="profile-card">
          <View className="profile-header">
            <View className="profile-copy">
              <Text className="profile-title">{user?.display_name || '用户'}</Text>
              <Text className="user-email">{user?.email}</Text>
            </View>
            <View className="logout-btn" onClick={handleLogout}>
              <Text className="logout-btn-text">退出</Text>
            </View>
          </View>

          <View className="profile-metrics">
            <View className="metric-card">
              <Text className="metric-value">{reviews.length}</Text>
              <Text className="metric-label">总点评</Text>
            </View>
            <View className="metric-card">
              <Text className="metric-value">{approvedCount}</Text>
              <Text className="metric-label">已发布</Text>
            </View>
            <View className="metric-card">
              <Text className="metric-value">{pendingCount}</Text>
              <Text className="metric-label">审核中</Text>
            </View>
          </View>
        </View>

        <View className="reviews-header">
          <Text className="reviews-title">我的点评</Text>
          <Text className="reviews-count">{reviews.length} 条记录</Text>
        </View>

        <View className="reviews-list">
          {loading ? (
            <View className="loading">加载中...</View>
          ) : reviews.length === 0 ? (
            <View className="empty">
              <Text className="empty-title">还没有点评记录</Text>
              <Text className="empty-desc">去发布页写下第一条美食反馈，这里就会开始累计。</Text>
            </View>
          ) : (
            reviews.map((review) => (
              <View
                className="review-item"
                key={review.id}
                onClick={() => handleReviewClick(review)}
              >
                <View className="review-main">
                  <View className="review-title-block">
                    <Text className="review-title">{review.title}</Text>
                    <Text className="review-address">{review.address}</Text>
                  </View>
                  <View className={`review-status ${getStatusClass(review.status)}`}>
                    <Text>{getStatusText(review.status)}</Text>
                  </View>
                </View>
                <Text className="review-desc">{review.description}</Text>
                <View className="review-meta">
                  <Text className="review-rating">评分：{review.rating.toFixed(1)}</Text>
                  <Text className="review-date">
                    {new Date(review.created_at).toLocaleDateString()}
                  </Text>
                </View>
              </View>
            ))
          )}
        </View>
      </View>
    </View>
  );
}
