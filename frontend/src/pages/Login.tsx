import { QqOutlined } from '@ant-design/icons';
import { Alert, Button, Card, Form, Input, Tabs, Typography, message } from 'antd';
import { useEffect, useState } from 'react';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';
import { getQQLoginURL } from '../api/client';
import { useAuth } from '../hooks/useAuth';

interface PasswordLoginValues {
  email: string;
  password: string;
}

interface SMSLoginValues {
  phone: string;
  code: string;
}

const readErrorMessage = (err: unknown, fallback: string): string => {
  if (typeof err !== 'object' || err === null || !('response' in err)) {
    return fallback;
  }
  const response = (err as { response?: { data?: { error?: string } } }).response;
  return response?.data?.error ?? fallback;
};

const Login = () => {
  const { login, loginWithQQ, sendSMSLoginCode, loginWithSMS } = useAuth();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const [smsForm] = Form.useForm<SMSLoginValues>();
  const [loadingPassword, setLoadingPassword] = useState(false);
  const [loadingSMS, setLoadingSMS] = useState(false);
  const [loadingQQ, setLoadingQQ] = useState(false);
  const [sendingSMSCode, setSendingSMSCode] = useState(false);
  const [countdown, setCountdown] = useState(0);
  const [handledQQCode, setHandledQQCode] = useState('');
  const [error, setError] = useState('');

  useEffect(() => {
    if (countdown <= 0) return undefined;
    const timer = window.setInterval(() => {
      setCountdown((prev) => (prev <= 1 ? 0 : prev - 1));
    }, 1000);
    return () => window.clearInterval(timer);
  }, [countdown]);

  useEffect(() => {
    const code = searchParams.get('code');
    const state = searchParams.get('state');
    if (!code || !state) return;

    const callbackToken = `${code}:${state}`;
    if (callbackToken === handledQQCode) return;

    const run = async () => {
      setHandledQQCode(callbackToken);
      setLoadingQQ(true);
      setError('');
      try {
        await loginWithQQ(code, state);
        navigate('/', { replace: true });
      } catch (err) {
        console.error(err);
        setError(readErrorMessage(err, 'QQ 登录失败，请重试'));
        navigate('/login', { replace: true });
      } finally {
        setLoadingQQ(false);
      }
    };

    void run();
  }, [searchParams, handledQQCode, loginWithQQ, navigate]);

  const handlePasswordSubmit = async (values: PasswordLoginValues) => {
    setLoadingPassword(true);
    setError('');
    try {
      await login(values.email, values.password);
      navigate('/');
    } catch (err) {
      console.error(err);
      setError(readErrorMessage(err, '登录失败，请检查账号或密码'));
    } finally {
      setLoadingPassword(false);
    }
  };

  const handleSendSMSCode = async () => {
    setError('');
    try {
      const values = await smsForm.validateFields(['phone']);
      setSendingSMSCode(true);
      const debugCode = await sendSMSLoginCode(values.phone);
      if (debugCode) {
        message.success(`验证码已发送（开发模式验证码：${debugCode}）`);
      } else {
        message.success('验证码已发送');
      }
      setCountdown(60);
    } catch (err) {
      if (typeof err === 'object' && err !== null && 'errorFields' in err) {
        return;
      }
      console.error(err);
      setError(readErrorMessage(err, '发送验证码失败，请稍后再试'));
    } finally {
      setSendingSMSCode(false);
    }
  };

  const handleSMSSubmit = async (values: SMSLoginValues) => {
    setLoadingSMS(true);
    setError('');
    try {
      await loginWithSMS(values.phone, values.code);
      navigate('/');
    } catch (err) {
      console.error(err);
      setError(readErrorMessage(err, '短信登录失败，请检查手机号和验证码'));
    } finally {
      setLoadingSMS(false);
    }
  };

  const handleQQEntry = async () => {
    setLoadingQQ(true);
    setError('');
    try {
      const { url } = await getQQLoginURL();
      window.location.href = url;
    } catch (err) {
      console.error(err);
      setError(readErrorMessage(err, 'QQ 登录暂不可用，请稍后再试'));
      setLoadingQQ(false);
    }
  };

  const tabItems = [
    {
      key: 'password',
      label: '账号密码登录',
      children: (
        <Form<PasswordLoginValues> layout="vertical" onFinish={handlePasswordSubmit}>
          <Form.Item label="邮箱" name="email" rules={[{ required: true, message: '请输入邮箱' }]}>
            <Input type="email" placeholder="name@example.com" />
          </Form.Item>
          <Form.Item label="密码" name="password" rules={[{ required: true, message: '请输入密码' }]}>
            <Input.Password placeholder="请输入密码" />
          </Form.Item>
          <Button type="primary" htmlType="submit" block loading={loadingPassword}>
            登录
          </Button>
        </Form>
      )
    },
    {
      key: 'sms',
      label: '手机号登录',
      children: (
        <Form<SMSLoginValues> form={smsForm} layout="vertical" onFinish={handleSMSSubmit}>
          <Form.Item
            label="手机号"
            name="phone"
            rules={[
              { required: true, message: '请输入手机号' },
              { pattern: /^(\+?86)?1[3-9]\d{9}$/, message: '请输入有效的手机号' }
            ]}
          >
            <Input placeholder="请输入手机号（支持 +86）" />
          </Form.Item>
          <Form.Item
            label="验证码"
            name="code"
            rules={[
              { required: true, message: '请输入验证码' },
              { len: 6, message: '验证码为 6 位数字' }
            ]}
          >
            <Input
              placeholder="请输入短信验证码"
              addonAfter={(
                <Button
                  type="link"
                  onClick={handleSendSMSCode}
                  loading={sendingSMSCode}
                  disabled={countdown > 0}
                  style={{ paddingInline: 0 }}
                >
                  {countdown > 0 ? `${countdown}s` : '获取验证码'}
                </Button>
              )}
            />
          </Form.Item>
          <Button type="primary" htmlType="submit" block loading={loadingSMS}>
            登录
          </Button>
        </Form>
      )
    }
  ];

  return (
    <Card style={{ maxWidth: 420, margin: '48px auto' }}>
      <Typography.Title level={3}>登录</Typography.Title>
      {error && <Alert type="error" message={error} style={{ marginBottom: 16 }} />}
      <Tabs defaultActiveKey="password" items={tabItems} />
      <Button icon={<QqOutlined />} block onClick={handleQQEntry} loading={loadingQQ}>
        QQ 一键登录
      </Button>
      <Typography.Paragraph style={{ marginTop: 16 }}>
        还没有账号？<Link to="/register">立即注册</Link>
      </Typography.Paragraph>
    </Card>
  );
};

export default Login;
