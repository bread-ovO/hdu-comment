import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { message } from 'antd';
import Register from './Register';

const mockRegister = vi.fn();
const mockSendRegistrationCode = vi.fn();
const mockNavigate = vi.fn();

vi.mock('../hooks/useAuth', () => ({
  useAuth: () => ({
    register: mockRegister
  })
}));

vi.mock('../api/auth', () => ({
  authApi: {
    sendRegistrationCode: (...args: unknown[]) => mockSendRegistrationCode(...args)
  }
}));

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom');
  return {
    ...actual,
    useNavigate: () => mockNavigate
  };
});

describe('Register page interactions', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(message, 'success').mockImplementation(() => undefined as any);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('sends registration code after validating email input', async () => {
    mockSendRegistrationCode.mockResolvedValueOnce({ message: 'ok' });

    render(
      <MemoryRouter>
        <Register />
      </MemoryRouter>
    );

    await userEvent.type(screen.getByPlaceholderText('name@example.com'), 'test@example.com');
    await userEvent.click(screen.getByText('获取验证码'));

    await waitFor(() => {
      expect(mockSendRegistrationCode).toHaveBeenCalledWith('test@example.com');
    });
    expect(message.success).toHaveBeenCalledWith('验证码已发送，请查收邮箱');
  });

  it('submits registration form and navigates home', async () => {
    mockRegister.mockResolvedValueOnce(undefined);

    render(
      <MemoryRouter>
        <Register />
      </MemoryRouter>
    );

    await userEvent.type(screen.getByPlaceholderText('展示名称'), '小明');
    await userEvent.type(screen.getByPlaceholderText('name@example.com'), 'test@example.com');
    await userEvent.type(screen.getByPlaceholderText('请输入邮箱验证码'), '123456');
    await userEvent.type(screen.getByPlaceholderText('请输入密码'), 'Passw0rd!');

    await userEvent.click(screen.getByRole('button', { name: /注\s*册/ }));

    await waitFor(() => {
      expect(mockRegister).toHaveBeenCalledWith('test@example.com', 'Passw0rd!', '小明', '123456');
    });
    expect(mockNavigate).toHaveBeenCalledWith('/');
  });
});
