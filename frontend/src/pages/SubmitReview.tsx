import { useCallback, useState } from 'react';
import { Button, Card, Form, Input, InputNumber, message, Typography, Upload } from 'antd';
import type { RcFile, UploadFile } from 'antd/es/upload/interface';
import { submitReview, uploadReviewImage } from '../api/client';
import type { Review } from '../types';

const { TextArea } = Input;

const SubmitReview = () => {
  const [submitting, setSubmitting] = useState(false);
  const [currentReview, setCurrentReview] = useState<Review | null>(null);
  const [fileList, setFileList] = useState<UploadFile[]>([]);
  const [uploadingImages, setUploadingImages] = useState(false);

  const uploadPendingImages = useCallback(async (reviewId: string, targets?: UploadFile[]) => {
    const filesToUpload = (targets ?? fileList).filter((item) => item.originFileObj && (item.status === undefined || item.status === 'uploading'));
    if (filesToUpload.length === 0) return;

    setUploadingImages(true);
    try {
      for (const item of filesToUpload) {
        const raw = item.originFileObj as File | undefined;
        if (!raw) continue;

        setFileList((prev) => prev.map((file) => (file.uid === item.uid ? { ...file, status: 'uploading' } : file)));

        try {
          await uploadReviewImage(reviewId, raw);
          setFileList((prev) => prev.map((file) => (file.uid === item.uid ? { ...file, status: 'done' } : file)));
          message.success(`${item.name} 上传成功`);
        } catch (error) {
          console.error(error);
          setFileList((prev) => prev.map((file) => (file.uid === item.uid ? { ...file, status: 'error' } : file)));
          message.error(`${item.name} 上传失败，请稍后再试`);
        }
      }
    } finally {
      setUploadingImages(false);
    }
  }, [fileList]);

  const handleSubmit = async (values: { title: string; address: string; description: string; rating: number }) => {
    setSubmitting(true);
    try {
      const created = await submitReview(values);
      setCurrentReview(created);
      message.success('点评提交成功，等待管理员审核');
      await uploadPendingImages(created.id, fileList.filter((file) => file.status === undefined));
    } catch (err) {
      console.error(err);
      message.error('提交失败，请稍后再试');
    } finally {
      setSubmitting(false);
    }
  };

  const handleBeforeUpload = (file: RcFile) => {
    const item: UploadFile = {
      uid: file.uid,
      name: file.name,
      status: currentReview ? 'uploading' : undefined,
      originFileObj: file
    };

    setFileList((prev) => [...prev, item]);

    if (currentReview) {
      void uploadPendingImages(currentReview.id, [item]);
    } else {
      message.info('图片将在提交点评后自动上传');
    }

    return false;
  };

  const handleRemove = (file: UploadFile) => {
    setFileList((prev) => prev.filter((item) => item.uid !== file.uid));
    return true;
  };

  return (
    <Card className="subpage-card">
      <Typography.Title level={3}>提交新的食物点评</Typography.Title>
      <Form
        layout="vertical"
        onFinish={handleSubmit}
        initialValues={{ rating: 3 }}
        style={{ maxWidth: 640 }}
      >
        <Form.Item label="菜品/店铺名称" name="title" rules={[{ required: true, message: '请输入名称' }]}> 
          <Input placeholder="如：学一食堂蛋包饭" />
        </Form.Item>
        <Form.Item label="地点" name="address" rules={[{ required: true, message: '请输入地点' }]}> 
          <Input placeholder="楼层或附近标志" />
        </Form.Item>
        <Form.Item label="点评内容" name="description"> 
          <TextArea rows={4} placeholder="详细描述口味、环境等（选填）" />
        </Form.Item>
        <Form.Item
          label="评分（0-5）"
          name="rating"
          rules={[{ required: true, message: '请输入评分' }]}
        >
          <InputNumber min={0} max={5} step={0.5} style={{ width: 120 }} />
        </Form.Item>
        <Button type="primary" htmlType="submit" loading={submitting}>
          提交点评
        </Button>
      </Form>

      <Card style={{ marginTop: 24 }} size="small" title="上传图片（可选）" className="subpage-inner-card">
        <Upload
          accept="image/*"
          multiple
          beforeUpload={handleBeforeUpload}
          onRemove={handleRemove}
          fileList={fileList}
          listType="picture"
        >
          <Button disabled={uploadingImages}>选择图片</Button>
        </Upload>
        <Typography.Paragraph type="secondary" style={{ marginTop: 12 }}>
          {currentReview ? '继续选择即可立即上传图片。' : '您可以先选择图片，点评提交成功后会自动上传。'}
        </Typography.Paragraph>
      </Card>
    </Card>
  );
};

export default SubmitReview;
