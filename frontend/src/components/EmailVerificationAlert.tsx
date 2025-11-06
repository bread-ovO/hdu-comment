import { useEffect, useState } from 'react';
import { Alert, Button, Space, message } from 'antd';
import { MailOutlined } from '@ant-design/icons';
import { authApi } from '../api/auth';
import { useAuth } from '../hooks/useAuth';

const EmailVerificationAlert = () => {
    const { user } = useAuth();
    const [loading, setLoading] = useState(false);
    const [showAlert, setShowAlert] = useState(false);

    useEffect(() => {
        if (user && user.email_verified === false) {
            setShowAlert(true);
        } else {
            setShowAlert(false);
        }
    }, [user]);

    const handleSendVerification = async () => {
        setLoading(true);
        try {
            await authApi.sendVerificationEmail();
            message.success('验证邮件已发送，请查收邮箱');
        } catch (error) {
            console.error('Failed to send verification email:', error);
            message.error('发送验证邮件失败，请稍后重试');
        } finally {
            setLoading(false);
        }
    };

    const handleDismiss = () => {
        setShowAlert(false);
    };

    if (!user || user.email_verified !== false || !showAlert) {
        return null;
    }

    return (
        <Alert
            message="邮箱未验证"
            description={
                <Space direction="vertical" size="small" style={{ width: '100%' }}>
                    <span>请验证您的邮箱地址以使用所有功能。</span>
                    <Space>
                        <Button
                            type="primary"
                            size="small"
                            icon={<MailOutlined />}
                            loading={loading}
                            onClick={handleSendVerification}
                        >
                            发送验证邮件
                        </Button>
                        <Button size="small" onClick={handleDismiss}>
                            稍后提醒
                        </Button>
                    </Space>
                </Space>
            }
            type="warning"
            showIcon
            style={{ marginBottom: 16 }}
        />
    );
};

export default EmailVerificationAlert;
