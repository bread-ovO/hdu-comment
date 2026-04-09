import type { UserConfigExport } from '@tarojs/cli';

export default {
  framework: 'react',
  projectName: 'hdu-comment-mini',
  date: '2026-04-08',
  designWidth: 750,
  deviceRatio: {
    640: 2.34 / 2,
    750: 1,
    828: 1.81 / 2,
  },
  sourceRoot: 'src',
  outputRoot: 'dist',
  plugins: [
    '@tarojs/plugin-framework-react',
    '@tarojs/plugin-platform-weapp',
    '@tarojs/plugin-platform-h5',
  ],
  defineConstants: {},
  compiler: {
    prebundle: { enable: false },
  },
  mini: {
    webpackChain(chain: any) {},
  },
  h5: {
    publicPath: '/',
    staticDirectory: 'static',
    webpackChain(chain: any) {},
  },
} as UserConfigExport;
