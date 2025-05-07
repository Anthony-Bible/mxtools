import React, { useState } from 'react';
import DiagnosticForm from '../components/DiagnosticForm';
import ResultView from '../components/ResultView';
import LoadingSpinner from '../components/LoadingSpinner';
import ErrorAlert from '../components/ErrorAlert';
import { smtpCheck, SMTPResponse, ErrorResponse } from '../api/diagnostics';

const SMTP: React.FC = () => {
  const [result, setResult] = useState<SMTPResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (value: string) => {
    setLoading(true);
    setError(null);
    setResult(null);
    try {
      // Use generic SMTP check endpoint
      const res = await smtpCheck(value);
      setResult(res);
    } catch (err: any) {
      const errorResponse = err as ErrorResponse;
      setError(errorResponse.error || err.toString());
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <h2>SMTP Diagnostics</h2>
      <DiagnosticForm onSubmit={handleSubmit} placeholder="Enter mail server hostname or IP" />
      {loading && <LoadingSpinner />}
      {error && <ErrorAlert message={error} />}
      {result && <ResultView result={result} />}
    </div>
  );
};

export default SMTP;
