import type { UserConfigExport } from '@tarojs/cli';

const apiBaseURL = process.env.TARO_APP_API_BASE_URL ?? 'http://127.0.0.1:8080/api/v1';
const assetBaseURL = process.env.TARO_APP_ASSET_BASE_URL ?? apiBaseURL.replace(/\/api\/v1\/?$/, '');

const devConfig: UserConfigExport = {
  defineConstants: {
    'process.env.TARO_APP_ENV': JSON.stringify('development'),
    'process.env.TARO_APP_API_BASE_URL': JSON.stringify(apiBaseURL),
    'process.env.TARO_APP_ASSET_BASE_URL': JSON.stringify(assetBaseURL),
    'process.env.TARO_APP_ENABLE_DEBUG': JSON.stringify('true'),
  },
};

export default devConfig;
