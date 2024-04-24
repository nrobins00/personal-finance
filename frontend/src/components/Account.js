import React from 'react'

export function Account({ account }) {
    return (
        <ul>
            <li>
                {account.Name}
            </li>
            <li>
                {account.AvailableBalance}
            </li>
            <li>
                {account.CurrentBalance}
            </li>
            <li>
                {account.Mask}
            </li>
        </ul>
    )
}
