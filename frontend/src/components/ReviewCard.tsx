import { Card, Tag, Typography, Space, Rate, Button, Popconfirm } from 'antd';
import { Link } from 'react-router-dom';
import {
    EnvironmentOutlined,
    CalendarOutlined,
    UserOutlined,
    EyeOutlined
} from '@ant-design/icons';
import ReviewStatsDisplay from './ReviewStatsDisplay';
import { statsApi } from '../api/stats';
import type { Review } from '../types';

const { Title, Paragraph, Text } = Typography;

interface ReviewCardProps {
    review: Review;
    onDelete?: (review: Review) => void;
    showStatus?: boolean;
    canDelete?: boolean;
}

const ReviewCard = ({ review, onDelete, showStatus = false, canDelete = false }: ReviewCardProps) => {

    const getStatusColor = (status: string) => {
        switch (status) {
            case 'approved': return 'success';
            case 'pending': return 'warning';
            case 'rejected': return 'error';
            default: return 'default';
        }
    };

    const getStatusText = (status: string) => {
        switch (status) {
            case 'approved': return '已发布';
            case 'pending': return '待审核';
            case 'rejected': return '已拒绝';
            default: return status;
        }
    };

    const handleViewDetail = () => {
        if (review.status === 'approved') {
            statsApi.recordView(review.id).catch((error) => {
                console.error('Failed to record review view:', error);
            });
        }
    };

    const actions = [
        <Link to={`/reviews/${review.id}`} key="view">
            <Button type="text" icon={<EyeOutlined />} onClick={handleViewDetail}>
                查看详情
            </Button>
        </Link>
    ];

    if (canDelete && onDelete) {
        actions.push(
            <Popconfirm
                key="delete"
                title="确认删除该点评吗？"
                okText="删除"
                cancelText="取消"
                okButtonProps={{ danger: true }}
                onConfirm={(e) => {
                    e?.preventDefault();
                    onDelete(review);
                }}
            >
                <Button type="text" danger>
                    删除
                </Button>
            </Popconfirm>
        );
    }

    return (
        <Card
            hoverable
            className="review-card"
            cover={
                review.images && review.images.length > 0 && (
                    <div className="review-card-image-container">
                        <img
                            alt={review.title}
                            src={review.images[0].url}
                            className="review-card-image"
                        />
                        {showStatus && (
                            <Tag
                                color={getStatusColor(review.status)}
                                className="review-status-tag"
                            >
                                {getStatusText(review.status)}
                            </Tag>
                        )}
                    </div>
                )
            }
            actions={actions}
        >
            <div className="review-card-content">
                <Title level={4} className="review-title" ellipsis={{ rows: 1 }}>
                    {review.title}
                </Title>

                <Space className="review-meta" size="small">
                    <Rate
                        disabled
                        defaultValue={review.rating}
                        className="review-rating"
                    />
                    <Text type="secondary" className="rating-text">
                        {review.rating.toFixed(1)}
                    </Text>
                </Space>

                <Paragraph
                    className="review-description"
                    ellipsis={{ rows: 2 }}
                >
                    {review.description || '暂无详细点评'}
                </Paragraph>

                <Space direction="vertical" size="small" className="review-info">
                    <Space className="info-item">
                        <EnvironmentOutlined className="info-icon" />
                        <Text type="secondary" className="info-text">
                            {review.address}
                        </Text>
                    </Space>

                    <Space className="info-item">
                        <UserOutlined className="info-icon" />
                        <Text type="secondary" className="info-text">
                            {review.author?.display_name || '匿名用户'}
                        </Text>
                    </Space>

                    <Space className="info-item">
                        <CalendarOutlined className="info-icon" />
                        <Text type="secondary" className="info-text">
                            {new Date(review.created_at).toLocaleDateString('zh-CN')}
                        </Text>
                    </Space>
                </Space>

                <div style={{ marginTop: 16 }}>
                    <ReviewStatsDisplay reviewId={review.id} />
                </div>
            </div>
        </Card>
    );
};

export default ReviewCard;
