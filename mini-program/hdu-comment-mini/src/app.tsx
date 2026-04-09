import { useEffect } from 'react';
import { useAuthStore } from './store/auth';
import './app.css';

function App(props: { children?: React.ReactNode }) {
  const init = useAuthStore((state) => state.init);

  useEffect(() => {
    void init();
  }, [init]);

  useEffect(() => {
    // 微信小程序环境检测
    if (process.env.TARO_ENV === 'weapp') {
      console.log('Running in WeChat Mini Program');
    }
  }, []);

  return props.children;
}

export default App;
