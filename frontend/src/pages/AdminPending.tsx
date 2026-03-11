import { useCallback, useEffect, useState } from 'react';
import { Button, Card, Descriptions, Form, Image, Input, Modal, Space, Table, Tag, Typography, message } from 'antd';
import type { ColumnsType, TablePaginationConfig } from 'antd/es/table';
import { approveReview, deleteReview, fetchPendingReviews, fetchReviewDetail, rejectReview } from '../api/client';
import type { Review } from '../types';

const AdminPending = () => {
  const [reviews, setReviews] = useState<Review[]>([]);
  const [loading, setLoading] = useState(true);
  const [pagination, setPagination] = useState<TablePaginationConfig>({ current: 1, pageSize: 10, total: 0 });
  const [query, setQuery] = useState('');
  const [rejectModalOpen, setRejectModalOpen] = useState(false);
  const [selectedReview, setSelectedReview] = useState<Review | null>(null);
  const [detailOpen, setDetailOpen] = useState(false);
  const [detailLoading, setDetailLoading] = useState(false);
  const [detailReview, setDetailReview] = useState<Review | null>(null);
  const [form] = Form.useForm();

  const load = useCallback(async (page = 1, pageSize = 10, keyword = query) => {
    setLoading(true);
    try {
      const data = await fetchPendingReviews({
        page,
        page_size: pageSize,
        query: keyword || undefined,
        sort: 'created_at',
        order: 'asc'
      });
      setReviews(data.data);
      setPagination({ current: data.pagination.page, pageSize: data.pagination.page_size, total: data.pagination.total });
    } catch (err) {
      console.error(err);
      message.error('加载待审核点评失败');
    } finally {
      setLoading(false);
    }
  }, [query]);

  useEffect(() => {
    load();
  }, [load]);

  const handleApprove = async (review: Review) => {
    try {
      await approveReview(review.id);
      message.success('已通过审核');
      load(pagination.current, pagination.pageSize);
    } catch (err) {
      console.error(err);
      message.error('通过失败，请稍后再试');
    }
  };

  const handleViewDetail = async (review: Review) => {
    setDetailOpen(true);
    setDetailLoading(true);
    try {
      const detail = await fetchReviewDetail(review.id);
      setDetailReview(detail);
    } catch (err) {
      console.error(err);
      message.error('获取点评详情失败');
      setDetailOpen(false);
    } finally {
      setDetailLoading(false);
    }
  };

  const openRejectModal = (review: Review) => {
    setSelectedReview(review);
    form.resetFields();
    setRejectModalOpen(true);
  };

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
          load(pagination.current, pagination.pageSize);
        } catch (err) {
          console.error(err);
          message.error('删除失败，请稍后再试');
        }
      }
    });
  };

  const handleReject = async () => {
    try {
      const values = await form.validateFields();
      if (!selectedReview) return;
      await rejectReview(selectedReview.id, values.reason);
      message.success('已驳回该点评');
      setRejectModalOpen(false);
      load(pagination.current, pagination.pageSize);
    } catch (err) {
      const validationError = err as { errorFields?: unknown } | undefined;
      if (validationError?.errorFields) return;
      console.error(err);
      message.error('驳回失败，请稍后再试');
    }
  };

  const columns: ColumnsType<Review> = [
    {
      title: '菜品/店铺',
      dataIndex: 'title',
      key: 'title'
    },
    {
      title: '地址',
      dataIndex: 'address',
      key: 'address'
    },
    {
      title: '评分',
      dataIndex: 'rating',
      key: 'rating',
      render: (rating: number) => rating.toFixed(1)
    },
    {
      title: '提交时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (value: string) => new Date(value).toLocaleString()
    },
    {
      title: '操作',
      key: 'actions',
      render: (_, record) => (
        <Space>
          <Button onClick={() => handleViewDetail(record)}>查看详情</Button>
          <Button type="primary" onClick={() => handleApprove(record)}>
            通过
          </Button>
          <Button danger onClick={() => openRejectModal(record)}>
            驳回
          </Button>
          <Button danger ghost onClick={() => handleDelete(record)}>
            删除
          </Button>
        </Space>
      )
    }
  ];

  return (
    <div className="subpage-page subpage-page-wide">
      <Card
        title={<Typography.Title level={4}>待审核点评</Typography.Title>}
        extra={
          <Space>
            <Input.Search
              placeholder="搜索"
              allowClear
              onSearch={(value) => {
                setQuery(value);
                load(1, pagination.pageSize ?? 10, value);
              }}
              style={{ width: 240 }}
            />
            <Tag color="orange">共 {pagination.total ?? 0} 条待处理</Tag>
          </Space>
        }
        className="subpage-card subpage-table-card"
      >
      <Table
        rowKey="id"
        columns={columns}
        dataSource={reviews}
        loading={loading}
        pagination={{
          ...pagination,
          showSizeChanger: true,
          onChange: (page, pageSize) => load(page, pageSize)
        }}
      />

      <Modal
        open={rejectModalOpen}
        title={`驳回点评：${selectedReview?.title ?? ''}`}
        onCancel={() => setRejectModalOpen(false)}
        onOk={handleReject}
        okText="确认驳回"
      >
        <Form form={form} layout="vertical">
          <Form.Item
            label="驳回原因"
            name="reason"
            rules={[{ required: true, message: '请输入驳回原因' }]}
          >
            <Input.TextArea rows={4} placeholder="请填写驳回原因" />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        open={detailOpen}
        title={detailReview?.title ?? '点评详情'}
        onCancel={() => {
          setDetailOpen(false);
          setDetailReview(null);
        }}
        footer={null}
        width={720}
      >
        {detailLoading ? (
          <Typography.Paragraph>加载中...</Typography.Paragraph>
        ) : !detailReview ? (
          <Typography.Text type="danger">未找到点评详情</Typography.Text>
        ) : (
          <Space direction="vertical" size="middle" style={{ width: '100%' }}>
            <Descriptions bordered column={1}>
              <Descriptions.Item label="地址">{detailReview.address}</Descriptions.Item>
              <Descriptions.Item label="评分">{detailReview.rating.toFixed(1)} 分</Descriptions.Item>
              <Descriptions.Item label="提交时间">
                {new Date(detailReview.created_at).toLocaleString()}
              </Descriptions.Item>
              <Descriptions.Item label="点评内容">
                {detailReview.description || '暂无详细描述'}
              </Descriptions.Item>
              {detailReview.rejection_reason && (
                <Descriptions.Item label="驳回原因">
                  <Typography.Text type="danger">{detailReview.rejection_reason}</Typography.Text>
                </Descriptions.Item>
              )}
            </Descriptions>
            {detailReview.images && detailReview.images.length > 0 && (
              <Image.PreviewGroup>
                <Space wrap>
                  {detailReview.images.map((image) => (
                    <Image
                      key={image.id}
                      src={image.url}
                      width={160}
                      height={120}
                      style={{ objectFit: 'cover' }}
                    />
                  ))}
                </Space>
              </Image.PreviewGroup>
            )}
          </Space>
        )}
      </Modal>
      </Card>
    </div>
  );
};

export default AdminPending;
