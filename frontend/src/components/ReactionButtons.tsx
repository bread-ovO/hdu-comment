import { useState, useEffect } from 'react';
import { Button, Space, message } from 'antd';
import { LikeOutlined, DislikeOutlined } from '@ant-design/icons';
import { useAuth } from '../hooks/useAuth';
import { statsApi } from '../api/stats';
import type { Review } from '../types';

interface ReactionButtonsProps {
    review: Review;
    onReactionUpdate?: (review: Review) => void;
}

const ReactionButtons = ({ review, onReactionUpdate }: ReactionButtonsProps) => {
    const { user } = useAuth();
    const [loading, setLoading] = useState(false);
    const [userReaction, setUserReaction] = useState<'like' | 'dislike' | null>(null);
    const [stats, setStats] = useState(review.stats || {
        id: '',
        review_id: review.id,
        views: 0,
        likes: 0,
        dislikes: 0,
        created_at: '',
        updated_at: ''
    });

    useEffect(() => {
        if (user && review.id) {
            loadUserReaction();
        }
        loadStats();
    }, [review.id, user]);

    useEffect(() => {
        if (review.stats) {
            setStats(review.stats);
        }
    }, [review.stats?.views, review.stats?.likes, review.stats?.dislikes, review.stats?.updated_at]);

    const loadUserReaction = async () => {
        try {
            const response = await statsApi.getUserReaction(review.id);
            setUserReaction(response.reaction);
        } catch (error) {
            console.error('Failed to load user reaction:', error);
        }
    };

    const loadStats = async () => {
        try {
            const stats = await statsApi.getReviewStats(review.id);
            setStats(stats);
        } catch (error) {
            console.error('Failed to load stats:', error);
        }
    };

    const handleReaction = async (type: 'like' | 'dislike') => {
        if (!user) {
            message.warning('请先登录后再进行操作');
            return;
        }

        setLoading(true);
        try {
            const response = await statsApi.toggleReaction(review.id, type);
            console.log('Reaction response:', response);

            // 重新加载用户反应和统计数据
            await Promise.all([loadUserReaction(), loadStats()]);

            message.success('操作成功');
        } catch (error: any) {
            console.error('Failed to toggle reaction:', error);
            if (error.response) {
                console.error('Error response:', error.response.data);
            }
            message.error(`操作失败: ${error.response?.data?.error || '请重试'}`);
        } finally {
            setLoading(false);
        }
    };

    return (
        <Space>
            <Button
                type={userReaction === 'like' ? 'primary' : 'default'}
                icon={<LikeOutlined />}
                onClick={() => handleReaction('like')}
                loading={loading}
                disabled={!user}
            >
                赞 ({stats.likes})
            </Button>
            <Button
                type={userReaction === 'dislike' ? 'primary' : 'default'}
                icon={<DislikeOutlined />}
                onClick={() => handleReaction('dislike')}
                loading={loading}
                disabled={!user}
            >
                踩 ({stats.dislikes})
            </Button>
            <span style={{ color: '#666', marginLeft: 8 }}>
                浏览量: {stats.views}
            </span>
        </Space>
    );
};

export default ReactionButtons;
