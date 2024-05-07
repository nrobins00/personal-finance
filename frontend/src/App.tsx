import "./App.css";
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
      <input type={"submit"} style={{ backgroundColor: "#a1eafb" }} />
    </form>
  );
}

function HomePage() {
  let [curBudget, setCurBudget] = useState(0.0);
  let [budget, setBudget] = useState("0");
  let [spendings, setSpendings] = useState(0.0);
  let [linkToken, setLinkToken] = useState(null);
  let [accounts, setAccounts] = useState([]);
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

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    let numBudget = parseFloat(budget)
    const response = await fetch("http://localhost:8080/api/budget/set", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      credentials: "include",
      body: JSON.stringify({ budget: numBudget }),
    });
    setCurBudget(numBudget)
  };

  useEffect(() => {
    fetchLinkTokenAndDoLink();
    getBudget();
    getSpending();
  }, []);

  return (
    <div>
      <header>
        <p>budget: {curBudget}</p>
        <form onSubmit={handleSubmit}>
          <label>
            Set new budget:
            <input value={budget} onChange={(e) => setBudget(e.target.value)} />
          </label>
          <input type="submit" />
        </form>
        <p>
          spendings: {spendings}
        </p>
        <p>
          {linkToken && (
            <LinkButton linkToken={linkToken} />
          )}
        </p>
        <p>
          <button onClick={getAllAccounts}>Get all items</button>
        </p>
        <TransactionDisplay />
      </header>
      <div style={{ display: "flex", gap: "90px" }}>
        {accounts.map((acc) => {
          return <Account account={acc} />;
        })}
      </div>
    </div>
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
