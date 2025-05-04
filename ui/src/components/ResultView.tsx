import React from 'react';

interface ResultViewProps {
  result: any;
}

const ResultView: React.FC<ResultViewProps> = ({ result }) => (
  <div>
    <pre>{JSON.stringify(result, null, 2)}</pre>
  </div>
);

export default ResultView;
