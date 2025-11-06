import { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Alert, Button, Card, Form, Input, Typography, message } from 'antd';
import { useAuth } from '../hooks/useAuth';
import { authApi } from '../api/auth';

const Register = () => {
  const { register } = useAuth();
  const navigate = useNavigate();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [sendingCode, setSendingCode] = useState(false);
  const [countdown, setCountdown] = useState(0);

  useEffect(() => {
    if (countdown <= 0) return;
    const timer = window.setTimeout(() => setCountdown((prev) => prev - 1), 1000);
    return () => window.clearTimeout(timer);
  }, [countdown]);

  const handleSendCode = async () => {
    const email = form.getFieldValue('email');
    try {
      await form.validateFields(['email']);
    } catch {
      return;
    }

    setError('');
    setSendingCode(true);
    setCountdown(60);

    try {
      await authApi.sendRegistrationCode(email);
      message.success('验证码已发送，请查收邮箱');
    } catch (err) {
      console.error(err);
    } finally {
      setSendingCode(false);
    }
  };

  const handleSubmit = async (values: { email: string; password: string; displayName: string; code: string }) => {
    setError('');
    setLoading(true);
    try {
      await register(values.email, values.password, values.displayName, values.code);
      navigate('/');
    } catch (err: any) {
      console.error(err);
      setError(err?.response?.data?.error || '注册失败，请稍后再试');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card style={{ maxWidth: 420, margin: '48px auto' }}>
      <Typography.Title level={3}>注册</Typography.Title>
      <Form layout="vertical" onFinish={handleSubmit} form={form}>
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
          name="code"
          rules={[
            { required: true, message: '请输入验证码' },
            { len: 6, message: '验证码为6位数字' }
          ]}
        >
          <Input
            placeholder="请输入邮箱验证码"
            maxLength={6}
            addonAfter={
              <Button
                type="link"
                onClick={handleSendCode}
                disabled={countdown > 0 || sendingCode || loading}
                loading={sendingCode}
              >
                {countdown > 0 ? `${countdown}s后重发` : '获取验证码'}
              </Button>
            }
          />
        </Form.Item>
        <Form.Item label="密码" name="password" rules={[{ required: true, message: '请输入密码' }]}> 
          <Input.Password placeholder="请输入密码" />
        </Form.Item>
        {error && <Alert type="error" message={error} style={{ marginBottom: 16 }} />}
        <Button type="primary" htmlType="submit" block loading={loading}>
          注册
        </Button>
      </Form>
      <Typography.Paragraph style={{ marginTop: 16 }}>
        已有账号？<Link to="/login">直接登录</Link>
      </Typography.Paragraph>
    </Card>
  );
};

export default Register;
