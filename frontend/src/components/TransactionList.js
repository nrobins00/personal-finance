import './styles/TransactionList.css'

function TransactionList({transactions}) {

    return (
        <>
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
