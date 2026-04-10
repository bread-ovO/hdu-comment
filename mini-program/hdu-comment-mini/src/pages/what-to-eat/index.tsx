import { useState } from 'react';
import { View, Text, Button } from '@tarojs/components';
import Taro, { useDidShow } from '@tarojs/taro';
import NavBar from '../../components/nav-bar';
import { request } from '../../adapters/request';
import type { PaginatedResponse, Review } from '../../types';
import './index.css';

function formatDate(dateString: string): string {
  return new Date(dateString).toLocaleDateString();
}

export default function WhatToEat() {
  const [review, setReview] = useState<Review | null>(null);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [totalReviews, setTotalReviews] = useState(0);

  useDidShow(() => {
    if (!review) {
      void drawRandomReview();
    }
  });

  const drawRandomReview = async () => {
    const shouldShowInitialLoading = !review;

    try {
      if (shouldShowInitialLoading) {
        setLoading(true);
      } else {
        setRefreshing(true);
      }

      const meta = await request<PaginatedResponse<Review>>({
        url: '/reviews?page=1&page_size=1&sort=created_at&order=desc',
        method: 'GET',
      });

      const total = meta.pagination?.total ?? 0;
      setTotalReviews(total);

      if (total === 0) {
        setReview(null);
        return;
      }

      const randomPage = Math.floor(Math.random() * total) + 1;
      const randomResult = await request<PaginatedResponse<Review>>({
        url: `/reviews?page=${randomPage}&page_size=1&sort=created_at&order=desc`,
        method: 'GET',
      });

      setReview(randomResult.data?.[0] ?? null);
    } catch (error) {
      console.error('随机推荐失败:', error);
      void Taro.showToast({
        title: '推荐失败，请重试',
        icon: 'none',
      });
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  };

  const handleReviewClick = () => {
    if (!review) {
      return;
    }

    void Taro.navigateTo({ url: `/pages/review-detail/index?id=${review.id}` });
  };

  return (
    <View className="what-page">
      <NavBar title="小面包今天吃什么" showBack />
      <View className="what-content">
        <View className="what-hero">
          <Text className="what-kicker">Random Pick</Text>
          <Text className="what-title">今天吃这个试试</Text>
          <Text className="what-desc">从已发布点评里随机抽一条，帮你在犹豫的时候更快做决定。</Text>
        </View>

        <View className="what-toolbar">
          <View className="what-toolbar-copy">
            <Text className="what-toolbar-label">可抽取点评</Text>
            <Text className="what-toolbar-value">{totalReviews || '--'}</Text>
          </View>
          <Button className="what-refresh-btn" onClick={() => void drawRandomReview()} loading={refreshing}>
            再抽一次
          </Button>
        </View>

        {loading ? (
          <View className="what-state-card">
            <Text className="what-state-title">小面包正在翻点评库</Text>
            <Text className="what-state-desc">稍等一下，马上给你一份随机推荐。</Text>
          </View>
        ) : !review ? (
          <View className="what-state-card">
            <Text className="what-state-title">还没有可推荐的点评</Text>
            <Text className="what-state-desc">等有人发布并审核通过后，这里就能帮你随机抽签了。</Text>
          </View>
        ) : (
          <View className="what-review-card" onClick={handleReviewClick}>
            <View className="what-review-header">
              <View className="what-review-main">
                <Text className="what-review-title">{review.title}</Text>
                <Text className="what-review-address">{review.address}</Text>
              </View>
              <View className="what-rating-badge">
                <Text className="what-rating-score">{review.rating.toFixed(1)}</Text>
                <Text className="what-rating-unit">分</Text>
              </View>
            </View>

            <View className="what-review-meta">
              <Text className="what-meta-chip">{review.author?.display_name || '匿名'}</Text>
              <Text className="what-meta-chip">{formatDate(review.created_at)}</Text>
            </View>

            <Text className="what-review-desc">{review.description}</Text>

            <View className="what-review-footer">
              <Text className="what-review-hint">点开查看完整点评和图片</Text>
              <Text className="what-review-arrow">{'>'}</Text>
            </View>
          </View>
        )}
      </View>
    </View>
  );
}
