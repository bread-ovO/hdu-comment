import { useEffect, useState } from 'react';
import {
    Input,
    Empty,
    Spin,
    Typography,
    Space,
    Select,
    Pagination,
    Row,
    Col,
    Button,
    message
} from 'antd';
import { SearchOutlined } from '@ant-design/icons';
import { Link } from 'react-router-dom';
import { deleteReview, fetchReviews } from '../api/client';
import ReviewCard from '../components/ReviewCard';
import type { PaginatedResponse, Review } from '../types';
import { useAuth } from '../hooks/useAuth';

const { Title, Text } = Typography;
const { Search } = Input;

const NewHome = () => {
    const [reviews, setReviews] = useState<Review[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string>('');
    const [page, setPage] = useState(1);
    const [pageSize, setPageSize] = useState(12);
    const [query, setQuery] = useState('');
    const [sort, setSort] = useState<'created_at' | 'rating'>('created_at');
    const [meta, setMeta] = useState<PaginatedResponse<Review>['pagination'] | null>(null);
    const { user } = useAuth();

    const load = async () => {
        setLoading(true);
        setError('');
        try {
            const data = await fetchReviews({
                page,
                page_size: pageSize,
                query: query || undefined,
                sort,
                order: sort === 'rating' ? 'desc' : 'desc'
            });
            setReviews(data.data);
            setMeta(data.pagination);
        } catch (err) {
            console.error(err);
            setError('加载点评失败，请稍后再试');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        load();
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [page, pageSize, sort, query]);

    const handleDelete = async (review: Review) => {
        if (user?.role !== 'admin') {
            message.info('删除功能需要管理员权限');
            return;
        }

        try {
            await deleteReview(review.id);
            message.success('点评已删除');
            await load();
        } catch (err) {
            console.error(err);
            message.error('删除失败，请稍后再试');
        }
    };

    return (
        <div className="home-container">
            <div className="home-header">
                <Title level={2} className="home-title">
                    杭电美食点评
                </Title>
                <Text type="secondary" className="home-subtitle">
                    发现校园里的美味佳肴
                </Text>
            </div>

            <div className="home-search-section">
                <Space direction="vertical" size="large" style={{ width: '100%' }}>
                    <Search
                        placeholder="搜索菜品、地点或关键词..."
                        allowClear
                        enterButton={<SearchOutlined />}
                        size="large"
                        onSearch={(value) => {
                            setPage(1);
                            setQuery(value);
                        }}
                        onChange={(e) => {
                            setPage(1);
                            setQuery(e.target.value);
                        }}
                        value={query}
                        className="home-search-input"
                    />

                    <div className="home-filters">
                        <Select
                            value={sort}
                            onChange={(value) => {
                                setSort(value);
                                setPage(1);
                            }}
                            options={[
                                { value: 'created_at', label: '最新发布' },
                                { value: 'rating', label: '评分最高' }
                            ]}
                            className="home-sort-select"
                        />
                    </div>
                </Space>
            </div>

            <div className="home-content">
                {loading ? (
                    <div className="loading-container">
                        <Spin size="large" />
                    </div>
                ) : error ? (
                    <div className="error-container">
                        <Empty
                            description={error}
                            image={Empty.PRESENTED_IMAGE_SIMPLE}
                        />
                    </div>
                ) : reviews.length === 0 ? (
                    <div className="empty-container">
                        <Empty
                            description={
                                <Space direction="vertical" size="middle">
                                    <Text>没有找到符合条件的点评</Text>
                                    <Link to="/submit">
                                        <Button type="primary">成为第一个点评的人</Button>
                                    </Link>
                                </Space>
                            }
                        />
                    </div>
                ) : (
                    <>
                        <Row gutter={[24, 24]} className="reviews-grid">
                            {reviews.map((review) => (
                                <Col
                                    key={review.id}
                                    xs={24}
                                    sm={12}
                                    md={8}
                                    lg={6}
                                >
                                    <ReviewCard
                                        review={review}
                                        onDelete={handleDelete}
                                        showStatus={user?.role === 'admin'}
                                    />
                                </Col>
                            ))}
                        </Row>

                        {meta && (
                            <div className="pagination-container">
                                <Pagination
                                    current={meta.page}
                                    pageSize={meta.page_size}
                                    total={meta.total}
                                    showSizeChanger
                                    showQuickJumper
                                    showTotal={(total, range) =>
                                        `第 ${range[0]}-${range[1]} 条/共 ${total} 条`
                                    }
                                    onChange={(p, size) => {
                                        setPage(p);
                                        setPageSize(size);
                                    }}
                                    className="home-pagination"
                                />
                            </div>
                        )}
                    </>
                )}
            </div>
        </div>
    );
};

export default NewHome;
