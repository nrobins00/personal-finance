import { useState } from "react";
//import { CategoryPie } from "./CategoryPie";
import TransactionList from "./TransactionList";
import "./styles/TransactionDisplay.css";
import type { Transaction } from "../types/types";

type transactionsResponse = {
  transactions: Transaction[]
}

export default function TransactionDisplay() {
  let [transactions, setTransactions] = useState<Transaction[]>([]);
  const getTransactions = async () => {
    const response = await fetch("http://localhost:8080/api/transactions", {
      method: "GET",
      credentials: "include",
    });
    const data: transactionsResponse = await response.json();
    let firstTenTrans: Transaction[] = [];
    let i = 0
    while (i < data.transactions?.length) { //&& firstTenTrans.length < 10) {
      firstTenTrans.push(data.transactions[i])
      i++;
    }
    console.log("firstTenTrans:", firstTenTrans);
    setTransactions(firstTenTrans);
  };

  return (
    <>
      <button onClick={getTransactions}>Get all transactions</button>
      <div className="row">
        <div className="col">
          {transactions.length > 0 && (
            <TransactionList transactions={transactions} />
          )}
        </div>
        <div className="col">
          {transactions.length > 0 //&&
            //<CategoryPie transactions={transactions} />
          }
        </div>
      </div>
    </>
  );
}
