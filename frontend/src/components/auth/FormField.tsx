import React, { useState } from 'react';
import { UseFormRegisterReturn } from 'react-hook-form';
import { Eye, EyeOff } from 'lucide-react';

interface FormFieldProps {
  label: string;
  type?: string;
  placeholder?: string;
  error?: string;
  registration: UseFormRegisterReturn;
  autoComplete?: string;
  disabled?: boolean;
}

export const FormField: React.FC<FormFieldProps> = ({
  label,
  type = 'text',
  placeholder,
  error,
  registration,
  autoComplete,
  disabled
}) => {
  const [showPassword, setShowPassword] = useState(false);
  const isPassword = type === 'password';
  const inputType = isPassword ? (showPassword ? 'text' : 'password') : type;

  return (
    <div className="form-group">
      <label className="form-label">{label}</label>
      <div className="input-wrapper" style={{ position: 'relative' }}>
        <input
          type={inputType}
          className={`form-input ${error ? 'error' : ''}`}
          placeholder={placeholder}
          autoComplete={autoComplete}
          disabled={disabled}
          {...registration}
          aria-invalid={!!error}
        />
        {isPassword && (
          <button
            type="button"
            className="password-toggle"
            onClick={() => setShowPassword(!showPassword)}
            style={{
              position: 'absolute',
              right: '12px',
              top: '50%',
              transform: 'translateY(-50%)',
              background: 'none',
              border: 'none',
              color: '#6b7280',
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center'
            }}
            tabIndex={-1}
          >
            {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
          </button>
        )}
      </div>
      {error && <span className="error-message" style={{ color: '#dc2626', fontSize: '12px', marginTop: '4px', display: 'block' }}>{error}</span>}
    </div>
  );
};
