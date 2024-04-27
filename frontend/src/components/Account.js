import React from 'react'

export function Account({ account }) {
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
