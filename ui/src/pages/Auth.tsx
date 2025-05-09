import React, { useState } from 'react';
import DiagnosticForm from '../components/DiagnosticForm';
import ResultView from '../components/ResultView';
import LoadingSpinner from '../components/LoadingSpinner';
import ErrorAlert from '../components/ErrorAlert';
import { spfCheck, dmarcCheck, dkimCheck, SPFResponse, DMARCResponse, DKIMResponse, ErrorResponse } from '../api/diagnostics';

interface AuthResult {
  spf?: SPFResponse;
  dmarc?: DMARCResponse;
  dkim?: DKIMResponse;
}

const Auth: React.FC = () => {
  const [result, setResult] = useState<AuthResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [dkimSelector, setDkimSelector] = useState<string>('default');
  const [useDefaultSelector, setUseDefaultSelector] = useState<boolean>(false);

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
      
      // Run DKIM check with or without selector based on user preference
      try {
        if (useDefaultSelector) {
          // Use backend default selectors
          authResult.dkim = await dkimCheck(domain);
          // Make sure the domain is set correctly for display
          if (authResult.dkim && !authResult.dkim.domain) {
            authResult.dkim.domain = domain;
          }
        } else {
          // Use user-provided selector
          authResult.dkim = await dkimCheck(domain, dkimSelector);
          // Make sure the domain is set correctly for display
          if (authResult.dkim && !authResult.dkim.domain) {
            authResult.dkim.domain = domain;
          }
        }
      } catch (dkimErr: any) {
        authResult.dkim = {
          domain,
          selector: useDefaultSelector ? 'default backend selectors' : dkimSelector,
          hasRecords: false,
          isValid: false,
          error: dkimErr.error || dkimErr.toString()
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
      <div className="mb-3 form-check">
        <input
          type="checkbox"
          className="form-check-input"
          id="useDefaultSelector"
          checked={useDefaultSelector}
          onChange={(e) => setUseDefaultSelector(e.target.checked)}
        />
        <label className="form-check-label" htmlFor="useDefaultSelector">
          Use backend default selectors (mail, google)
        </label>
      </div>
      
      {!useDefaultSelector && (
        <div className="mb-3">
          <label htmlFor="dkimSelector" className="form-label">DKIM Selector:</label>
          <input 
            type="text" 
            id="dkimSelector" 
            className="form-control" 
            value={dkimSelector} 
            onChange={(e) => setDkimSelector(e.target.value)}
            placeholder="Enter DKIM selector (e.g. default, google, mail)"
          />
          <small className="text-muted">Common selectors: default, google, mail, selector1, selector2, k1</small>
        </div>
      )}
      
      <DiagnosticForm onSubmit={handleSubmit} placeholder="Enter domain or email address" />
      {loading && <LoadingSpinner />}
      {error && <ErrorAlert message={error} />}
      {result && <ResultView result={result} />}
    </div>
  );
};

export default Auth;
