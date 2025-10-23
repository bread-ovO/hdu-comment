import { useEffect, useState } from 'react';
import { Button, Card, Col, Empty, Input, List, Modal, Row, Select, Space, Spin, Typography, message } from 'antd';
import { Link } from 'react-router-dom';
import { deleteReview, fetchReviews } from '../api/client';
import SiteStats from '../components/SiteStats';
import ReviewStatsDisplay from '../components/ReviewStatsDisplay';
import type { PaginatedResponse, Review } from '../types';
import { useAuth } from '../hooks/useAuth';
import { statsApi } from '../api/stats';

const { Title, Paragraph, Text } = Typography;

const Home = () => {
  const [reviews, setReviews] = useState<Review[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string>('');
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(8);
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

  const handleDelete = (review: Review) => {
    Modal.confirm({
      title: `删除点评：${review.title}`,
      content: '删除后不可恢复，确认继续吗？',
      okText: '确认删除',
      okButtonProps: { danger: true },
      cancelText: '取消',
      onOk: async () => {
        try {
          await deleteReview(review.id);
          message.success('已删除该点评');
          await load();
        } catch (err) {
          console.error(err);
          message.error('删除失败，请稍后再试');
        }
      }
    });
  };

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <SiteStats />

      <Row justify="space-between" align="middle">
        <Col>
          <Title level={3}>最新点评</Title>
        </Col>
        <Col>
          <Space>
            <Input.Search
              allowClear
              placeholder="搜索菜品或地点"
              onSearch={(value) => {
                setPage(1);
                setQuery(value);
              }}
              onChange={(e) => {
                setPage(1);
                setQuery(e.target.value);
              }}
              value={query}
              style={{ width: 260 }}
            />
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
            />
          </Space>
        </Col>
      </Row>

      {loading ? (
        <Spin style={{ width: '100%' }} />
      ) : error ? (
        <Card><Text type="danger">{error}</Text></Card>
      ) : reviews.length === 0 ? (
        <Empty description="目前还没有符合条件的点评" />
      ) : (
        <List
          grid={{ gutter: 16, column: 2 }}
          dataSource={reviews}
          pagination={meta ? {
            current: meta.page,
            pageSize: meta.page_size,
            total: meta.total,
            onChange: (p, size) => {
              setPage(p);
              setPageSize(size);
            }
          } : false}
          renderItem={(review) => (
            <List.Item key={review.id}>
              <Card
                title={review.title}
                extra={<Text strong>{review.rating.toFixed(1)} 分</Text>}
                hoverable
              >
                <Paragraph ellipsis={{ rows: 3 }}>{review.description || '暂无详细点评'}</Paragraph>
                <Paragraph type="secondary">地址：{review.address}</Paragraph>
                {review.images && review.images.length > 0 && (
                  <img
                    src={review.images[0].url}
                    alt={review.title}
                    style={{ width: '100%', height: 180, objectFit: 'cover', borderRadius: 8 }}
                  />
                )}
                <div style={{ marginTop: 12 }}>
                  <ReviewStatsDisplay reviewId={review.id} />
                  <div style={{ marginTop: 8, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <Link
                      to={`/reviews/${review.id}`}
                      onClick={() => {
                        if (review.status === 'approved') {
                          statsApi.recordView(review.id).catch((err) => {
                            console.error('Failed to record review view:', err);
                          });
                        }
                      }}
                    >
                      查看详情
                    </Link>
                    {user?.role === 'admin' && (
                      <Button type="link" danger onClick={() => handleDelete(review)}>
                        删除
                      </Button>
                    )}
                  </div>
                </div>
              </Card>
            </List.Item>
          )}
        />
      )}
    </Space>
  );
};

export default Home;
