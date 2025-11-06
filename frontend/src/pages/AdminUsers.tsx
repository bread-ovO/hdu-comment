import { useCallback, useEffect, useState } from 'react';
import { Button, Card, message, Popconfirm, Space, Table, Tag } from 'antd';
import type { TablePaginationConfig } from 'antd/es/table';
import dayjs from 'dayjs';
import { deleteAdminUser, fetchAdminUsers } from '../api/client';
import type { User } from '../types';

interface AdminUserRecord extends User {
  email_verified: boolean;
  email_verified_at?: string | null;
}

interface AdminUserTableRecord extends AdminUserRecord {
  created_at?: string;
}

const AdminUsers = () => {
  const [loading, setLoading] = useState(false);
  const [users, setUsers] = useState<AdminUserTableRecord[]>([]);
  const [pagination, setPagination] = useState<TablePaginationConfig>({
    current: 1,
    pageSize: 20,
    total: 0,
    showSizeChanger: true,
    pageSizeOptions: ['10', '20', '50', '100']
  });

  const loadUsers = useCallback(
    async (page: number = pagination.current || 1, pageSize: number = pagination.pageSize || 20) => {
      setLoading(true);
      try {
        const response = await fetchAdminUsers({ page, page_size: pageSize });
        setUsers(
          response.data.map((user) => ({
            ...user,
            email_verified_at: user.email_verified_at,
            created_at: user.created_at
          }))
        );
        setPagination((prev) => ({
          ...prev,
          current: response.pagination.page,
          pageSize: response.pagination.page_size,
          total: response.pagination.total
        }));
      } catch (err: any) {
        console.error(err);
        message.error(err?.response?.data?.error || '加载用户列表失败');
      } finally {
        setLoading(false);
      }
    },
    [pagination.current, pagination.pageSize]
  );

  useEffect(() => {
    loadUsers();
  }, [loadUsers]);

  const handleTableChange = (paginationConfig: TablePaginationConfig) => {
    loadUsers(paginationConfig.current, paginationConfig.pageSize);
  };

  const handleDelete = async (id: string) => {
    try {
      await deleteAdminUser(id);
      message.success('用户已删除');
      loadUsers();
    } catch (err: any) {
      console.error(err);
      message.error(err?.response?.data?.error || '删除用户失败');
    }
  };

  const columns = [
    {
      title: '邮箱',
      dataIndex: 'email',
      key: 'email'
    },
    {
      title: '昵称',
      dataIndex: 'display_name',
      key: 'display_name'
    },
    {
      title: '角色',
      dataIndex: 'role',
      key: 'role',
      render: (role: string) => (
        <Tag color={role === 'admin' ? 'geekblue' : 'default'}>{role === 'admin' ? '管理员' : '用户'}</Tag>
      )
    },
    {
      title: '邮箱验证',
      dataIndex: 'email_verified',
      key: 'email_verified',
      render: (verified: boolean) => (
        <Tag color={verified ? 'success' : 'error'}>{verified ? '已验证' : '未验证'}</Tag>
      )
    },
    {
      title: '注册时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (value?: string) => (value ? dayjs(value).format('YYYY-MM-DD HH:mm') : '-')
    },
    {
      title: '操作',
      key: 'actions',
      render: (_: unknown, record: AdminUserTableRecord) => (
        <Space>
          <Popconfirm
            title="确认删除该用户？"
            okText="删除"
            cancelText="取消"
            okButtonProps={{ danger: true }}
            onConfirm={() => handleDelete(record.id)}
          >
            <Button danger size="small">
              删除
            </Button>
          </Popconfirm>
        </Space>
      )
    }
  ];

  return (
    <div style={{ maxWidth: 960, margin: '32px auto', padding: '0 16px' }}>
      <Card title="用户管理">
        <Table
          rowKey="id"
          columns={columns}
          dataSource={users}
          loading={loading}
          pagination={pagination}
          onChange={handleTableChange}
        />
      </Card>
    </div>
  );
};

export default AdminUsers;
