import type { UserConfigExport } from '@tarojs/cli';

const apiBaseURL = process.env.TARO_APP_API_BASE_URL ?? 'https://api.example.com/api/v1';
const assetBaseURL = process.env.TARO_APP_ASSET_BASE_URL ?? apiBaseURL.replace(/\/api\/v1\/?$/, '');
const appEnv = process.env.TARO_APP_ENV ?? 'production';
const enableDebug = process.env.TARO_APP_ENABLE_DEBUG ?? 'false';

const prodConfig: UserConfigExport = {
  defineConstants: {
    'process.env.TARO_APP_ENV': JSON.stringify(appEnv),
    'process.env.TARO_APP_API_BASE_URL': JSON.stringify(apiBaseURL),
    'process.env.TARO_APP_ASSET_BASE_URL': JSON.stringify(assetBaseURL),
    'process.env.TARO_APP_ENABLE_DEBUG': JSON.stringify(enableDebug),
  },
};

export default prodConfig;
