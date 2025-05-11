let plaidLinkHandler;
const hostname = window.location.host;
const linkTokenResp = fetch("http://" + hostname + "/api/linktoken", {
  method: "POST",
})
  .then((resp) => resp.json())
  .then((data) => {
    console.log(data);
    const linkToken = data.link_token;
    plaidLinkHandler = Plaid.create({
      token: linkToken,
      onSuccess: (public_token, metadata) => {
        exchangePublicToken(public_token, metadata);
      },
      onLoad: () => {},
      onExit: (err, metadata) => {},
      onEvent: (eventName, metadata) => {},
    });
  });

async function exchangePublicToken(public_token, metadata) {
  const url = "http://" + hostname + "/api/publicToken";
  const response = await fetch(url, {
    // TODO: figure out how to handle IP
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ Public_token: public_token }),
  });
  const data = await response.json();
  console.log(data.access_token);
  //window.location.replace("http://" + hostname + "/home")
}

function linkClickHandler() {
  console.log(plaidLinkHandler);
  plaidLinkHandler.open();
}