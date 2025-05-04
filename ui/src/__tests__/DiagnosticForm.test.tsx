import { render, screen, fireEvent } from '@testing-library/react';
import DiagnosticForm from '../components/DiagnosticForm';

test('renders DiagnosticForm and submits input', () => {
  const handleSubmit = jest.fn();
  render(<DiagnosticForm onSubmit={handleSubmit} placeholder="Test input" />);
  const input = screen.getByPlaceholderText('Test input');
  fireEvent.change(input, { target: { value: 'example.com' } });
  fireEvent.submit(input.closest('form')!);
  expect(handleSubmit).toHaveBeenCalledWith('example.com');
});
