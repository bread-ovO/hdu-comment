import { useState } from 'react';
import { View, Text, Button } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { useAuthStore } from '../../store/auth';
import NavBar from '../../components/nav-bar';
import './index.css';

export default function Login() {
  const { loginByWechat } = useAuthStore();
  const [loading, setLoading] = useState(false);

  const handleWxLogin = async () => {
    try {
      setLoading(true);
      await loginByWechat();
      // 登录成功后返回上一页
      await Taro.navigateBack();
    } catch (error) {
      console.error('登录失败:', error);
      void Taro.showToast({
        title: '登录失败，请重试',
        icon: 'none',
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <View className="login-page">
      <NavBar title="微信登录" showBack />
      <View className="login-content">
        <View className="login-container">
          <View className="login-header">
            <Text className="login-kicker">Dining Login</Text>
            <Text className="app-name">杭电美食点评</Text>
            <Text className="app-desc">延续 Web 端的简洁信息卡风格，在小程序里直接登录和查看食堂反馈。</Text>
          </View>

          <View className="login-panel">
            <Text className="panel-title">微信快捷登录</Text>
            <Text className="panel-desc">登录后可以发布点评、查看自己的美食记录，并同步保存访问状态。</Text>
            <Button
              className="wx-login-btn"
              type="primary"
              loading={loading}
              onClick={handleWxLogin}
            >
              微信一键登录
            </Button>

            <Text className="login-tip">
              登录即表示同意《用户协议》和《隐私政策》
            </Text>
          </View>
        </View>
      </View>
    </View>
  );
}
