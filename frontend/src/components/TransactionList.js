import {useState} from 'react'
import './TransactionList.css'
import { CategoryPie } from './CategoryPie';
import { CategoryList } from './CategoryList';

function TransactionList() {
   let [transactions, setTransactions] = useState([]);
    const getTransactions = async () => {
        const response = await fetch("http://localhost:8080/api/transactions")
        const data = await response.json();
        setTransactions([...data.added]);
    }

    return (
        <>
            {transactions && <CategoryPie transactions={transactions} /> }
            {transactions && <CategoryList transactions={transactions} />}
            <button onClick={getTransactions}>Get those things!</button>
            <ul className='transaction-container'>
                {transactions.map((item) => {
                    return (
                        <li>
                            <div >
                                <ul className='transaction-inner-container'>
                                    <li>
                                        Name: {item.name}
                                    </li>
                                    <li>
                                        Category: {item.category[0]}
                                    </li>
                                    <li>
                                        Amount: {item.amount}
                                    </li>
                                </ul>
                            </div>
                        </li>
                    );
                    }
                )}
            </ul>
        </>
    );
}

export default TransactionList;
