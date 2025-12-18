import React, { useEffect, useState } from 'react';

interface Transaction {
  id: number;
  user_id: number;
  amount: number;
  description: string;
  is_paid: boolean;
  created_at: string;
}

interface TransactionListProps {
  userId: number;
  token: string;
  refreshKey: number;
  onRefresh: () => void;
}

const TransactionList: React.FC<TransactionListProps> = ({ userId, token, refreshKey, onRefresh }) => {
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [paying, setPaying] = useState(false);

  const apiUrl = process.env.REACT_APP_API_URL || 'http://localhost:8080';

  const fetchTransactions = async () => {
    setLoading(true);
    setError('');

    try {
      const response = await fetch(`${apiUrl}/payment/transactions/list?user_id=${userId}`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || 'Failed to fetch transactions');
      }

      setTransactions(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setLoading(false);
    }
  };

  const handlePayAll = async () => {
    setPaying(true);
    setError('');

    try {
      const response = await fetch(`${apiUrl}/payment/transactions/pay?user_id=${userId}`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || 'Failed to pay transactions');
      }

      await fetchTransactions();
      onRefresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setPaying(false);
    }
  };

  useEffect(() => {
    fetchTransactions();
  }, [refreshKey]);

  const totalAmount = transactions.reduce((sum, t) => sum + t.amount, 0);
  const unpaidCount = transactions.filter(t => !t.is_paid).length;

  return (
    <div className="transaction-list">
      <h2>Your Transactions</h2>

      {error && <div className="error-message">{error}</div>}

      <div className="transaction-stats">
        <p>Total Transactions: {transactions.length}</p>
        <p>Unpaid Transactions: {unpaidCount}</p>
        <p className="total">Total Amount: ${totalAmount.toFixed(2)} / $1000.00</p>
      </div>

      {unpaidCount > 0 && (
        <button 
          onClick={handlePayAll} 
          className="btn btn-secondary"
          disabled={paying}
        >
          {paying ? 'Processing...' : `Pay All Unpaid Transactions (${unpaidCount})`}
        </button>
      )}

      {loading ? (
        <p>Loading transactions...</p>
      ) : transactions.length === 0 ? (
        <div className="no-transactions">
          No transactions yet. Create your first transaction!
        </div>
      ) : (
        <div style={{ marginTop: '20px' }}>
          {transactions.map((transaction) => (
            <div 
              key={transaction.id} 
              className={`transaction-item ${transaction.is_paid ? 'paid' : ''}`}
            >
              <p className="amount">${transaction.amount.toFixed(2)}</p>
              <p><strong>Description:</strong> {transaction.description}</p>
              <p><strong>Status:</strong> {transaction.is_paid ? '✓ Paid' : '○ Unpaid'}</p>
              <p><strong>Created:</strong> {new Date(transaction.created_at).toLocaleString()}</p>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default TransactionList;
