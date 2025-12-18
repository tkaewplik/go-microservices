import React, { useState } from 'react';

interface TransactionFormProps {
  userId: number;
  token: string;
  onTransactionCreated: () => void;
}

const TransactionForm: React.FC<TransactionFormProps> = ({ userId, token, onTransactionCreated }) => {
  const [amount, setAmount] = useState('');
  const [description, setDescription] = useState('');
  const [customAuthHeader, setCustomAuthHeader] = useState(`Bearer ${token}`);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [loading, setLoading] = useState(false);

  const apiUrl = process.env.REACT_APP_API_URL || 'http://localhost:8080';

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setSuccess('');
    setLoading(true);

    try {
      const response = await fetch(`${apiUrl}/payment/transactions`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': customAuthHeader,
        },
        body: JSON.stringify({
          user_id: userId,
          amount: parseFloat(amount),
          description,
        }),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || 'Failed to create transaction');
      }

      setSuccess('Transaction created successfully!');
      setAmount('');
      setDescription('');
      onTransactionCreated();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <h2>Create Transaction</h2>
      
      {error && <div className="error-message">{error}</div>}
      {success && <div className="success-message">{success}</div>}

      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label htmlFor="amount">Amount</label>
          <input
            type="number"
            id="amount"
            step="0.01"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            required
          />
          <small style={{ color: '#666' }}>Max total: 1000</small>
        </div>

        <div className="form-group">
          <label htmlFor="description">Description</label>
          <textarea
            id="description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            required
          />
        </div>

        <div className="form-group">
          <label htmlFor="authHeader">Authorization Header (customizable)</label>
          <input
            type="text"
            id="authHeader"
            value={customAuthHeader}
            onChange={(e) => setCustomAuthHeader(e.target.value)}
            placeholder="Bearer your-token-here"
          />
          <small style={{ color: '#666' }}>You can modify the auth header for testing</small>
        </div>

        <button type="submit" className="btn" disabled={loading}>
          {loading ? 'Creating...' : 'Create Transaction'}
        </button>
      </form>
    </div>
  );
};

export default TransactionForm;
