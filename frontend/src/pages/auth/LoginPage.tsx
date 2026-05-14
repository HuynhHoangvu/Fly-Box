import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { GoogleLogin, CredentialResponse } from '@react-oauth/google';
import { motion, AnimatePresence } from 'framer-motion';
import { AlertCircle, Loader2 } from 'lucide-react';
import { useAuthStore } from '../../store/useAuthStore';
import { loginSchema, registerSchema, LoginInput, RegisterInput } from '../../validations/auth';
import { FormField } from '../../components/auth/FormField';
import './LoginPage.css';

export const LoginPage: React.FC = () => {
  const [isRegister, setIsRegister] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);
  
  const navigate = useNavigate();
  const { 
    isLoading, 
    isAuthenticated, 
    loginWithCredentials, 
    loginWithGoogle, 
    register 
  } = useAuthStore();

  // 1. Redirect if already authenticated
  useEffect(() => {
    if (isAuthenticated) {
      navigate('/inbox', { replace: true });
    }
  }, [isAuthenticated, navigate]);

  // 2. Form Setup
  const loginForm = useForm<LoginInput>({
    resolver: zodResolver(loginSchema),
  });

  const registerForm = useForm<RegisterInput>({
    resolver: zodResolver(registerSchema),
  });

  // Reset error when switching modes
  useEffect(() => {
    setFormError(null);
    loginForm.reset();
    registerForm.reset();
  }, [isRegister, loginForm, registerForm]);

  // 3. Handlers
  const onLoginSubmit = async (data: LoginInput) => {
    setFormError(null);
    try {
      await loginWithCredentials(data.email, data.password);
    } catch (err: any) {
      const msg = err?.response?.data?.error || 'Đăng nhập thất bại. Vui lòng thử lại.';
      setFormError(msg);
      loginForm.setValue('password', ''); // Clear password on fail
    }
  };

  const onRegisterSubmit = async (data: RegisterInput) => {
    setFormError(null);
    try {
      await register(data.name, data.email, data.password);
    } catch (err: any) {
      const msg = err?.response?.data?.error || 'Đăng ký thất bại. Email có thể đã tồn tại.';
      setFormError(msg);
    }
  };

  const handleGoogleSuccess = async (credentialResponse: CredentialResponse) => {
    if (!credentialResponse.credential) return;
    setFormError(null);
    try {
      await loginWithGoogle(credentialResponse.credential);
    } catch (err) {
      setFormError('Xác thực Google thất bại.');
    }
  };

  return (
    <div className="login-page">
      <motion.div 
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="login-card"
      >
        <div className="auth-header">
          <h1>Fly Box</h1>
          <p>{isRegister ? 'Khởi tạo hành trình mới' : 'Chào mừng bạn quay lại'}</p>
        </div>

        {/* Google Auth hidden for now */}
        {/* 
        <div className="google-auth-section">
          <GoogleLogin
            onSuccess={handleGoogleSuccess}
            onError={() => setFormError('Không thể kết nối với Google.')}
            useOneTap
            shape="rectangular"
            width="100%"
          />
        </div>

        <div className="divider">
          <span>hoặc sử dụng Email</span>
        </div>
        */}

        <AnimatePresence mode="wait">
          {formError && (
            <motion.div 
              initial={{ opacity: 0, height: 0 }}
              animate={{ opacity: 1, height: 'auto' }}
              exit={{ opacity: 0, height: 0 }}
              className="form-alert error"
            >
              <AlertCircle size={16} />
              <span>{formError}</span>
            </motion.div>
          )}
        </AnimatePresence>

        <form 
          className="auth-form" 
          onSubmit={isRegister ? registerForm.handleSubmit(onRegisterSubmit) : loginForm.handleSubmit(onLoginSubmit)}
        >
          {isRegister && (
            <FormField
              label="Họ tên"
              placeholder="Nguyễn Văn A"
              disabled={isLoading}
              registration={registerForm.register('name')}
              error={registerForm.formState.errors.name?.message}
            />
          )}

          <FormField
            label="Email"
            type="email"
            placeholder="name@company.com"
            autoComplete="email"
            disabled={isLoading}
            registration={isRegister ? registerForm.register('email') : loginForm.register('email')}
            error={isRegister ? registerForm.formState.errors.email?.message : loginForm.formState.errors.email?.message}
          />

          <FormField
            label="Mật khẩu"
            type="password"
            placeholder="••••••••"
            autoComplete={isRegister ? 'new-password' : 'current-password'}
            disabled={isLoading}
            registration={isRegister ? registerForm.register('password') : loginForm.register('password')}
            error={isRegister ? registerForm.formState.errors.password?.message : loginForm.formState.errors.password?.message}
          />

          {isRegister && (
            <FormField
              label="Xác nhận mật khẩu"
              type="password"
              placeholder="••••••••"
              autoComplete="new-password"
              disabled={isLoading}
              registration={registerForm.register('confirmPassword')}
              error={registerForm.formState.errors.confirmPassword?.message}
            />
          )}

          {!isRegister && (
            <div className="form-options">
              <label className="remember-me">
                <input type="checkbox" />
                <span>Ghi nhớ đăng nhập</span>
              </label>
              <button type="button" className="btn-link">Quên mật khẩu?</button>
            </div>
          )}

          <button 
            type="submit" 
            className="btn-primary" 
            disabled={isLoading}
          >
            {isLoading ? (
              <><Loader2 className="spinner" size={18} /> Đang xử lý...</>
            ) : (
              isRegister ? 'Đăng ký ngay' : 'Đăng nhập'
            )}
          </button>
        </form>

        <div className="auth-footer">
          <p>
            {isRegister ? 'Bạn đã có tài khoản?' : 'Chưa có tài khoản?'}
            <button 
              type="button" 
              className="btn-link-highlight" 
              onClick={() => setIsRegister(!isRegister)}
            >
              {isRegister ? 'Đăng nhập' : 'Tham gia ngay'}
            </button>
          </p>
        </div>
      </motion.div>
    </div>
  );
};
