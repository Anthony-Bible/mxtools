import React, { useState } from 'react';
import DiagnosticForm from '../components/DiagnosticForm';
import ResultView from '../components/ResultView';
import LoadingSpinner from '../components/LoadingSpinner';
import ErrorAlert from '../components/ErrorAlert';
import { spfCheck, dmarcCheck, SPFResponse, DMARCResponse, ErrorResponse } from '../api/diagnostics';

interface AuthResult {
  spf?: SPFResponse;
  dmarc?: DMARCResponse;
}

const Auth: React.FC = () => {
  const [result, setResult] = useState<AuthResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (value: string) => {
    setLoading(true);
    setError(null);
    setResult(null);
    try {
      // Run SPF and DMARC checks for the domain
      const authResult: AuthResult = {};
      
      // Extract domain from email if provided
      const domain = value.includes('@') ? value.split('@')[1] : value;
      
      // Run SPF check
      try {
        authResult.spf = await spfCheck(domain);
      } catch (spfErr: any) {
        authResult.spf = {
          domain,
          hasRecord: false,
          isValid: false,
          error: spfErr.error || spfErr.toString()
        };
      }
      
      // Run DMARC check
      try {
        authResult.dmarc = await dmarcCheck(domain);
      } catch (dmarcErr: any) {
        authResult.dmarc = {
          domain,
          hasRecord: false,
          isValid: false,
          error: dmarcErr.error || dmarcErr.toString()
        };
      }
      
      setResult(authResult);
    } catch (err: any) {
      const errorResponse = err as ErrorResponse;
      setError(errorResponse.error || err.toString());
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <h2>Email Authentication Diagnostics</h2>
      <DiagnosticForm onSubmit={handleSubmit} placeholder="Enter domain or email address" />
      {loading && <LoadingSpinner />}
      {error && <ErrorAlert message={error} />}
      {result && <ResultView result={result} />}
    </div>
  );
};

export default Auth;
