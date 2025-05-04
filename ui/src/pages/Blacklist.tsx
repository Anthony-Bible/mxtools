import React, { useState } from 'react';
import DiagnosticForm from '../components/DiagnosticForm';
import ResultView from '../components/ResultView';
import LoadingSpinner from '../components/LoadingSpinner';
import ErrorAlert from '../components/ErrorAlert';
import { runDiagnostic } from '../api/diagnostics';

const Blacklist: React.FC = () => {
  const [result, setResult] = useState<any>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (value: string) => {
    setLoading(true);
    setError(null);
    setResult(null);
    try {
      const res = await runDiagnostic('blacklist', value);
      setResult(res);
    } catch (err: any) {
      setError(err.toString());
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
