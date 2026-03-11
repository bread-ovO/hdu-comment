import { useState } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { Button, Drawer, Menu, Typography, Avatar, Dropdown, Tooltip } from 'antd';
import {
    MenuOutlined,
    UserOutlined,
    LogoutOutlined,
    PlusOutlined,
    FileTextOutlined,
    AuditOutlined,
    HomeOutlined,
    MoonOutlined,
    SunOutlined,
    TeamOutlined,
    BarChartOutlined
} from '@ant-design/icons';
import { useAuth } from '../../hooks/useAuth';
import { useTheme } from '../../contexts/ThemeContext';
import type { MenuProps } from 'antd';

const { Title, Text } = Typography;

const AppHeader = () => {
    const { user, logout } = useAuth();
    const location = useLocation();
    const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
    const { theme, toggleTheme } = useTheme();

    const menuItems: MenuProps['items'] = [
        {
            key: '/',
            label: <Link to="/">首页</Link>,
            icon: <HomeOutlined />
        }
    ];

    if (user) {
        menuItems.push(
            {
                key: '/submit',
                label: <Link to="/submit">提交点评</Link>,
                icon: <PlusOutlined />
            },
            {
                key: '/my',
                label: <Link to="/my">我的点评</Link>,
                icon: <FileTextOutlined />
            }
        );
    }

    if (user?.role === 'admin') {
        menuItems.push(
            {
                key: '/admin/stats',
                label: <Link to="/admin/stats">流量统计</Link>,
                icon: <BarChartOutlined />
            },
            {
                key: '/admin/users',
                label: <Link to="/admin/users">用户管理</Link>,
                icon: <TeamOutlined />
            },
            {
                key: '/admin/reviews',
                label: <Link to="/admin/reviews">审核中心</Link>,
                icon: <AuditOutlined />
            }
        );
    }

    const selectedKey = location.pathname.startsWith('/admin/stats')
        ? '/admin/stats'
        : location.pathname.startsWith('/admin/users')
        ? '/admin/users'
        : location.pathname.startsWith('/admin/reviews')
            ? '/admin/reviews'
            : location.pathname.startsWith('/submit')
            ? '/submit'
            : location.pathname.startsWith('/my')
                ? '/my'
                : '/';

    const userMenuItems: MenuProps['items'] = [
        {
            key: 'user-info',
            label: (
                <div style={{ padding: '8px 0' }}>
                    <Text strong>{user?.display_name}</Text>
                    <br />
                    <Text type="secondary" style={{ fontSize: 12 }}>
                        {user?.email}
                    </Text>
                </div>
            ),
            disabled: true,
        },
        { type: 'divider' },
        {
            key: 'logout',
            label: '退出登录',
            icon: <LogoutOutlined />,
            onClick: () => logout(),
        },
    ];

    const MobileMenu = () => (
        <Drawer
            title="菜单"
            placement="left"
            closable={false}
            onClose={() => setMobileMenuOpen(false)}
            open={mobileMenuOpen}
            width={280}
            className="mobile-menu-drawer"
            bodyStyle={{ padding: 0, background: 'var(--bg-primary)', color: 'var(--text-primary)' }}
        >
            <div style={{ padding: 16 }}>
                <Menu
                    mode="vertical"
                    selectedKeys={[selectedKey]}
                    items={menuItems}
                    style={{ border: 'none', background: 'transparent', color: 'var(--text-primary)' }}
                    theme={theme === 'dark' ? 'dark' : 'light'}
                />
            </div>
        </Drawer>
    );

    return (
        <>
            <MobileMenu />

            <header className="site-header">
                <div className="site-header-inner">
                    <Button
                        type="text"
                        icon={<MenuOutlined />}
                        onClick={() => setMobileMenuOpen(true)}
                        className="mobile-menu-btn"
                    />

                    <Link to="/" className="brand-link">
                        <div className="brand-mark" aria-hidden />
                        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                            <Title level={3} className="brand-title">
                                杭电点评
                            </Title>
                        </div>
                    </Link>

                    <nav className="desktop-nav">
                        <Menu
                            mode="horizontal"
                            selectedKeys={[selectedKey]}
                            items={menuItems}
                            style={{ border: 'none', background: 'transparent', lineHeight: '64px', flex: 1 }}
                            theme={theme === 'dark' ? 'dark' : 'light'}
                        />
                    </nav>

                    <div className="header-actions">
                        <Tooltip title={theme === 'dark' ? '切换到浅色模式' : '切换到深色模式'}>
                            <Button
                                type="text"
                                icon={theme === 'dark' ? <SunOutlined /> : <MoonOutlined />}
                                onClick={toggleTheme}
                                aria-label="切换主题模式"
                                className="theme-toggle-btn"
                            />
                        </Tooltip>
                        {user ? (
                            <Dropdown
                                menu={{ items: userMenuItems }}
                                placement="bottomRight"
                                trigger={['click']}
                            >
                                <div className="user-dropdown-trigger">
                                    <Avatar
                                        size={32}
                                        icon={<UserOutlined />}
                                        style={{ backgroundColor: 'var(--primary-color)' }}
                                    />
                                    <Text className="desktop-only">
                                        {user.display_name}
                                    </Text>
                                </div>
                            </Dropdown>
                        ) : (
                            <div className="auth-links">
                                <Link to="/login">
                                    <Button type="text">
                                        登录
                                    </Button>
                                </Link>
                                <Link to="/register">
                                    <Button type="primary">注册</Button>
                                </Link>
                            </div>
                        )}
                    </div>
                </div>
            </header>
        </>
    );
};

export default AppHeader;
