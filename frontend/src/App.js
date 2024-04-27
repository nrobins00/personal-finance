import logo from './logo.svg';
import './App.css';
import {
    usePlaidLink,
    PlaidLinkOptions,
    PlaidLinkOnSuccess,
} from 'react-plaid-link';
import { useEffect, useState, useCallback } from 'react';
import TransactionDisplay from './components/TransactionDisplay';
import { Account } from './components/Account';

function App() {
    const [loggedIn, setLoggedIn] = useState(false);
    return (
        <div className="App">
            <header className="App-header">
                {loggedIn ? <HomePage /> : <LoginForm setLoggedIn={setLoggedIn} />}
            </header>
        </div>
    );
}

function LoginForm({ setLoggedIn }) {
    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");
    const handleSubmit = async (e) => {
        e.preventDefault();
        console.log(username + ":" + password);

        const response = await fetch("http://localhost:8080/signin", {
            method: 'POST',
            headers: {
                "Authorization": btoa(username + ":" + password)
            },
            credentials: "include",
        });
        if (response.status === 200) {
            setLoggedIn(true);
        } else {
            //TODO
        }
    }
    return (
        <form onSubmit={handleSubmit}>
            <label>
                Username:
                <input type="text" value={username}
                    onChange={(e) => setUsername(e.target.value)} />
            </label>
            <label>
                Password:
                <input type="password" value={password}
                    onChange={(e) => setPassword(e.target.value)} />
            </label>
            <input type={"submit"}
                style={{ backgroundColor: "#a1eafb" }} />
        </form>

    );

}

function HomePage() {
    let [linkToken, setLinkToken] = useState(null)
    let [publicToken, setPublicToken] = useState(null)
    let [err, setErr] = useState(null)
    let [accounts, setAccounts] = useState([])
    const fetchLinkTokenAndDoLink = async () => {
        if (linkToken) return;
        const response = await fetch("http://localhost:8080/api/linktoken", { method: 'POST' });
        const data = await response.json();
        console.log(data.link_token);
        setLinkToken(data.link_token)
    }
    const getAllAccounts = async () => {
        const response = await fetch("http://localhost:8080/api/accounts", {
            method: 'GET',
            credentials: "include",
        });
        const data = await response.json()
        setAccounts(data.accounts)
        console.log(data);
    }

    useEffect(() => { fetchLinkTokenAndDoLink() }, [])

    return (
        <div>
            <header>
                <p>
                    {linkToken && <LinkButton linkToken={linkToken} setPublicToken={setPublicToken} />}
                </p>
                <p>
                    <button onClick={getAllAccounts}>
                        Get all items
                    </button>
                </p>
                <TransactionDisplay />
            </header>
            <div style={{ display: 'flex', gap: '90px' }}>
                {accounts.map(acc => {
                    return <Account account={acc} />
                })
                }
            </div>
        </div>
    );
}

function LinkButton({ linkToken, setPublicToken }) {
    let [accessToken, setAccessToken] = useState(null);
    const onSuccess = async (public_token, metadata) => {
        const response = await fetch("http://localhost:8080/api/publicToken", {
            method: 'POST',
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({ "Public_token": public_token }),
            credentials: "include",
        });
        const data = await response.json();
        console.log(data.access_token);
        setAccessToken(data.access_token);
    };
    const config = {
        onSuccess: (public_token, metadata) => { onSuccess(public_token, metadata) },
        onExit: (err, metadata) => { console.log("err: " + err + "; metadata: " + metadata) },
        onEvent: (eventName, metadata) => { },
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
