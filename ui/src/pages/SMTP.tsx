import React, { useState } from 'react';
import DiagnosticForm from '../components/DiagnosticForm';
import ResultView from '../components/ResultView';
import LoadingSpinner from '../components/LoadingSpinner';
import ErrorAlert from '../components/ErrorAlert';
import { smtpConnect, smtpStartTLS, SMTPConnectionResponse, ErrorResponse } from '../api/diagnostics';

interface CombinedSMTPResult {
  domain: string;
  connections: {
    [port: number]: SMTPConnectionResponse;
  };
  startTLS: {
    [port: number]: any;  // Using 'any' here as the type is not clearly defined in the API
  };
  error?: string;
}

const SMTP: React.FC = () => {
  const [result, setResult] = useState<CombinedSMTPResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (value: string) => {
    setLoading(true);
    setError(null);
    setResult(null);
    
    try {
      // Make concurrent connections to common SMTP ports
      const commonPorts = [25, 465, 587];
      
      // Run connection checks
      const connectionPromises = commonPorts.map(port => 
        smtpConnect(value, port)
          .catch(err => {
            // Return failed connection with error info
            return {
              host: value,
              port: port,
              connected: false,
              error: err.error || 'Connection failed'
            } as SMTPConnectionResponse;
          })
      );

      // Run STARTTLS checks
      const startTLSPromises = commonPorts.map(() => 
        smtpStartTLS(value)
          .catch(err => {
            // Return failed STARTTLS check with error info
            return {
              host: value,
              error: err.error || 'STARTTLS check failed'
            };
          })
      );

      // Wait for all checks to complete
      const [connectionResults, startTLSResults] = await Promise.all([
        Promise.all(connectionPromises),
        Promise.all(startTLSPromises)
      ]);
      
      // Combine the results
      const combinedResult: CombinedSMTPResult = {
        domain: value,
        connections: {},
        startTLS: {}
      };
      
      // Add results for each port
      connectionResults.forEach((portResult, index) => {
        combinedResult.connections[commonPorts[index]] = portResult;
      });
      
      // Add STARTTLS results for each port
      startTLSResults.forEach((tlsResult, index) => {
        combinedResult.startTLS[commonPorts[index]] = tlsResult;
      });
      
      setResult(combinedResult);
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
