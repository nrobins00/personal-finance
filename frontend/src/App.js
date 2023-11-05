import logo from './logo.svg';
import './App.css';
import {
    usePlaidLink,
    PlaidLinkOptions,
    PlaidLinkOnSuccess,
} from 'react-plaid-link';
import { useEffect, useState, useCallback } from 'react';
import './components/TransactionList';
import TransactionList from './components/TransactionList';

function App() {
    let [linkToken, setLinkToken] = useState(null)
    let [publicToken, setPublicToken] = useState(null)
    let [err, setErr] = useState(null)
    const fetchLinkTokenAndDoLink = async () => {
        if (linkToken) return;
        const response = await fetch("http://localhost:8080/api/linktoken", {method: 'POST'});
        const data = await response.json();
        console.log(data.link_token);
        setLinkToken(data.link_token)
    }
    useEffect(() => {fetchLinkTokenAndDoLink()}, [])
  return (
    <div className="App">
      <header className="App-header">
        <p>
      {linkToken && <LinkButton linkToken={linkToken} setPublicToken={setPublicToken}/>} 
        </p>
      <TransactionList/>
      </header>
    </div>
  );
}

function LinkButton({linkToken, setPublicToken}) {
    let [accessToken, setAccessToken] = useState(null);
    const onSuccess = async (public_token, metadata) => {
        const response = await fetch("http://localhost:8080/api/publicToken", {
            method: 'POST',
            headers: { 
                "Content-Type": "application/json",
            },
            body: JSON.stringify({"Public_token": public_token}),
        });
        const data = await response.json();
        console.log(data.access_token);
        setAccessToken(data.access_token);
    };
    const config = {
        onSuccess: (public_token, metadata) => {onSuccess(public_token, metadata)},
        onExit: (err, metadata) => {}, 
        onEvent: (eventName, metadata) => {},
        token: linkToken,
    };
    const { open, ready } = usePlaidLink(config);
    const clickHandler = () => {
        if (ready) {
            open();
        }
    }
    return (
        <>
            <button onClick={clickHandler}>
               Link your bank
            </button>
            {accessToken}
        </>
    );
}

export default App;
