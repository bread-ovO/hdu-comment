export default {
  lazyCodeLoading: 'requiredComponents',
  pages: [
    'pages/index/index',
    'pages/what-to-eat/index',
    'pages/login/index',
    'pages/review-detail/index',
    'pages/submit-review/index',
    'pages/my-reviews/index',
  ],
  window: {
    backgroundTextStyle: 'light',
    navigationStyle: 'custom',
    navigationBarBackgroundColor: '#0f766e',
    navigationBarTitleText: '小面包今天吃什么',
    navigationBarTextStyle: 'white',
  },
  tabBar: {
    color: '#64748b',
    selectedColor: '#0f766e',
    backgroundColor: '#fffdf8',
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
