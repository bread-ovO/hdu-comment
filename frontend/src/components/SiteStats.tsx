import { useEffect, useState } from 'react';
import { Card, Statistic, Space } from 'antd';
import { EyeOutlined } from '@ant-design/icons';
import { statsApi } from '../api/stats';
import type { SiteStats } from '../types';

const SiteStats = () => {
    const [stats, setStats] = useState<SiteStats | null>(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        loadStats();
    }, []);

    const loadStats = async () => {
        try {
            const data = await statsApi.getSiteStats();
            setStats(data);
        } catch (error) {
            console.error('Failed to load site stats:', error);
        } finally {
            setLoading(false);
        }
    };

    if (loading) {
        return <Card loading />;
    }

    return (
        <Card title="网站统计" size="small">
            <Space size="large">
                <Statistic
                    title="总浏览量"
                    value={stats?.total_views || 0}
                    prefix={<EyeOutlined />}
                />
            </Space>
        </Card>
    );
};

export default SiteStats;