import { useEffect, useState } from 'react';
import { Space, Typography } from 'antd';
import { EyeOutlined, LikeOutlined, DislikeOutlined } from '@ant-design/icons';
import { statsApi } from '../api/stats';
import type { ReviewStats } from '../types';

const { Text } = Typography;

interface ReviewStatsDisplayProps {
    reviewId: string;
}

const ReviewStatsDisplay = ({ reviewId }: ReviewStatsDisplayProps) => {
    const [stats, setStats] = useState<ReviewStats | null>(null);

    useEffect(() => {
        loadStats();
    }, [reviewId]);

    const loadStats = async () => {
        try {
            const data = await statsApi.getReviewStats(reviewId);
            setStats(data);
        } catch (error) {
            console.error('Failed to load review stats:', error);
        }
    };

    if (!stats) {
        return (
            <Space size="middle">
                <Text type="secondary">
                    <EyeOutlined /> 浏览 0
                </Text>
                <Text type="secondary">
                    <LikeOutlined /> 赞 0
                </Text>
                <Text type="secondary">
                    <DislikeOutlined /> 踩 0
                </Text>
            </Space>
        );
    }

    return (
        <Space size="middle">
            <Text type="secondary">
                <EyeOutlined /> 浏览 {stats.views}
            </Text>
            <Text type="secondary">
                <LikeOutlined /> 赞 {stats.likes}
            </Text>
            <Text type="secondary">
                <DislikeOutlined /> 踩 {stats.dislikes}
            </Text>
        </Space>
    );
};

export default ReviewStatsDisplay;