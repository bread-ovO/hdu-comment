export default {
  lazyCodeLoading: 'requiredComponents',
  pages: [
    'pages/index/index',
    'pages/login/index',
    'pages/review-detail/index',
    'pages/submit-review/index',
    'pages/my-reviews/index',
  ],
  window: {
    backgroundTextStyle: 'light',
    navigationStyle: 'custom',
    navigationBarBackgroundColor: '#2563eb',
    navigationBarTitleText: '杭电美食点评',
    navigationBarTextStyle: 'white',
  },
  tabBar: {
    color: '#64748b',
    selectedColor: '#2563eb',
    backgroundColor: '#ffffff',
    borderStyle: 'black',
    list: [
      {
        pagePath: 'pages/index/index',
        text: '首页',
        iconPath: 'assets/home.png',
        selectedIconPath: 'assets/home-active.png',
      },
      {
        pagePath: 'pages/submit-review/index',
        text: '发布',
        iconPath: 'assets/edit.png',
        selectedIconPath: 'assets/edit-active.png',
      },
      {
        pagePath: 'pages/my-reviews/index',
        text: '我的',
        iconPath: 'assets/user.png',
        selectedIconPath: 'assets/user-active.png',
      },
    ],
  },
};
