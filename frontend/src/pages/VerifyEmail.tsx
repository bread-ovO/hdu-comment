import { useEffect, useState } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { Button, Card, Result, Spin, Typography } from 'antd';
import { CheckCircleOutlined, CloseCircleOutlined } from '@ant-design/icons';
import { authApi } from '../api/auth';
import { useAuth } from '../hooks/useAuth';

const { Title, Paragraph } = Typography;

const VerifyEmail = () => {
    const [searchParams] = useSearchParams();
    const navigate = useNavigate();
    const { user, refreshProfile } = useAuth();
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    const [success, setSuccess] = useState(false);

    const token = searchParams.get('token');

    useEffect(() => {
        if (!token) {
            setLoading(false);
            setError('无效的验证链接');
            return;
        }

        const verifyEmail = async () => {
            try {
                await authApi.verifyEmail(token);
                if (user) {
                    try {
                        await refreshProfile();
                    } catch (refreshErr) {
                        console.warn('Failed to refresh profile after verification', refreshErr);
                    }
                }
                setSuccess(true);
            } catch (err: any) {
                setError(err.response?.data?.error || '验证失败');
            } finally {
                setLoading(false);
            }
        };

        verifyEmail();
    }, [token, user, refreshProfile]);

    if (loading) {
        return (
            <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '60vh' }}>
                <Spin size="large" />
            </div>
        );
    }

    if (success) {
        return (
            <div style={{ maxWidth: 600, margin: '48px auto', padding: '0 16px' }}>
                <Result
                    status="success"
                    icon={<CheckCircleOutlined />}
                    title="邮箱验证成功！"
                    subTitle="您的邮箱地址已成功验证，现在可以开始使用所有功能了。"
                    extra={[
                        <Button type="primary" key="home" onClick={() => navigate('/')}>
                            返回首页
                        </Button>,
                        <Button key="profile" onClick={() => navigate('/profile')}>
                            查看个人资料
                        </Button>,
                    ]}
                />
            </div>
        );
    }

    return (
        <div style={{ maxWidth: 600, margin: '48px auto', padding: '0 16px' }}>
            <Card>
                <Title level={3} style={{ textAlign: 'center', marginBottom: 24 }}>
                    邮箱验证
                </Title>

                <Result
                    status="error"
                    icon={<CloseCircleOutlined />}
                    title="验证失败"
                    subTitle={error}
                    extra={
                        <div style={{ textAlign: 'center' }}>
                            <Paragraph>
                                验证链接可能已过期或无效。请重新发送验证邮件。
                            </Paragraph>
                            <Button type="primary" onClick={() => navigate('/login')}>
                                前往登录
                            </Button>
                        </div>
                    }
                />
            </Card>
        </div>
    );
};

export default VerifyEmail;
