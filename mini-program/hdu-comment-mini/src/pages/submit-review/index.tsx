import { useState } from 'react';
import { View, Text, Input, Textarea, Button, Image } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { useAuthStore } from '../../store/auth';
import { request, uploadReviewImage } from '../../adapters/request';
import NavBar from '../../components/nav-bar';
import './index.css';

export default function SubmitReview() {
  const { isLoggedIn, loginByWechat } = useAuthStore();
  const [title, setTitle] = useState('');
  const [address, setAddress] = useState('');
  const [description, setDescription] = useState('');
  const [rating, setRating] = useState(5);
  const [images, setImages] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);

  const handleChooseImage = () => {
    if (!isLoggedIn) {
      void Taro.showToast({ title: '请先登录', icon: 'none' });
      return;
    }

    void Taro.chooseImage({
      count: 9 - images.length,
      sizeType: ['compressed'],
      sourceType: ['album', 'camera'],
    }).then((res) => {
      setImages((prev) => [...prev, ...res.tempFilePaths]);
    });
  };

  const handleRemoveImage = (index: number) => {
    const newImages = [...images];
    newImages.splice(index, 1);
    setImages(newImages);
  };

  const handleSubmit = async () => {
    if (!isLoggedIn) {
      await loginByWechat();
      return;
    }

    if (!title.trim()) {
      void Taro.showToast({ title: '请输入菜品名称', icon: 'none' });
      return;
    }

    if (!address.trim()) {
      void Taro.showToast({ title: '请输入购买地点', icon: 'none' });
      return;
    }

    if (!description.trim()) {
      void Taro.showToast({ title: '请输入点评内容', icon: 'none' });
      return;
    }

    try {
      setLoading(true);

      // 提交点评
      const review = await request<{ id: string }>({
        url: '/reviews',
        method: 'POST',
        data: {
          title: title.trim(),
          address: address.trim(),
          description: description.trim(),
          rating,
        },
      });

      // 上传图片
      for (const imagePath of images) {
        await uploadReviewImage(review.id, imagePath);
      }

      void Taro.showToast({ title: '发布成功', icon: 'success' });

      // 清空表单
      setTitle('');
      setAddress('');
      setDescription('');
      setRating(5);
      setImages([]);

      // 跳转到列表页
      setTimeout(() => {
        void Taro.switchTab({ url: '/pages/index/index' });
      }, 1500);
    } catch (error) {
      console.error('发布失败:', error);
      void Taro.showToast({ title: '发布失败，请重试', icon: 'none' });
    } finally {
      setLoading(false);
    }
  };

  const handleRatingClick = (score: number) => {
    setRating(score);
  };

  if (!isLoggedIn) {
    return (
      <View className="submit-page">
        <NavBar title="发布点评" />
        <View className="submit-content">
          <View className="submit-hero">
            <Text className="hero-title">发布美食点评</Text>
            <Text className="hero-desc">保持和 Web 一致的卡片式录入体验，登录后即可开始记录。</Text>
          </View>

          <View className="login-card">
            <Text className="login-tip">登录后可以发布点评、上传图片并保存自己的美食档案。</Text>
            <Button className="login-btn" type="primary" onClick={loginByWechat}>
              微信一键登录
            </Button>
          </View>
        </View>
      </View>
    );
  }

  return (
    <View className="submit-page">
      <NavBar title="发布点评" />
      <View className="submit-content">
        <View className="submit-hero">
          <Text className="hero-title">发布美食点评</Text>
          <Text className="hero-desc">写下口味、分量、价格和排队体验，让后来的同学先看到真实反馈。</Text>
        </View>

        <View className="form-section">
          <Text className="label">菜品名称 *</Text>
          <Text className="hint">和 Web 端一样，优先填写大家最常搜索的菜品名。</Text>
          <Input
            className="input"
            placeholder="请输入菜品名称"
            value={title}
            onInput={(e) => setTitle(e.detail.value)}
          />
        </View>

        <View className="form-section">
          <Text className="label">购买地点 *</Text>
          <Text className="hint">例如一餐二楼、东门档口或具体窗口名称。</Text>
          <Input
            className="input"
            placeholder="如：一餐二楼麻辣香锅"
            value={address}
            onInput={(e) => setAddress(e.detail.value)}
          />
        </View>

        <View className="form-section">
          <Text className="label">评分 *</Text>
          <Text className="hint">用 1-5 分快速概括你的整体体验。</Text>
          <View className="rating-row">
            {[1, 2, 3, 4, 5].map((score) => (
              <Text
                key={score}
                className={`star ${score <= rating ? 'active' : ''}`}
                onClick={() => handleRatingClick(score)}
              >
                ★
              </Text>
            ))}
            <Text className="rating-text">{rating.toFixed(1)}</Text>
          </View>
        </View>

        <View className="form-section">
          <Text className="label">点评内容 *</Text>
          <Text className="hint">写清楚口味、分量、价格、排队时间或值不值得回购。</Text>
          <Textarea
            className="textarea"
            placeholder="分享你的口味感受、分量、性价比和踩雷点..."
            value={description}
            onInput={(e) => setDescription(e.detail.value)}
            maxlength={2000}
          />
        </View>

        <View className="form-section">
          <Text className="label">上传图片（可选）</Text>
          <Text className="hint">支持上传菜品照片或档口环境图，注意避免包含敏感信息。</Text>
          <View className="image-grid">
            {images.map((url, index) => (
              <View key={index} className="image-item">
                <Image className="preview-img" src={url} mode="aspectFill" />
                <Text
                  className="remove-btn"
                  onClick={() => handleRemoveImage(index)}
                >
                  ×
                </Text>
              </View>
            ))}
            {images.length < 9 && (
              <View className="add-image" onClick={handleChooseImage}>
                <Text className="add-icon">+</Text>
                <Text className="add-text">添加图片</Text>
              </View>
            )}
          </View>
        </View>

        <Button
          className="submit-btn"
          type="primary"
          loading={loading}
          onClick={handleSubmit}
        >
          发布点评
        </Button>
      </View>
    </View>
  );
}
