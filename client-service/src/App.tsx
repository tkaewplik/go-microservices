import React, { useState } from 'react';
import './App.css';
import LoginForm from './components/LoginForm';
import TransactionForm from './components/TransactionForm';
import TransactionList from './components/TransactionList';

interface User {
  id: number;
  username: string;
  token: string;
}

function App() {
  const [user, setUser] = useState<User | null>(null);
  const [refreshKey, setRefreshKey] = useState(0);

  const handleLogin = (userData: User) => {
    setUser(userData);
  };

  const handleLogout = () => {
    setUser(null);
  };

  const handleTransactionCreated = () => {
    setRefreshKey(prev => prev + 1);
  };

  return (
    <div className="App">
      <header className="App-header">
        <h1>Go Microservices Payment System</h1>
        {user && (
          <div className="user-info">
            <span>Welcome, {user.username}!</span>
            <button onClick={handleLogout} className="logout-btn">Logout</button>
          </div>
        )}
      </header>
      <main className="App-main">
        {!user ? (
          <LoginForm onLogin={handleLogin} />
        ) : (
          <div className="dashboard">
            <div className="left-panel">
              <TransactionForm 
                userId={user.id} 
                token={user.token}
                onTransactionCreated={handleTransactionCreated}
              />
            </div>
            <div className="right-panel">
              <TransactionList 
                userId={user.id} 
                token={user.token}
                refreshKey={refreshKey}
                onRefresh={handleTransactionCreated}
              />
            </div>
          </div>
        )}
      </main>
    </div>
  );
}

export default App;
