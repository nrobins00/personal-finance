import React from 'react'
import type { Account } from '../types/types'

export default function AccountDisplay(props: { account: Account }) {
    let account = props.account;
    console.log("account here")
    return (
        <div style={{ border: 'solid' }}>
            <p>{account.Name}</p>
            <ul>
                <li>
                    Available balance: {account.AvailableBalance}
                </li>
                <li>
                    Current balance: {account.CurrentBalance}
                </li>
                {account.Mask &&
                    <li>
                        Mask: {account.Mask}
                    </li>
                }
            </ul>
        </div>
    )
}
