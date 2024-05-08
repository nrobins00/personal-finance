//import "./App.css";
import {
  usePlaidLink,
  PlaidLinkOptions,
  PlaidLinkOnSuccess,
  PlaidLinkOnSuccessMetadata,
  PlaidLinkOnExitMetadata,
  PlaidLinkOnEventMetadata,
  PlaidLinkError,
} from "react-plaid-link";
import { useEffect, useState, FormEvent } from "react";
import TransactionDisplay from "./components/TransactionDisplay";
import Account from "./components/Account";
import BudgetUpdateForm from "./components/BudgetUpdateForm";


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

type LoginProps = {
  setLoggedIn: React.Dispatch<React.SetStateAction<boolean>>
}

function LoginForm({ setLoggedIn }: LoginProps) {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    console.log(username + ":" + password);

    const response = await fetch("http://localhost:8080/signin", {
      method: "POST",
      headers: {
        Authorization: btoa(username + ":" + password),
      },
      credentials: "include",
    });
    if (response.status === 200) {
      setLoggedIn(true);
    } else {
      //TODO
    }
  };
  return (
    <form onSubmit={handleSubmit}>
      <label>
        Username:
        <input
          type="text"
          value={username}
          onChange={(e) => setUsername(e.target.value)}
        />
      </label>
      <label>
        Password:
        <input
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
        />
      </label>
      <input type={"submit"} />
    </form>
  );
}

function HomePage() {
  let [curBudget, setCurBudget] = useState(0.0);
  let [spendings, setSpendings] = useState(0.0);
  let [linkToken, setLinkToken] = useState(null);
  let [accounts, setAccounts] = useState([]);
  let [showUpdateBudget, setShowUpdateBudget] = useState(false);
  const fetchLinkTokenAndDoLink = async () => {
    if (linkToken) return;
    const response = await fetch("http://localhost:8080/api/linktoken", {
      method: "POST",
    });
    const data = await response.json();
    console.log(data.link_token);
    setLinkToken(data.link_token);
  };
  const getBudget = async () => {
    const response = await fetch("http://localhost:8080/api/budget", {
      method: "GET",
      credentials: "include",
    });
    if (response.status === 200) {
      const data = await response.json();
      setCurBudget(parseFloat(data.budget));
    }
  };
  const getSpending = async () => {
    const response = await fetch("http://localhost:8080/api/spendings", {
      method: "GET",
      credentials: "include",
    });
    if (response.status === 200) {
      const data = await response.json();
      setSpendings(parseFloat(data.spendings));
    }
  }
  const getAllAccounts = async () => {
    const response = await fetch("http://localhost:8080/api/accounts", {
      method: "GET",
      credentials: "include",
    });
    const data = await response.json();
    setAccounts(data.accounts);
    console.log(data);
  };
  const handleBudgetUpdateSubmit = (newBudget: number) => {
    setCurBudget(newBudget)
    setShowUpdateBudget(false)
  }

  useEffect(() => {
    fetchLinkTokenAndDoLink();
    getBudget();
    getSpending();
  }, []);

  return (
    <div>
      <header>
        <div style={{ display: 'flex' }}>
          <p>budget: {curBudget}</p>
          <button onClick={() => setShowUpdateBudget(true)}>Update Budget</button>
          {showUpdateBudget && <BudgetUpdateForm handleBudgetUpdate={handleBudgetUpdateSubmit} />}
        </div>
        <p>spent: {spendings}</p>
        <p>
          {linkToken && (
            <LinkButton linkToken={linkToken} />
          )}
        </p>
        <p>
          <button onClick={getAllAccounts}>Get all items</button>
        </p>
        <TransactionDisplay />
      </header >
      <div style={{ display: "flex", gap: "90px" }}>
        {accounts.map((acc) => {
          return <Account account={acc} />;
        })}
      </div>
    </div >
  );
}

function LinkButton(props: { linkToken: string }) {
  let [accessToken, setAccessToken] = useState(null);
  const onSuccess = async (public_token: string, metadata: PlaidLinkOnSuccessMetadata) => {
    const response = await fetch("http://localhost:8080/api/publicToken", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ Public_token: public_token }),
      credentials: "include",
    });
    const data = await response.json();
    console.log(data.access_token);
    setAccessToken(data.access_token);
  };
  const config: PlaidLinkOptions = {
    onSuccess: (public_token: string, metadata: PlaidLinkOnSuccessMetadata) => {
      onSuccess(public_token, metadata);
    },
    onExit: (err: null | PlaidLinkError, metadata: PlaidLinkOnExitMetadata) => {
      console.log("err: " + err + "; metadata: " + metadata);
    },
    onEvent: (eventName: string, metadata: PlaidLinkOnEventMetadata) => { },
    token: props.linkToken,
  };
  const { open, ready } = usePlaidLink(config);
  const clickHandler = () => {
    if (ready) {
      open();
    }
  };
  return (
    <>
      <button onClick={clickHandler}>Link your bank</button>
      {accessToken}
    </>
  );
}

export default App;
