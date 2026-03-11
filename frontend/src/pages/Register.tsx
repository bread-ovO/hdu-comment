import { QqOutlined } from '@ant-design/icons';
import { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Alert, Button, Card, Form, Input, Tabs, Typography, message } from 'antd';
import { useAuth } from '../hooks/useAuth';
import { authApi } from '../api/auth';
import { getQQLoginURL } from '../api/client';

interface EmailRegisterValues {
  email: string;
  password: string;
  displayName: string;
  code: string;
}

interface PhoneRegisterValues {
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

const Register = () => {
  const { register, sendSMSLoginCode, loginWithSMS } = useAuth();
  const navigate = useNavigate();
  const [emailForm] = Form.useForm<EmailRegisterValues>();
  const [phoneForm] = Form.useForm<PhoneRegisterValues>();
  const [loadingEmail, setLoadingEmail] = useState(false);
  const [loadingPhone, setLoadingPhone] = useState(false);
  const [loadingQQ, setLoadingQQ] = useState(false);
  const [error, setError] = useState('');
  const [sendingEmailCode, setSendingEmailCode] = useState(false);
  const [emailCountdown, setEmailCountdown] = useState(0);
  const [sendingSMSCode, setSendingSMSCode] = useState(false);
  const [smsCountdown, setSMSCountdown] = useState(0);

  useEffect(() => {
    if (emailCountdown <= 0) return;
    const timer = window.setTimeout(() => setEmailCountdown((prev) => prev - 1), 1000);
    return () => window.clearTimeout(timer);
  }, [emailCountdown]);

  useEffect(() => {
    if (smsCountdown <= 0) return;
    const timer = window.setTimeout(() => setSMSCountdown((prev) => prev - 1), 1000);
    return () => window.clearTimeout(timer);
  }, [smsCountdown]);

  const handleSendEmailCode = async () => {
    const email = emailForm.getFieldValue('email');
    try {
      await emailForm.validateFields(['email']);
    } catch {
      return;
    }

    setError('');
    setSendingEmailCode(true);
    setEmailCountdown(60);

    try {
      await authApi.sendRegistrationCode(email);
      message.success('验证码已发送，请查收邮箱');
    } catch (err) {
      console.error(err);
      setError(readErrorMessage(err, '发送邮箱验证码失败，请稍后再试'));
    } finally {
      setSendingEmailCode(false);
    }
  };

  const handleEmailSubmit = async (values: EmailRegisterValues) => {
    setError('');
    setLoadingEmail(true);
    try {
      await register(values.email, values.password, values.displayName, values.code);
      navigate('/');
    } catch (err) {
      console.error(err);
      setError(readErrorMessage(err, '注册失败，请稍后再试'));
    } finally {
      setLoadingEmail(false);
    }
  };

  const handleSendSMSCode = async () => {
    setError('');
    try {
      const values = await phoneForm.validateFields(['phone']);
      setSendingSMSCode(true);
      const debugCode = await sendSMSLoginCode(values.phone);
      if (debugCode) {
        message.success(`验证码已发送（开发模式验证码：${debugCode}）`);
      } else {
        message.success('验证码已发送');
      }
      setSMSCountdown(60);
    } catch (err) {
      if (typeof err === 'object' && err !== null && 'errorFields' in err) {
        return;
      }
      console.error(err);
      setError(readErrorMessage(err, '发送短信验证码失败，请稍后再试'));
    } finally {
      setSendingSMSCode(false);
    }
  };

  const handlePhoneSubmit = async (values: PhoneRegisterValues) => {
    setError('');
    setLoadingPhone(true);
    try {
      // Backend auto-creates account when phone not registered.
      await loginWithSMS(values.phone, values.code);
      navigate('/');
    } catch (err) {
      console.error(err);
      setError(readErrorMessage(err, '手机号注册失败，请检查验证码'));
    } finally {
      setLoadingPhone(false);
    }
  };

  const handleQQEntry = async () => {
    setError('');
    setLoadingQQ(true);
    try {
      const { url } = await getQQLoginURL();
      window.location.href = url;
    } catch (err) {
      console.error(err);
      setError(readErrorMessage(err, 'QQ 注册暂不可用，请稍后再试'));
      setLoadingQQ(false);
    }
  };

  const tabItems = [
    {
      key: 'email',
      label: '邮箱注册',
      children: (
        <Form<EmailRegisterValues> layout="vertical" onFinish={handleEmailSubmit} form={emailForm}>
          <Form.Item label="昵称" name="displayName" rules={[{ required: true, message: '请输入昵称' }]}>
            <Input placeholder="展示名称" />
          </Form.Item>
          <Form.Item
            label="邮箱"
            name="email"
            rules={[
              { required: true, message: '请输入邮箱' },
              { type: 'email', message: '邮箱格式不正确' }
            ]}
          >
            <Input type="email" placeholder="name@example.com" />
          </Form.Item>
          <Form.Item
            label="验证码"
            required
          >
            <div className="auth-code-row">
              <Form.Item
                name="code"
                rules={[
                  { required: true, message: '请输入验证码' },
                  { len: 6, message: '验证码为6位数字' }
                ]}
                style={{ flex: 1, marginBottom: 0 }}
              >
                <Input placeholder="请输入邮箱验证码" maxLength={6} />
              </Form.Item>
              <Button
                onClick={handleSendEmailCode}
                disabled={emailCountdown > 0 || sendingEmailCode || loadingEmail}
                loading={sendingEmailCode}
                className="auth-code-button"
                style={{ minWidth: 128, flexShrink: 0 }}
              >
                {emailCountdown > 0 ? `${emailCountdown}s后重发` : '获取验证码'}
              </Button>
            </div>
          </Form.Item>
          <Form.Item label="密码" name="password" rules={[{ required: true, message: '请输入密码' }]}>
            <Input.Password placeholder="请输入密码" />
          </Form.Item>
          <Button type="primary" htmlType="submit" block loading={loadingEmail}>
            注册
          </Button>
        </Form>
      )
    },
    {
      key: 'phone',
      label: '手机号注册',
      children: (
        <Form<PhoneRegisterValues> layout="vertical" onFinish={handlePhoneSubmit} form={phoneForm}>
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
            required
          >
            <div className="auth-code-row">
              <Form.Item
                name="code"
                rules={[
                  { required: true, message: '请输入验证码' },
                  { len: 6, message: '验证码为 6 位数字' }
                ]}
                style={{ flex: 1, marginBottom: 0 }}
              >
                <Input placeholder="请输入短信验证码" maxLength={6} />
              </Form.Item>
              <Button
                onClick={handleSendSMSCode}
                disabled={smsCountdown > 0 || sendingSMSCode || loadingPhone}
                loading={sendingSMSCode}
                className="auth-code-button"
                style={{ minWidth: 128, flexShrink: 0 }}
              >
                {smsCountdown > 0 ? `${smsCountdown}s后重发` : '获取验证码'}
              </Button>
            </div>
          </Form.Item>
          <Typography.Paragraph type="secondary" style={{ marginTop: -4 }}>
            手机号首次验证会自动创建账号并完成登录。
          </Typography.Paragraph>
          <Button type="primary" htmlType="submit" block loading={loadingPhone}>
            注册并登录
          </Button>
        </Form>
      )
    }
  ];

  return (
    <Card className="auth-card">
      <Typography.Title level={3}>注册</Typography.Title>
      {error && <Alert type="error" message={error} style={{ marginBottom: 16 }} />}
      <Tabs defaultActiveKey="email" items={tabItems} animated={false} destroyOnHidden />
      <Button
        icon={<QqOutlined />}
        block
        onClick={handleQQEntry}
        loading={loadingQQ}
        style={{ marginTop: 14 }}
      >
        QQ 一键接入
      </Button>
      <Typography.Paragraph style={{ marginTop: 16 }}>
        已有账号？<Link to="/login">直接登录</Link>
      </Typography.Paragraph>
    </Card>
  );
};

export default Register;
