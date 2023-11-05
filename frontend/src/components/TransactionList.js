import {useState} from 'react';

function TransactionList() {
   let [transactions, setTransactions] = useState([]);
    const getTransactions = async () => {
        const response = await fetch("http://localhost:8080/api/transactions")
        const data = await response.json();
        setTransactions([...data.added]);
    }

    return (
        <>
            <button onClick={getTransactions}>Get those things!</button>
            <ul>
                {transactions.map((item) => {
                    return (
                        <li>

                            {item.name}
                        </li>
                    );
                    }
                )}
            </ul>
        </>
    );
}

export default TransactionList;
