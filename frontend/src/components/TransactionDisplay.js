import {useState} from 'react'
import { CategoryPie } from "./CategoryPie"
import TransactionList from "./TransactionList"
import './styles/TransactionDisplay.css'

export default function TransactionDisplay() {
    
   let [transactions, setTransactions] = useState();
    const getTransactions = async () => {
        const response = await fetch("http://localhost:8080/api/transactions")
        const data = await response.json();
        let firstTenTrans = []
        for (let i = 0; i < 10; i++) {
            firstTenTrans.push(data.added[i])
        }
        console.log("firstTenTrans:", firstTenTrans)
        setTransactions(firstTenTrans)
    }



    return (
        <>
            <button onClick={getTransactions}>Get those things!</button>
            <div className="row">
                <div className="col">{transactions && <TransactionList transactions={transactions}/>}</div>
                <div className="col">{transactions && <CategoryPie transactions={transactions}/>}</div>
            </div>
        </>
    )

}
