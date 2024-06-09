let plaidLinkHandler;
const linkTokenResp = fetch("http://localhost:8080/api/linktoken", {
	method: "POST",
})
	.then((resp) => resp.json())
	.then((data) => {
		console.log(data);
		const linkToken = data.link_token;
		plaidLinkHandler = Plaid.create({
			token: linkToken,
			onSuccess: () => { exchangePublicToken() },
			onLoad: () => { },
			onExit: (err, metadata) => { },
			onEvent: (eventName, metadata) => { },
		})
	});

async function exchangePublicToken(public_token, metadata) {
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
};

function linkClickHandler() {
	console.log(plaidLinkHandler);
	plaidLinkHandler.open();
}
