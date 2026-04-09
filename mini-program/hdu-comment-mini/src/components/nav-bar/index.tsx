import { View, Text } from '@tarojs/components';
import Taro from '@tarojs/taro';
import './index.css';

interface NavMetrics {
  statusBarHeight: number;
  navBarHeight: number;
  totalHeight: number;
  sideWidth: number;
  horizontalPadding: number;
}

interface NavBarProps {
  title: string;
  showBack?: boolean;
}

const FALLBACK_SIDE_WIDTH = 96;
const FALLBACK_NAV_HEIGHT = 44;
const HORIZONTAL_PADDING = 16;

const getNavMetrics = (): NavMetrics => {
  const systemInfo = Taro.getSystemInfoSync();
  const statusBarHeight = systemInfo.statusBarHeight ?? 20;

  if (process.env.TARO_ENV !== 'weapp' || typeof Taro.getMenuButtonBoundingClientRect !== 'function') {
    const navBarHeight = FALLBACK_NAV_HEIGHT;
    return {
      statusBarHeight,
      navBarHeight,
      totalHeight: statusBarHeight + navBarHeight,
      sideWidth: FALLBACK_SIDE_WIDTH,
      horizontalPadding: HORIZONTAL_PADDING,
    };
  }

  const menuButton = Taro.getMenuButtonBoundingClientRect();
  const gap = Math.max(menuButton.top - statusBarHeight, 6);
  const navBarHeight = menuButton.height + gap * 2;
  const rightReserved = Math.max(systemInfo.windowWidth - menuButton.left + HORIZONTAL_PADDING, FALLBACK_SIDE_WIDTH);

  return {
    statusBarHeight,
    navBarHeight,
    totalHeight: statusBarHeight + navBarHeight,
    sideWidth: rightReserved,
    horizontalPadding: HORIZONTAL_PADDING,
  };
};

const navigateBack = () => {
  const pages = Taro.getCurrentPages();
  if (pages.length > 1) {
    void Taro.navigateBack();
    return;
  }

  void Taro.switchTab({ url: '/pages/index/index' });
};

export default function NavBar({ title, showBack = false }: NavBarProps) {
  const metrics = getNavMetrics();

  return (
    <View
      className="nav-bar"
      style={{
        paddingTop: `${metrics.statusBarHeight}px`,
        height: `${metrics.totalHeight}px`,
      }}
    >
      <View
        className="nav-bar-content"
        style={{
          height: `${metrics.navBarHeight}px`,
          paddingLeft: `${metrics.horizontalPadding}px`,
          paddingRight: `${metrics.horizontalPadding}px`,
        }}
      >
        <View className="nav-bar-side" style={{ width: `${metrics.sideWidth}px` }}>
          {showBack ? (
            <View className="nav-back-button" onClick={navigateBack}>
              <Text className="nav-back-icon">‹</Text>
              <Text className="nav-back-text">返回</Text>
            </View>
          ) : null}
        </View>

        <View className="nav-bar-center">
          <Text className="nav-bar-title">{title}</Text>
        </View>

        <View className="nav-bar-side" style={{ width: `${metrics.sideWidth}px` }} />
      </View>
    </View>
  );
}
