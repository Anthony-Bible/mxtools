import React from 'react';

interface DiagnosticFormProps {
  onSubmit: (input: string) => void;
  placeholder?: string;
}

const DiagnosticForm: React.FC<DiagnosticFormProps> = ({ onSubmit, placeholder }) => {
  const [input, setInput] = React.useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit(input);
  };

  return (
    <form onSubmit={handleSubmit}>
      <input
        type="text"
        value={input}
        onChange={e => setInput(e.target.value)}
        placeholder={placeholder || 'Enter value'}
      />
      <button type="submit">Run</button>
    </form>
  );
};

export default DiagnosticForm;
