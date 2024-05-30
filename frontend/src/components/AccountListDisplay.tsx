import { useEffect, useState } from "react";
import { Account } from '../types/types'
import AccountDisplay from "./AccountDisplay";
import {
    usePlaidLink,
    PlaidLinkOptions,
    PlaidLinkOnSuccess,
    PlaidLinkOnSuccessMetadata,
    PlaidLinkOnExitMetadata,
    PlaidLinkOnEventMetadata,
    PlaidLinkError,
} from "react-plaid-link";

export default function AccountListDisplay() {
    let [accounts, setAccounts] = useState<Account[]>([]);
    const getAllAccounts = async () => {
        const response = await fetch("http://localhost:8080/api/accounts", {
            method: "GET",
            credentials: "include",
        });
        const data = await response.json();
        setAccounts(data.accounts);
        console.log(data);
    };
    useEffect(() => {
        getAllAccounts();
    }, [])
    return <><div style={{ display: 'flex', gap: '40px', marginTop: '20px' }}>
        {accounts.map((acc) => {
            return <AccountDisplay account={acc} />
        })}

    </div>
        <LinkButton />
    </>;

}


function LinkButton() {
    let [linkToken, setLinkToken] = useState<string>("")
    let [accessToken, setAccessToken] = useState(null);
    const fetchLinkTokenAndDoLink = async (curLinkToken: string) => {
        //if (curLinkToken) return;
        const response = await fetch("http://localhost:8080/api/linktoken", {
            method: "POST",
        });
        const data = await response.json();
        console.log(data.link_token);
        setLinkToken(data.link_token);
    };
    const onSuccess = async (public_token: string, metadata: PlaidLinkOnSuccessMetadata) => {
        const response = await fetch("http://localhost:8080/api/publicToken", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({ Public_token: public_token }),
            credentials: "include",
        });
        const data = await response.json();
        console.log(data.access_token);
        setAccessToken(data.access_token);
    };
    //TODO: stick this all inside the button onClick handler
    const config: PlaidLinkOptions = {
        onSuccess: (public_token: string, metadata: PlaidLinkOnSuccessMetadata) => {
            onSuccess(public_token, metadata);
        },
        onExit: (err: null | PlaidLinkError, metadata: PlaidLinkOnExitMetadata) => {
            console.log("err: " + err + "; metadata: " + metadata);
        },
        onEvent: (eventName: string, metadata: PlaidLinkOnEventMetadata) => {
            console.log("event! " + eventName);
        },
        token: linkToken || null,
    };
    const { error, ready, exit, open } = usePlaidLink(config);
    console.log(error, ready, exit, open);
    useEffect(() => {
        console.log("useEffect!")
        if (ready) {
            exit();
            console.log("ready!");
            open();
        }
    });
    return (
        <>
            <button onClick={() => fetchLinkTokenAndDoLink(linkToken)}>Link new bank</button>
            {accessToken}
        </>
    );
}
