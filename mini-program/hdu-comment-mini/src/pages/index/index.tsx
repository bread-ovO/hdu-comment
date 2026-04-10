import { useState } from 'react';
import { View, Text, Input } from '@tarojs/components';
import Taro, { useDidShow } from '@tarojs/taro';
import { useAuthStore } from '../../store/auth';
import { request } from '../../adapters/request';
import NavBar from '../../components/nav-bar';
import type { Review } from '../../types';
import './index.css';

export default function Index() {
  const { isLoggedIn, user, loginByWechat } = useAuthStore();
  const [reviews, setReviews] = useState<Review[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');

  useDidShow(() => {
    void loadReviews();
  });

  const normalizedQuery = searchQuery.trim().toLowerCase();
  const filteredReviews = reviews.filter((review) => {
    if (!normalizedQuery) {
      return true;
    }

    return [review.title, review.address, review.description]
      .join(' ')
      .toLowerCase()
      .includes(normalizedQuery);
  });
  const averageRating = reviews.length
    ? (reviews.reduce((sum, review) => sum + review.rating, 0) / reviews.length).toFixed(1)
    : '--';

  const loadReviews = async () => {
    try {
      setLoading(true);
      const res = await request<{ data: Review[] }>({
        url: '/reviews',
        method: 'GET',
      });
      setReviews(res.data || []);
    } catch (error) {
      console.error('加载点评失败:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSearch = () => {
    console.log('搜索:', searchQuery);
  };

  const handleReviewClick = (review: Review) => {
    void Taro.navigateTo({ url: `/pages/review-detail/index?id=${review.id}` });
  };

  const handleWhatToEat = () => {
    void Taro.navigateTo({ url: '/pages/what-to-eat/index' });
  };

  const handleLogin = async () => {
    try {
      await loginByWechat();
    } catch (error) {
      console.error('登录失败:', error);
      void Taro.showToast({
        title: '登录失败，请重试',
        icon: 'none',
      });
    }
  };

  return (
    <View className="index-page">
      <NavBar title="小面包今天吃什么" />
      <View className="index-content">
        <View className="home-header">
          <Text className="home-kicker">What's Little Bread Eating Today?</Text>
          <Text className="home-title">小面包今天吃什么</Text>
          <Text className="home-subtitle">是啊今天吃什么呢？</Text>

          <View className="home-metrics">
            <View className="metric-card">
              <Text className="metric-value">{reviews.length}</Text>
              <Text className="metric-label">收录点评</Text>
            </View>
            <View className="metric-card">
              <Text className="metric-value">{averageRating}</Text>
              <Text className="metric-label">平均评分</Text>
            </View>
            <View className="metric-card">
              <Text className="metric-value">{isLoggedIn ? '已登录' : '游客'}</Text>
              <Text className="metric-label">当前状态</Text>
            </View>
          </View>
        </View>

        <View className="search-card">
          <Text className="search-label">搜索菜品、档口或关键词</Text>
          <Input
            className="search-input"
            placeholder="例如：鸡腿饭、一餐二楼、分量足"
            value={searchQuery}
            onInput={(e) => setSearchQuery(e.detail.value)}
            onConfirm={handleSearch}
          />
        </View>

        <View className="what-to-eat-card" onClick={handleWhatToEat}>
          <View className="what-to-eat-copy">
            <Text className="what-to-eat-kicker">Random Pick</Text>
            <Text className="what-to-eat-title">小面包今天吃什么</Text>
            <Text className="what-to-eat-desc">不知道吃什么的时候，就交给小面包吧！</Text>
          </View>
          <View className="what-to-eat-action">
            <Text className="what-to-eat-action-text">去抽卡！</Text>
            <Text className="what-to-eat-arrow">{'>'}</Text>
          </View>
        </View>

        {!isLoggedIn ? (
          <View className="home-banner">
            <View className="home-banner-copy">
              <Text className="home-banner-title">登录后可以发布美食点评</Text>
              <Text className="home-banner-desc">记录口味、分量、价格和排队情况，帮后来人少踩雷。</Text>
            </View>
            <View className="login-btn" onClick={handleLogin}>
              <Text className="login-btn-text">微信一键登录</Text>
            </View>
          </View>
        ) : (
          <View className="home-banner home-banner-active">
            <Text className="user-name">{user?.display_name || '用户'}</Text>
            <Text className="user-meta">已连接到你的个人美食记录</Text>
          </View>
        )}

        <View className="section-header">
          <Text className="section-title">最新点评</Text>
          <Text className="section-subtitle">{filteredReviews.length} 条结果</Text>
        </View>

        <View className="review-list">
          {loading ? (
            <View className="loading">加载中...</View>
          ) : filteredReviews.length === 0 ? (
            <View className="empty">
              <Text className="empty-title">没有找到符合条件的点评</Text>
              <Text className="empty-desc">换个关键词试试，或者登录后成为第一个写下反馈的人。</Text>
            </View>
          ) : (
            filteredReviews.map((review) => (
              <View
                className="review-card"
                key={review.id}
                onClick={() => handleReviewClick(review)}
              >
                <View className="review-header">
                  <View className="review-title-group">
                    <Text className="review-title">{review.title}</Text>
                    <Text className="review-address">{review.address}</Text>
                  </View>
                  <View className="rating">
                    <Text className="rating-score">{review.rating.toFixed(1)}</Text>
                    <Text className="rating-star">分</Text>
                  </View>
                </View>
                <Text className="review-desc">{review.description}</Text>

                <View className="review-meta">
                  <Text className="meta-chip">{review.author?.display_name || '匿名'}</Text>
                  <Text className="meta-chip">{new Date(review.created_at).toLocaleDateString()}</Text>
                </View>

                <View className="review-footer">
                  <Text className="review-link">查看详情</Text>
                  <Text className="review-arrow">{'>'}</Text>
                </View>
              </View>
            ))
          )}
        </View>
      </View>
    </View>
  );
}
