import { defineConfig, type UserConfigExport } from '@tarojs/cli';
import devConfig from './dev';
import prodConfig from './prod';

const baseConfig: UserConfigExport = {
  framework: 'react',
  projectName: 'little-bread-eats-today-mini',
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
};

export default defineConfig(async (merge) => {
  const envConfig = process.env.NODE_ENV === 'development' ? devConfig : prodConfig;
  return merge({}, baseConfig, envConfig);
});
