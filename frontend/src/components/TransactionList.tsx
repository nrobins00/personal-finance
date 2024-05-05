import { useState } from "react";
import "./styles/TransactionList.css";
import type { Transaction } from "../types/types";

function TransactionList(props: { transactions: Transaction[] }) {
  let transactions = props.transactions
  let [visibleTrans, setVisibleTrans] = useState([...transactions])
  let [filteredCats, setFilteredCats] = useState(new Set<string>())

  function filterTransactions(newFilteredCats: Set<string>) {
    let newTransactions = transactions.filter((tr) => {
      return !newFilteredCats.has(tr.CategoryPrimary)
    })
    setVisibleTrans(newTransactions);
  }
  function toggleCategory(category: string) {
    let newFilteredCats = new Set(filteredCats);
    if (filteredCats.has(category)) {
      newFilteredCats.delete(category)
    } else {
      newFilteredCats.add(category);
    }
    setFilteredCats(newFilteredCats);
    filterTransactions(newFilteredCats);
  }
  return (
    <>
      <div>
        <button
          onClick={() => {
            toggleCategory("Travel");
          }}
        >
          Travel: {filteredCats.has("Travel") ? "filtered" : "showing"}
        </button>
        <button
          onClick={() => {
            toggleCategory("Food and Drink");
          }}
        >
          Food and Drink:{" "}
          {filteredCats.has("Food and Drink") ? "filtered" : "showing"}
        </button>
        <button
          onClick={() => {
            toggleCategory("Payment");
          }}
        >
          Payment: {filteredCats.has("Payment") ? "filtered" : "showing"}
        </button>
      </div>
      <ul className="transaction-container">
        {visibleTrans.map((item) => {
          return (
            <li>
              <div>
                <ul className="transaction-inner-container">
                  <li>Category: {item.CategoryDetailed}</li>
                  <li>Amount: {item.Amount}</li>
                </ul>
              </div>
            </li>
          );
        })}
      </ul>
    </>
  );
}

export default TransactionList;
