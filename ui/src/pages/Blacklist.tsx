import React, { useState } from 'react';
import DiagnosticForm from '../components/DiagnosticForm';
import ResultView from '../components/ResultView';
import LoadingSpinner from '../components/LoadingSpinner';
import ErrorAlert from '../components/ErrorAlert';
import { blacklistCheck, BlacklistResponse, ErrorResponse } from '../api/diagnostics';

const Blacklist: React.FC = () => {
  const [result, setResult] = useState<BlacklistResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (value: string) => {
    setLoading(true);
    setError(null);
    setResult(null);
    try {
      const res = await blacklistCheck(value);
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
      <h2>Blacklist Diagnostics</h2>
      <DiagnosticForm onSubmit={handleSubmit} placeholder="Enter domain or IP address" />
      {loading && <LoadingSpinner />}
      {error && <ErrorAlert message={error} />}
      {result && <ResultView result={result} />}
    </div>
  );
};

export default Blacklist;
