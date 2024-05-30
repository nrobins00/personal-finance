import "./App.css";
import { useEffect, useState, FormEvent } from "react";
import TransactionDisplay from "./components/TransactionDisplay";
import BudgetUpdateForm from "./components/BudgetUpdateForm";
import BudgetSpendingPie from "./components/BudgetSpendingPie";
import AccountListDisplay from "./components/AccountListDisplay";
import { getCookieValue } from "./utils/utils";


function App() {
  const [loggedIn, setLoggedIn] = useState(false);
  const [curPage, setCurPage] = useState('Home');
  useEffect(() => {
    let userId = getCookieValue(document.cookie, "userId");
    console.log(userId);
    if (userId != null) {
      setLoggedIn(true);
    }
  }, [])
  return (
    <div className="App">
      <div className="topnav">
        <button
          className={curPage === 'Home' && "active" || ""}
          onClick={() => { setCurPage('Home') }}
        >
          Home
        </button>
        <button
          className={curPage === 'Accounts' && "active" || ""}
          onClick={() => { setCurPage('Accounts') }}
        >
          Accounts
        </button>
        <button
          className={curPage === 'Budget' && "active" || ""}
          onClick={() => { setCurPage('Budget') }}
        >Budget Config
        </button>
      </div>
      <header className="App-header">
      </header>
      {loggedIn ?
        (curPage === 'Home' && <HomePage />) ||
        (curPage === 'Accounts' && <AccountListDisplay />) ||
        (curPage === 'Budget' &&
          <BudgetUpdateForm handleBudgetUpdate={(num) => console.log(num)} />)

        : <LoginForm setLoggedIn={setLoggedIn} />}
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

  useEffect(() => {
    getBudget();
    getSpending();
  }, []);

  return (
    <div>
      <header>
        <BudgetSpendingPie budget={curBudget} spending={spendings} />
        <p>budget: {curBudget}</p>
        <p>spent: {spendings}</p>
        <TransactionDisplay />
      </header >
    </div >
  );
}


export default App;
