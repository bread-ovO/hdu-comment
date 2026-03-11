import { useCallback, useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { Button, Card, Col, Grid, Row, Space, Statistic, Table, Typography, message } from 'antd';
import { EyeOutlined, LikeOutlined, DislikeOutlined, ReloadOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import { fetchReviews } from '../api/client';
import { statsApi } from '../api/stats';
import type { SiteStats } from '../types';

interface ReviewTrafficRow {
  id: string;
  title: string;
  address: string;
  views: number;
  likes: number;
  dislikes: number;
  updated_at: string;
}

const AdminStats = () => {
  const screens = Grid.useBreakpoint();
  const [loading, setLoading] = useState(true);
  const [siteStats, setSiteStats] = useState<SiteStats | null>(null);
  const [rows, setRows] = useState<ReviewTrafficRow[]>([]);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const [site, reviews] = await Promise.all([
        statsApi.getSiteStats(),
        fetchReviews({ page: 1, page_size: 20, sort: 'created_at', order: 'desc' })
      ]);

      const reviewStats = await Promise.all(
        reviews.data.map(async (review) => {
          try {
            const stats = await statsApi.getReviewStats(review.id);
            return {
              id: review.id,
              title: review.title,
              address: review.address,
              views: stats.views,
              likes: stats.likes,
              dislikes: stats.dislikes,
              updated_at: stats.updated_at
            } satisfies ReviewTrafficRow;
          } catch {
            return {
              id: review.id,
              title: review.title,
              address: review.address,
              views: 0,
              likes: 0,
              dislikes: 0,
              updated_at: review.updated_at
            } satisfies ReviewTrafficRow;
          }
        })
      );

      setSiteStats(site);
      setRows(reviewStats.sort((a, b) => b.views - a.views));
    } catch (err) {
      console.error(err);
      message.error('加载流量统计失败，请稍后再试');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  return (
    <div className="subpage-page subpage-page-wide">
      <Space direction="vertical" size="large" style={{ width: '100%' }}>
        <Card
          title="流量总览"
          className="subpage-card"
          extra={
            <Button icon={<ReloadOutlined />} onClick={load} loading={loading}>
              刷新
            </Button>
          }
        >
          <Row gutter={[16, 16]}>
            <Col xs={24} sm={12} lg={8}>
              <Statistic
                title="网站总浏览量"
                value={siteStats?.total_views ?? 0}
                prefix={<EyeOutlined />}
              />
            </Col>
            <Col xs={24} sm={12} lg={8}>
              <Statistic
                title="统计最后更新时间"
                value={siteStats?.updated_at ? dayjs(siteStats.updated_at).format('YYYY-MM-DD HH:mm:ss') : '-'}
              />
            </Col>
            <Col xs={24} sm={24} lg={8}>
              <Typography.Text type="secondary">
                浏览量在用户进入点评详情时累加（查看详情触发）。
              </Typography.Text>
            </Col>
          </Row>
        </Card>

        <Card title="点评热度（近 20 条已发布点评）" className="subpage-card subpage-table-card">
          <Table<ReviewTrafficRow>
            rowKey="id"
            loading={loading}
            dataSource={rows}
            size={screens.xs ? 'small' : 'middle'}
            scroll={{ x: 980 }}
            pagination={false}
            columns={[
              {
                title: '点评',
                dataIndex: 'title',
                key: 'title',
                ellipsis: true,
                render: (_value, record) => <Link to={`/reviews/${record.id}`}>{record.title}</Link>
              },
              {
                title: '地点',
                dataIndex: 'address',
                key: 'address',
                ellipsis: true
              },
              {
                title: '浏览',
                dataIndex: 'views',
                key: 'views',
                width: 120,
                render: (value: number) => (
                  <Space size={4}>
                    <EyeOutlined />
                    <span>{value}</span>
                  </Space>
                )
              },
              {
                title: '点赞',
                dataIndex: 'likes',
                key: 'likes',
                width: 120,
                render: (value: number) => (
                  <Space size={4}>
                    <LikeOutlined />
                    <span>{value}</span>
                  </Space>
                )
              },
              {
                title: '点踩',
                dataIndex: 'dislikes',
                key: 'dislikes',
                width: 120,
                render: (value: number) => (
                  <Space size={4}>
                    <DislikeOutlined />
                    <span>{value}</span>
                  </Space>
                )
              },
              {
                title: '更新时间',
                dataIndex: 'updated_at',
                key: 'updated_at',
                width: 180,
                render: (value: string) => dayjs(value).format('MM-DD HH:mm')
              }
            ]}
          />
        </Card>
      </Space>
    </div>
  );
};

export default AdminStats;
