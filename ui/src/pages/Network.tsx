import React, { useState } from 'react';
import DiagnosticForm from '../components/DiagnosticForm';
import ResultView from '../components/ResultView';
import LoadingSpinner from '../components/LoadingSpinner';
import ErrorAlert from '../components/ErrorAlert';
import { runDiagnostic } from '../api/diagnostics';

const NETWORK_TOOLS = [
  { label: 'Ping', value: 'ping', placeholder: 'Enter domain or IP for Ping' },
  { label: 'Traceroute', value: 'traceroute', placeholder: 'Enter domain or IP for Traceroute' },
  { label: 'WHOIS', value: 'whois', placeholder: 'Enter domain or IP for WHOIS' },
];

const Network: React.FC = () => {
  const [tool, setTool] = useState(NETWORK_TOOLS[0].value);
  const [result, setResult] = useState<any>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleToolChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    setTool(e.target.value);
    setResult(null);
    setError(null);
  };

  const handleSubmit = async (value: string) => {
    setLoading(true);
    setError(null);
    setResult(null);
    try {
      const res = await runDiagnostic(`network/${tool}`, value);
      setResult(res);
    } catch (err: any) {
      setError(err.toString());
    } finally {
      setLoading(false);
    }
  };

  const currentTool = NETWORK_TOOLS.find(t => t.value === tool);

  return (
    <div>
      <h2>Network Tools</h2>
      <label htmlFor="network-tool-select">Select Tool:</label>
      <select id="network-tool-select" value={tool} onChange={handleToolChange} style={{ marginBottom: '1rem' }}>
        {NETWORK_TOOLS.map(t => (
          <option key={t.value} value={t.value}>{t.label}</option>
        ))}
      </select>
      <DiagnosticForm onSubmit={handleSubmit} placeholder={currentTool?.placeholder || 'Enter domain or IP'} />
      {loading && <LoadingSpinner />}
      {error && <ErrorAlert message={error} />}
      {result && <ResultView result={result} />}
    </div>
  );
};

export default Network;
