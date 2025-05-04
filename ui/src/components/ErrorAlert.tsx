import React from 'react';

interface ErrorAlertProps {
  message: string;
}

const ErrorAlert: React.FC<ErrorAlertProps> = ({ message }) => (
  <div style={{ color: 'red', margin: '1em 0' }}>
    <strong>Error:</strong> {message}
  </div>
);

export default ErrorAlert;
